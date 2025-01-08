package main

import (
	"HydraServer/constant"
	"HydraServer/gameserver/config"
	protos "HydraServer/gameserver/protos"
	"HydraServer/pkg/errors"
	log "HydraServer/pkg/log"
	"HydraServer/pkg/msgpack"
	"HydraServer/pkg/redisbackend"
	"HydraServer/pkg/utils"
	"context"
	"flag"
	"fmt"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/cluster"
	"github.com/topfreegames/pitaya/v2/component"
	pitayaConfig "github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/constants"
	"github.com/topfreegames/pitaya/v2/groups"
	logruswrapper "github.com/topfreegames/pitaya/v2/logger/logrus"
	"github.com/topfreegames/pitaya/v2/modules"
	"github.com/topfreegames/pitaya/v2/serialize/json"
	"github.com/topfreegames/pitaya/v2/session"
	"github.com/topfreegames/pitaya/v2/tracing/jaeger"
	"os"
	"strings"
	"time"
)

var (
	App pitaya.Pitaya
)

func configureBackend(redisBackend *redisbackend.RedisBackend, log logrus.FieldLogger) func() {
	injector, cleanup, err := BuildInjector(&App, redisBackend, log)
	if err != nil {
		panic(err)
	}

	roleService := injector.RoleService
	App.Register(roleService,
		component.WithName("role"),
		component.WithNameFunc(strings.ToLower),
	)

	App.RegisterRemote(injector.RoleRemoteService,
		component.WithName("role"),
		component.WithNameFunc(strings.ToLower),
	)
	return cleanup
}

func configureJaeger() {
	if config.C.JAEGER.ServiceName == "" {
		return
	}
	options := jaeger.Options{
		Disabled:    config.C.JAEGER.Disabled,
		Probability: config.C.JAEGER.SamplerParam,
		ServiceName: config.C.JAEGER.ServiceName,
	}
	_, err := jaeger.Configure(options)
	if err != nil {
		panic(err)
	}
}

func beforeHandler(ctx context.Context, in interface{}) (context.Context, interface{}, error) {
	s := App.GetSessionFromCtx(ctx)

	switch in.(type) {
	case nil:
	case *protos.EnterServerReq, *protos.CreateRoleReq:
		if s.UID() == "" {
			return ctx, in, errors.NewResponseError(constant.UserNotInConnect, nil)
		}
	default:
		rid := cast.ToInt64(s.Get("rid"))
		if rid == 0 {
			return ctx, in, errors.NewResponseError(constant.RoleNotInConnect, nil)
		}
	}

	return ctx, in, nil
}

func afterHandler(ctx context.Context, resp interface{}, err error) (interface{}, error) {
	if resp != nil {

	}
	return resp, err
}

// 初始化游戏数据表
func initGameData(logger logrus.FieldLogger) {
	logger.Infoln("load gamedata")
}

func initRedisBackend(logger logrus.FieldLogger) *redisbackend.RedisBackend {
	logger.Infoln("initRedisBackend")

	redisBackendConfig := config.C.RedisBackend
	redisBackend := redisbackend.NewRedisBackend(&redis.UniversalOptions{
		Addrs:           redisBackendConfig.Addrs,
		Username:        redisBackendConfig.Username,
		Password:        redisBackendConfig.Password,
		DB:              redisBackendConfig.DB,
		MaxRetries:      redisBackendConfig.MaxRetries,
		DialTimeout:     redisBackendConfig.DialTimeout * time.Second,     // 连接超时时间
		ReadTimeout:     redisBackendConfig.ReadTimeout * time.Second,     // 读取超时时间
		WriteTimeout:    redisBackendConfig.WriteTimeout * time.Second,    // 写入超时时间
		ConnMaxIdleTime: redisBackendConfig.ConnMaxIdleTime * time.Second, // 连接池中连接的最大闲置时间
		ConnMaxLifetime: redisBackendConfig.ConnMaxLifetime * time.Second, // 连接最大生命周期
		PoolSize:        redisBackendConfig.PoolSize,                      // 连接池大小
		MinIdleConns:    redisBackendConfig.MinIdleConns,                  // 最小空闲连接数
		TLSConfig:       nil,
	})
	return redisBackend
}

