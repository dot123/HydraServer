package services

import (
	"HydraServer/constant"
	"HydraServer/gameserver/logic"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/errors"
	"HydraServer/pkg/token"
	"context"
	"fmt"
	"github.com/spf13/cast"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"github.com/topfreegames/pitaya/v2/timer"
	"time"
)

type Role struct {
	component.Base
	app        pitaya.Pitaya
	timer      *timer.Timer
	roleLogic  *logic.RoleLogic
	roleRemote *RoleRemote
}

func NewRoleService(app *pitaya.Pitaya, roleMgr *logic.RoleLogic, roleRemote *RoleRemote) *Role {
	m := &Role{
		app:        *app,
		roleLogic:  roleMgr,
		roleRemote: roleRemote,
	}
	return m
}

func (m *Role) Init() {

}

func (m *Role) AfterInit() {

}

func (m *Role) EnterServer(ctx context.Context, msg *protos.EnterServerReq) (*protos.EnterServerRsp, error) {
	s := m.app.GetSessionFromCtx(ctx)

	user, err := token.ParseToken(msg.Token)
	if err != nil {
		return nil, errors.NewResponseError(constant.TokenInvalid, err)
	}

	addr := s.RemoteAddr()
	ip := "127.0.0.1"
	if addr != nil {
		ip = addr.String()
	}

	role, err := m.roleLogic.Enter(ctx, user.UId, user.Username, ip)
	if err != nil {
		return nil, err
	}

	// 取消该角色的退出定时器
	if m.roleRemote != nil {
		m.roleRemote.CancelLogoutTimer(role.RId)
	}

	s.Set(constant.RIdKey, role.RId)

	if err = s.PushToFront(ctx); err != nil {
		return nil, fmt.Errorf("PushToFront() error: %v", err)
	}

	return &protos.EnterServerRsp{
		Role:     role.ToProto(),
		Time:     time.Now().UnixNano() / 1e6,
		ServerID: m.app.GetServerID(),
	}, nil
}

func (m *Role) Create(ctx context.Context, msg *protos.CreateRoleReq) (*protos.CreateRoleRsp, error) {
	session := m.app.GetSessionFromCtx(ctx)
	uid := cast.ToInt64(session.Get(constant.UIdKey))
	role, err := m.roleLogic.Create(ctx, uid, msg)
	if err != nil {
		return nil, err
	}
	return &protos.CreateRoleRsp{Role: role.ToProto()}, nil
}

func (m *Role) ChangeNickName(ctx context.Context, msg *protos.ChangeNickNameReq) (*protos.ChangeNickNameRsp, error) {
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get(constant.RIdKey))

	err := m.roleLogic.ChangeNickName(ctx, rid, msg.NickName)

	return &protos.ChangeNickNameRsp{}, err
}

func (m *Role) Ping(ctx context.Context, msg *protos.Ping) (*protos.Pong, error) {
	return &protos.Pong{Delay: time.Now().UnixMilli() - msg.Time}, nil
}
