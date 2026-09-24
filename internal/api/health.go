package api

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

type healthOutput struct {
	Body struct {
		Status string `json:"status" enum:"ok" example:"ok"`
	}
}

func (a *API) registerHealth() {
	huma.Register(a.api, huma.Operation{
		OperationID: "health",
		Method:      http.MethodGet,
		Path:        "/health",
		Summary:     "Liveness and database check",
	}, func(ctx context.Context, _ *struct{}) (*healthOutput, error) {
		sqlDB, err := a.db.DB()
		if err != nil || sqlDB.PingContext(ctx) != nil {
			return nil, huma.Error503ServiceUnavailable("database unreachable")
		}
		out := &healthOutput{}
		out.Body.Status = "ok"
		return out, nil
	})
}
