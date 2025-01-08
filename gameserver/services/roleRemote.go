package services

import (
	"HydraServer/constant"
	"HydraServer/gameserver/models"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/errors"
	"context"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
)

type RoleRemote struct {
	component.Base
	app       pitaya.Pitaya
	roleDBMgr *models.RoleDBMgr
}

func NewRoleRemoteService(app *pitaya.Pitaya, roleMgr *models.RoleDBMgr) *RoleRemote {
	return &RoleRemote{
		app:       *app,
		roleDBMgr: roleMgr,
	}
}

func (m *RoleRemote) GetRole(ctx context.Context, msg *protos.GetRoleReq) (*protos.GetRoleRsp, error) {
	role, err := m.roleDBMgr.Get(ctx, msg.RId)
	if err != nil {
		return nil, err
	}

	if role == nil {
		return nil, errors.NewResponseError(constant.RoleNotExist, err)
	}

	rsp := &protos.GetRoleRsp{Role: role.ToProto()}
	return rsp, nil
}
