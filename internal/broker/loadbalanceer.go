package broker

import (
	"GolemCore/internal/rpc/protocol"
	"errors"
	"log"
	"math/rand"
	"sync"
)

// 负载均衡器接口
type LoadBalancer interface {
	AddNode(addr string)                                   //添加节点
	RemoveNode(addr string)                                //移除节点
	SelectNode(task *protocol.ComputeTask) (string, error) //选择节点
	SyncNodes(availableNodes []string)                     // 同步节点列表
}

// 基于加权轮询的负载均衡
type WeightedRoundRobin struct {
	mu     sync.Mutex
	nodes  []string       //节点列表
	index  int            //当前索引
	weight map[string]int //权重
}

// 创建加权轮询负载均衡器
func NewWeightedRoundRobin() *WeightedRoundRobin {
	return &WeightedRoundRobin{
		weight: make(map[string]int),
	}
}

// 添加节点
func (w *WeightedRoundRobin) AddNode(addr string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.nodes = append(w.nodes, addr)
	w.weight[addr] = 10 // 初始权重
}

// 移除节点
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

// 选择节点
func (w *WeightedRoundRobin) SelectNode(task *protocol.ComputeTask) (string, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.nodes) == 0 {
		log.Printf("没有可用的节点")
		return "", errors.New("no nodes available")
	}

	log.Printf("当前节点列表: %v", w.nodes)
	log.Printf("当前权重: %v", w.weight)

	// 动态调整权重算法
	totalWeight := 0
	weights := make([]int, len(w.nodes))
	for i, node := range w.nodes {
		weights[i] = w.weight[node]
		totalWeight += weights[i]
	}

	if totalWeight == 0 {
		log.Printf("所有节点权重为0")
		return "", errors.New("all nodes have zero weight")
	}

	// 确保至少有一个节点有正权重
	hasPositiveWeight := false
	for _, weight := range weights {
		if weight > 0 {
			hasPositiveWeight = true
			break
		}
	}

	if !hasPositiveWeight {
		log.Printf("没有节点有正权重")
		return "", errors.New("no nodes with positive weight")
	}

	selected := rand.Intn(totalWeight)
	current := 0
	for i, weight := range weights {
		if weight <= 0 {
			continue // 跳过权重为0或负数的节点
		}
		current += weight
		if selected < current {
			w.index = (i + 1) % len(w.nodes)
			selectedNode := w.nodes[i]
			log.Printf("选择节点: %s (权重: %d)", selectedNode, weight)
			return selectedNode, nil
		}
	}

	// 如果所有节点权重都为0，返回第一个有正权重的节点
	for i, weight := range weights {
		if weight > 0 {
			selectedNode := w.nodes[i]
			log.Printf("选择第一个有正权重的节点: %s", selectedNode)
			return selectedNode, nil
		}
	}

	log.Printf("无法选择节点：所有节点权重都为0")
	return "", errors.New("unable to select node: all nodes have zero weight")
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

// 添加节点
func (l *LeastConnLB) AddNode(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.nodes[addr] = 0
}

// 移除节点
func (l *LeastConnLB) RemoveNode(addr string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.nodes, addr)
}

// 选择节点
func (l *LeastConnLB) SelectNode(task *protocol.ComputeTask) (string, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if len(l.nodes) == 0 {
		return "", nil
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
	return selected, nil
}

// 同步节点列表
// SyncNodes 同步节点列表
func (l *LeastConnLB) SyncNodes(availableNodes []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 创建新的节点集合
	newNodes := make(map[string]struct{})
	for _, node := range availableNodes {
		newNodes[node] = struct{}{}
	}

	// 移除不存在的节点
	for addr := range l.nodes {
		if _, exists := newNodes[addr]; !exists {
			delete(l.nodes, addr)
		}
	}

	// 添加新节点
	for _, node := range availableNodes {
		if _, exists := l.nodes[node]; !exists {
			l.nodes[node] = 0
		}
	}

	log.Printf("同步后的节点列表: %v", l.nodes)
}

// SyncNodes 同步节点列表
func (w *WeightedRoundRobin) SyncNodes(availableNodes []string) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 创建新的节点集合
	newNodes := make(map[string]struct{})
	for _, node := range availableNodes {
		newNodes[node] = struct{}{}
	}

	// 移除不存在的节点
	for _, node := range w.nodes {
		if _, exists := newNodes[node]; !exists {
			delete(w.weight, node)
		}
	}

	// 添加新节点
	for _, node := range availableNodes {
		if _, exists := w.weight[node]; !exists {
			w.weight[node] = 10 // 初始权重
		}
	}

	// 更新节点列表
	w.nodes = availableNodes
	log.Printf("同步后的节点列表: %v", w.nodes)
}
