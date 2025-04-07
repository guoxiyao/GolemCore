package broker

import (
	"bytes"
	"container/list"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"GolemCore/internal/game"
	"GolemCore/internal/rpc/protocol"
)

//包含任务队列和状态管理

var (
	ErrNoAvailableWorkers = errors.New("no available workers")
)

// TaskStatus 任务状态
type TaskStatus int

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusProcessing
	TaskStatusCompleted
	TaskStatusFailed
)

// Task 任务结构
type Task struct {
	ID          string
	Grid        *game.Grid
	Generations int
	Status      TaskStatus
	Result      *game.Grid
	CreatedAt   time.Time
}

// Node 节点结构
type Node struct {
	Address  string
	Capacity int
	Weight   int
}

// Broker 代理节点
type Broker struct {
	mu              sync.RWMutex
	workers         map[string]*WorkerNode
	taskQueue       *list.List
	taskStatusCache map[string]protocol.TaskStatusResponse_TaskStatus
	taskResultCache map[string]*protocol.ComputeResult
	cond            *sync.Cond
	lbStrategy      LoadBalancer
	healthCheck     *HealthChecker
	tasks           map[string]*Task
}

// WorkerNode 工作节点
type WorkerNode struct {
	Address    string
	LastActive time.Time
	Capacity   int
	Throughput float64
}

// NewBroker 创建代理节点
func NewBroker(strategy LoadBalancer, numWorkers int) *Broker {
	b := &Broker{
		workers:         make(map[string]*WorkerNode),
		taskQueue:       list.New(),
		taskStatusCache: make(map[string]protocol.TaskStatusResponse_TaskStatus),
		taskResultCache: make(map[string]*protocol.ComputeResult),
		lbStrategy:      strategy,
		tasks:           make(map[string]*Task),
	}
	b.cond = sync.NewCond(&b.mu)

	// 初始化健康检查器，每30秒检查一次
	b.healthCheck = NewHealthChecker(b, 30*time.Second)

	// 启动健康检查
	go b.healthCheck.Start()

	return b
}

// 添加工作节点
// 当工作节点启动时，它会调用 RegisterWorker 方法向 Broker 注册自己。
// Broker 会将其添加到 workers 映射中，并通知负载均衡器。
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
// 当工作节点失效或健康检查器检测到节点长时间未活动时，会调用此方法移除节点。
func (b *Broker) RemoveWorker(addr string) {
	b.mu.Lock()
	defer b.mu.Unlock()

	delete(b.workers, addr)
	b.lbStrategy.RemoveNode(addr)
}

// 分发任务到工作节点
// 当有新的任务时，Broker 会根据负载均衡策略选择一个工作节点，并将任务分发给该节点。
// 它会确保选中的节点仍然可用，并更新节点的最后活动时间。
func (b *Broker) DispatchTask(task *protocol.ComputeTask) (string, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	if len(b.workers) == 0 {
		log.Printf("没有可用的工作节点")
		return "", ErrNoAvailableWorkers
	}

	// 打印当前可用的工作节点
	log.Printf("可用工作节点: %v", b.workers)

	// 获取所有可用节点的地址列表
	availableNodes := make([]string, 0, len(b.workers))
	for addr := range b.workers {
		availableNodes = append(availableNodes, addr)
	}

	// 确保负载均衡器中的节点列表与可用节点同步
	b.lbStrategy.SyncNodes(availableNodes)

	target, err := b.lbStrategy.SelectNode(task)
	if err != nil {
		log.Printf("选择节点失败: %v", err)
		return "", fmt.Errorf("failed to select node: %v", err)
	}

	// 验证选中的节点是否仍然可用
	if _, exists := b.workers[target]; !exists {
		log.Printf("选中的节点 %s 已不可用", target)
		return "", fmt.Errorf("selected node %s is no longer available", target)
	}

	// 更新节点的最后活动时间
	b.workers[target].LastActive = time.Now()
	log.Printf("选择节点: %s", target)
	return target, nil
}

// 添加任务到队列并通知等待的消费者
// 当客户端提交任务时，Broker 会将任务添加到队列，并更新任务状态为 PENDING
func (b *Broker) PushTask(task *protocol.ComputeTask) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.taskQueue.PushBack(task)
	b.taskStatusCache[task.TaskId] = protocol.TaskStatusResponse_PENDING
	b.cond.Signal() // 通知等待的StreamResults
}

