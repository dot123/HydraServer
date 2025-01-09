package services

import (
	protos "HydraServer/chatserver/protos"
	"HydraServer/constant"
	"HydraServer/gameserver/models"
	"HydraServer/pkg/errors"
	"HydraServer/pkg/redisbackend"
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/component"
	"time"
)

type Chat struct {
	component.Base
	app    pitaya.Pitaya
	client redis.UniversalClient
}

const (
	world                   = 0
	union                   = 1
	redisMsgListKey         = "chat:messages"
	redisUnionGroupByRidKey = "union:group:rid"
	redisUnionGroupKey      = "union:groups"
)

func NewChatService(app *pitaya.Pitaya, redisBackend *redisbackend.RedisBackend) *Chat {
	return &Chat{
		app:    *app,
		client: redisBackend.Client(),
	}
}

func (m *Chat) Init() {
	m.app.GroupCreate(context.Background(), "world")
}

func (m *Chat) AfterInit() {

}

// Enter 玩家进入聊天时的处理
func (m *Chat) Enter(ctx context.Context, msg *protos.LoginReq) (*protos.LoginRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	uid := cast.ToString(session.Get(constant.UIdKey))
	rid := cast.ToInt64(session.Get(constant.RIdKey))

	if uid == "" || rid == 0 {
		return nil, errors.NewResponseError(constant.InvalidParam, nil)
	}

	// 检查是否已经加入世界聊天组
	b, err := m.app.GroupContainsMember(ctx, "world", uid)
	if err != nil {
		logger.Errorf("GroupContainsMember(%s,%s) error: %v\n", "world", uid, err)
		return nil, err
	}

	// 如果没有加入世界组，加入该组
	if !b {
		if err := m.app.GroupAddMember(ctx, "world", uid); err != nil {
			logger.Errorf("GroupAddMember(%s,%s) error: %v\n", "world", uid, err)
			return nil, err
		}
	}
	return &protos.LoginRsp{}, nil
}

// Join 玩家加入联盟频道
func (m *Chat) Join(ctx context.Context, msg *protos.JoinReq) (*protos.JoinRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get(constant.RIdKey))
	if msg.Type == union {
		// 锁住 Redis 中的组信息
		m.client.Watch(ctx, func(tx *redis.Tx) error {
			groupName, _ := tx.HGet(ctx, redisUnionGroupByRidKey, fmt.Sprintf("%d", rid)).Result()
			if groupName != "" {
				// 从当前联盟频道移除
				if err := m.app.GroupRemoveMember(ctx, groupName, session.UID()); err != nil {
					logger.Errorf("GroupRemoveMember(%s,%s) error: %v\n", groupName, session.UID(), err)
					return err
				}
			}

			// 创建新的联盟频道
			groupName = fmt.Sprintf("%d", msg.Id)
			ret, _ := tx.HGet(ctx, redisUnionGroupKey, groupName).Result()
			if ret == "" {
				if err := m.app.GroupCreate(ctx, groupName); err != nil {
					logger.Errorf("GroupCreate(%s) error: %v\n", groupName, err)
					return err
				}
				// 更新联盟信息到 Redis
				if err := tx.HSet(ctx, redisUnionGroupKey, groupName, 1).Err(); err != nil {
					logger.Errorf("Failed to set group name to Redis: %v", err)
					return err
				}
			}

			// 加入联盟频道
			if err := m.app.GroupAddMember(ctx, groupName, session.UID()); err != nil {
				logger.Errorf("GroupAddMember(%s,%s) error: %v\n", groupName, session.UID(), err)
				return err
			}

			// 更新联盟信息到 Redis
			if err := tx.HSet(ctx, redisUnionGroupByRidKey, fmt.Sprintf("%d", rid), groupName).Err(); err != nil {
				logger.Errorf("Failed to set group name to Redis: %v", err)
				return err
			}
			return nil
		}, fmt.Sprintf("group:%d", rid)) // Redis 事务处理
	}

	return &protos.JoinRsp{
		Type: msg.Type,
		Id:   msg.Id,
	}, nil
}