func main() {
	serializer := flag.String("serializer", "json", "json or protobuf")
	grpc := flag.Int("grpc", 0, "turn on grpc")
	grpcHost := flag.String("grpchost", "127.0.0.1", "the grpc server host")
	grpcPort := flag.Int("grpcport", 3434, "the grpc server port")
	serverId := flag.Int64("serverid", 1, "the server unique id")

	flag.Parse()

	config.ServerID = *serverId

	serverName := fmt.Sprintf("gameserver-%d", config.ServerID)

	utils.SetConsoleTitle(serverName)

	plog := logrus.New()
	c := config.C.Log
	lcleanup, err := log.InitLogger(&log.Config{
		Level:         c.Level,
		Format:        c.Format,
		Output:        c.Output,
		OutputFile:    c.OutputFile,
		RotationCount: c.RotationCount,
		RotationTime:  c.RotationTime,
	}, plog)

	if err != nil {
		panic(err)
	}

	plog.Infof("Start server,#pid %d", os.Getpid())

	pitaya.SetLogger(logruswrapper.NewWithFieldLogger(plog))

	configureJaeger()

	redisBackend := initRedisBackend(plog)

	initGameData(plog)

	app, bs, _ := createApp(*serializer, *grpc == 1, "game", pitaya.Cluster, *grpcPort, map[string]string{
		constants.GRPCHostKey: *grpcHost,
		constants.GRPCPortKey: fmt.Sprintf("%d", *grpcPort),
		"serverId":            fmt.Sprintf("%d", config.ServerID),
		"serverName":          serverName,
	}, config.CFG)

	App = app

	if *grpc == 1 {
		if err := app.RegisterModule(bs, "bindingsStorage"); err != nil {
			panic(err)
		}
	}

	cleanup := configureBackend(redisBackend, plog)

	defer func() {
		app.Shutdown()
		cleanup()
		lcleanup()
	}()

	app.Start()
}

func createApp(serializer string, grpc bool, svType string, serverMode pitaya.ServerMode, rpcServerPort int, metadata map[string]string, cfg ...*viper.Viper) (pitaya.Pitaya, *modules.ETCDBindingStorage, session.SessionPool) {
	conf := pitayaConfig.NewConfig(cfg...)
	builder := pitaya.NewBuilderWithConfigs(false, svType, serverMode, metadata, conf)

	builder.HandlerHooks.BeforeHandler.PushBack(beforeHandler)
	builder.HandlerHooks.AfterHandler.PushBack(afterHandler)

	builder.Groups = groups.NewMemoryGroupService(*pitayaConfig.NewDefaultMemoryGroupConfig())

	if serializer == "json" {
		builder.Serializer = json.NewSerializer()
	} else if serializer == "msgpack" {
		builder.Serializer = msgpack.NewSerializer()
	} else {
		panic("serializer should be either json or msgpack")
	}

	var bs *modules.ETCDBindingStorage
	if grpc {
		grpcServerConfig := pitayaConfig.NewDefaultGRPCServerConfig()
		grpcServerConfig.Port = rpcServerPort
		gs, err := cluster.NewGRPCServer(*grpcServerConfig, builder.Server, builder.MetricsReporters)
		if err != nil {
			panic(err)
		}

		bs = modules.NewETCDBindingStorage(builder.Server, builder.SessionPool, *pitayaConfig.NewETCDBindingConfig(conf))

		gc, err := cluster.NewGRPCClient(
			*pitayaConfig.NewGRPCClientConfig(conf),
			builder.Server,
			builder.MetricsReporters,
			bs,
			cluster.NewInfoRetriever(*pitayaConfig.NewInfoRetrieverConfig(conf)),
		)
		if err != nil {
			panic(err)
		}
		builder.RPCServer = gs
		builder.RPCClient = gc
	}

	return builder.Build(), bs, builder.SessionPool
}
