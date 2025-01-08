package main

import (
	"HydraServer/gateserver/models"
	"HydraServer/httpserver/controller"
	"HydraServer/httpserver/logic"
	"github.com/google/wire"
	"github.com/topfreegames/pitaya/v2/cluster"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

type Injector struct {
	AccountController *controller.AccountController
	ServiceDiscovery  cluster.ServiceDiscovery
}

var ControllerSet = wire.NewSet(
	controller.AccountControllerSet,
)

var LogicSet = wire.NewSet(
	logic.UserLogicSet,
)

var RepoSet = wire.NewSet(
	models.UserInfoRepoSet,
)
