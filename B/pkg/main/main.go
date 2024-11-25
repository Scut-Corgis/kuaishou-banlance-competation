package main

import (
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/conf"
	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/server"
	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/web"
	"github.com/rs/zerolog/log"
	//"git.corp.kuaishou.com/infra/infra-framework-go.git/runtimemetrics"
)

func main() {

	conf.InitAll()
	// 开始report运行中的任务
	if conf.DebugMode {
		//go server.ReportLog()
		go web.StartPprof()
	}

	// 开启web服务
	go web.Start(conf.HttpPort)
	log.Info().Msgf("fasthttp启动成功,端口:%v", conf.HttpPort)

	// 开启gRPC server
	bser := server.NewServer()
	bser.Init()
	go bser.Start()
	log.Info().Msgf("grpc服务启动成功, port:%v", conf.GrpcPort)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGKILL)
	select {
	case sig := <-sigChan:
		log.Info().Msgf("received signal: %v", sig)
		bser.Stop()
		log.Info().Msg("bye...")
		return
	}
}
