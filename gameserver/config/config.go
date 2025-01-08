package config

import (
	"HydraServer/pkg/config"
	"fmt"
	"github.com/spf13/viper"
	"time"
)

var (
	C        Config
	CFG      *viper.Viper
	ServerID int64
)

func init() {
	CFG = config.Load("./data/conf", "gameserver", &C)
	if C.PrintConfig {
		config.PrintWithJSON(&C)
	}
}

type Config struct {
	PrintConfig  bool
	Gorm         Gorm
	MySQL        MySQL
	JAEGER       JAEGER
	Log          Log
	RedisBackend RedisBackend
}

type Gorm struct {
	Debug             bool
	MaxLifetime       time.Duration
	MaxOpenConns      int
	MaxIdleConns      int
	TablePrefix       string
	EnableAutoMigrate bool
	LogOutputFile     string
}

type MySQL struct {
	Host       string
	Port       int
	User       string
	Password   string
	DBName     string
	Parameters string
}

func (a MySQL) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?%s", a.User, a.Password, a.Host, a.Port, a.DBName, a.Parameters)
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
