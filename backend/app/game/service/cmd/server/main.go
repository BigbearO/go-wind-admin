package main

import (
	"context"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/transport/http"

	//_ "github.com/tx7do/kratos-bootstrap/config/nacos"

	_ "github.com/tx7do/kratos-bootstrap/registry/nacos"

	conf "github.com/tx7do/kratos-bootstrap/api/gen/go/conf/v1"
	"github.com/tx7do/kratos-bootstrap/bootstrap"
)

var version = "1.0.0"

func newApp(
	ctx *bootstrap.Context,
	hs *http.Server,
) *kratos.App {
	return bootstrap.NewApp(ctx, hs)
}

func runApp() error {
	ctx := bootstrap.NewContext(
		context.Background(),
		&conf.AppInfo{
			Project: "game",
			AppId:   "game",
			Version: version,
		},
	)
	return bootstrap.RunApp(ctx, initApp)
}

func main() {
	if err := runApp(); err != nil {
		panic(err)
	}
}
