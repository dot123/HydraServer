package main

import (
	"HydraServer/constant"
	"HydraServer/gateserver/config"
	"HydraServer/gateserver/logic"
	"HydraServer/pkg/errors"
	"HydraServer/pkg/log"
	"HydraServer/pkg/msgpack"
	"HydraServer/pkg/utils"
	"context"
	"flag"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/topfreegames/pitaya/v2"
	"github.com/topfreegames/pitaya/v2/acceptor"
	"github.com/topfreegames/pitaya/v2/acceptorwrapper"
	"github.com/topfreegames/pitaya/v2/cluster"
	"github.com/topfreegames/pitaya/v2/component"
	pitayaConfig "github.com/topfreegames/pitaya/v2/config"
	"github.com/topfreegames/pitaya/v2/constants"
	"github.com/topfreegames/pitaya/v2/groups"
	logruswrapper "github.com/topfreegames/pitaya/v2/logger/logrus"
	"github.com/topfreegames/pitaya/v2/metrics"
	"github.com/topfreegames/pitaya/v2/modules"
	"github.com/topfreegames/pitaya/v2/route"
	"github.com/topfreegames/pitaya/v2/serialize/json"
	"github.com/topfreegames/pitaya/v2/session"
	"github.com/topfreegames/pitaya/v2/tracing/jaeger"
	"os"
	"strings"
	"time"
)

var (
	App       pitaya.Pitaya
	gateLogic *logic.GateLogic
)

func configureFrontend(port int, log logrus.FieldLogger) func() {
	injector, cleanup, err := BuildInjector(&App, log)
	if err != nil {
		panic(err)
	}

	accountService := injector.AccountService
	App.Register(accountService,
		component.WithName("account"),
		component.WithNameFunc(strings.ToLower),
	)

	if err := App.AddRoute("game", func(
		ctx context.Context,
		route *route.Route,
		payload []byte,
		servers map[string]*cluster.Server,
	) (*cluster.Server, error) {
		s := App.GetSessionFromCtx(ctx)
		serverKey := cast.ToString(s.Get(constant.ServerKey))
		if serverKey != "" {
			return servers[serverKey], nil
		}
		return nil, errors.NewResponseError(constant.NoServersAvailable, nil)
	}); err != nil {
		panic(err)
	}

	configureDictionary()

	fmt.Printf("gateserver is running at 0.0.0.0:%d.\n", port)

	return cleanup
}

func configureDictionary() {
	dict := make(map[string]uint16)
	idx := uint16(1)
	for _, v := range config.RouteDict {
		dict[v] = idx
		idx++
	}

	if err := App.SetDictionary(dict); err != nil {
		fmt.Printf("error setting route dictionary %s\n", err.Error())
	}
}

func createAcceptor(conf *pitayaConfig.Config, port int, reporters []metrics.Reporter) acceptor.Acceptor {
	rateLimitConfig := pitayaConfig.NewRateLimitingConfig(conf)

	ws := acceptor.NewWSAcceptor(fmt.Sprintf(":%d", port))
	return acceptorwrapper.WithWrappers(
		ws,
		acceptorwrapper.NewRateLimitingWrapper(reporters, *rateLimitConfig))
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

func main() {
	ip := flag.String("ip", "127.0.0.1", "the ip to listen")
	port := flag.Int("port", 3250, "the port to listen")
	serializer := flag.String("serializer", "json", "json or protobuf")
	grpc := flag.Int("grpc", 0, "turn on grpc")
	grpcHost := flag.String("grpchost", "127.0.0.1", "the grpc server host")
	grpcPort := flag.Int("grpcport", 3434, "the grpc server port")
	serverId := flag.Int64("serverid", 1, "the server unique id")

	flag.Parse()

	config.ServerID = *serverId

	serverName := fmt.Sprintf("gateserver-%d", config.ServerID)

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

	app, bs, sessionPool := createApp(*serializer, *port, *grpc == 1, "gate", pitaya.Cluster, *grpcPort, map[string]string{
		constants.GRPCHostKey: *grpcHost,
		constants.GRPCPortKey: fmt.Sprintf("%d", *grpcPort),
		constant.ServerId:     fmt.Sprintf("%d", config.ServerID),
		"serverName":          serverName,
		"ip":                  *ip,
		"port":                fmt.Sprintf("%d", *port),
	}, config.CFG)

	App = app
	gateLogic = logic.NewGateLogic(&App, plog)

	if *grpc == 1 {
		if err := app.RegisterModule(bs, "bindingsStorage"); err != nil {
			panic(err)
		}
	}

	// 握手校验
	sessionPool.AddHandshakeValidator("MyCustomValidator", func(data *session.HandshakeData) error {
		if data.Sys.Version != "1.0.0" {
			return errors.New("unknown client version")
		}
		return nil
	})

	sessionPool.OnSessionBind(func(ctx context.Context, s session.Session) error {
		if s.UID() != "" {
			gateLogic.OnEnter(cast.ToInt64(s.UID()))
		}
		return nil
	})

	sessionPool.OnSessionClose(func(s session.Session) {
		if s.UID() != "" {
			gateLogic.OnExit(cast.ToInt64(s.UID()))
		}
	})

	pitaya.NewTimer(time.Minute, func() {
		plog.Infof("在线人数：%d", sessionPool.GetSessionCount())
	})

	cleanup := configureFrontend(*port, plog)

	defer func() {
		app.Shutdown()
		cleanup()
		lcleanup()
	}()

	app.Start()
}

func createApp(serializer string, port int, grpc bool, svType string, serverMode pitaya.ServerMode, rpcServerPort int, metadata map[string]string, cfg ...*viper.Viper) (pitaya.Pitaya, *modules.ETCDBindingStorage, session.SessionPool) {
	conf := pitayaConfig.NewConfig(cfg...)
	builder := pitaya.NewBuilderWithConfigs(true, svType, serverMode, metadata, conf)

	builder.AddAcceptor(createAcceptor(conf, port, builder.MetricsReporters))

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
