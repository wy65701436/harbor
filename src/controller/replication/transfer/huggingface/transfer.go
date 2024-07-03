// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package image

import (
	"context"
	"errors"
	"github.com/goharbor/harbor/src/controller/artifact/processor/hf"
	"github.com/goharbor/harbor/src/lib/config"
	"oras.land/oras-go/v2/registry/remote/auth"
	"oras.land/oras-go/v2/registry/remote/retry"
	"os"
	"strconv"
	"strings"

	trans "github.com/goharbor/harbor/src/controller/replication/transfer"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/reg/adapter"
	"github.com/goharbor/harbor/src/pkg/reg/model"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/seasonjs/hf-hub/api"
	oras "oras.land/oras-go/v2"
	"oras.land/oras-go/v2/content/file"
	"oras.land/oras-go/v2/registry/remote"
)

var (
	blobRetryCnt, chunkRetryCnt int
	replicationChunkSize        int64
	errStopped                  = errors.New("stopped")
	// default chunk size is 10MB
	defaultChunkSize = 10 * 1024 * 1024
)

func init() {
	blobRetryCnt, _ = strconv.Atoi(os.Getenv("COPY_BLOB_RETRY_COUNT"))
	if blobRetryCnt <= 0 {
		blobRetryCnt = 5
	}

	chunkRetryCnt, _ = strconv.Atoi(os.Getenv("COPY_CHUNK_RETRY_COUNT"))
	if chunkRetryCnt <= 0 {
		chunkRetryCnt = 5
	}

	replicationChunkSize, _ = strconv.ParseInt(os.Getenv("REPLICATION_CHUNK_SIZE"), 10, 64)
	if replicationChunkSize <= 0 {
		replicationChunkSize = int64(defaultChunkSize)
	}

	if err := trans.RegisterFactory(model.ResourceTypeHF, factory); err != nil {
		log.Errorf("failed to register transfer factory: %v", err)
	}
}

type repository struct {
	repository string
	tags       []string
}

func factory(logger trans.Logger, stopFunc trans.StopFunc) (trans.Transfer, error) {
	return &transfer{
		logger:    logger,
		isStopped: stopFunc,
	}, nil
}

type transfer struct {
	logger    trans.Logger
	isStopped trans.StopFunc
	src       adapter.ArtifactRegistry
	dst       adapter.ArtifactRegistry
}

func (t *transfer) Transfer(src *model.Resource, dst *model.Resource, opts *trans.Options) error {
	// initialize
	if err := t.initialize(src, dst); err != nil {
		return err
	}

	// copy the repository from source registry to the destination
	return t.copy(t.convert(src), t.convert(dst), dst.Override, opts)
}

func (t *transfer) convert(resource *model.Resource) *repository {
	repository := &repository{
		repository: resource.Metadata.Repository.Name,
	}
	for _, artifact := range resource.Metadata.Artifacts {
		if len(artifact.Tags) > 0 {
			repository.tags = append(repository.tags, artifact.Tags...)
			continue
		}
		// no tags
		if len(artifact.Digest) > 0 {
			repository.tags = append(repository.tags, artifact.Digest)
		}
	}
	if len(repository.tags) > 0 {
		return repository
	}
	// fallback to vtags
	repository.tags = resource.Metadata.Vtags
	return repository
}

func (t *transfer) initialize(src *model.Resource, dst *model.Resource) error {
	// create client for source registry
	srcReg, err := createRegistry(src.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for source registry: %v", err)
		return err
	}
	t.src = srcReg
	t.logger.Infof("client for source registry [type: %s, URL: %s, insecure: %v] created",
		src.Registry.Type, src.Registry.URL, src.Registry.Insecure)

	// create client for destination registry
	dstReg, err := createRegistry(dst.Registry)
	if err != nil {
		t.logger.Errorf("failed to create client for destination registry: %v", err)
		return err
	}
	t.dst = dstReg
	t.logger.Infof("client for destination registry [type: %s, URL: %s, insecure: %v] created",
		dst.Registry.Type, dst.Registry.URL, dst.Registry.Insecure)

	return nil
}

func createRegistry(reg *model.Registry) (adapter.ArtifactRegistry, error) {
	factory, err := adapter.GetFactory(reg.Type)
	if err != nil {
		return nil, err
	}
	ad, err := factory.Create(reg)
	if err != nil {
		return nil, err
	}
	registry, ok := ad.(adapter.ArtifactRegistry)
	if !ok {
		return nil, errors.New("the adapter doesn't implement the \"ArtifactRegistry\" interface")
	}
	return registry, nil
}

func (t *transfer) shouldStop() bool {
	isStopped := t.isStopped()
	if isStopped {
		t.logger.Info("the job is stopped")
	}
	return isStopped
}

