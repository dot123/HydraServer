package logic

import (
	"HydraServer/constant"
	"HydraServer/gameserver/models"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/errors"
	"HydraServer/pkg/utils"
	"context"
	"github.com/topfreegames/pitaya/v2"
	"time"
)

type RoleLogic struct {
	app       pitaya.Pitaya
	roleDBMgr *models.RoleDBMgr
}

func NewRoleLogic(app *pitaya.Pitaya, roleDBMgr *models.RoleDBMgr) *RoleLogic {
	m := &RoleLogic{
		app:       *app,
		roleDBMgr: roleDBMgr,
	}

	return m
}

// Create 创建角色
func (m *RoleLogic) Create(ctx context.Context, uid int64, msg *protos.CreateRoleReq) (*models.Role, error) {
	role, err := m.roleDBMgr.Get(ctx, uid)
	if err != nil {
		return nil, err
	}

	if role != nil {
		return nil, errors.NewResponseError(constant.RoleAlreadyCreate, nil)
	}

	newRole := &models.Role{
		UId:       uid,
		HeadId:    msg.HeadId,
		Sex:       uint(msg.Sex),
		NickName:  msg.NickName,
		CreatedAt: time.Now().Unix(),
	}

	if err := m.roleDBMgr.Create(ctx, newRole); err != nil {
		return nil, err
	}
	return newRole, nil
}

// Enter 角色进入
func (m *RoleLogic) Enter(ctx context.Context, uid int64, username string, ip string) (*models.Role, error) {
	roles, err := m.roleDBMgr.GetAll(ctx, uid)
	if err != nil {
		return nil, err
	}

	if len(roles) == 0 {
		return nil, errors.NewResponseError(constant.RoleNotExist, nil)
	}

	// 选择第一个角色
	role := roles[0]
	now := time.Now().Unix()
	role.PreLoginTime = role.LastLoginTime
	role.LastLoginTime = time.Now().Unix()

	// 执行每日任务
	if role.UpdateDailyTime == 0 || utils.IsDifferentDays(time.Unix(role.UpdateDailyTime, 0), time.Unix(now, 0), "Asia/Shanghai") {
		role.UpdateDailyTime = now
		m.onDaily(ctx, role.RId)
	}

	err = m.roleDBMgr.Update(ctx, role)

	return role, err
}

func (m *RoleLogic) ChangeNickName(ctx context.Context, rid int64, nickName string) error {
	role, err := m.roleDBMgr.Get(ctx, rid)
	if err != nil {
		return err
	}

	if role == nil {
		return errors.NewResponseError(constant.RoleNotExist, nil)
	}
	role.NickName = nickName

	return m.roleDBMgr.Update(ctx, role)
}

func (m *RoleLogic) onDaily(ctx context.Context, rid int64) {

}
