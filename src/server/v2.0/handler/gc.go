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

func (g *gcAPI) ParseSchedule(ctx context.Context, params gc.parseScheduleParams) middleware.Responder {
	switch params.Schedule.Type {
	case "Manual":
		return g.Start(ctx, params)
	case "None":
		return g.DeleteSchedule(ctx)
	case "Hourly", "Daily", "Weekly", "Custom":
		return g.UpdateSchedule(ctx, params)
	}
	return nil
}

func (g *gcAPI) Start(ctx context.Context, params gc.startParams) middleware.Responder {
	if err := g.gcCtr.Start(ctx, params); err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewStartOK()
}

func (g *gcAPI) CreateSchedule(ctx context.Context, params gc.createScheduleParams) middleware.Responder {
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

func (g *gcAPI) GetSchedule(ctx context.Context) middleware.Responder {
	schedule, err := g.gcCtr.GetSchedule(ctx)
	if err != nil {
		return g.SendError(ctx, err)
	}

	return operation.NewGetScheduleOK().WithPayload(model.NewSchedule(schedule).ToSwagger())
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

func (g *gcAPI) GetLog(ctx context.Context, id int64) middleware.Responder {
	log, err := g.gcCtr.GetLog(ctx, id)
	if err != nil {
		return g.SendError(ctx, err)
	}
	return operation.NewGetLogOK().WithPayload(string(log))
}
