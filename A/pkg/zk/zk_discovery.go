package zk

import (
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"github.com/go-zookeeper/zk"
	"github.com/rs/zerolog/log"
)

// ServiceDiscovery 是服务发现的结构体
type ServiceDiscovery struct {
	zkConn      *zk.Conn
	zkAddr      []string
	servicePath string

	NodesPre map[string]struct{}
}

func NewServiceDiscovery(servicePath string) *ServiceDiscovery {
	return &ServiceDiscovery{
		zkAddr:      []string{conf.ZkAddr},
		servicePath: servicePath,
		NodesPre:    make(map[string]struct{}),
	}
}

// Connect 连接到 ZooKeeper
func (sd *ServiceDiscovery) Init() error {
	// 连接
	var err error
	sd.zkConn, _, err = zk.Connect(sd.zkAddr, time.Second)
	if err != nil {
		log.Fatal().Msg("zk连接失败")
	}
	return err
}

// DiscoverServices 获取已注册的所有服务地址
func (sd *ServiceDiscovery) DiscoverServices() ([]string, error) {
	children, _, err := sd.zkConn.Children(sd.servicePath)
	if err != nil {
		return nil, err
	}

	// 过滤一下，有时候zk删节点不及时，导致重复注册
	filter := map[string]bool{}
	for _, child := range children {
		serviceNodePath := sd.servicePath + "/" + child
		data, _, err := sd.zkConn.Get(serviceNodePath)
		if err != nil {
			log.Warn().Msgf("Failed to get service data from %s: %v", serviceNodePath, err)
			continue
		}
		filter[string(data)] = true
	}
	var services []string
	for key := range filter {
		services = append(services, key)
	}
	return services, nil
}

// Close 关闭 ZooKeeper 连接
func (sd *ServiceDiscovery) Stop() {
	if sd.zkConn != nil {
		sd.zkConn.Close()
		log.Info().Msg("ZooKeeper connection closed")
	}
}
