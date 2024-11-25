package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "net/http/pprof"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/clientA"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/clientB"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/web"
	"github.com/rs/zerolog/log"
	//"git.corp.kuaishou.com/infra/infra-framework-go.git/runtimemetrics"
)

func main() {

	conf.InitAll()

	if conf.DebugMode {
		go web.ReportQps()
	}
	// grpc客户端初始化
	clientB.NewClient().Init()
	// http客户端初始化
	clientA.NewClient().Init()
	// 开启web服务
	go web.Start(conf.HttpPort)
	// A服务注册

	fmt.Println("conf.HttpPort:", conf.HttpPort)
	log.Info().Msgf("fasthttp启动成功,端口:%d", conf.HttpPort)

	// 如果debug，开启pprof服务器
	// if conf.DebugMode {
	// 	go web.StartPprof()
	// }

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	select {
	case sig := <-sigChan:
		log.Info().Msgf("received signal: %v", sig)
		clientB.NewClient().Stop()
		log.Info().Msg("bye...")
		return
	}
}
