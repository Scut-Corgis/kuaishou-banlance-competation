package web

import (
	//"github.com/gin-contrib/pprof"

	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"strconv"
	"sync/atomic"
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/clientA"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/clientB"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"github.com/rs/zerolog/log"

	"github.com/valyala/fasthttp"
)

// 定义请求体结构体
type RequestBody struct {
	InputValue string `json:"inputValue"`
	Number     int    `json:"number"`
}

// 定义响应体结构体
type ResponseBody struct {
	ResultValue string `json:"resultValue"`
}

// 用于统计qps
var succReqQps int64
var failReqQps int64
var succTotal int64
var failTotal int64

func Start(port int) {
	s := &fasthttp.Server{
		Handler:     requestHandler,
		Concurrency: 9999,
	}
	// 启动 HTTP 服务器
	if err := s.ListenAndServe(":" + strconv.Itoa(port)); err != nil {
		log.Fatal().Msgf("Error in ListenAndServe: %s", err)
	}

}

// 主请求处理函数
func requestHandler(ctx *fasthttp.RequestCtx) {
	switch string(ctx.Path()) {
	case "/init":
		healthCheck(ctx)
	case "/sha256/calc":
		fasthttp.TimeoutHandler(handleSha256Request, time.Second, "request timed out")(ctx)
		//handleSha256Request(ctx)
	case "/report":
		getReport(ctx)
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

// 处理 SHA256 请求
func handleSha256Request(ctx *fasthttp.RequestCtx) {
	var requestBody RequestBody

	// 解析 JSON 请求体
	if err := json.Unmarshal(ctx.PostBody(), &requestBody); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}
	// if requestBody.InputValue == "" || requestBody.Number == 0 {
	// 	ctx.SetStatusCode(fasthttp.StatusBadRequest)
	// 	ctx.SetBody([]byte(`{"error":"参数不正确"}`))
	// 	return
	// }
	// 调用外部请求
	resultVal, err := clientB.NewClient().Sha256ReqForward(requestBody.InputValue, requestBody.Number)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusServiceUnavailable)
		if conf.DebugMode {
			atomic.AddInt64(&failReqQps, 1)
			atomic.AddInt64(&failTotal, 1)
		}
		return
	}

	// 增加qps
	if conf.DebugMode {
		atomic.AddInt64(&succTotal, 1)
		atomic.AddInt64(&succReqQps, 1)
	}

	// 	// 构建响应体
	// responseBody := ResponseBody{
	// 	ResultValue: resultVal,
	// }
	// // 返回响应
	// response, err := json.Marshal(responseBody)
	// if err != nil {
	// 	ctx.SetStatusCode(fasthttp.StatusInternalServerError)
	// 	return
	// }
	var buf bytes.Buffer
	buf.WriteString(`{"resultValue": "`)
	buf.WriteString(resultVal)
	buf.WriteString(`"}`)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody(buf.Bytes())
}

func getReport(ctx *fasthttp.RequestCtx) {
	var aReq clientA.AReq
	if err := json.Unmarshal(ctx.PostBody(), &aReq); err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBody([]byte(`{"error":"` + err.Error() + `"}`))
		return
	}

	clientB.NewClient().Wr.UpdateNodeReqNum(aReq.Addr, aReq.Num, aReq.TaskNum)
	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBody([]byte("good"))
}

func StartPprof() {
	var pprofPort int = 10005
	// 启动 pprof 服务器
	log.Debug().Msgf("pprof启动成功, port: %d", pprofPort)
	http.ListenAndServe(":"+strconv.Itoa(pprofPort), nil)
}
func ReportQps() {
	for {
		time.Sleep(5 * time.Second)                  // 每10秒执行一次
		scReqQps := atomic.SwapInt64(&succReqQps, 0) // 获取当前访问量并重置为 0
		flReqQps := atomic.SwapInt64(&failReqQps, 0)
		scTotal := atomic.LoadInt64(&succTotal)
		flTotal := atomic.LoadInt64(&failTotal)
		log.Info().Msgf("{succQps: %d, failQps: %d, succTotal: %d, failTotal: %d", scReqQps/5, flReqQps/5, scTotal, flTotal) // 打印当前的 QPS

		nodes := clientB.NewClient().Wr.CopyNodes()
		logstr := "各节点任务数与请求数："
		for _, node := range nodes {
			logstr += fmt.Sprintf("%s: %d, %d;  ", node.Name, node.TaskNum, node.ReqNum)
		}
		log.Info().Msg(logstr)
	}
}
