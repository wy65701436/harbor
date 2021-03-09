package handler

import (
	"context"
	"fmt"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	"github.com/goharbor/harbor/src/controller/event"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	policy_model "github.com/goharbor/harbor/src/pkg/notification/policy/model"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
	"strings"
	"time"
)

func newNotificationPolicyAPI() *notificationPolicyAPI {
	return &notificationPolicyAPI{
		webhookjobMgr:    job.Mgr,
		webhookPolicyMgr: policy.Mgr,
		supportedEvents:  initSupportedEvents(),
	}
}

type notificationPolicyAPI struct {
	BaseAPI
	webhookjobMgr    job.Manager
	webhookPolicyMgr policy.Manager
	supportedEvents  map[string]struct{}
}

func (n *notificationPolicyAPI) ListWebhookPolicy(ctx context.Context, params webhook.ListWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	projectID, err := getProjectID(ctx, projectNameOrID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	query := &q.Query{
		Keywords: q.KeyWords{
			"ProjectID": projectID,
		},
		PageNumber: *params.Page,
		PageSize:   *params.PageSize,
	}

	total, err := n.webhookPolicyMgr.Count(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}

	policies, err := n.webhookPolicyMgr.List(ctx, query)
	if err != nil {
		return n.SendError(ctx, err)
	}
	var results []*models.WebhookPolicy
	for _, p := range policies {
		results = append(results, model.NewNotifiactionPolicy(p).ToSwagger())
	}

	return operation.NewListWebhookPolicyOK().
		WithXTotalCount(total).
		WithLink(n.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (n *notificationPolicyAPI) CreateWebhookPolicy(ctx context.Context, params webhook.CreateWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy := &policy_model.Policy{}
	lib.JSONCopy(policy, params.Policy)

	if ok, err := n.validateEventTypes(policy); !ok {
		return n.SendError(ctx, err)
	}
	if ok, err := n.validateTargets(policy); !ok {
		return n.SendError(ctx, err)
	}

	id, err := n.webhookPolicyMgr.Create(ctx, policy)
	if err != nil {
		return n.SendError(ctx, err)
	}

	location := fmt.Sprintf("%s/%d", strings.TrimSuffix(params.HTTPRequest.URL.Path, "/"), id)
	return operation.NewCreateWebhookPolicyCreated().WithLocation(location)
}

func (n *notificationPolicyAPI) UpdateWebhookPolicy(ctx context.Context, params webhook.UpdateWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy := &policy_model.Policy{}
	lib.JSONCopy(policy, params.Policy)

	if ok, err := n.validateEventTypes(policy); !ok {
		return n.SendError(ctx, err)
	}
	if ok, err := n.validateTargets(policy); !ok {
		return n.SendError(ctx, err)
	}

	if err := n.webhookPolicyMgr.Update(ctx, policy); err != nil {
		return n.SendError(ctx, err)
	}

	return operation.NewUpdateWebhookPolicyOK()
}

func (n *notificationPolicyAPI) DeleteWebhookPolicy(ctx context.Context, params webhook.DeleteWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	if err := n.webhookPolicyMgr.Delete(ctx, params.WebhookPolicyID); err != nil {
		return n.SendError(ctx, err)
	}
	return operation.NewDeleteWebhookPolicyOK()
}

func (n *notificationPolicyAPI) GetWebhookPolicy(ctx context.Context, params webhook.GetWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	policy, err := n.webhookPolicyMgr.Get(ctx, params.WebhookPolicyID)
	if err != nil {
		return n.SendError(ctx, err)
	}

	return operation.NewGetWebhookPolicyOK().WithPayload(model.NewNotifiactionPolicy(policy).ToSwagger())
}

func (n *notificationPolicyAPI) LastTrigger(ctx context.Context, params webhook.LastTriggerParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
}

func (n *notificationPolicyAPI) GetSupportedEventTypes(ctx context.Context, params webhook.GetSupportedEventTypesParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
}

func (n *notificationPolicyAPI) getLastTriggerTimeGroupByEventType(ctx context.Context, eventType string, policyID int64) (time.Time, error) {
	jobs, err := n.webhookjobMgr.ListJobsGroupByEventType(ctx, policyID)
	if err != nil {
		return time.Time{}, err
	}

	for _, job := range jobs {
		if eventType == job.EventType {
			return job.CreationTime, nil
		}
	}
	return time.Time{}, nil
}

func (n *notificationPolicyAPI) validateTargets(policy *policy_model.Policy) (bool, error) {
	if len(policy.Targets) == 0 {
		return false, errors.New(nil).WithMessage("empty notification target with policy %s", policy.Name).WithCode(errors.BadRequestCode)
	}
	for _, target := range policy.Targets {
		url, err := utils.ParseEndpoint(target.Address)
		if err != nil {
			return false, errors.New(err).WithCode(errors.BadRequestCode)
		}
		// Prevent SSRF security issue #3755
		target.Address = url.Scheme + "://" + url.Host + url.Path

		_, ok := notification.SupportedNotifyTypes[target.Type]
		if !ok {
			return false, errors.New(nil).WithMessage("unsupported target type %s with policy %s", target.Type, policy.Name).WithCode(errors.BadRequestCode)
		}
	}
	return true, nil
}

func (n *notificationPolicyAPI) validateEventTypes(policy *policy_model.Policy) (bool, error) {
	if len(policy.EventTypes) == 0 {
		return false, errors.New(nil).WithMessage("empty event type").WithCode(errors.BadRequestCode)
	}
	for _, eventType := range policy.EventTypes {
		_, ok := n.supportedEvents[eventType]
		if !ok {
			return false, errors.New(nil).WithMessage("unsupported event type %s", eventType).WithCode(errors.BadRequestCode)
		}
	}
	return true, nil
}

func initSupportedEvents() map[string]struct{} {
	eventTypes := []string{
		event.TopicPushArtifact,
		event.TopicPullArtifact,
		event.TopicDeleteArtifact,
		event.TopicUploadChart,
		event.TopicDeleteChart,
		event.TopicDownloadChart,
		event.TopicQuotaExceed,
		event.TopicQuotaWarning,
		event.TopicScanningFailed,
		event.TopicScanningCompleted,
		event.TopicReplication,
		event.TopicTagRetention,
	}

	var supportedEventTypes = make(map[string]struct{})
	for _, eventType := range eventTypes {
		supportedEventTypes[eventType] = struct{}{}
	}

	return supportedEventTypes
}

// constructPolicyWithTriggerTime construct notification policy information displayed in UI
// including event type, enabled, creation time, last trigger time
func constructPolicyWithTriggerTime(policies []*models.NotificationPolicy) ([]*notificationPolicyForUI, error) {
	res := []*notificationPolicyForUI{}
	if policies != nil {
		for _, policy := range policies {
			for _, t := range policy.EventTypes {
				ply := &notificationPolicyForUI{
					PolicyName:   policy.Name,
					EventType:    t,
					Enabled:      policy.Enabled,
					CreationTime: &policy.CreationTime,
				}
				if !policy.CreationTime.IsZero() {
					ply.CreationTime = &policy.CreationTime
				}

				ltTime, err := getLastTriggerTimeGroupByEventType(t, policy.ID)
				if err != nil {
					return nil, err
				}
				if !ltTime.IsZero() {
					ply.LastTriggerTime = &ltTime
				}
				res = append(res, ply)
			}
		}
	}
	return res, nil
}
