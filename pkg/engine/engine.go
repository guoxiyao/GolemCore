package engine

import (
	"context"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"sync"
	"time"

	"GolemCore/internal/rpc/protocol"
	"GolemCore/internal/storage"
	"GolemCore/pkg/concurrency"
)

type TaskMeta struct {
	ID        string
	Submitted time.Time
}

// 添加指标结构
type SchedulerMetrics struct {
	Queued int
	Active int
}

// 引擎运行状态
type EngineStatus int

const (
	StatusBooting EngineStatus = iota
	StatusRunning
	StatusDraining
	StatusShutdown
)

// 工作节点状态
type WorkerState struct {
	Address    string
	LastActive time.Time
	Capacity   int32
	Running    int32
}

// 引擎核心结构
type Engine struct {
	mu          sync.RWMutex
	config      Config
	status      EngineStatus
	taskQueue   *concurrency.Scheduler
	workerPool  map[string]*WorkerState
	storage     *storage.SnapshotManager
	logger      storage.Logger
	heartbeatTk *time.Ticker
	eventChan   chan Event
	semaphore   *concurrency.Semaphore
}

// 引擎配置
type Config struct {
	MaxConcurrentTasks int           // 最大并发任务数
	HeartbeatInterval  time.Duration // 节点心跳间隔
	TaskTimeout        time.Duration // 任务超时时间
	SnapshotInterval   time.Duration // 快照间隔
}

// 系统事件类型
type Event struct {
	Type    EventType
	Payload interface{}
}

type EventType int

const (
	EventTaskReceived EventType = iota
	EventTaskCompleted
	EventWorkerOnline
	EventWorkerOffline
	EventSystemError
)

// 创建新引擎实例
func NewEngine(cfg Config, store *storage.SnapshotManager, logger storage.Logger) *Engine {
	e := &Engine{
		config:     cfg,
		storage:    store,
		logger:     logger,
		taskQueue:  concurrency.NewScheduler(cfg.MaxConcurrentTasks),
		workerPool: make(map[string]*WorkerState),
		eventChan:  make(chan Event, 100),
		semaphore:  concurrency.NewSemaphore(cfg.MaxConcurrentTasks),
	}

	// 恢复系统状态
	if snap, err := store.LoadLatest(); err == nil {
		e.restoreFromSnapshot(snap)
	}

	// 启动后台协程
	go e.monitorSystem()
	go e.processEvents()
	go e.autoSnapshot()

	return e
}

// 提交新任务
func (e *Engine) SubmitTask(task *protocol.ComputeTask) (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.status != StatusRunning {
		return "", errors.New("引擎未就绪")
	}

	// 生成唯一任务ID
	taskID := generateTaskID()

	// 创建调度任务
	schedTask := &concurrency.Task{
		ID:       taskID,
		Priority: concurrency.Medium,
		Timeout:  e.config.TaskTimeout,
		Execute:  e.createTaskExecutor(task),
	}

	// 提交到调度器
	if err := e.taskQueue.Submit(schedTask); err != nil {
		return "", fmt.Errorf("任务提交失败: %w", err)
	}

	// 发送系统事件
	e.eventChan <- Event{
		Type: EventTaskReceived,
		Payload: TaskMeta{
			ID:        taskID,
			Submitted: time.Now(),
		},
	}

	return taskID, nil
}

// 注册工作节点
func (e *Engine) RegisterWorker(addr string, capacity int32) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if _, exists := e.workerPool[addr]; exists {
		return errors.New("节点已注册")
	}

	e.workerPool[addr] = &WorkerState{
		Address:    addr,
		LastActive: time.Now(),
		Capacity:   capacity,
		Running:    0,
	}

	e.eventChan <- Event{
		Type:    EventWorkerOnline,
		Payload: addr,
	}

	return nil
}

// 核心任务执行逻辑
func (e *Engine) createTaskExecutor(task *protocol.ComputeTask) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		worker, err := e.selectWorker()
		if err != nil {
			return fmt.Errorf("选择工作节点失败: %w", err)
		}

		// 获取资源许可
		e.semaphore.Acquire(1)
		defer e.semaphore.Release(1)

		// 执行RPC调用
		conn, err := grpc.Dial(worker.Address, grpc.WithInsecure())
		if err != nil {
			return fmt.Errorf("连接节点失败: %w", err)
		}
		defer conn.Close()

		client := protocol.NewConwayServiceClient(conn)
		result, err := client.ComputeSync(ctx, task)

		// 更新节点状态
		e.updateWorkerStatus(worker.Address, func(ws *WorkerState) {
			ws.Running--
		})

		if err != nil {
			return fmt.Errorf("任务执行失败: %w", err)
		}

		// 保存计算结果
		e.storeResult(task.TaskId, result)

		return nil
	}
}

