package models

import (
	"HydraServer/constant"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/cache"
	"HydraServer/pkg/errors"
	"context"
	"github.com/jinzhu/copier"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// Role 玩家表
type Role struct {
	RId             int64  `gorm:"column:rid;primary_key;AUTO_INCREMENT"` // roleId
	UId             int64  `gorm:"column:uid;NOT NULL"`                   // 用户UId
	HeadId          int64  `gorm:"column:head_id;default:0;NOT NULL"`     // 头像Id
	Sex             uint   `gorm:"column:sex;default:0;NOT NULL"`         // 性别，0:女 1男
	NickName        string `gorm:"column:nick_name"`                      // nick_name
	CreatedAt       int64  `gorm:"column:created_at;"`
	UpdateDailyTime int64  `gorm:"column:update_daily_time;"`
	PreLoginTime    int64  `gorm:"column:pre_login_time;"`
	LastLoginTime   int64  `gorm:"column:last_login_time;"`
}

func (m *Role) TableName() string {
	return "role"
}

func (m *Role) ToProto() *protos.Role {
	p := &protos.Role{}
	p.Sex = int32(m.Sex)
	copier.Copy(p, m)
	return p
}

type RoleDBMgr struct {
	db    *gorm.DB
	cache *cache.Cache
	log   logrus.FieldLogger
}

func NewRoleDBMgr(db *gorm.DB, cache *cache.Cache, log logrus.FieldLogger) *RoleDBMgr {
	roleDBMgr := &RoleDBMgr{
		db:    db,
		cache: cache,
		log:   log,
	}
	return roleDBMgr
}

// Create 创建角色
func (m *RoleDBMgr) Create(ctx context.Context, newRole *Role) error {
	if err := m.db.Create(newRole).Error; err != nil {
		return errors.NewResponseError(constant.DBError, err)
	}

	if err := m.Update(ctx, newRole); err != nil {
		return errors.NewResponseError(constant.DBError, err)
	}
	return nil
}

// Update 更新角色数据
func (m *RoleDBMgr) Update(ctx context.Context, role *Role) error {
	if err := m.db.Save(role).Error; err != nil {
		return errors.NewResponseError(constant.DBError, err)
	}

	if err := m.cache.Update(ctx, role.UId, role); err != nil {
		return errors.NewResponseError(constant.DBError, err)
	}
	return nil
}

// Get 获取角色数据
func (m *RoleDBMgr) Get(ctx context.Context, uid int64) (*Role, error) {
	ret, err := m.cache.GetOrSet(ctx, uid, new(Role), func() (interface{}, error) {
		role := new(Role)
		err := m.db.Where("uid =?", uid).First(role).Error
		return role, err
	})

	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewResponseError(constant.DBError, err)
	}

	if ret != nil {
		return ret.(*Role), nil
	}
	return nil, nil
}
