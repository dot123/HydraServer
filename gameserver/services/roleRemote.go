package services

import (
	"HydraServer/constant"
	"HydraServer/gameserver/models"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/cache"
	"HydraServer/pkg/errors"
	"context"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
)

type RoleRemote struct {
	component.Base
	app       pitaya.Pitaya
	cache     *cache.Cache
	roleDBMgr *models.RoleDBMgr
}

func NewRoleRemoteService(app *pitaya.Pitaya, cache *cache.Cache, roleMgr *models.RoleDBMgr) *RoleRemote {
	return &RoleRemote{
		app:       *app,
		cache:     cache,
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

// Logout 玩家退出游戏
func (m *RoleRemote) Logout(ctx context.Context, msg *protos.LogoutReq) (*protos.LogoutRsp, error) {
	rsp := new(protos.LogoutRsp)
	role, err := m.roleDBMgr.Get(ctx, msg.RId)
	if err != nil {
		return rsp, err
	}

	err = m.roleDBMgr.Save(ctx, role)
	if err != nil {
		return rsp, err
	}

	cacheKey := m.cache.GenCacheKey(msg.RId, new(models.Role))
	err = m.cache.Del(ctx, cacheKey)
	if err != nil {
		return rsp, err
	}
	return rsp, err
}
