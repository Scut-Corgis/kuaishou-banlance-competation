package web

import (
	//"github.com/gin-contrib/pprof"

	"net/http"
	_ "net/http/pprof"
	"strconv"

	"github.com/rs/zerolog/log"

	"github.com/valyala/fasthttp"
)

func Start(port int) {
	// 启动 HTTP 服务器
	if err := fasthttp.ListenAndServe(":"+strconv.Itoa(port), requestHandler); err != nil {
		log.Fatal().Msgf("Error in ListenAndServe: %s", err)
	}
}

// 主请求处理函数
func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/init":
		healthCheck(ctx)
	default:
		ctx.Error("Unsupported path", fasthttp.StatusNotFound)
	}
}

// 健康检查接口
func healthCheck(ctx *fasthttp.RequestCtx) {
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("ok")) // 返回 "ok"
	log.Debug().Msg("健康检查成功")
}

func StartPprof() {
	var pprofPort int = 10006
	// 启动 pprof 服务器
	log.Debug().Msgf("pprof启动成功, port: %d", pprofPort)
	http.ListenAndServe(":"+strconv.Itoa(pprofPort), nil)
}
