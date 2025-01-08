package main

import (
	"HydraServer/chatserver/services"
	"github.com/google/wire"
	"github.com/topfreegames/pitaya/v2"
)

var InjectorSet = wire.NewSet(wire.Struct(new(Injector), "*"))

type Injector struct {
	ChatService *services.Chat
	App         *pitaya.Pitaya
}

var ServiceSet = wire.NewSet(
	services.NewChatService,
)
