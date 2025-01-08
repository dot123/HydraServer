//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2"
)

func BuildInjector(*pitaya.Pitaya, logrus.FieldLogger) (*Injector, func(), error) {
	wire.Build(InitGormDB, RepoSet, ServiceSet, InjectorSet)
	return new(Injector), nil, nil
}
