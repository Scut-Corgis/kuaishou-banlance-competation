package clientB

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/api/apiB"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/clientA"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/zk"
	"github.com/rs/zerolog/log"
)

var (
	cc             *Client
	ccInitOnce     sync.Once
	allColdTimeout = 320 * time.Millisecond
	requestTimeout = 320 * time.Millisecond
	coldInterval   = 1500 * time.Millisecond
)

type Client struct {
	sd               *zk.ServiceDiscovery
	sdUpdateInterval time.Duration

	ClientPools map[string]*ClientPool
	CpMutex     sync.Mutex

	Wr *WeightedRandom // 内部带锁

	done chan struct{}
}

func NewClient() *Client {
	ccInitOnce.Do(func() {
		cc = &Client{
			ClientPools:      make(map[string]*ClientPool),
			sd:               zk.NewServiceDiscovery(conf.ZkServiceBPath),
			sdUpdateInterval: 5 * time.Second,
			done:             make(chan struct{}),
			Wr:               NewWeightedRandom(),
		}
	})
	return cc
}
func (c *Client) Init() {
	c.sd.Init()
	c.doOnceServerDiscover()
	go c.discoverAndUpdate()

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
	log.Debug().Msgf("当前发现的B节点列表:%v", nodes)
	if err != nil {
		log.Error().Msgf("服务发现更新失败！err:%v", err)
	}
	for _, node := range nodes {
		nodesCur[node] = struct{}{}
	}
	// 对比新老服务发现，更新连接池
	nodesAdd := make(map[string]struct{})
	nodesDelete := make(map[string]struct{})
	for node := range nodesCur {
		if _, ok := c.sd.NodesPre[node]; !ok {
			nodesAdd[node] = struct{}{}
		}
	}
	for node := range c.sd.NodesPre {
		if _, ok := nodesCur[node]; !ok {
			nodesDelete[node] = struct{}{}
		}
	}
	// TODO：锁粒度不小，一旦服务发现列表变化，就会性能抖动.
	// 解决方法，单个clientPool也有锁，直接从pools里面取出pool处理
	c.CpMutex.Lock()
	for node := range nodesDelete {
		c.ClientPools[node].Close()
		delete(c.ClientPools, node)
	}
	for node := range nodesAdd {
		clientPool := NewClientPool(node, NewDefaultClientOption())
		clientPool.init()
		c.ClientPools[node] = clientPool
		log.Info().Msgf("发现新节点%v，连接池建立", node)
	}
	c.CpMutex.Unlock()
	c.sd.NodesPre = make(map[string]struct{})
	for node := range nodesCur {
		c.sd.NodesPre[node] = struct{}{}
	}

	if len(nodesAdd) != 0 || len(nodesDelete) != 0 {
		c.ResetWr()
		log.Info().Msgf("服务发现发生更新, 当前节点列表:%v", nodes)
	}
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

// 超时的情况不再重选，超时时间延长
func (c *Client) Sha256ReqForward(req string, num int) (string, error) {

	// 最多选3次, 防止部分过载
	for i := 0; i < 3; i++ {
		node := c.Wr.Choose()
		if node == "" {
			log.Error().Msgf("轮次:%d . 选不出B节点, 全在冷却", i)
			// 等一会下一轮试试
			time.Sleep(allColdTimeout)
			continue
		}

		// TODO: 并发优化，这里加的锁粒度不小，因为getConn的过程可能伴随连接重连,出现全部等待的现象
		// 解决方法： 用pool级别的锁，不用pools
		c.CpMutex.Lock()
		if _, ok := c.ClientPools[node]; !ok {
			errStr := fmt.Sprintf("连接池不存在NodeSelector选取的节点,node:%v", node)
			log.Warn().Msg(errStr)
			c.CpMutex.Unlock()
			log.Error().Msg(errStr)
			return "", errors.New(errStr)
		}
		conn, err := c.ClientPools[node].getConn()
		c.CpMutex.Unlock()

		if err != nil {
			log.Error().Msgf("节点node:%v, 获取连接失败err:%v", node, err)
			return "", err
		}

		bcli := apiB.NewBClient(conn)

		ctx, _ := context.WithTimeout(context.Background(), requestTimeout*time.Duration(3-i))
		go clientA.NewClient().SendComputeTimesChange(node, num, 1)
		c.Wr.UpdateNodeReqNum(node, num, 1)
		r, err := bcli.ComputeSha256(ctx, &apiB.Sha256Request{ShaReq: req, Num: int32(num)})
		c.Wr.UpdateNodeReqNum(node, num*(-1), -1)
		go clientA.NewClient().SendComputeTimesChange(node, num*(-1), -1)
		if err != nil {
			log.Error().Msgf("轮次: %d. 请求失败: %v", i, err)
			// 冷却
			//随机值实测没意义
			//randomFactor := 0.5 + rand.Float64()*(1.5-0.5)
			//randomCold := time.Duration(float64(coldInterval.Milliseconds())*randomFactor) * time.Millisecond
			c.Wr.UpdateNodeColdTime(node, time.Now().Add(coldInterval).Unix())

			// 超时还是别重选了
			return "", errors.New("超时")
		}
		return r.GetShaResp(), nil
	}
	log.Error().Msg("三次尝试均失败了")
	return "", errors.New("三次尝试均失败了")
}
