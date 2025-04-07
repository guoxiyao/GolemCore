// E:\go\GolemCore\internal\broker\loadbalanceer.go
package broker

import (
	"GolemCore/internal/rpc/protocol"
	"math/rand"
	"sync"
)

type LoadBalancer interface {
	AddNode(addr string)
	RemoveNode(addr string)
	SelectNode(task *protocol.ComputeTask) string
}

// 基于加权轮询的负载均衡
type WeightedRoundRobin struct {
	mu     sync.Mutex
	nodes  []string
	index  int
	weight map[string]int
}

func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		weight: make(map[string]int),
	}
}

func (w *WeightedRoundRobin) AddNode(addr string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.nodes = append(w.nodes, addr)
	w.weight[addr] = 10 // 初始权重
}

func (w *WeightedRoundRobin) RemoveNode(addr string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	for i, node := range w.nodes {
		if node == addr {
			w.nodes = append(w.nodes[:i], w.nodes[i+1:]...)
			delete(w.weight, addr)
			break
		}
	}
}

func (w *WeightedRoundRobin) SelectNode(task *protocol.ComputeTask) string {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.nodes) == 0 {
		return ""
	}

	// 动态调整权重算法
	totalWeight := 0
	weights := make([]int, len(w.nodes))
	for i, node := range w.nodes {
		weights[i] = w.weight[node]
		totalWeight += weights[i]
	}

	selected := rand.Intn(totalWeight)
	current := 0
	for i, weight := range weights {
		current += weight
		if selected < current {
			w.index = (i + 1) % len(w.nodes)
			return w.nodes[i]
		}
	}
	return w.nodes[0]
}

// 添加最少连接负载均衡器
type LeastConnLB struct {
	mu    sync.Mutex
	nodes map[string]int // 节点地址 -> 当前连接数
}

func NewLeastConnLB() *LeastConnLB {
	return &LeastConnLB{
		nodes: make(map[string]int),
	}
}

func (l *LeastConnLB) AddNode(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nodes[addr] = 0
}

func (l *LeastConnLB) RemoveNode(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.nodes, addr)
}

func (l *LeastConnLB) SelectNode(task *protocol.ComputeTask) string {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.nodes) == 0 {
		return ""
	}

	// 找到连接数最少的节点
	min := int(^uint(0) >> 1)
	var selected string
	for addr, conn := range l.nodes {
		if conn < min {
			min = conn
			selected = addr
		}
	}
	if selected != "" {
		l.nodes[selected]++
	}
	return selected
}