func (t *transfer) copy(src *repository, dst *repository, override bool, opts *trans.Options) error {
	srcRepo := src.repository
	dstRepo := dst.repository
	t.logger.Infof("copying %s:[%s](source registry) to %s:[%s](destination registry)...",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))

	// download model files from hugging face
	modelId := "google/owlvit-base-patch32"
	filesInModels, err := t.download(modelId)
	if err != nil {
		return err
	}
	t.logger.Infof("the model: %s is completely downloaded successfully.", modelId)

	// create the local file store for OCI object
	fs, err := file.New("/tmp/demo")
	if err != nil {
		return err
	}
	defer fs.Close()

	// compose the oci object for the downloaded files
	ctx := context.Background()
	manifestDescriptor, err := t.composeOCI(ctx, fs, filesInModels)
	if err != nil {
		return err
	}
	t.logger.Infof("manifest descriptor:", manifestDescriptor.Digest)

	// push the manifest to the destination registry
	if err := t.pushManifest(ctx, fs, manifestDescriptor, dstRepo, dst.tags[0]); err != nil {
		return err
	}

	t.logger.Infof("copy %s:[%s](source registry) to %s:[%s](destination registry) completed",
		srcRepo, strings.Join(src.tags, ","), dstRepo, strings.Join(dst.tags, ","))
	return nil
}

func (t *transfer) download(modelID string) ([]string, error) {
	hapi, err := api.NewApi()
	if err != nil {
		return nil, err
	}

	r, err := hapi.Model(modelID).Info()
	if err != nil {
		return nil, err
	}
	var filesInModels []string
	if len(r.Siblings) != 0 {
		for _, s := range r.Siblings {
			modelPath, err := hapi.Model("google/owlvit-base-patch32").Get(s.Rfilename)
			if err != nil {
				return nil, err
			}
			filesInModels = append(filesInModels, modelPath)
			t.logger.Infof("%s ", modelPath, "is copied successfully.")
		}
	}

	return filesInModels, nil
}

func (t *transfer) composeOCI(ctx context.Context, fs *file.Store, files []string) (v1.Descriptor, error) {
	mediaType := "application/vnd.goharbor.huggingface.v1"
	fileDescriptors := make([]v1.Descriptor, 0, len(files))
	for _, name := range files {
		fileDescriptor, err := fs.Add(ctx, name, mediaType, "")
		if err != nil {
			return v1.Descriptor{}, err
		}
		fileDescriptors = append(fileDescriptors, fileDescriptor)
		log.Infof("file descriptor for %s: %v\n", name, fileDescriptor)
	}

	// 2. Pack the files and tag the packed manifest
	artifactType := hf.MediaType
	orasOpts := oras.PackManifestOptions{
		Layers: fileDescriptors,
		ConfigDescriptor: &v1.Descriptor{
			MediaType: hf.MediaType,
		},
		ManifestAnnotations: map[string]string{
			"type":              "hugging-face-object",
			"author":            "Stability AI",
			"model":             "google/owlvit-base-patch32",
			"model description": "This is a model that can be used to generate images based on text prompts. It is a Multimodal Diffusion Transformer (https://arxiv.org/abs/2403.03206) that uses three fixed, pretrained text encoders (OpenCLIP-ViT/G, CLIP-ViT/L and T5-xxl)",
			"model type":        "MMDiT text-to-image generative model",
			"license":           "stabilityai-nc-research-community",
		},
	}
	manifestDescriptor, err := oras.PackManifest(ctx, fs, oras.PackManifestVersion1_1, artifactType, orasOpts)
	if err != nil {
		return v1.Descriptor{}, err
	}
	return manifestDescriptor, nil
}

func (t *transfer) pushManifest(ctx context.Context, fs *file.Store, manifest v1.Descriptor, repository, tag string) error {
	if t.shouldStop() {
		return errStopped
	}

	t.logger.Infof("pushing the manifest of artifact %s:%s ...", repository, tag)
	if err := fs.Tag(ctx, manifest, tag); err != nil {
		return err
	}
	reg := config.GetCoreURL()
	reg = strings.TrimPrefix(reg, "http://")
	repo, err := remote.NewRepository(reg + "/library/demo")
	if err != nil {
		return err
	}
	repo.PlainHTTP = true
	repo.Client = &auth.Client{
		Client: retry.DefaultClient,
		Cache:  auth.NewCache(),
		Credential: auth.StaticCredential(reg, auth.Credential{
			Username: "admin",
			Password: "Harbor12345",
		}),
	}
	t.logger.Info("===========")
	_, err = oras.Copy(ctx, fs, tag, repo, tag, oras.DefaultCopyOptions)
	if err != nil {
		t.logger.Infof("%v ", err)
		return err
	}
	t.logger.Infof("the manifest of artifact %s:%s pushed",
		repository, tag)
	return nil
}
