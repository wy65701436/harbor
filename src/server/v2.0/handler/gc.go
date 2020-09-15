package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/lib/errors"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/gc"
)

type gcAPI struct {
	BaseAPI
	gcCtr gc.Controller
}

func newGCAPI() *gcAPI {
	return &gcAPI{
		gcCtr: gc.NewController(),
	}
}

func (g *gcAPI) Start(ctx context.Context, params gc.startParams) middleware.Responder {
	if err := g.gcCtr.Start(ctx, params); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewStartOK()
}

func (g *gcAPI) CreateSchedule(ctx context.Context, params gc.createScheduleParams) middleware.Responder {
	cron := params.Cron
	if cron == "" {
		return g.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("empty cron string for gc schedule"))
	}
	_, err := g.gcCtr.CreateSchedule(ctx, params.Cron, params)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewCreateScheduleOK()
}

func (g *gcAPI) GetSchedule(ctx context.Context) middleware.Responder {
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetScheduleOK().WithPayload(schedule)
}

func (g *gcAPI) UpdateSchedule(ctx context.Context, params gc.createScheduleParams) middleware.Responder {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return g.SendError(ctx, err)
	}
	return g.CreateSchedule(ctx, params)
}

func (g *gcAPI) DeleteSchedule(ctx context.Context) middleware.Responder {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewDeleteScheduleOK()
}

func (g *gcAPI) History(ctx context.Context, params gc.historyParams) middleware.Responder {
	query, err := g.BuildQuery(ctx, params.Q, params.Page, params.PageSize)
	if err != nil {
		return g.SendError(ctx, err)
	}
	total, err := g.gcCtr.Count(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}
	hs, err := g.gcCtr.History(ctx, query)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewHistoryOK().
		WithXTotalCount(total).
		WithLink(g.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(hs)
}

func (g *gcAPI) GetLog(ctx context.Context, id int64) middleware.Responder {
	log, err := g.gcCtr.GetLog(ctx, id)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetLogOK().WithPayload(string(log))
}
