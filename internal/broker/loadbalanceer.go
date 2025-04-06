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
