package notification

import (
	"testing"

	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notifier/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedPolicyMgr struct {
}

func (f *fakedPolicyMgr) Create(*models.NotificationPolicy) (int64, error) {
	return 0, nil
}

func (f *fakedPolicyMgr) List(id int64) ([]*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *fakedPolicyMgr) Get(id int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *fakedPolicyMgr) GetByNameAndProjectID(string, int64) (*models.NotificationPolicy, error) {
	return nil, nil
}

func (f *fakedPolicyMgr) Update(*models.NotificationPolicy) error {
	return nil
}

func (f *fakedPolicyMgr) Delete(int64) error {
	return nil
}

func (f *fakedPolicyMgr) Test(*models.NotificationPolicy) error {
	return nil
}

func (f *fakedPolicyMgr) GetRelatedPolices(id int64, eventType string) ([]*models.NotificationPolicy, error) {
	return []*models.NotificationPolicy{
		{
			ID: 1,
			EventTypes: []string{
				model.EventTypeUploadChart,
				model.EventTypeDownloadChart,
				model.EventTypeDeleteChart,
				model.EventTypeScanningCompleted,
				model.EventTypeScanningFailed,
			},
			Targets: []models.EventTarget{
				{
					Type:    "http",
					Address: "http://127.0.0.1:8080",
				},
			},
		},
	}, nil
}

func TestChartPreprocessHandler_Handle(t *testing.T) {
	PolicyMgr := notification.PolicyMgr
	defer func() {
		notification.PolicyMgr = PolicyMgr
	}()
	notification.PolicyMgr = &fakedPolicyMgr{}

	handler := &ChartHandler{}
	config.Init()

	name := "project_for_test_chart_event_preprocess"
	id, _ := config.GlobalProjectMgr.Create(&models.Project{
		Name:    name,
		OwnerID: 1,
		Metadata: map[string]string{
			models.ProMetaEnableContentTrust:   "true",
			models.ProMetaPreventVul:           "true",
			models.ProMetaSeverity:             "Low",
			models.ProMetaReuseSysCVEWhitelist: "false",
		},
	})
	defer func(id int64) {
		if err := config.GlobalProjectMgr.Delete(id); err != nil {
			t.Logf("failed to delete project %d: %v", id, err)
		}
	}(id)

	type args struct {
		data interface{}
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "ChartHandler Want Error 1",
			args: args{
				data: nil,
			},
			wantErr: true,
		},
		{
			name: "ChartHandler Want Error 2",
			args: args{
				data: &model.ChartEvent{},
			},
			wantErr: true,
		},
		{
			name: "ChartHandler Want Error 3",
			args: args{
				data: &model.ChartEvent{
					Versions: []string{
						"v1.2.1",
					},
					ProjectName: "project_for_test_chart_event_preprocess",
				},
			},
			wantErr: true,
		},
		{
			name: "ChartHandler Want Error 4",
			args: args{
				data: &model.ChartEvent{
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
			name: "ChartHandler Want Error 5",
			args: args{
				data: &model.ChartEvent{
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
			err := handler.Handle(tt.args.data)
			if tt.wantErr {
				require.NotNil(t, err, "Error: %s", err)
				return
			}
			assert.Nil(t, err)
		})
	}
}

func TestChartPreprocessHandler_IsStateful(t *testing.T) {
	handler := &ChartHandler{}
	assert.False(t, handler.IsStateful())
}
