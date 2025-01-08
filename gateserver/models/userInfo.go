package models

import (
	"github.com/google/wire"
	"gorm.io/gorm"
	"time"
)

// UserInfo 用户信息表
type UserInfo struct {
	UId       int64     `gorm:"column:uid;primary_key;AUTO_INCREMENT"`
	Username  string    `gorm:"column:username;NOT NULL"`         // 用户名
	Password  string    `gorm:"column:password;NOT NULL"`         // md5密码
	Status    uint      `gorm:"column:status;default:0;NOT NULL"` // 用户账号状态。0-默认；1-冻结；2-停号
	Hardware  string    `gorm:"column:hardware;NOT NULL"`
	CreatedAt time.Time `gorm:"column:created_at;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time `gorm:"column:updated_at;default:CURRENT_TIMESTAMP"`
}

func (m *UserInfo) TableName() string {
	return "user_info"
}

func GetUserInfoDB(defDB *gorm.DB) *gorm.DB {
	return defDB.Table("user_info")
}

var UserInfoRepoSet = wire.NewSet(wire.Struct(new(UserInfoRepo), "*"))

type UserInfoRepo struct {
	DB *gorm.DB
}

func (m *UserInfoRepo) Insert(model *UserInfo) error {
	err := GetUserInfoDB(m.DB).Create(model).Error
	return err
}

func (m *UserInfoRepo) Get(username string) (*UserInfo, error) {
	model := new(UserInfo)
	err := GetUserInfoDB(m.DB).Where("username =?", username).First(model).Error
	return model, err
}

func (m *UserInfoRepo) UpdatePwd(username string, password string) (rowsAffected int64, err error) {
	db := GetUserInfoDB(m.DB).Where("username =?", username).Update("password", password)
	return db.RowsAffected, db.Error
}

func (m *UserInfoRepo) ResetPwd(username string, password string) error {
	err := GetUserInfoDB(m.DB).Where("username =? ", username).
		Updates(map[string]interface{}{"password": password}).Error
	return err
}
