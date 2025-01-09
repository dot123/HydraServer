package locate

import (
	"context"
	"github.com/redis/go-redis/v9"
)

const (
	defaultPrefix = "Hydra"
)

type Config struct {
	Addrs      []string
	DB         int
	MaxRetries int
	Prefix     string
	Username   string
	Password   string
}

type options struct {
	ctx context.Context

	// 客户端连接地址
	// 内建客户端配置，默认为[]string{"127.0.0.1:6379"}
	addrs []string

	// 数据库号
	// 内建客户端配置，默认为0
	db int

	// 用户名
	// 内建客户端配置，默认为空
	username string

	// 密码
	// 内建客户端配置，默认为空
	password string

	// 最大重试次数
	// 内建客户端配置，默认为3次
	maxRetries int

	// 客户端
	// 外部客户端配置，存在外部客户端时，优先使用外部客户端，默认为nil
	client redis.UniversalClient

	// 前缀
	// key前缀，默认为magic
	prefix string
}
