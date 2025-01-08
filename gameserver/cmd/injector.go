package main

import (
	"HydraServer/gameserver/logic"
	"HydraServer/gameserver/models"
	"HydraServer/gameserver/services"
	"HydraServer/pkg/cache"
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

type Injector struct {
	RoleService       *services.Role
	RoleRemoteService *services.RoleRemote
	App               *pitaya.Pitaya
	Log               logrus.FieldLogger
}

var ServiceSet = wire.NewSet(
	services.NewRoleService,
	services.NewRoleRemoteService,
)

var RepoSet = wire.NewSet(
	cache.NewCache,
	models.NewRoleDBMgr,
)

var LogicSet = wire.NewSet(
	logic.NewRoleLogic,
)
