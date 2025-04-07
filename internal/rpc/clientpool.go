package rpc

import (
	"errors"
	"google.golang.org/grpc"
	"sync"

	"GolemCore/internal/rpc/protocol"
)

type ClientPool struct {
	target   string
	pool     sync.Pool
	mu       sync.Mutex
	maxConns int
	active   int
}

func NewClientPool(target string, maxConns int) *ClientPool {
	return &ClientPool{
		target:   target,
		maxConns: maxConns,
		pool: sync.Pool{
			New: func() interface{} {
				conn, _ := grpc.Dial(target, grpc.WithInsecure())
				return protocol.NewConwayServiceClient(conn)
			},
		},
	}
}

func (p *ClientPool) Get() (protocol.ConwayServiceClient, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active < p.maxConns {
		p.active++
		return p.pool.Get().(protocol.ConwayServiceClient), nil
	}

	// 等待可用连接（带超时）
	// 实际实现需要更复杂的等待策略
	return nil, ErrConnectionLimit
}

func (p *ClientPool) Put(client protocol.ConwayServiceClient) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.pool.Put(client)
	p.active--
}

func (p *ClientPool) HealthCheck() {
	// 定期检查连接健康状态
	// 自动重建失效连接
}

var ErrConnectionLimit = errors.New("connection pool limit reached")
