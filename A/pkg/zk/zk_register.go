package zk

import (
	"os"
	"strconv"
	"strings"
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"github.com/go-zookeeper/zk"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// ServiceRegistrar 是服务注册的结构体
type ServiceRegistrar struct {
	zkConn         *zk.Conn
	zkAddr         []string
	servicePath    string
	serviceID      string
	done           chan struct{}
	updateInterval time.Duration
}

// NewServiceRegistrar 创建一个新的 ServiceRegistrar 实例
func NewServiceRegistrar(path string) *ServiceRegistrar {
	return &ServiceRegistrar{
		zkAddr:         []string{conf.ZkAddr},
		servicePath:    path,
		updateInterval: 5 * time.Second,
		done:           make(chan struct{}),
	}
}

func (sr *ServiceRegistrar) Init() {
	// 使用 UUID 作为节点名称
	sr.serviceID = uuid.New().String()
	serviceFullPath := sr.servicePath + "/" + sr.serviceID
	sr.Register(serviceFullPath)
	go sr.healthCheck(serviceFullPath)
	// 启动健康检查协程
	//go sr.healthCheck(path)
}

// Register 连接到 ZooKeeper 并注册服务
func (sr *ServiceRegistrar) Register(serviceFullPath string) string {
	var err error
	sr.zkConn, _, err = zk.Connect(sr.zkAddr, time.Second)
	if err != nil {
		log.Fatal().Msgf("Failed to connect to ZooKeeper: %v", err)
	}

	// 创建服务注册的父节点
	err = createPathRecursively(sr.zkConn, sr.servicePath)
	if err != nil {
		log.Fatal().Msgf("Failed to create service path: %v", err)
	}

	// 注册服务，创建临时节点
	_, err = sr.zkConn.Create(serviceFullPath, []byte(GetMyAddr()), zk.FlagEphemeral, zk.WorldACL(zk.PermAll))
	if err != nil {
		log.Fatal().Msgf("Failed to register service: %v", err)
	}

	log.Info().Msgf("Service registered at %s", serviceFullPath)
	return serviceFullPath

}

// healthCheck 定期更新 ZooKeeper 节点的状态
func (sr *ServiceRegistrar) healthCheck(path string) {
	healthCheckTicker := time.NewTicker(sr.updateInterval)
	defer healthCheckTicker.Stop()

	for {
		select {
		case <-healthCheckTicker.C:
			// 更新节点数据以保持活跃
			_, err := sr.zkConn.Set(path, []byte(GetMyAddr()), -1)
			if err != nil {
				log.Fatal().Msgf("Failed to zk heartbeat: %v", err)
			} else {
				log.Debug().Msgf("zk heartbeat updated at %s", path)
			}
		case <-sr.done:
			log.Info().Msg("Stopping health check goroutine")
			return
		}
	}
}

func createPathRecursively(zkConn *zk.Conn, path string) error {
	if path == "" {
		return nil
	}
	// 检查当前路径是否存在
	exists, _, err := zkConn.Exists(path)
	if err != nil {
		return err
	}
	// 如果路径存在，直接返回
	if exists {
		return nil
	}

	// 获取父路径
	parentPath := path[:strings.LastIndex(path, "/")]
	// 递归创建父路径
	if err := createPathRecursively(zkConn, parentPath); err != nil {
		return err
	}

	// 创建当前节点
	_, err = zkConn.Create(path, []byte{}, 0, zk.WorldACL(zk.PermAll))
	return err
}

// Stop 优雅关闭 ZooKeeper 连接和健康检查协程
func (sr *ServiceRegistrar) Stop() {
	close(sr.done) // 通知健康检查协程停止
	sr.zkConn.Close()
	log.Info().Msg("ZooKeeper connection closed")
}

// GetMyAddr 获取服务地址
func GetMyAddr() string {
	ip := os.Getenv("MY_POD_IP")
	if ip != "" {
		return ip + ":" + strconv.Itoa(conf.HttpPort)
	}
	return "127.0.0.1:" + strconv.Itoa(conf.HttpPort)
}
