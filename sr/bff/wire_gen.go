// Code generated by Wire. DO NOT EDIT.

//go:generate go run -mod=mod github.com/google/wire/cmd/wire
//go:build !wireinject
// +build !wireinject

package main

import (
	"github.com/a1270107629/studyroom/sr/bff/ioc"
	"github.com/a1270107629/studyroom/sr/bff/web"
	"github.com/a1270107629/studyroom/sr/bff/web/jwt"
	"github.com/a1270107629/studyroom/sr/pkg/wego"
)

// Injectors from wire.go:

func InitApp() *wego.App {
	loggerV1 := ioc.InitLogger()
	cmdable := ioc.InitRedis()
	handler := jwt.NewRedisHandler(cmdable)
	client := ioc.InitEtcdClient()
	userServiceClient := ioc.InitUserClient(client)
	userHandler := web.NewUserHandler(userServiceClient, handler)
	server := ioc.InitGinServer(loggerV1, handler, userHandler)
	app := &wego.App{
		WebServer: server,
	}
	return app
}