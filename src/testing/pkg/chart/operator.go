package chart

import (
	"github.com/goharbor/harbor/src/pkg/chart"
	"github.com/stretchr/testify/mock"
	"k8s.io/helm/pkg/proto/hapi/chart"
)

// FakeOpertaor ...
type FakeOpertaor struct {
	mock.Mock
}

// Match ...
func (f *FakeOpertaor) GetDetails(content []byte) (*chartserver.ChartVersionDetails, error) {
	args := f.Called()
	var chartDetails *chartserver.ChartVersionDetails
	if args.Get(1) != nil {
		chartDetails = args.Get(1).(*chartserver.ChartVersionDetails)
	}
	return chartDetails, args.Error(1)
}

func (f *FakeOpertaor) GetData(content []byte) (*chart.Chart, error) {
	args := f.Called()
	var chartData *chart.Chart
	if args.Get(1) != nil {
		chartData = args.Get(1).(*chart.Chart)
	}
	return chartData, args.Error(1)
}
