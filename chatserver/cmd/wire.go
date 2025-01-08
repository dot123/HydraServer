//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"
	"github.com/topfreegames/pitaya/v2"
)

func BuildInjector(*pitaya.Pitaya) (*Injector, func(), error) {
	wire.Build(ServiceSet, InjectorSet)
	return new(Injector), nil, nil
}
