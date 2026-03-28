package server

import (
	"context"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/selector"
	"github.com/go-kratos/kratos/v2/transport/http"

	swaggerUI "github.com/tx7do/kratos-swagger-ui"

	"github.com/tx7do/kratos-bootstrap/bootstrap"
	"github.com/tx7do/kratos-bootstrap/rpc"

	"go-wind-admin/app/game/service/cmd/server/assets"
	"go-wind-admin/app/game/service/internal/service"

	gameV1 "go-wind-admin/api/gen/go/game/service/v1"
)

func newRestWhiteListMatcher() selector.MatchFunc {
	whiteList := make(map[string]bool)
	return func(ctx context.Context, operation string) bool {
		if _, ok := whiteList[operation]; ok {
			return false
		}
		return true
	}
}

func newRestMiddleware(logger log.Logger) []middleware.Middleware {
	var ms []middleware.Middleware

	ms = append(ms, logging.Server(logger))
	ms = append(ms, selector.Server().Match(newRestWhiteListMatcher()).Build())

	return ms
}

func NewRestServer(
	ctx *bootstrap.Context,
	gameAccountService *service.GameAccountService,
) (*http.Server, error) {
	cfg := ctx.GetConfig()

	if cfg == nil || cfg.Server == nil || cfg.Server.Rest == nil {
		return nil, nil
	}

	srv, err := rpc.CreateRestServer(cfg, newRestMiddleware(ctx.GetLogger())...)
	if err != nil {
		return nil, err
	}

	gameV1.RegisterGameAccountServiceHTTPServer(srv, gameAccountService)

	if cfg.GetServer().GetRest().GetEnableSwagger() {
		swaggerUI.RegisterSwaggerUIServerWithOption(
			srv,
			swaggerUI.WithTitle("Game Service"),
			swaggerUI.WithMemoryData(assets.OpenApiData, "yaml"),
		)
	}

	return srv, nil
}
