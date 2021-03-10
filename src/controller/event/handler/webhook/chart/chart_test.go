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

package chart

import (
	"context"
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/controller/project"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	projecttesting "github.com/goharbor/harbor/src/testing/controller/project"
	"github.com/goharbor/harbor/src/testing/mock"
	testingnotification "github.com/goharbor/harbor/src/testing/pkg/notification/policy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChartPreprocessHandler_Handle(t *testing.T) {
	PolicyMgr := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = PolicyMgr
	}()
	notification.PolicyMgr = &testingnotification.Manager{}

	ProjectCtl := project.Ctl
	defer func() {
		project.Ctl = ProjectCtl
	}()
	projectCtl := &projecttesting.Controller{}
	project.Ctl = projectCtl

	name := "project_for_test_chart_event_preprocess"
	mock.OnAnything(projectCtl, "Get").Return(func(ctx context.Context, projectIDOrName interface{}, options ...project.Option) *models.Project {
		return &models.Project{
			Name:    name,
			OwnerID: 1,
			Metadata: map[string]string{
				models.ProMetaEnableContentTrust:   "true",
				models.ProMetaPreventVul:           "true",
				models.ProMetaSeverity:             "Low",
				models.ProMetaReuseSysCVEAllowlist: "false",
			},
		}
	}, nil)
	projectCtl.On("Get")

	handler := &Handler{}
	config.Init()

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "Handler Want Error 1",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
		{
			name: "Handler Want Error 2",
			args: args{
				data: &event.ChartEvent{},
			},
			wantErr: true,
		},
		{
			name: "Handler Want Error 3",
			args: args{
				data: &event.ChartEvent{
					Versions: []string{
						"v1.2.1",
					},
					ProjectName: "project_for_test_chart_event_preprocess",
				},
			},
			wantErr: true,
		},
		{
			name: "Handler Want Error 4",
			args: args{
				data: &event.ChartEvent{
					Versions: []string{
						"v1.2.1",
					},
					ProjectName: "project_for_test_chart_event_preprocess_not_exists",
					ChartName:   "testChart",
				},
			},
			wantErr: true,
		},
		{
			name: "Handler Want Error 5",
			args: args{
				data: &event.ChartEvent{
					Versions: []string{
						"v1.2.1",
					},
					ProjectName: "project_for_test_chart_event_preprocess",
					ChartName:   "testChart",
					EventType:   "uploadChart",
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := handler.Handle(context.TODO(), tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestChartPreprocessHandler_IsStateful(t *testing.T) {
	handler := &Handler{}
	assert.False(t, handler.IsStateful())
}

func TestChartPreprocessHandler_Name(t *testing.T) {
	handler := &Handler{}
	assert.Equal(t, "ChartWebhook", handler.Name())
}
