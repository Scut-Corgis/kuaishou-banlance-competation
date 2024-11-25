package clientB

import (
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"git.corp.kuaishou.com/kuaishou-starter/A/pkg/conf"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

var (
	// ip:port -> 连接池

	ErrStringSplit    = errors.New("err string split")
	ErrNotFoundClient = errors.New("not found grpc conn")
	ErrConnShutdown   = errors.New("grpc conn shutdown")

	//defaultClientPoolCap //默认连接数

	defaultDialTimeout      = 5 * time.Second
	defaultKeepAlive        = 30 * time.Second
	defaultKeepAliveTimeout = 10 * time.Second
)

type ClientOption struct {
	DialTimeout      time.Duration
	KeepAlive        time.Duration
	KeepAliveTimeout time.Duration
	ClientPoolSize   int
}

func NewDefaultClientOption() *ClientOption {
	return &ClientOption{
		DialTimeout:      defaultDialTimeout,
		KeepAlive:        defaultKeepAlive,
		KeepAliveTimeout: defaultKeepAliveTimeout,
	}
}

type ClientPool struct {
	option   *ClientOption
	capacity int64
	next     int64
	target   string

	sync.Mutex

	conns []*grpc.ClientConn
}

func NewClientPool(target string, option *ClientOption) *ClientPool {
	if option.ClientPoolSize <= 0 {
		option.ClientPoolSize = conf.ConnectPerPoolNum
	}

	return &ClientPool{
		target:   target,
		conns:    make([]*grpc.ClientConn, option.ClientPoolSize),
		capacity: int64(option.ClientPoolSize),
		option:   option,
	}
}

func (cc *ClientPool) init() {
	for idx := range cc.conns {
		conn, _ := cc.connect()
		cc.conns[idx] = conn
	}
	log.Info().Msgf("节点:%v, 连接池数量:%v", cc.target, len(cc.conns))
}

func (cc *ClientPool) checkState(conn *grpc.ClientConn) error {
	state := conn.GetState()
	switch state {
	case connectivity.TransientFailure, connectivity.Shutdown:
		return ErrConnShutdown
	}

	return nil
}

func (cc *ClientPool) getConn() (*grpc.ClientConn, error) {
	cc.Lock()
	defer cc.Unlock()
	var (
		idx  int64
		next int64

		err error
	)

	next = atomic.AddInt64(&cc.next, 1)
	idx = next % cc.capacity
	conn := cc.conns[idx]
	if conn != nil && cc.checkState(conn) == nil {
		return conn, nil
	}

	// gc old conn
	if conn != nil {
		conn.Close()
	}

	conn, err = cc.connect()
	if err != nil {
		return nil, err
	}
	log.Info().Msgf("原连接断开，一条连接重新建立:%v", cc.target)
	cc.conns[idx] = conn
	return conn, nil
}

func (cc *ClientPool) connect() (*grpc.ClientConn, error) {
	conn, err := grpc.NewClient(cc.target, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:    cc.option.KeepAlive,
		Timeout: cc.option.KeepAliveTimeout},
	))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func (cc *ClientPool) Close() {
	cc.Lock()
	defer cc.Unlock()

	for _, conn := range cc.conns {
		if conn == nil {
			continue
		}

		conn.Close()
	}
}