// Exit 玩家退出联盟频道
func (m *Chat) Exit(ctx context.Context, msg *protos.ExitReq) (*protos.ExitRsp, error) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get(constant.RIdKey))
	if msg.Type == union {
		groupName, _ := m.client.HGet(ctx, redisUnionGroupByRidKey, fmt.Sprintf("%d", rid)).Result()
		if groupName != "" {
			// 从联盟频道中移除玩家
			if err := m.app.GroupRemoveMember(ctx, groupName, session.UID()); err != nil {
				logger.Errorf("GroupRemoveMember(%s,%s) error: %v\n", groupName, session.UID(), err)
				return nil, err
			}

			// 删除 Redis 中的频道信息
			m.client.HDel(ctx, redisUnionGroupByRidKey, fmt.Sprintf("%d", rid))
		}
	}

	return &protos.ExitRsp{Type: msg.Type}, nil
}

// Chat 处理玩家发送的聊天消息
func (m *Chat) Chat(ctx context.Context, msg *protos.ChatReq) {
	logger := pitaya.GetDefaultLoggerFromCtx(ctx)
	session := m.app.GetSessionFromCtx(ctx)
	rid := cast.ToInt64(session.Get(constant.RIdKey))

	// 构建聊天消息
	rsp := &protos.ChatMsg{
		RId:  rid,
		Type: msg.Type,
		Msg:  msg.Msg,
		Time: time.Now().Unix(),
	}

	// 获取玩家的昵称（从 Redis 中获取角色信息）
	roleJson, err := m.client.Get(ctx, fmt.Sprintf("%d-role", rid)).Result()
	if err != nil {
		logger.Errorf("Failed to get role data from Redis: %v", err)
	} else {
		role := new(models.Role)
		if err := json.Unmarshal([]byte(roleJson), role); err != nil {
			logger.Errorf("Failed to unmarshal role data: %v", err)
		} else {
			rsp.NickName = role.NickName
		}
	}

	groupName := ""
	if msg.Type == world {
		groupName = "world"
	} else if msg.Type == union {
		groupName, _ = m.client.HGet(ctx, redisUnionGroupByRidKey, fmt.Sprintf("%d", rid)).Result()
		if groupName == "" {
			logger.Warnf("chat not found rid(%d) in chatGroup\n", rid)
		}
	}

	// 广播消息到指定的聊天组
	if groupName != "" {
		err := m.app.GroupBroadcast(ctx, "gate", groupName, "onMessage", rsp)
		if err != nil {
			logger.Errorf("GroupBroadcast(%s,%s,%s) error: %v\n", "gate", groupName, "onMessage", err)
		}
	}

	// 将消息序列化并存储到 Redis
	msgJson, err := json.Marshal(rsp)
	if err != nil {
		logger.Errorf("消息序列化为 JSON 时失败: %v", err)
		return
	}

	// 存储消息到 Redis
	_, err = m.client.LPush(ctx, redisMsgListKey, msgJson).Result()
	if err != nil {
		logger.Errorf("Failed to store chat message in Redis: %v", err)
	}

	// 修剪消息列表，确保只存储最新的100条消息
	m.client.LTrim(ctx, redisMsgListKey, 0, 100)
}

// History 获取聊天历史记录
func (m *Chat) History(ctx context.Context) (*[]*protos.ChatMsg, error) {
	chatMsgs, err := m.client.LRange(ctx, redisMsgListKey, 0, 99).Result()
	if err != nil {
		return nil, fmt.Errorf("从 Redis 获取聊天历史失败: %v", err)
	}

	var msgList []*protos.ChatMsg
	for _, msgJson := range chatMsgs {
		chatMsg := &protos.ChatMsg{}
		if err := json.Unmarshal([]byte(msgJson), chatMsg); err != nil {
			return nil, fmt.Errorf("反序列化消息失败: %v", err)
		}
		msgList = append(msgList, chatMsg)
	}

	return &msgList, nil
}
