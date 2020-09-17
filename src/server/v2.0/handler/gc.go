package handler

import (
	"context"
	"github.com/go-openapi/runtime/middleware"
	"github.com/goharbor/harbor/src/common/security/local"
	"github.com/goharbor/harbor/src/controller/gc"
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/goharbor/harbor/src/server/v2.0/handler/model"
	"github.com/goharbor/harbor/src/server/v2.0/models"
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

func (g *gcAPI) PostSchedule(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	return g.createSchedule(ctx, params)
}

func (g *gcAPI) PutSchedule(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	switch params.Schedule.Type {
	case "Manual":
		return g.start(ctx, params)
	case "None":
		return g.deleteSchedule(ctx)
	case "Hourly", "Daily", "Weekly", "Custom":
		return g.updateSchedule(ctx, params)
	}
	return nil
}

func (g *gcAPI) start(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	if err := g.gcCtr.Start(ctx, params); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewStartOK()
}

func (g *gcAPI) createSchedule(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	cron := params.Schedule.Cron
	if cron == "" {
		return g.SendError(ctx, errors.New(nil).WithCode(errors.BadRequestCode).
			WithMessage("empty cron string for gc schedule"))
	}
	_, err := g.gcCtr.CreateSchedule(ctx, cron, params)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewCreateScheduleOK()
}

func (g *gcAPI) updateSchedule(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return g.SendError(ctx, err)
	}
	return g.createSchedule(ctx, params)
}

func (g *gcAPI) deleteSchedule(ctx context.Context) middleware.Responder {
	if err := g.gcCtr.DeleteSchedule(ctx); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewDeleteScheduleOK()
}

func (g *gcAPI) GetSchedule(ctx context.Context) middleware.Responder {
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewGetScheduleOK().WithPayload(model.NewSchedule(schedule).ToSwagger())
}

func (g *gcAPI) GetGCHistory(ctx context.Context, params gc.getGCHistoryParams) middleware.Responder {
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
	var results []*models.History
	for _, h := range hs {
		res := &model.History{}
		res.History = h
		results = append(results, res.ToSwagger())
	}
	return operation.NewHistoryOK().
		WithXTotalCount(total).
		WithLink(g.Links(ctx, params.HTTPRequest.URL, total, query.PageNumber, query.PageSize).String()).
		WithPayload(results)
}

func (g *gcAPI) GetGCLog(ctx context.Context, params operation.getGCLogParams) middleware.Responder {
	log, err := g.gcCtr.GetLog(ctx, params.GcID)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetLogOK().WithPayload(string(log))
}
