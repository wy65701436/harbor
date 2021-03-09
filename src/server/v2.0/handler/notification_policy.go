package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/lib/q"
	"github.com/goharbor/harbor/src/pkg/notification/job"
	"github.com/goharbor/harbor/src/pkg/notification/policy"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
	"github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/webhook"
)

func newNotificationPolicyAPI() *notificationPolicyAPI {
	return &notificationJobAPI{
		webhookjobMgr:    job.Mgr,
		webhookPolicyMgr: policy.Mgr,
	}
}

type notificationPolicyAPI struct {
	BaseAPI
	webhookjobMgr    job.Manager
	webhookPolicyMgr policy.Manager
}

func (n *notificationPolicyAPI) ListWebhookPolicy(ctx context.Context, params webhook.ListWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionList, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}

	query := &q.Query{
		Keywords: q.KeyWords{
			"ProjectID": policy.ID,
		},
		PageNumber: *params.Page,
		PageSize:   *params.PageSize,
	}

	return nil
}

func (n *notificationPolicyAPI) CreateWebhookPolicy(ctx context.Context, params webhook.CreateWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionCreate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
}

func (n *notificationPolicyAPI) UpdateWebhookPolicy(ctx context.Context, params webhook.UpdateWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionUpdate, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
}

func (n *notificationPolicyAPI) DeleteWebhookPolicy(ctx context.Context, params webhook.DeleteWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionDelete, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
}

func (n *notificationPolicyAPI) GetWebhookPolicy(ctx context.Context, params webhook.GetWebhookPolicyParams) middleware.Responder {
	projectNameOrID := parseProjectNameOrID(params.ProjectNameOrID, params.XIsResourceName)
	if err := n.RequireProjectAccess(ctx, projectNameOrID, rbac.ActionRead, rbac.ResourceNotificationPolicy); err != nil {
		return n.SendError(ctx, err)
	}
	return nil
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
