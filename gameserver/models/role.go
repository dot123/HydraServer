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

// Role 角色模型
type Role struct {
	RId             int64  `gorm:"column:rid;primary_key;AUTO_INCREMENT"` // 角色ID
	UId             int64  `gorm:"column:uid;NOT NULL"`                   // 用户ID
	HeadId          int64  `gorm:"column:head_id;default:0;NOT NULL"`     // 头像ID
	Sex             uint   `gorm:"column:sex;default:0;NOT NULL"`         // 性别
	NickName        string `gorm:"column:nick_name"`                      // 昵称
	CreatedAt       int64  `gorm:"column:created_at;"`                    // 创建时间
	UpdateDailyTime int64  `gorm:"column:update_daily_time;"`             // 最后每日更新时间
	PreLoginTime    int64  `gorm:"column:pre_login_time;"`                // 上次登录时间
	LastLoginTime   int64  `gorm:"column:last_login_time;"`               // 最后一次登录时间
}

func (m *Role) TableName() string {
	return "role"
}

// ToProto 转换为Proto模型
func (m *Role) ToProto() *protos.Role {
	p := &protos.Role{}
	p.Sex = int32(m.Sex)
	copier.Copy(p, m)
	return p
}

type RoleDBMgr struct {
	db    *gorm.DB
	cache *cache.Cache
	roles chan *Role
	log   logrus.FieldLogger
}

func NewRoleDBMgr(db *gorm.DB, cache *cache.Cache, log logrus.FieldLogger) *RoleDBMgr {
	roleDBMgr := &RoleDBMgr{
		db:    db,
		cache: cache,
		roles: make(chan *Role, 100), // 缓存通道
		log:   log,
	}

	// 启动后台 goroutine 处理角色更新
	go roleDBMgr.running()

	return roleDBMgr
}

// running 异步处理角色更新
func (m *RoleDBMgr) running() {
	for {
		select {
		case role := <-m.roles:
			if role.RId > 0 {
				// 更新缓存
				if err := m.cache.Update(context.TODO(), role.RId, role); err != nil {
					m.log.Errorf("cache error:%v\n", err)
				}
			}
		}
	}
}

// Create 创建角色
func (m *RoleDBMgr) Create(ctx context.Context, newRole *Role) error {
	if err := m.db.Create(newRole).Error; err != nil {
		return errors.NewResponseError(constant.DBError, err)
	}

	// 异步更新
	m.Push(ctx, newRole)

	return nil
}

// Push 将角色推送到更新通道
func (m *RoleDBMgr) Push(ctx context.Context, role *Role) {
	m.roles <- role
}

// Get 获取角色数据，支持缓存
func (m *RoleDBMgr) Get(ctx context.Context, rid int64) (*Role, error) {
	ret, err := m.cache.GetOrSet(ctx, rid, new(Role), func() (interface{}, error) {
		role := new(Role)
		err := m.db.Where("rid =?", rid).First(role).Error
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

// GetAll 获取所有角色数据
func (m *RoleDBMgr) GetAll(ctx context.Context, rid int64) ([]*Role, error) {
	roles := make([]*Role, 0)
	err := m.db.Where("rid =?", rid).Find(&roles).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return roles, errors.NewResponseError(constant.DBError, err)
	}

	return roles, nil
}

// Save 保存角色数据
func (m *RoleDBMgr) Save(ctx context.Context, role *Role) error {
	err := m.db.Save(role).Error
	if err != nil {
		m.log.Errorf("db error:%v\n", err)
	}
	return err
}
