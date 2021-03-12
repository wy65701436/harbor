package nydus

import (
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

type NydusifyConverter struct {
	logger logger.Interface
}

// MaxFails implements the interface in job/Interface
func (n *NydusifyConverter) MaxFails() uint {
	return 1
}

// MaxCurrency is implementation of same method in Interface.
func (n *NydusifyConverter) MaxCurrency() uint {
	return 1
}

// ShouldRetry implements the interface in job/Interface
func (n *NydusifyConverter) ShouldRetry() bool {
	return false
}

// Validate implements the interface in job/Interface
func (n *NydusifyConverter) Validate(params job.Parameters) error {
	return nil
}

// init ...
func (n *NydusifyConverter) init(ctx job.Context, params job.Parameters) error {
	n.logger = ctx.GetLogger()
	return nil
}

// Run implements the interface in job/Interface
func (n *NydusifyConverter) Run(ctx job.Context, params job.Parameters) error {
	n.init(ctx, params)

	wordDir := "./tmp"
	nydusImagePath := "/path/to/nydus-image"
	source := "localhost:5000/ubuntu:latest"
	target := "localhost:5000/ubuntu:latest-nydus"
	auth := "<base64_encoded_auth>"
	insecure := true

	// Create remote with auth string for registry communication
	sourceRemote, err := provider.DefaultRemoteWithAuth(source, insecure, auth)
	if err != nil {
		panic(err)
	}
	// Or we can create with docker config
	// sourceRemote, err := provider.DefaultRemote(source, insecure)
	// if err != nil {
	// 	panic(err)
	// }
	targetRemote, err := provider.DefaultRemoteWithAuth(target, insecure, auth)
	if err != nil {
		panic(err)
	}

	// Source provider gets source image manifest, config, and layer
	sourceProvider, err := provider.DefaultSource(context.Background(), sourceRemote, wordDir)
	if err != nil {
		panic(err)
	}

	opt := converter.Opt{
		Logger:         logger,
		SourceProvider: sourceProvider,
		TargetRemote:   targetRemote,

		WorkDir:        wordDir,
		PrefetchDir:    "/",
		NydusImagePath: nydusImagePath,
		MultiPlatform:  false,
		DockerV2Format: true,
		WhiteoutSpec:   "oci",
	}

	cvt, err := converter.New(opt)
	if err != nil {
		panic(err)
	}

	if err := cvt.Convert(context.Background()); err != nil {
		panic(err)
	}

	return nil
}
