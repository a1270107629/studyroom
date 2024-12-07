//go:build wireinject
package main

import (
	"github.com/a1270107629/studyroom/sr/pkg/wego"
	"github.com/a1270107629/studyroom/sr/user/grpc"
	"github.com/a1270107629/studyroom/sr/user/ioc"
	"github.com/a1270107629/studyroom/sr/user/repository"
	"github.com/a1270107629/studyroom/sr/user/repository/cache"
	"github.com/a1270107629/studyroom/sr/user/repository/dao"
	"github.com/a1270107629/studyroom/sr/user/service"
	"github.com/google/wire"
)

var thirdProvider = wire.NewSet(
	ioc.InitLogger,
	ioc.InitDB,
	ioc.InitRedis,
	ioc.InitEtcdClient,
)

func Init() *wego.App {
	wire.Build(
		thirdProvider,
		cache.NewRedisUserCache,
		dao.NewGORMUserDAO,
		repository.NewCachedUserRepository,
		service.NewUserService,
		grpc.NewUserServiceServer,
		ioc.InitGRPCxServer,
		wire.Struct(new(wego.App), "GRPCServer"),
	)
	return new(wego.App)
}
