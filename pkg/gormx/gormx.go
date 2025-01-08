package gormx

import (
	"HydraServer/pkg/log"
	"database/sql"
	"fmt"
	mysqlDriver "github.com/go-sql-driver/mysql"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
	"time"
)

type Config struct {
	Debug         bool
	DSN           string
	MaxLifetime   time.Duration
	MaxOpenConns  int
	MaxIdleConns  int
	TablePrefix   string
	LogOutputFile string
}

func New(c *Config) (*gorm.DB, func(), error) {
	var dialector gorm.Dialector

	cfg, err := mysqlDriver.ParseDSN(c.DSN)
	if err != nil {
		return nil, nil, err
	}

	err = createDatabaseWithMySQL(cfg)
	if err != nil {
		return nil, nil, err
	}

	dialector = mysql.New(mysql.Config{
		DSN:                       c.DSN,
		DefaultStringSize:         256,   // string 类型字段的默认长度
		DisableDatetimePrecision:  true,  // 禁用 datetime 精度，MySQL 5.6 之前的数据库不支持
		DontSupportRenameIndex:    true,  // 重命名索引时采用删除并新建的方式，MySQL 5.7 之前的数据库和 MariaDB 不支持重命名索引
		DontSupportRenameColumn:   true,  // 用 `change` 重命名列，MySQL 8 之前的数据库和 MariaDB 不支持重命名列
		SkipInitializeWithVersion: false, // 根据当前 MySQL 版本自动配置
	})

	plog := logrus.New()
	cleanFunc, err := log.InitLogger(&log.Config{
		Level:         4,
		Format:        "json",
		Output:        "file",
		OutputFile:    c.LogOutputFile,
		RotationCount: 48,
		RotationTime:  1800,
	}, plog)

	newLogger := NewLogger(plog, LogConfig{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormLogger.Warn,
		IgnoreRecordNotFoundError: true,
		Colorful:                  false,
	})

	db, err := gorm.Open(dialector, &gorm.Config{
		Logger: newLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   c.TablePrefix,
			SingularTable: true,
		},
	})

	if err != nil {
		return nil, cleanFunc, err
	}

	if c.Debug {
		db = db.Debug()
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, cleanFunc, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, cleanFunc, err
	}

	sqlDB.SetMaxIdleConns(c.MaxIdleConns)
	sqlDB.SetMaxOpenConns(c.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(c.MaxLifetime * time.Second)

	//redisClient := redis.NewClient(&redis.Options{
	//	Addr: "localhost:6379",
	//})

	//cache, _ := cache.NewGorm2Cache(&config.CacheConfig{
	//	CacheLevel:           config.CacheLevelAll,
	//	CacheStorage:         config.CacheStorageRedis,
	//	RedisConfig:          cache.NewRedisConfigWithClient(redisClient),
	//	InvalidateWhenUpdate: true,      // when you create/update/delete objects, invalidate cache
	//	CacheTTL:             60000 * 5, // 5 m
	//	CacheMaxItemCnt:      5,         // if length of objects retrieved one single time
	//	DebugMode:            true,
	//	// exceeds this number, then don't cache
	//})
	//
	//db.Use(cache) // use gorm plugin

	return db, func() {
		sqlDB.Close()
		cleanFunc()
	}, nil
}

func createDatabaseWithMySQL(cfg *mysqlDriver.Config) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/", cfg.User, cfg.Passwd, cfg.Addr)
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	defer db.Close()

	query := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS `%s` DEFAULT CHARACTER SET = `utf8mb4`;", cfg.DBName)
	_, err = db.Exec(query)
	return err
}
