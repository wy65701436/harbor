package nydus

import (
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter"
	"github.com/dragonflyoss/image-service/contrib/nydusify/pkg/converter/provider"
	"github.com/goharbor/harbor/src/jobservice/job"
)

type NydusifyConverter struct {
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

// Run implements the interface in job/Interface
func (n *NydusifyConverter) Run(ctx job.Context, params job.Parameters) error {

}
