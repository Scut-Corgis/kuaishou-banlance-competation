package clientA

import (
	"sync"
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/zk"
	"github.com/rs/zerolog/log"
)

var (
	cc         *Client
	ccInitOnce sync.Once
)

type Client struct {
	sd               *zk.ServiceDiscovery
	sr               *zk.ServiceRegistrar
	sdUpdateInterval time.Duration

	addrs     []string
	addrMutex sync.Mutex

	done chan struct{}
}

func NewClient() *Client {
	ccInitOnce.Do(func() {
		cc = &Client{
			sd:               zk.NewServiceDiscovery(conf.ZkServiceAPath),
			sr:               zk.NewServiceRegistrar(conf.ZkServiceAPath),
			sdUpdateInterval: 5 * time.Second,
		}
	})
	return cc
}
func (c *Client) Init() {
	c.sd.Init()
	c.doOnceServerDiscover()
	go c.discoverAndUpdate()
	c.sr.Init()

	//c.doOnceMetrics()
	//go c.startMetrics()
}

func (c *Client) Stop() {
	c.sd.Stop()
	close(c.done)
}

func (c *Client) doOnceServerDiscover() {
	nodesCur := make(map[string]struct{})
	nodes, err := c.sd.DiscoverServices()
	log.Debug().Msgf("当前发现的A节点列表:%v", nodes)
	if err != nil {
		log.Error().Msgf("服务发现更新失败！err:%v", err)
	}
	for _, node := range nodes {
		// 跳过自己
		if node == zk.GetMyAddr() {
			continue
		}
		nodesCur[node] = struct{}{}
	}

	c.addrMutex.Lock()
	c.addrs = []string{}
	for node := range nodesCur {
		c.addrs = append(c.addrs, node)
	}
	c.addrMutex.Unlock()
}

func (c *Client) discoverAndUpdate() {
	// 独立协程持续服务发现，并更新相关数据结构，如连接池，metrics表
	discoverTicker := time.NewTicker(c.sdUpdateInterval)
	defer discoverTicker.Stop()

	for {
		select {
		case <-discoverTicker.C:
			cc.doOnceServerDiscover()
		case <-c.done:
			log.Warn().Msg("服务发现goroutine退出")
			return
		}
	}
}
