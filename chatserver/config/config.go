package config

import (
	"HydraServer/pkg/config"
	"github.com/spf13/viper"
	"time"
)

var (
	C        Config
	CFG      *viper.Viper
	ServerID int64
)

func init() {
	CFG = config.Load("./data/conf", "chatserver", &C)
	if C.PrintConfig {
		config.PrintWithJSON(&C)
	}
}

type Config struct {
	PrintConfig  bool
	JAEGER       JAEGER
	Log          Log
	RedisBackend RedisBackend
}

type JAEGER struct {
	ServiceName  string
	Disabled     bool
	SamplerParam float64
}

type Log struct {
	Level         int
	Format        string
	Output        string
	OutputFile    string
	RotationCount int
	RotationTime  int
}

type RedisBackend struct {
	Addrs           []string
	DB              int
	MaxRetries      int
	Username        string
	Password        string
	PoolSize        int
	MinIdleConns    int
	ConnMaxLifetime time.Duration
	ConnMaxIdleTime time.Duration
	DialTimeout     time.Duration
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}
