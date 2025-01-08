package logic

import (
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2"
)

type GateLogic struct {
	log logrus.FieldLogger
	app pitaya.Pitaya
}

func NewGateLogic(app *pitaya.Pitaya, log logrus.FieldLogger) *GateLogic {
	m := &GateLogic{
		log: log,
		app: *app,
	}
	return m
}

func (m *GateLogic) OnEnter(uid int64) {
	m.log.Infof("uid(%d)玩家进入", uid)
}

func (m *GateLogic) OnExit(uid int64) {
	m.log.Infof("uid(%d)玩家退出", uid)
}
