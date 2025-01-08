package models

import (
	"context"
	"github.com/google/wire"
	"gorm.io/gorm"
	"time"
)

// LoginLast 最后一次用户登录表
type LoginLast struct {
	ID         int64     `gorm:"column:id;primary_key;AUTO_INCREMENT"`
	UId        int64     `gorm:"column:uid;default:0;NOT NULL"`       // 用户UID
	ServerId   string    `gorm:"column:serverId;size:255;NOT NULL"`   // 服务器id
	LoginTime  time.Time `gorm:"column:login_time"`                   // 登录时间
	LogoutTime time.Time `gorm:"column:logout_time;default:null"`     // 登出时间
	IP         string    `gorm:"column:ip;NOT NULL"`                  // ip
	IsLogout   int64     `gorm:"column:is_logout;default:0;NOT NULL"` // 是否logout,1:logout，0:login
	Token      string    `gorm:"column:token"`                        // 会话
	Hardware   string    `gorm:"column:hardware;NOT NULL"`            // hardware
}

func (m *LoginLast) TableName() string {
	return "login_last"
}

func GetLoginLastDB(defDB *gorm.DB) *gorm.DB {
	return defDB.Table("login_last")
}

var LoginLastRepoSet = wire.NewSet(wire.Struct(new(LoginLastRepo), "*"))

type LoginLastRepo struct {
	DB *gorm.DB
}

func (m *LoginLastRepo) Get(Uid int64) (*LoginLast, error) {
	model := new(LoginLast)
	err := GetLoginLastDB(m.DB).Where("uid =?", Uid).First(model).Error
	return model, err
}

func (m *LoginLastRepo) Save(ctx context.Context, model *LoginLast) error {
	err := GetLoginLastDB(m.DB).Save(model).Error
	return err
}
