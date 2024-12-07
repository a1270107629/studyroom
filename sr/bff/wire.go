//go:build wireinject

package main

import (
	"github.com/a1270107629/studyroom/sr/bff/ioc"
	"github.com/a1270107629/studyroom/sr/bff/web"
	"github.com/a1270107629/studyroom/sr/bff/web/jwt"
	"github.com/a1270107629/studyroom/sr/pkg/wego"
	"github.com/google/wire"
)

func InitApp() *wego.App {
	wire.Build(
		ioc.InitLogger,
		ioc.InitRedis,
		ioc.InitEtcdClient,
		ioc.InitUserClient,
		web.NewUserHandler,

		jwt.NewRedisHandler,

		ioc.InitGinServer,
		wire.Struct(new(wego.App), "WebServer"),
	)
	return new(wego.App)
}
