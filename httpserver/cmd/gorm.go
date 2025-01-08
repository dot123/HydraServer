package main

import (
	"HydraServer/gateserver/models"
	"HydraServer/httpserver/config"
	"HydraServer/pkg/gormx"
	"gorm.io/gorm"
)

func InitGormDB() (*gorm.DB, func(), error) {
	db, cleanFunc, err := NewGormDB()
	if err != nil {
		return nil, nil, err
	}
	if config.C.Gorm.EnableAutoMigrate {
		err = db.AutoMigrate(
			new(models.UserInfo),
		)
		if err != nil {
			return nil, cleanFunc, err
		}
	}

	return db, cleanFunc, nil
}

func NewGormDB() (*gorm.DB, func(), error) {
	return gormx.New(&gormx.Config{
		Debug:         config.C.Gorm.Debug,
		DSN:           config.C.MySQL.DSN(),
		MaxIdleConns:  config.C.Gorm.MaxIdleConns,
		MaxLifetime:   config.C.Gorm.MaxLifetime,
		MaxOpenConns:  config.C.Gorm.MaxOpenConns,
		TablePrefix:   config.C.Gorm.TablePrefix,
		LogOutputFile: config.C.Gorm.LogOutputFile,
	})
}
