package logic

import (
	"HydraServer/gateserver/models"
	"HydraServer/httpserver/errors"
	"HydraServer/httpserver/pkg/logger"
	"context"
	"github.com/google/wire"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var UserLogicSet = wire.NewSet(wire.Struct(new(UserLogic), "*"))

type UserLogic struct {
	UserInfoRepo *models.UserInfoRepo
}

func (m *UserLogic) Insert(ctx context.Context, username, password, hardware string) error {
	if username == "" || password == "" {
		return errors.NewDefaultResponse("用户名或密码不能为空")
	}

	user, err := m.UserInfoRepo.Get(username)
	if user.Username == username {
		return errors.NewDefaultResponse("账号已经存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithContext(ctx).Errorf("密码加密失败: %v", err.Error())
		return err
	}

	model := &models.UserInfo{
		Username: username,
		Password: string(hash),
		Hardware: hardware,
	}

	if err := m.UserInfoRepo.Insert(model); err != nil {
		return errors.NewDefaultResponse("创建账号失败")
	}
	return nil
}

func (m *UserLogic) UpdatePwd(ctx context.Context, username, password, newPassword string) error {
	if username == "" || newPassword == "" || password == "" {
		return errors.NewDefaultResponse("用户名或密码不能为空")
	}

	model, err := m.UserInfoRepo.Get(username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NewDefaultResponse("账号不存在")
		}
		return err
	}

	if err = bcrypt.CompareHashAndPassword([]byte(model.Password), []byte(password)); err != nil {
		return errors.NewDefaultResponse("密码不正确")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logger.WithContext(ctx).Errorf("密码加密失败: %v", err.Error())
		return err
	}

	rowsAffected, err := m.UserInfoRepo.UpdatePwd(username, string(hash))
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		err = errors.New("数据库更新密码错误")
		return err
	}
	return nil
}

func (m *UserLogic) ResetPwd(ctx context.Context, username, password string) error {
	if username == "" || password == "" {
		return errors.NewDefaultResponse("用户名或密码不能为空")
	}

	_, err := m.UserInfoRepo.Get(username)
	if err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		return errors.NewDefaultResponse("账号不存在")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		logger.WithContext(ctx).Errorf("密码加密失败: %v", err.Error())
		return err
	}

	err = m.UserInfoRepo.ResetPwd(username, string(hash))
	if err != nil {
		return errors.NewDefaultResponse("重置用户密码失败")
	}
	return nil
}
