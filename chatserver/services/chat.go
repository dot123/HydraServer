package services

import (
	protos "HydraServer/chatserver/protos"
	"HydraServer/constant"
	game_protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/errors"
	"context"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/spf13/cast"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"sync"
	"time"
)

type Chat struct {
	component.Base
	app            pitaya.Pitaya
	mu             sync.RWMutex
	groupNameByRId map[int64]string // rid对应的联盟频道
	groupNameById  map[string]bool
	roleByRId      map[int64]*game_protos.Role
	msgList        []*protos.ChatMsg
}

const (
	world = 0
	union = 1
)

func NewChatService(app *pitaya.Pitaya) *Chat {
	return &Chat{
		app:            *app,
		groupNameByRId: make(map[int64]string),
		groupNameById:  make(map[string]bool),
		roleByRId:      make(map[int64]*game_protos.Role),
	}
}

func (m *Chat) Init() {
	m.app.GroupCreate(context.Background(), "world")
}

func (m *Chat) AfterInit() {

}

func (m *Chat) Enter(ctx context.Context, msg *protos.LoginReq) (*protos.LoginRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	uid := cast.ToString(session.Get("uid"))
	rid := cast.ToInt64(session.Get("rid"))

	if uid == "" || rid == 0 {
		return nil, errors.NewResponseError(constant.InvalidParam, nil)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	rsp := &game_protos.GetRoleRsp{}
	if err := m.app.RPCTo(ctx, msg.ServerID, "game.role.getrole", rsp, &game_protos.GetRoleReq{RId: rid}); err != nil {
		logger.Errorln(err)
	} else {
		role := new(game_protos.Role)
		copier.Copy(role, rsp.Role)
		m.roleByRId[rid] = role
	}

	b, err := m.app.GroupContainsMember(ctx, "world", uid)
	if err != nil {
		logger.Errorf("GroupContainsMember(%s,%s) error: %v\n", "world", uid, err)
		return nil, err
	}

	if !b {
		if err := m.app.GroupAddMember(ctx, "world", uid); err != nil {
			logger.Errorf("GroupAddMember(%s,%s) error: %v\n", "world", uid, err)
			return nil, err
		}
	}
	return &protos.LoginRsp{}, nil
}

func (m *Chat) Join(ctx context.Context, msg *protos.JoinReq) (*protos.JoinRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get("rid"))
	if msg.Type == union {
		m.mu.Lock()
		defer m.mu.Unlock()

		// 先移除组
		groupName, ok := m.groupNameByRId[rid]
		if ok {
			if err := m.app.GroupRemoveMember(ctx, groupName, session.UID()); err != nil {
				logger.Errorf("GroupRemoveMember(%s,%s) error: %v\n", groupName, session.UID(), err)
				return nil, err
			}
		}

		// 不存在则创建组
		groupName = fmt.Sprintf("%d", msg.Id)
		if !m.groupNameById[groupName] {
			if err := m.app.GroupCreate(context.Background(), groupName); err != nil {
				logger.Errorf("GroupCreate(%s) error: %v\n", groupName, err)
				return nil, err
			}
			m.groupNameById[groupName] = true
		}

		// 加入组
		if err := m.app.GroupAddMember(ctx, groupName, session.UID()); err != nil {
			logger.Errorf("GroupAddMember(%s,%s) error: %v\n", groupName, session.UID(), err)
			return nil, err
		}
		m.groupNameByRId[rid] = groupName
	}

	return &protos.JoinRsp{
		Type: msg.Type,
		Id:   msg.Id,
	}, nil
}

func (m *Chat) Exit(ctx context.Context, msg *protos.ExitReq) (*protos.ExitRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get("rid"))
	if msg.Type == union {
		m.mu.Lock()
		defer m.mu.Unlock()

		groupName, ok := m.groupNameByRId[rid]
		if ok {
			if err := m.app.GroupRemoveMember(ctx, groupName, session.UID()); err != nil {
				logger.Errorf("GroupRemoveMember(%s,%s) error: %v\n", groupName, session.UID(), err)
				return nil, err
			}
		}
		delete(m.groupNameByRId, rid)
	}

	return &protos.ExitRsp{Type: msg.Type}, nil
}

func (m *Chat) Chat(ctx context.Context, msg *protos.ChatReq) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get("rid"))

	m.mu.Lock()
	defer m.mu.Unlock()

	rsp := &protos.ChatMsg{
		RId:  rid,
		Type: msg.Type,
		Msg:  msg.Msg,
		Time: time.Now().Unix(),
	}

	if role, ok := m.roleByRId[rid]; ok {
		rsp.NickName = role.NickName
	}

	groupName := ""
	if msg.Type == world {
		groupName = "world"
	} else if msg.Type == union {
		name, ok := m.groupNameByRId[rid]
		if ok {
			groupName = name
		} else {
			logger.Warnf("chat not found rid(%d) in groupNameByRId\n", rid)
		}
	}

	if groupName != "" {
		err := m.app.GroupBroadcast(ctx, "gate", groupName, "onMessage", rsp)
		if err != nil {
			logger.Errorf("GroupBroadcast(%s,%s,%s) error: %v\n", "gate", groupName, "onMessage", err)
		}
	}

	chatMsg := new(protos.ChatMsg)
	copier.Copy(chatMsg, rsp)
	if len(m.msgList) > 100 {
		m.msgList = m.msgList[1:len(m.msgList)]
	}
	m.msgList = append(m.msgList, chatMsg)
}

func (m *Chat) History(ctx context.Context) (*[]*protos.ChatMsg, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var msgList []*protos.ChatMsg
	for _, msg := range m.msgList {
		c := new(protos.ChatMsg)
		copier.Copy(c, msg)
		msgList = append(msgList, c)
	}

	return &msgList, nil
}
