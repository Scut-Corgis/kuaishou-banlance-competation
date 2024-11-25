package conf

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var (
	HttpPort int

	LogFilePath string

	CpuCoreNum int

	DebugMode         bool = false
	ConnectPerPoolNum int
	ZkAddr            string
	ZkServiceAPath    = "/KSChallenge2024/public-hjg-a"
	ZkServiceBPath    = "/KSChallenge2024/public-hjg-b"
)

func InitConf() {
	flag.IntVar(&HttpPort, "hp", 8000, "http port")
	flag.StringVar(&LogFilePath, "log_dir", "/home/huangjiegang/project/kuaishou-starter/A/logs/A.log", "日志文件路径")
	flag.IntVar(&CpuCoreNum, "cpu_num", 4, "cpu核心数")

	flag.BoolVar(&DebugMode, "debug", false, "是否debug模式, 性能更差")
	flag.IntVar(&ConnectPerPoolNum, "conn_per_pool_num", 1, "grpc连接池连接数")
	flag.StringVar(&ZkAddr, "zk_addr", "127.0.0.1:2181", "zk地址")

	flag.Parse()
	// 打印一下
	fmt.Println("HttpPort:", HttpPort, "  cpu_num:", CpuCoreNum)
}

func InitLimit() {
	// 创建一个 rlimit 结构体来设置新的限制
	var limit syscall.Rlimit

	// 设置软限制和硬限制为 65536
	limit.Cur = 65536
	limit.Max = 65536

	// 修改进程的文件描述符限制
	err := syscall.Setrlimit(syscall.RLIMIT_NOFILE, &limit)
	if err != nil {
		fmt.Printf("Error setting file limit: %v\n", err)
		return
	}

	// 打印当前的文件描述符限制
	var currentLimit syscall.Rlimit
	err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &currentLimit)
	if err != nil {
		fmt.Printf("Error getting file limit: %v\n", err)
		return
	}
	fmt.Println("当前文件描述符限制为:", currentLimit.Max)
}
func InitLog() {
	logFile := LogFilePath + "/A.log"
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
	InitLimit()
}
