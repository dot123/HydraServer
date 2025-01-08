package locate

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"golang.org/x/sync/singleflight"
	"sort"
	"strings"
	"sync"
)

const (
	userGateKey            = "%s:locate:user:%d:gate"     // string
	userNodeKey            = "%s:locate:user:%d:node"     // hash
	clusterEventKey        = "%s:locate:cluster:%s:event" // channel
	serverOnlineUserNumKey = "%s:locate:server:%s:num"    // num
)

type Event struct {
	// 用户ID
	UID int64 `json:"uid"`
	// 事件类型
	Type EventType `json:"type"`
	// 实例ID
	ServerID string `json:"serverID"`
	// 实例类型
	ServerType string `json:"serverType"`
	// 实例名称
	ServerName string `json:"serverName"`
}

type EventType int

const (
	BindGate   EventType = iota + 1 // 绑定网关
	BindNode                        // 绑定节点
	UnbindGate                      // 解绑网关
	UnbindNode                      // 解绑节点
)

type Locator struct {
	ctx      context.Context
	cancel   context.CancelFunc
	opts     *options
	sfg      singleflight.Group // singleFlight
	watchers sync.Map
}

func NewLocator(ctx context.Context, config *Config) *Locator {
	o := &options{
		ctx:        ctx,
		addrs:      config.Addrs,
		db:         config.DB,
		maxRetries: config.MaxRetries,
		prefix:     config.Prefix,
		username:   config.Username,
		password:   config.Password,
	}

	if o.prefix == "" {
		o.prefix = defaultPrefix
	}

	if o.client == nil {
		o.client = redis.NewUniversalClient(&redis.UniversalOptions{
			Addrs:      o.addrs,
			DB:         o.db,
			Username:   o.username,
			Password:   o.password,
			MaxRetries: o.maxRetries,
		})

		pong, err := o.client.Ping(context.Background()).Result()
		if err != nil {
			panic(err)
		}
		fmt.Println(pong)
	}

	l := &Locator{}
	l.ctx, l.cancel = context.WithCancel(o.ctx)
	l.opts = o

	return l
}

// GetOnlineUserNum 获取在线人数
func (l *Locator) GetOnlineUserNum(ctx context.Context, serverType string, serverID string) (int64, error) {
	key := fmt.Sprintf(serverOnlineUserNumKey, l.opts.prefix, serverType)
	val, err, _ := l.sfg.Do(key, func() (interface{}, error) {
		val, err := l.opts.client.HGet(ctx, key, serverID).Int64()
		if err != nil && err != redis.Nil {
			return -1, err
		}

		return val, nil
	})
	if err != nil {
		return -1, err
	}

	return val.(int64), nil
}

// SetOnlineUserNum 设置在线人数
func (l *Locator) SetOnlineUserNum(ctx context.Context, serverType string, serverID string, num int64) error {
	key := fmt.Sprintf(serverOnlineUserNumKey, l.opts.prefix, serverType)
	err := l.opts.client.HSet(ctx, key, serverID, num).Err()

	return err
}

// LocateGate 定位用户所在网关
func (l *Locator) LocateGate(ctx context.Context, uid int64) (string, error) {
	key := fmt.Sprintf(userGateKey, l.opts.prefix, uid)
	val, err, _ := l.sfg.Do(key, func() (interface{}, error) {
		val, err := l.opts.client.Get(ctx, key).Result()
		if err != nil && err != redis.Nil {
			return "", err
		}

		return val, nil
	})
	if err != nil {
		return "", err
	}

	return val.(string), nil
}

// LocateNode 定位用户所在节点
func (l *Locator) LocateNode(ctx context.Context, uid int64, name string) (string, error) {
	key := fmt.Sprintf(userNodeKey, l.opts.prefix, uid)
	val, err, _ := l.sfg.Do(key+name, func() (interface{}, error) {
		val, err := l.opts.client.HGet(ctx, key, name).Result()
		if err != nil && err != redis.Nil {
			return "", err
		}

		return val, nil
	})
	if err != nil {
		return "", err
	}

	return val.(string), nil
}

