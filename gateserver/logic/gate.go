package logic

import (
	game_protos "HydraServer/gameserver/protos"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/cluster"
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

func (m *GateLogic) OnExit(uid int64, rid int64, server *cluster.Server) {
	m.log.Infof("uid(%d)玩家退出", uid)

	ctx := context.TODO()
	rsp := &game_protos.LogoutRsp{}
	if err := m.app.RPCTo(ctx, server.ID, "game.role.logout", rsp, &game_protos.LogoutReq{RId: rid}); err != nil {
		m.log.Errorf("uid(%d)玩家error %v", uid, err)
	}
}
