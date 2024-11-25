package conf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	HttpPort       int
	GrpcPort       int
	LogFilePath    string
	CpuCoreNum     int
	DebugMode      bool = false
	ZkAddr         string
	ZkServiceAPath = "/KSChallenge2024/public-hjg-a"
	ZkServiceBPath = "/KSChallenge2024/public-hjg-b"
	//Mock           bool
)

func InitConf() {
	flag.IntVar(&HttpPort, "hp", 8000, "http port")
	flag.IntVar(&GrpcPort, "gp", 8001, "gRPC port")
	flag.StringVar(&LogFilePath, "log_dir", "/home/huangjiegang/project/kuaishou-starter/A/logs/A.log", "日志文件路径")
	flag.IntVar(&CpuCoreNum, "cpu_num", runtime.NumCPU(), "cpu核心数")
	flag.BoolVar(&DebugMode, "debug", false, "是否debug模式, 性能更差")
	flag.StringVar(&ZkAddr, "zk_addr", "127.0.0.1:2181", "zk地址")
	//flag.BoolVar(&Mock, "mock", false, "sha256太快了, 整一个cpu密集型的")

	flag.Parse()

	fmt.Println("CpuCoreNum: ", CpuCoreNum)
}

func InitLog() {
	logFile := LogFilePath + "/B.log"
	// 创建日志文件
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Err(err).Msg("无法打开日志文件")
	}

	// 设置zerolog的输出
	multi := zerolog.MultiLevelWriter(file)
	if DebugMode {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		return filepath.Base(file) + ":" + strconv.Itoa(line)
	}
	log.Logger = log.With().Caller().Logger().Output(multi)
	// 自定义时间格式
	zerolog.TimeFieldFormat = time.RFC3339 // 使用默认的RFC3339格式，也可以自定义
}

func InitCpu() {
	runtime.GOMAXPROCS(CpuCoreNum)
	log.Info().Msgf("设置最大cpu核心数为:%v", CpuCoreNum)
}

func InitAll() {
	InitConf()
	InitLog()
	InitCpu()
}