// BindGate 绑定网关
func (l *Locator) BindGate(ctx context.Context, uid int64, serverID string) error {
	key := fmt.Sprintf(userGateKey, l.opts.prefix, uid)
	err := l.opts.client.Set(ctx, key, serverID, redis.KeepTTL).Err()
	if err != nil {
		return err
	}

	err = l.publish(ctx, BindGate, uid, serverID)
	if err != nil {
		fmt.Printf("location event publish failed: %v\n", err)
	}

	return nil
}

// BindNode 绑定节点
func (l *Locator) BindNode(ctx context.Context, uid int64, name, nid string) error {
	key := fmt.Sprintf(userNodeKey, l.opts.prefix, uid)
	err := l.opts.client.HSet(ctx, key, name, nid).Err()
	if err != nil {
		return err
	}

	err = l.publish(ctx, BindNode, uid, nid, name)
	if err != nil {
		fmt.Printf("location event publish failed: %v\n", err)
	}

	return nil
}

// UnbindGate 解绑网关
func (l *Locator) UnbindGate(ctx context.Context, uid int64, gid string) error {
	oldGID, err := l.LocateGate(ctx, uid)
	if err != nil {
		return err
	}

	if oldGID == "" || oldGID != gid {
		return nil
	}

	key := fmt.Sprintf(userGateKey, l.opts.prefix, uid)
	err = l.opts.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	err = l.publish(ctx, UnbindGate, uid, gid)
	if err != nil {
		fmt.Printf("location event publish failed: %v\n", err)
	}

	return nil
}

// UnbindNode 解绑节点
func (l *Locator) UnbindNode(ctx context.Context, uid int64, name string, nid string) error {
	oldNID, err := l.LocateNode(ctx, uid, name)
	if err != nil {
		return err
	}

	if oldNID == "" || oldNID != nid {
		return nil
	}

	key := fmt.Sprintf(userNodeKey, l.opts.prefix, uid)
	err = l.opts.client.Del(ctx, key).Err()
	if err != nil {
		return err
	}

	err = l.publish(ctx, UnbindNode, uid, nid, name)
	if err != nil {
		fmt.Printf("location event publish failed: %v\n", err)
	}

	return nil
}

func (l *Locator) publish(ctx context.Context, typ EventType, uid int64, serverID string, serverName ...string) error {
	var (
		serverType string
		name       string
	)
	switch typ {
	case BindGate, UnbindGate:
		serverType = "gate"
	case BindNode, UnbindNode:
		serverType = "node"
	}

	if len(serverName) > 0 {
		name = serverName[0]
	}

	msg, err := marshal(&Event{
		UID:        uid,
		Type:       typ,
		ServerID:   serverID,
		ServerType: serverType,
		ServerName: name,
	})
	if err != nil {
		return err
	}

	return l.opts.client.Publish(ctx, fmt.Sprintf(clusterEventKey, l.opts.prefix, serverType), msg).Err()
}

func (l *Locator) toUniqueKey(serverType ...string) string {
	sort.Slice(serverType, func(i, j int) bool {
		return serverType[i] < serverType[j]
	})

	keys := make([]string, 0, len(serverType))
	for _, insKind := range serverType {
		keys = append(keys, insKind)
	}

	return strings.Join(keys, "&")
}

// Watch 监听用户定位变化
func (l *Locator) Watch(ctx context.Context, serverType ...string) (*Watcher, error) {
	key := l.toUniqueKey(serverType...)

	v, ok := l.watchers.Load(key)
	if ok {
		return v.(*watcherMgr).fork(), nil
	}

	w, err := newWatcherMgr(ctx, l, key, serverType...)
	if err != nil {
		return nil, err
	}

	l.watchers.Store(key, w)

	return w.fork(), nil
}

func marshal(event *Event) (string, error) {
	buf, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return string(buf), nil
}

func unmarshal(data []byte) (*Event, error) {
	event := &Event{}
	if err := json.Unmarshal(data, event); err != nil {
		return nil, err
	}
	return event, nil
}
