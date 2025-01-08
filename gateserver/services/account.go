package services

import (
	"HydraServer/constant"
	"HydraServer/gateserver/models"
	protos "HydraServer/gateserver/protos"
	"HydraServer/pkg/errors"
	"HydraServer/pkg/token"
	"context"
	"fmt"
	"github.com/spf13/cast"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"time"
)

type Account struct {
	component.Base
	app           pitaya.Pitaya
	userInfoRepo  *models.UserInfoRepo
	loginLastRepo *models.LoginLastRepo
}

func NewAccountService(app *pitaya.Pitaya, userInfoRepo *models.UserInfoRepo, loginLastRepo *models.LoginLastRepo) *Account {
	return &Account{
		app:           *app,
		userInfoRepo:  userInfoRepo,
		loginLastRepo: loginLastRepo,
	}
}

func (m *Account) Init() {

}

func (m *Account) AfterInit() {

}

func (m *Account) Login(ctx context.Context, msg *protos.LoginReq) (*protos.LoginRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)

	user, err := m.userInfoRepo.Get(msg.Username)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NewResponseError(constant.RoleNotExist, nil)
	} else if err != nil {
		return nil, errors.NewResponseError(constant.DBError, nil)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(msg.Password)); err != nil {
		return nil, errors.NewResponseError(constant.PwdIncorrect, nil)
	}

	addr := session.RemoteAddr()
	ip := "127.0.0.1"
	if addr != nil {
		ip = addr.String()
	}

	// 生成token
	newToken, err := token.GenerateToken(user.UId, msg.Username)
	if err != nil {
		logger.Errorf("token.GenerateToken(%d,%s) error: %v\n", user.UId, msg.Username, err)
		return nil, errors.NewResponseError(constant.GenerateTokenFail, nil)
	}

	serverId := m.app.GetServer().Metadata["serverId"]
	loginLast := &models.LoginLast{
		UId:       user.UId,
		ServerId:  serverId,
		LoginTime: time.Now(),
		IP:        ip,
		IsLogout:  0,
		Token:     newToken,
		Hardware:  msg.Hardware,
	}

	if err := m.loginLastRepo.Save(ctx, loginLast); err != nil {
		return nil, errors.NewResponseError(constant.DBError, nil)
	}

	// 绑定uid
	if err = m.bind(ctx, user.UId); err != nil {
		return nil, err
	}

	return &protos.LoginRsp{
		Username: msg.Username,
		Password: msg.Password,
		Token:    newToken,
		UId:      user.UId,
	}, nil
}

func (m *Account) ReLogin(ctx context.Context, msg *protos.ReLoginReq) (*protos.ReLoginRsp, error) {
	if msg.Token == "" {
		return nil, errors.NewResponseError(constant.TokenInvalid, nil)
	}

	user, err := token.ParseToken(msg.Token)
	if err != nil {
		return nil, errors.NewResponseError(constant.ParseTokenFail, nil)
	}

	loginLast, err := m.loginLastRepo.Get(user.UId)
	if err != nil {
		return nil, errors.NewResponseError(constant.ReLoginFail, nil)
	}

	if loginLast.Token == msg.Token {
		if loginLast.Hardware == msg.Hardware {
			if err = m.bind(ctx, user.UId); err != nil {
				return nil, err
			}
			return &protos.ReLoginRsp{Token: msg.Token}, nil
		} else {
			return nil, errors.NewResponseError(constant.HardwareIncorrect, nil)
		}
	}
	return nil, errors.NewResponseError(constant.TokenInvalid, nil)
}

func (m *Account) Logout(ctx context.Context, msg *protos.LogoutReq) (*protos.LogoutRsp, error) {
	session := m.app.GetSessionFromCtx(ctx)
	uid := cast.ToInt64(session.Get(constant.UIdKey))

	if err := m.updateLogoutLastRecord(ctx, uid); err != nil {
		return nil, err
	}

	if err := m.unbind(ctx); err != nil {
		return nil, err
	}

	return &protos.LogoutRsp{}, nil
}

// Bind 绑定uid
func (m *Account) bind(ctx context.Context, uid int64) error {
	serverId := m.app.GetServer().Metadata["serverId"]
	servers, err := m.app.GetServersByType("game")
	if err != nil {
		return err
	}

	serverKey := ""
	for key := range servers {
		server := servers[key]
		if server.Metadata[constant.ServerId] == serverId {
			serverKey = server.ID
		}
	}

	s := m.app.GetSessionFromCtx(ctx)
	s.Set(constant.UIdKey, uid)
	s.Set(constant.ServerKey, serverKey)
	s.Bind(ctx, fmt.Sprintf("%d", uid))

	return nil
}

func (m *Account) unbind(ctx context.Context) error {
	s := m.app.GetSessionFromCtx(ctx)
	s.Remove(constant.UIdKey)
	s.Remove(constant.RIdKey)
	return nil
}

func (m *Account) updateLogoutLastRecord(ctx context.Context, uid int64) error {
	loginLast, err := m.loginLastRepo.Get(uid)
	if err != nil {
		return errors.NewResponseError(constant.DBError, nil)
	}
	loginLast.IsLogout = 1
	loginLast.LogoutTime = time.Now()

	if err := m.loginLastRepo.Save(ctx, loginLast); err != nil {
		return errors.NewResponseError(constant.DBError, nil)
	}
	return nil
}
