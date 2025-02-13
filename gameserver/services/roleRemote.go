package services

import (
	"HydraServer/constant"
	"HydraServer/gameserver/models"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/cache"
	"HydraServer/pkg/errors"
	"context"
	"github.com/sirupsen/logrus"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"sync"
	"time"
)

type RoleRemote struct {
	component.Base
	app       pitaya.Pitaya
	cache     *cache.Cache
	roleDBMgr *models.RoleDBMgr
	log       logrus.FieldLogger
	logoutCh  chan *LogoutTask // 退出任务通道
	timers    sync.Map         // 存储角色ID对应的定时器
}

// LogoutTask 玩家退出任务结构
type LogoutTask struct {
	RId int64
	ctx context.Context
}

// timerInfo 定时器信息
type timerInfo struct {
	timer  *time.Timer
	cancel chan struct{}
}

func NewRoleRemoteService(app *pitaya.Pitaya, cache *cache.Cache, roleMgr *models.RoleDBMgr, log logrus.FieldLogger) *RoleRemote {
	r := &RoleRemote{
		app:       *app,
		cache:     cache,
		roleDBMgr: roleMgr,
		log:       log,
		logoutCh:  make(chan *LogoutTask, 100), // 设置缓冲区大小为100
	}

	// 启动任务处理协程
	go r.processLogoutTasks()

	return r
}

// CancelLogoutTimer 取消玩家的退出定时器
func (m *RoleRemote) CancelLogoutTimer(rid int64) {
	if value, exists := m.timers.Load(rid); exists {
		info := value.(*timerInfo)
		info.timer.Stop()
		close(info.cancel)
		m.timers.Delete(rid)
		m.log.Infof("Role %d logout timer cancelled", rid)
	}
}

// Logout 玩家退出游戏
func (m *RoleRemote) Logout(ctx context.Context, msg *protos.LogoutReq) (*protos.LogoutRsp, error) {
	rsp := new(protos.LogoutRsp)

	// 如果已存在定时器，先取消它
	m.CancelLogoutTimer(msg.RId)

	// 创建退出任务
	task := &LogoutTask{
		RId: msg.RId,
		ctx: ctx,
	}

	// 发送到任务通道，如果通道满则等待
	m.logoutCh <- task
	return rsp, nil
}

// processLogoutTasks 处理退出任务的协程
func (m *RoleRemote) processLogoutTasks() {
	for task := range m.logoutCh {
		// 创建取消通道
		cancel := make(chan struct{})

		// 创建定时器
		timer := time.NewTimer(5 * time.Second)

		// 保存定时器信息
		m.timers.Store(task.RId, &timerInfo{
			timer:  timer,
			cancel: cancel,
		})

		go func(t *LogoutTask, cancel chan struct{}) {
			select {
			case <-timer.C:
				// 定时器到期，执行保存
				m.saveRoleData(t)
				m.timers.Delete(t.RId)
			case <-cancel:
				// 定时器被取消
				return
			}
		}(task, cancel)
	}
}

// saveRoleData 保存角色数据
func (m *RoleRemote) saveRoleData(task *LogoutTask) {
	// 获取角色数据
	role, err := m.roleDBMgr.Get(task.ctx, task.RId)
	if err != nil {
		m.log.Errorf("get role error: %v", err)
		return
	}

	// 保存角色数据
	if err := m.roleDBMgr.Save(task.ctx, role); err != nil {
		m.log.Errorf("save role error: %v", err)
		return
	}

	// 清除缓存
	cacheKey := m.cache.GenCacheKey(task.RId, new(models.Role))
	if err := m.cache.Del(task.ctx, cacheKey); err != nil {
		m.log.Errorf("delete cache error: %v", err)
		return
	}

	m.log.Infof("Role %d data saved after 5 minutes", task.RId)
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