// 从队列头部获取任务（线程安全）
// 当有新的任务时，Broker 会从队列头部获取任务
// 如果队列为空，会等待新任务到达
func (b *Broker) PopTask() *protocol.ComputeTask {
	b.mu.Lock()
	defer b.mu.Unlock()

	for b.taskQueue.Len() == 0 {
		b.cond.Wait() // 等待新任务到达
	}

	front := b.taskQueue.Front()
	task := front.Value.(*protocol.ComputeTask)
	b.taskQueue.Remove(front)
	return task
}

// 更新任务状态
// 任务状态会随着处理过程更新：PENDING -> PROCESSING -> COMPLETED 或 FAILED。
func (b *Broker) UpdateTaskStatus(taskID string, status protocol.TaskStatusResponse_TaskStatus) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.taskStatusCache[taskID] = status
	log.Printf("任务 %s 状态更新为: %v", taskID, status)
}

// 获取客户端的所有任务
// 包括待处理、正在处理和已完成的任务。
func (b *Broker) GetClientTasks(clientId string) []*protocol.ComputeTask {
	b.mu.RLock()
	defer b.mu.RUnlock()

	taskMap := make(map[string]*protocol.ComputeTask)

	// 检查任务队列中的任务
	for e := b.taskQueue.Front(); e != nil; e = e.Next() {
		task := e.Value.(*protocol.ComputeTask)
		// 使用更精确的前缀匹配
		if strings.HasPrefix(task.TaskId, clientId+"-") {
			taskMap[task.TaskId] = task
		}
	}

	// 检查任务状态缓存中的任务
	for taskID := range b.taskStatusCache {
		// 使用更精确的前缀匹配
		if strings.HasPrefix(taskID, clientId+"-") {
			// 如果任务不在 map 中，说明它已经从队列中移除
			if _, exists := taskMap[taskID]; !exists {
				// 从结果缓存中获取任务信息
				if result, exists := b.taskResultCache[taskID]; exists {
					taskMap[taskID] = &protocol.ComputeTask{
						TaskId:   taskID,
						GridData: result.ResultData,
					}
				}
			}
		}
	}

	// 将 map 转换为 slice
	tasks := make([]*protocol.ComputeTask, 0, len(taskMap))
	for _, task := range taskMap {
		tasks = append(tasks, task)
	}

	return tasks
}

// 存储任务结果
// 当工作节点完成任务后，会将结果存储在 Broker 中，以便客户端获取
func (b *Broker) StoreTaskResult(taskID string, result *protocol.ComputeResult) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.taskResultCache[taskID] = result
}

// 获取任务结果
// 当客户端需要获取任务结果时，Broker 会从结果缓存中获取结果
func (b *Broker) GetTaskResult(taskID string) (*protocol.ComputeResult, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	// 检查任务状态
	status, exists := b.taskStatusCache[taskID]
	if !exists {
		return nil, errors.New("task not found")
	}

	if status != protocol.TaskStatusResponse_COMPLETED {
		return nil, errors.New("task not completed")
	}

	// 获取任务结果
	result, exists := b.taskResultCache[taskID]
	if !exists {
		return nil, errors.New("task result not found")
	}

	return result, nil
}

// RegisterNode 注册节点
func (b *Broker) RegisterNode(ctx context.Context, req *protocol.WorkerRegistration) (*protocol.RegistrationResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	node := &WorkerNode{
		Address:    req.Address,
		Capacity:   int(req.Capacity),
		LastActive: time.Now(),
	}

	b.workers[req.Address] = node
	b.lbStrategy.AddNode(req.Address)

	log.Printf("节点 %s 已注册，容量: %d", req.Address, req.Capacity)
	return &protocol.RegistrationResponse{
		Status: protocol.RegistrationResponse_SUCCESS,
	}, nil
}