// 选择最优工作节点
func (e *Engine) selectWorker() (*WorkerState, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	var best *WorkerState
	for _, ws := range e.workerPool {
		if ws.Running < ws.Capacity {
			if best == nil || (ws.Capacity-ws.Running) > (best.Capacity-best.Running) {
				best = ws
			}
		}
	}

	if best == nil {
		return nil, errors.New("无可用工作节点")
	}

	// 更新节点状态
	e.updateWorkerStatus(best.Address, func(ws *WorkerState) {
		ws.Running++
	})

	return best, nil
}

// 后台监控系统
func (e *Engine) monitorSystem() {
	e.heartbeatTk = time.NewTicker(e.config.HeartbeatInterval)
	defer e.heartbeatTk.Stop()

	for range e.heartbeatTk.C {
		e.checkWorkerHealth()
		e.logSystemStatus()
	}
}

// 检查节点健康状态
func (e *Engine) checkWorkerHealth() {
	e.mu.Lock()
	defer e.mu.Unlock()

	cutoff := time.Now().Add(-3 * e.config.HeartbeatInterval)

	for addr, ws := range e.workerPool {
		if ws.LastActive.Before(cutoff) {
			delete(e.workerPool, addr)
			e.eventChan <- Event{
				Type:    EventWorkerOffline,
				Payload: addr,
			}
		}
	}
}

// 处理系统事件
func (e *Engine) processEvents() {
	for event := range e.eventChan {
		switch event.Type {
		case EventTaskReceived:
			meta := event.Payload.(TaskMeta)
			e.logger.Info("engine", "收到新任务", storage.Fields{
				"task_id": meta.ID,
			})

		case EventTaskCompleted:
			// 处理任务完成逻辑

		case EventWorkerOnline:
			addr := event.Payload.(string)
			e.logger.Info("engine", "节点上线", storage.Fields{
				"address": addr,
			})

		case EventSystemError:
			err := event.Payload.(error)
			e.logger.Error("engine", "系统错误", storage.Fields{
				"error": err.Error(),
			})
		}
	}
}

// 自动保存快照
func (e *Engine) autoSnapshot() {
	tk := time.NewTicker(e.config.SnapshotInterval)
	defer tk.Stop()

	for range tk.C {
		snap := e.createSnapshot()
		if err := e.storage.Save(snap); err != nil {
			e.eventChan <- Event{
				Type:    EventSystemError,
				Payload: err,
			}
		}
	}
}

// 创建系统快照
func (e *Engine) createSnapshot() *storage.SystemSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return &storage.SystemSnapshot{
		Timestamp:  time.Now(),
		TaskQueue:  e.getPendingTaskIDs(),
		WorkerList: e.getWorkerAddresses(),
	}
}

// 从快照恢复
func (e *Engine) restoreFromSnapshot(snap *storage.SystemSnapshot) {
	// 实现恢复逻辑
}

// 关闭引擎
func (e *Engine) Shutdown() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.status = StatusDraining
	e.taskQueue.Shutdown()

	// 等待所有任务完成
	e.semaphore.Acquire(e.config.MaxConcurrentTasks)
	defer e.semaphore.Release(e.config.MaxConcurrentTasks)

	e.status = StatusShutdown
}

// 辅助方法
func generateTaskID() string {
	return fmt.Sprintf("task-%d", time.Now().UnixNano())
}

func (e *Engine) updateWorkerStatus(addr string, update func(*WorkerState)) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if ws, exists := e.workerPool[addr]; exists {
		update(ws)
		ws.LastActive = time.Now()
	}
}

func (e *Engine) logSystemStatus() {
	e.mu.RLock()
	defer e.mu.RUnlock()

	stats := storage.Fields{
		"workers_online": len(e.workerPool),
		"tasks_queued":   e.taskQueue.Metrics().Queued,
		"tasks_active":   e.taskQueue.Metrics().Active,
	}

	e.logger.Info("engine", "系统状态", stats)
}

// 存储任务结果
func (e *Engine) storeResult(taskID string, result *protocol.ComputeResult) {
	// 实际存储逻辑需要根据项目需求实现
	// 示例：保存到数据库或文件系统
}

// 获取待处理任务ID列表
func (e *Engine) getPendingTaskIDs() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	// 示例实现（需根据实际调度器接口调整）
	return make([]string, 0)
}

// 获取工作节点地址列表
func (e *Engine) getWorkerAddresses() []string {
	e.mu.RLock()
	defer e.mu.RUnlock()

	addresses := make([]string, 0, len(e.workerPool))
	for addr := range e.workerPool {
		addresses = append(addresses, addr)
	}
	return addresses
}
