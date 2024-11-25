package server

import (
	"context"
	"crypto/sha256"

	//"encoding/hex"
	"fmt"
	"net"
	"sync"

	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/server/hex"

	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/api/apiB"
	"git.corp.kuaishou.com/kuaishou-starter/B/pkg/conf"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
)

var (
	srv         *Server
	srcInitOnce sync.Once
)

type Server struct {
	apiB.UnimplementedBServer

	grpcServer *grpc.Server

	sr *ServiceRegistrar

	//wp *WorkerPool
	workerPool chan struct{}
}

// NewServer 创建一个新的 gRPC 服务器
func NewServer() *Server {
	srcInitOnce.Do(func() {
		srv = &Server{
			sr: NewServiceRegistrar(),
			//wp: NewWorkerPool(conf.CpuCoreNum), // 初始化信号量
			workerPool: make(chan struct{}, conf.CpuCoreNum),
		}
	})
	return srv
}

func (s *Server) Init() {
	//NewMetrics().Init()
	s.sr.Init()
	//s.wp.Init()
}

func (s *Server) ComputeSha256(_ context.Context, in *apiB.Sha256Request) (*apiB.Sha256Reply, error) {
	s.workerPool <- struct{}{}        // 获取一个工作位
	defer func() { <-s.workerPool }() // 释放工作位

	res := DoMultiSha256MyUgly(in.GetShaReq(), int(in.GetNum()))

	return &apiB.Sha256Reply{
		ShaResp: res,
	}, nil
}

// Start 启用gRPC server
func (s *Server) Start() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.GrpcPort))
	if err != nil {
		log.Fatal().Msgf("failed to listen: %v", err)
	}
	s.grpcServer = grpc.NewServer(
	// grpc.NumStreamWorkers(4),
	// grpc.MaxRecvMsgSize(1024*1024*10),
	// grpc.MaxSendMsgSize(1024*1024*10),
	// grpc.InitialWindowSize(1024*1024),
	// grpc.InitialConnWindowSize(1024*1024*10),
	// grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
	// 	MinTime:             5 * time.Second,
	// 	PermitWithoutStream: true,
	// }),
	)
	apiB.RegisterBServer(s.grpcServer, s)
	if err := s.grpcServer.Serve(lis); err != nil {
		log.Fatal().Msgf("grpc的serve停止: %v", err)
	}
}

// Stop 停止服务
func (s *Server) Stop() {
	s.sr.Stop()
	s.grpcServer.Stop()
}

// func DoMultiSha256(req string, times int) string {
// 	// 预分配一个字节切片用于存储 SHA-256 哈希结果
// 	hashBytes := make([]byte, sha256.Size)
// 	// 预分配一个字节切片用于存储十六进制字符串的字节
// 	hexBytes := make([]byte, hex.EncodedLen(len(hashBytes)))

// 	curbytes := []byte(req)
// 	hash := sha256.New()
// 	for i := 0; i < times; i++ {
// 		// 计算 SHA-256 哈希值
// 		hash.Reset()
// 		hash.Write(curbytes)
// 		hash.Sum(hashBytes[:0]) // 将结果写入预分配的字节切片
// 		// 将哈希值转为十六进制字符串
// 		hex.Encode(hexBytes, hashBytes)
// 		// 将十六进制字符串的字节切片作为下次计算的输入
// 		curbytes = hexBytes
// 	}

// 	return string(curbytes)
// }

func DoMultiSha256My(req string, times int) string {
	hash := sha256.New()
	// 预分配一个字节切片用于存储 SHA-256 哈希结果
	hashBytes := make([]byte, 32)
	// 预分配一个字节切片用于存储十六进制字符串的字节
	hexBytes := make([]byte, 64)

	curbytes := []byte(req)
	// 计算 SHA-256 哈希值
	hash.Write(curbytes)
	hash.Sum(hashBytes[:0]) // 将结果写入预分配的字节切片
	// 将哈希值转为十六进制字符串
	//hex.Encode(hexBytes, hashBytes)
	hex.EncodeAVX(&hexBytes[0], &hashBytes[0], 32, &hex.Lower[0])
	// 将十六进制字符串的字节切片作为下次计算的输入
	//curbytes = hexBytes

	// 自定义hash
	myhash := sha256.New2()
	for i := 1; i < times; i++ {
		myhash.Reset()
		myhash.Write2(hexBytes)
		myhash.Sum2(hashBytes[:0]) // 将结果写入预分配的字节切片
		// 将哈希值转为十六进制字符串
		//hex.Encode(hexBytes, hashBytes)
		hex.EncodeAVX(&hexBytes[0], &hashBytes[0], 32, &hex.Lower[0])

		// 将十六进制字符串的字节切片作为下次计算的输入
		//curbytes = hexBytes
	}
	return string(hexBytes)
}

// 丑陋的实现，可读性极差，纯为了性能
func DoMultiSha256MyUgly(req string, times int) string {
	hash := sha256.New()
	// 预分配一个字节切片用于存储 SHA-256 哈希结果
	hashBytes := make([]byte, 32)
	// 预分配一个字节切片用于存储十六进制字符串的字节
	hexBytes := make([]byte, 64)

	curbytes := []byte(req)
	// 计算 SHA-256 哈希值
	hash.Write(curbytes)
	hash.Sum(hashBytes[:0]) // 将结果写入预分配的字节切片
	// 将哈希值转为十六进制字符串
	//hex.Encode(hexBytes, hashBytes)
	hex.EncodeAVX(&hexBytes[0], &hashBytes[0], 32, &hex.Lower[0])
	// 将十六进制字符串的字节切片作为下次计算的输入
	//curbytes = hexBytes

	// 自定义hash
	myhash := sha256.New2()
	for i := 1; i < times; i++ {
		myhash.Reset()
		//myhash.Write2(hexBytes)
		sha256.Block(myhash, hexBytes[:64])

		//myhash.Sum2(hashBytes[:0]) // 将结果写入预分配的字节切片
		sha256.Block2(myhash)

		// 将哈希值转为十六进制字符串
		//hex.Encode(hexBytes, hashBytes)
		hex.EncodeAVX(&hexBytes[0], myhash.GetH(), 32, &hex.Lower[0])

		// 将十六进制字符串的字节切片作为下次计算的输入
		//curbytes = hexBytes
	}
	return string(hexBytes)
}

// func StringToBytes(s string) []byte {
// 	return unsafe.Slice(unsafe.StringData(s), len(s))
// }
