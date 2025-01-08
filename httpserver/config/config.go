package config

import (
	"HydraServer/pkg/config"
	"fmt"
	"time"
)

var (
	C        Config
	ServerID int64
)

func init() {
	config.Load("./data/conf", "httpserver", &C)
	if C.PrintConfig {
		config.PrintWithJSON(&C)
	}
}

type Config struct {
	RunMode     string
	Swagger     bool
	PrintConfig bool
	HTTP        HTTP
	Log         Log
	RateLimiter RateLimiter
	CORS        CORS
	GZIP        GZIP
	Gorm        Gorm
	MySQL       MySQL
	Etcd        Etcd
}

func (c *Config) IsDebugMode() bool {
	return c.RunMode == "debug"
}

type Log struct {
	Level         int
	Format        string
	Output        string
	OutputFile    string
	RotationCount int
	RotationTime  int
}

type HTTP struct {
	Host               string
	Port               int
	CertFile           string
	KeyFile            string
	ShutdownTimeout    int
	MaxContentLength   int64
	MaxReqLoggerLength int `default:"1024"`
	MaxResLoggerLength int
}

type RateLimiter struct {
	Enable bool
	Count  int
}

type CORS struct {
	Enable           bool
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	AllowCredentials bool
	MaxAge           time.Duration
}

type GZIP struct {
	Enable             bool
	ExcludedExtentions []string
	ExcludedPaths      []string
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

type Etcd struct {
	Endpoints []string
}
