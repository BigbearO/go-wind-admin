package data

import (
	"context"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tx7do/kratos-bootstrap/bootstrap"

	adminV1 "go-wind-admin/api/gen/go/admin/service/v1"
)

type AdminClient struct {
	UserService adminV1.UserServiceHTTPClient
}

func NewAdminClient(
	ctx *bootstrap.Context,
) (*AdminClient, error) {
	cfg := ctx.GetConfig()
	if cfg == nil || cfg.Client == nil {
		return nil, nil
	}

	cli, err := http.NewClient(
		context.Background(),
		http.WithEndpoint("http://127.0.0.1:7788"),
	)
	if err != nil {
		return nil, err
	}

	return &AdminClient{
		UserService: adminV1.NewUserServiceHTTPClient(cli),
	}, nil
}
