// E:\go\GolemCore\internal\broker\broker.go
package broker

import (
	"container/list"
	"errors"
	"sync"
	"time"

	"GolemCore/internal/rpc/protocol"
)

var (
	ErrNoAvailableWorkers = errors.New("no available workers")
)

type Broker struct {
	mu          sync.RWMutex
	workers     map[string]*WorkerNode // 可用工作节点
	taskQueue   *list.List             // 待处理任务队列
	lbStrategy  LoadBalancer           // 负载均衡策略
	healthCheck *HealthChecker
}

type WorkerNode struct {
	Address    string
	LastActive time.Time
	Capacity   int     // 剩余计算能力
	Throughput float64 // 最近一分钟吞吐量
}

func NewBroker(strategy LoadBalancer) *Broker {
	b := &Broker{
		workers:    make(map[string]*WorkerNode),
		taskQueue:  list.New(),
		lbStrategy: strategy,
	}
	b.healthCheck = NewHealthChecker(b, 30*time.Second)
	go b.healthCheck.Start()
	return b
}

// 添加工作节点
func (b *Broker) AddWorker(addr string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if _, exists := b.workers[addr]; !exists {
		b.workers[addr] = &WorkerNode{
			Address:    addr,
			LastActive: time.Now(),
			Capacity:   100, // 初始容量
		}
		b.lbStrategy.AddNode(addr)
	}
}

// 移除失效节点
func (b *Broker) RemoveWorker(addr string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.workers, addr)
	b.lbStrategy.RemoveNode(addr)
}

// 分发任务到工作节点
func (b *Broker) DispatchTask(task *protocol.ComputeTask) (string, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if len(b.workers) == 0 {
		return "", ErrNoAvailableWorkers
	}

	target := b.lbStrategy.SelectNode(task)
	return target, nil
}
