package main

import (
	"HydraServer/gateserver/models"
	"HydraServer/gateserver/services"
	"github.com/google/wire"
	"github.com/topfreegames/pitaya/v2"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

type Injector struct {
	AccountService *services.Account
	App            *pitaya.Pitaya
}

var ServiceSet = wire.NewSet(
	services.NewAccountService,
)

var RepoSet = wire.NewSet(
	models.UserInfoRepoSet,
	models.LoginLastRepoSet,
)
