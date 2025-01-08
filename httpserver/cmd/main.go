package main

import (
	"HydraServer/httpserver/config"
	_ "HydraServer/httpserver/docs"
	"HydraServer/httpserver/ginx"
	middleware "HydraServer/httpserver/middleware"
	"HydraServer/httpserver/pkg/logger"
	"HydraServer/pkg/log"
	"HydraServer/pkg/utils"
	"context"
	"crypto/tls"
	"errors"
	"flag"
	"fmt"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/cluster"
	pitayaConfig "github.com/topfreegames/pitaya/v2/config"
	logruswrapper "github.com/topfreegames/pitaya/v2/logger/logrus"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// https://github.com/swaggo/swag/blob/master/README_zh-CN.md

// @title          httpserver API
// @version        1.0
// @description    This is a game management background. you can use the api key `ApiKeyAuth` to test the authorization filters.
// @termsOfService https://github.com

// @contact.name  conjurer
// @contact.url   https:/github.com/dot123
// @contact.email conjurer888888@gmail.com

// @license.name Apache 2.0
// @license.url  http://www.apache.org/licenses/LICENSE-2.0.html

// @host     127.0.0.1:8000
// @BasePath /api/v1

// @securityDefinitions.apikey ApiKeyAuth
// @in                         header
// @name                       Authorization

// VERSION Usage: go build -ldflags "-X main.VERSION=x.x.x"
var VERSION = "1.0.0"

func main() {
	port := flag.Int("port", 9000, "the port to listen")
	serverId := flag.Int64("serverid", 1, "the server unique id")

	flag.Parse()

	config.ServerID = *serverId

	serverName := fmt.Sprintf("httpserver-%d", config.ServerID)

	utils.SetConsoleTitle(serverName)

	c := config.C.Log
	lcleanup, err := log.InitLogger(&log.Config{
		Level:         c.Level,
		Format:        c.Format,
		Output:        c.Output,
		OutputFile:    c.OutputFile,
		RotationCount: c.RotationCount,
		RotationTime:  c.RotationTime,
	}, logrus.StandardLogger())

	if err != nil {
		panic(err)
	}

	ctx := logger.NewTagContext(context.Background(), "__main__")
	logger.WithContext(ctx).Printf("Start server,#run_mode %s,#version %s,#pid %d", config.C.RunMode, VERSION, os.Getpid())

	dieChan := make(chan bool)

	serviceDiscovery := newEtcdServiceDiscovery(ctx, dieChan, config.C.Etcd.Endpoints)
	err = serviceDiscovery.Init()
	if err != nil {
		panic(err)
	}

	injector, _, err := BuildInjector(serviceDiscovery, logger.StandardLogger())
	if err != nil {
		panic(err)
	}

	gin.SetMode(config.C.RunMode)

	app := gin.New()

	// Recover
	app.Use(middleware.RecoveryMiddleware())

	// CORS
	if config.C.CORS.Enable {
		app.Use(middleware.CORSMiddleware())
	}

	// RateLimiter
	if config.C.RateLimiter.Enable {
		app.Use(middleware.MaxAllowed(config.C.RateLimiter.Count))
	}

	// GZIP
	if config.C.GZIP.Enable {
		app.Use(gzip.Gzip(gzip.BestCompression,
			gzip.WithExcludedExtensions(config.C.GZIP.ExcludedExtentions),
			gzip.WithExcludedPaths(config.C.GZIP.ExcludedPaths),
		))
	}

	// Router register
	app.GET("/ping", func(c *gin.Context) {
		ginx.ResOk(c)
	})

	g := app.Group("/api/v1")

	injector.AccountController.RegisterRoute(g)

	// Swagger
	if config.C.Swagger {
		app.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		fmt.Printf("visit http://127.0.0.1:%d/swagger/index.html\n", *port)
	}

	go func() {
		cfg := config.C.HTTP
		addr := fmt.Sprintf("%s:%d", cfg.Host, *port)
		srv := &http.Server{
			Addr:         addr,
			Handler:      app,
			ReadTimeout:  10 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  30 * time.Second,
		}

		logger.WithContext(ctx).Printf("HTTP server is running at %s.", addr)

		if cfg.CertFile != "" && cfg.KeyFile != "" {
			srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12}
			err = srv.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
		} else {
			err = srv.ListenAndServe()
		}
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	sg := make(chan os.Signal)
	signal.Notify(sg, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGKILL, syscall.SIGTERM)

	// stop server
	select {
	case <-dieChan:
		logger.WithContext(ctx).Warn("the app will shutdown in a few seconds")
	case s := <-sg:
		logger.WithContext(ctx).Warn("got signal: ", s, ", shutting down...")
		close(dieChan)
	}

	logger.WithContext(ctx).Warn("server is stopping...")

	lcleanup()
}

func newEtcdServiceDiscovery(ctx context.Context, dieChan chan bool, endpoints []string) cluster.ServiceDiscovery {
	etcdSDConfig := pitayaConfig.NewDefaultEtcdServiceDiscoveryConfig()
	etcdSDConfig.Endpoints = endpoints
	pitaya.SetLogger(logruswrapper.NewWithFieldLogger(logger.StandardLogger()))
	server := cluster.NewServer(uuid.New().String(), "http", true, map[string]string{})
	serviceDiscovery, err := cluster.NewEtcdServiceDiscovery(*etcdSDConfig, server, dieChan)
	if err != nil {
		logger.WithContext(ctx).Fatalf("newEtcdServiceDiscovery error: %s", err.Error())
	}
	return serviceDiscovery
}