// SubmitTask 提交任务
func (b *Broker) SubmitTask(ctx context.Context, task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
	// 记录开始时间
	startTime := time.Now()
	log.Printf("开始处理任务 %s，演化代数: %d", task.TaskId, task.Generations)

	// 检查输入数据
	if task.GridData == nil {
		log.Printf("任务 %s 的网格数据为空", task.TaskId)
		return nil, errors.New("网格数据不能为空")
	}

	// 反序列化网格数据
	grid := deserializeGrid(task.GridData)

	// 创建任务
	t := &Task{
		ID:          task.TaskId,
		Grid:        grid,
		Generations: int(task.Generations),
		Status:      TaskStatusProcessing,
		CreatedAt:   time.Now(),
	}

	// 存储任务
	b.mu.Lock()
	b.tasks[task.TaskId] = t
	b.mu.Unlock()

	// 执行任务
	result, err := b.executeTask(t)
	if err != nil {
		t.Status = TaskStatusFailed
		return nil, err
	}

	t.Status = TaskStatusCompleted
	t.Result = result

	// 计算执行时间
	elapsed := time.Since(startTime)
	log.Printf("任务 %s 处理完成，耗时: %v", task.TaskId, elapsed)

	// 返回结果
	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: serializeGrid(result),
		ExecTime:   elapsed.Seconds(),
	}, nil
}

// executeTask 执行任务
func (b *Broker) executeTask(task *Task) (*game.Grid, error) {
	// 分发任务到工作节点
	target, err := b.DispatchTask(&protocol.ComputeTask{
		TaskId:      task.ID,
		GridData:    serializeGrid(task.Grid),
		Generations: int32(task.Generations),
	})
	if err != nil {
		return nil, err
	}

	// 获取工作节点
	b.mu.RLock()
	worker, exists := b.workers[target]
	b.mu.RUnlock()
	if !exists {
		return nil, errors.New("worker not found")
	}

	// 更新工作节点状态
	worker.LastActive = time.Now()

	return task.Grid, nil
}

// GetTaskStatus 获取任务状态
func (b *Broker) GetTaskStatus(ctx context.Context, req *protocol.TaskStatusRequest) (*protocol.TaskStatusResponse, error) {
	b.mu.RLock()
	defer b.mu.RUnlock()

	task, exists := b.tasks[req.TaskId]
	if !exists {
		return nil, errors.New("任务不存在")
	}

	var status protocol.TaskStatusResponse_TaskStatus
	switch task.Status {
	case TaskStatusPending:
		status = protocol.TaskStatusResponse_PENDING
	case TaskStatusProcessing:
		status = protocol.TaskStatusResponse_PROCESSING
	case TaskStatusCompleted:
		status = protocol.TaskStatusResponse_COMPLETED
	case TaskStatusFailed:
		status = protocol.TaskStatusResponse_FAILED
	}

	return &protocol.TaskStatusResponse{
		Status: status,
	}, nil
}

// SyncNodes 同步节点列表
func (b *Broker) SyncNodes(ctx context.Context, req *protocol.SyncNodesRequest) (*protocol.SyncNodesResponse, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 移除不可用的节点
	for addr := range b.workers {
		found := false
		for _, node := range req.Nodes {
			if node.Address == addr {
				found = true
				break
			}
		}
		if !found {
			delete(b.workers, addr)
			b.lbStrategy.RemoveNode(addr)
		}
	}

	// 添加新节点
	for _, node := range req.Nodes {
		if _, exists := b.workers[node.Address]; !exists {
			b.workers[node.Address] = &WorkerNode{
				Address:    node.Address,
				Capacity:   int(node.Capacity),
				LastActive: time.Now(),
			}
			b.lbStrategy.AddNode(node.Address)
		}
	}

	return &protocol.SyncNodesResponse{Success: true}, nil
}

// 序列化网格为二进制格式
func serializeGrid(grid *game.Grid) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	type ExportGrid struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	err := enc.Encode(ExportGrid{
		Width:  grid.Width,
		Height: grid.Height,
		Cells:  grid.Cells,
	})

	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

// 反序列化二进制数据为网格
func deserializeGrid(data []byte) *game.Grid {
	var export struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&export); err != nil {
		return game.NewGrid(0, 0, game.Periodic)
	}

	return &game.Grid{
		Width:    export.Width,
		Height:   export.Height,
		Cells:    export.Cells,
		Boundary: game.Periodic,
	}
}
