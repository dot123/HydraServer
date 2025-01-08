package config

import (
	"HydraServer/pkg/config"
	"encoding/json"
	"fmt"
	"github.com/spf13/viper"
	"log"
	"os"
	"time"
)

var (
	C         Config
	CFG       *viper.Viper
	ServerID  int64
	RouteDict []string
)

func init() {
	CFG = config.Load("./data/conf", "gateserver", &C)
	if C.PrintConfig {
		config.PrintWithJSON(&C)
	}

	// 读取json配置
	data, err := os.ReadFile("./data/conf/routedict.json")
	if err != nil {
		log.Fatalf("%v\n", err)
		return
	}

	if err = json.Unmarshal(data, &RouteDict); err != nil {
		log.Fatalf("%v\n", err)
		return
	}
}

type Config struct {
	PrintConfig bool
	JAEGER      JAEGER
	Log         Log
	Gorm        Gorm
	MySQL       MySQL
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
