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

package hf

import (
	"context"
	"encoding/json"
	"io"

	v1 "github.com/opencontainers/image-spec/specs-go/v1"

	ps "github.com/goharbor/harbor/src/controller/artifact/processor"
	"github.com/goharbor/harbor/src/controller/artifact/processor/base"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/log"
	"github.com/goharbor/harbor/src/pkg/artifact"
)

// const definitions
const (
	AdditionTypeReadme = "README.MD"
	ArtifactTypeModel  = "MODEL"
	MediaType          = "application/vnd.goharbor.aiartifact.v1+json"
)

// const definitions
const ()

func init() {
	pc := &processor{}
	pc.ManifestProcessor = base.NewManifestProcessor()
	if err := ps.Register(pc, MediaType); err != nil {
		log.Errorf("failed to register processor for media type %s: %v", MediaType, err)
		return
	}
}

type processor struct {
	*base.ManifestProcessor
}

//func (p *processor) AbstractMetadata(ctx context.Context, art *artifact.Artifact, _ []byte) error {
//	cfgManiDgt := ""
//	// try to get the digest of the manifest that the config layer is referenced by
//	for _, reference := range art.References {
//		if reference.Annotations != nil &&
//			reference.Annotations["io.cnab.manifest.type"] == "config" {
//			cfgManiDgt = reference.ChildDigest
//		}
//	}
//	if len(cfgManiDgt) == 0 {
//		return nil
//	}
//
//	// get the manifest that the config layer is referenced by
//	mani, _, err := p.RegCli.PullManifest(art.RepositoryName, cfgManiDgt)
//	if err != nil {
//		return err
//	}
//	_, payload, err := mani.Payload()
//	if err != nil {
//		return err
//	}
//
//	// abstract the metadata from config layer
//	return p.manifestProcessor.AbstractMetadata(ctx, art, payload)
//}

func (p *processor) AbstractAddition(_ context.Context, artifact *artifact.Artifact, addition string) (*ps.Addition, error) {
	if addition != AdditionTypeReadme {
		return nil, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("addition %s isn't supported for %s", addition, ArtifactTypeModel)
	}

	m, _, err := p.RegCli.PullManifest(artifact.RepositoryName, artifact.Digest)
	if err != nil {
		return nil, err
	}
	_, payload, err := m.Payload()
	if err != nil {
		return nil, err
	}
	manifest := &v1.Manifest{}
	if err := json.Unmarshal(payload, manifest); err != nil {
		return nil, err
	}

	for _, layer := range manifest.Layers {
		layerDgst := layer.Digest.String()
		// currently, we only handle readme addition for model artifact.
		if layerDgst != manifest.Config.Digest.String() &&
			(layer.Annotations != nil && layer.Annotations["org.cnai.model.readme"] == "true") {
			_, blob, err := p.RegCli.PullBlob(artifact.RepositoryName, layerDgst)
			if err != nil {
				return nil, err
			}

			defer blob.Close()
			content, err := io.ReadAll(blob)
			if err != nil {
				return nil, err
			}

			var additionContent []byte
			var additionContentType string

			switch addition {
			case AdditionTypeReadme:
				additionContent = []byte(content)
				additionContentType = "text/markdown; charset=utf-8"
			}

			return &ps.Addition{
				Content:     additionContent,
				ContentType: additionContentType,
			}, nil
		}
	}
	return nil, nil
}

func (p *processor) GetArtifactType(_ context.Context, _ *artifact.Artifact) string {
	return ArtifactTypeModel
}
