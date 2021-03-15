package nydus

import (
	"context"
	"encoding/base64"
	"fmt"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/goharbor/harbor/src/jobservice/job"
	"github.com/goharbor/harbor/src/jobservice/logger"
)

type NydusifyConverter struct {
	repository string
	tag        string
	username   string
	password   string
	coreUrl    string
	logger     logger.Interface
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
	n.coreUrl = params["core_url"].(string)
	n.username = params["username"].(string)
	n.password = params["password"].(string)
	n.repository = params["repository"].(string)
	n.tag = params["tag"].(string)
	return nil
}

func basicAuth(username, password string) string {
	auth := username + ":" + password
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

// Run implements the interface in job/Interface
func (n *NydusifyConverter) Run(ctx job.Context, params job.Parameters) error {
	n.init(ctx, params)

	// TODO needs to define these two parameters.
	wordDir := "./tmp"
	nydusImagePath := "/var/log/jobs"

	source := fmt.Sprintf("%s/%s:%s", n.coreUrl, n.repository, n.tag)
	target := fmt.Sprintf("%s/%s:%s-nydus", n.coreUrl, n.repository, n.tag)
	auth := basicAuth(n.username, n.password)
	insecure := true

	logger, err := provider.DefaultLogger()
	if err != nil {
		return err
	}

	// Create remote with auth string for registry communication
	sourceRemote, err := provider.DefaultRemoteWithAuth(source, insecure, auth)
	if err != nil {
		return err
	}

	targetRemote, err := provider.DefaultRemoteWithAuth(target, insecure, auth)
	if err != nil {
		return err
	}

	// Source provider gets source image manifest, config, and layer
	sourceProvider, err := provider.DefaultSource(context.Background(), sourceRemote, wordDir)
	if err != nil {
		return err
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
		return err
	}

	if err := cvt.Convert(context.Background()); err != nil {
		return err
	}

	return nil
}
