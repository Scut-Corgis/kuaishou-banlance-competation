package clientA

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/valyala/fasthttp"
)

type AReq struct {
	Addr    string `json:"addr"` // B实例的地址
	Num     int    `json:"num"`
	TaskNum int    `json:"task_num"`
}

var headerContentTypeJson = []byte("application/json")

var client *fasthttp.Client

func init() {
	readTimeout, _ := time.ParseDuration("500ms")
	writeTimeout, _ := time.ParseDuration("500ms")
	maxIdleConnDuration, _ := time.ParseDuration("1h")
	maxConnsPerHost := 3000 // 最多3000并发连接
	client = &fasthttp.Client{
		ReadTimeout:                   readTimeout,
		WriteTimeout:                  writeTimeout,
		MaxIdleConnDuration:           maxIdleConnDuration,
		NoDefaultUserAgentHeader:      true, // Don't send: User-Agent: fasthttp
		DisableHeaderNamesNormalizing: true, // If you set the case on your headers correctly you can enable this
		DisablePathNormalizing:        true,
		// increase DNS cache time to an hour instead of default minute
		Dial: (&fasthttp.TCPDialer{
			Concurrency:      4096,
			DNSCacheDuration: time.Hour,
		}).Dial,
		MaxConnsPerHost: maxConnsPerHost,
	}
	fmt.Println("clientA初始化成功")
}

func (c *Client) SendComputeTimesChange(bAddr string, num int, taskNum int) {
	c.addrMutex.Lock()
	addrsCopy := make([]string, len(c.addrs))
	copy(addrsCopy, c.addrs)
	c.addrMutex.Unlock()

	for _, addr := range addrsCopy {
		reqTimeout := time.Duration(100) * time.Millisecond
		aReq := &AReq{
			Num:     num,
			Addr:    bAddr,
			TaskNum: taskNum,
		}
		reqBytes, _ := json.Marshal(aReq)
		req := fasthttp.AcquireRequest()
		req.SetRequestURI("http://" + addr + "/report")
		req.Header.SetMethod(fasthttp.MethodPost)
		req.Header.SetContentTypeBytes(headerContentTypeJson)
		req.SetBodyRaw(reqBytes)

		resp := fasthttp.AcquireResponse()
		err := client.DoTimeout(req, resp, reqTimeout)
		fasthttp.ReleaseRequest(req)
		defer fasthttp.ReleaseResponse(resp)
		if err != nil {
			errName, known := httpConnError(err)
			if known {
				fmt.Fprintf(os.Stderr, "WARN conn error: %v\n", errName)
			} else {
				fmt.Fprintf(os.Stderr, "ERR conn failure: %v %v\n", errName, err)
			}
			return
		}
		statusCode := resp.StatusCode()
		if statusCode != http.StatusOK {
			fmt.Fprintf(os.Stderr, "ERR invalid HTTP response code: %d\n", statusCode)
			return
		}
	}
}

func httpConnError(err error) (string, bool) {
	var (
		errName string
		known   = true
	)

	switch {
	case errors.Is(err, fasthttp.ErrTimeout):
		errName = "timeout"
	case errors.Is(err, fasthttp.ErrNoFreeConns):
		errName = "conn_limit"
	case errors.Is(err, fasthttp.ErrConnectionClosed):
		errName = "conn_close"
	case reflect.TypeOf(err).String() == "*net.OpError":
		errName = "timeout"
	default:
		known = false
	}

	return errName, known
}
