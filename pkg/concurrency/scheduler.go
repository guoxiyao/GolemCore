package concurrency

import (
	"container/heap"
	"context"
	"sync"
	"time"
)

// 任务优先级类型
type Priority int

const (
	Low Priority = iota
	Medium
	High
	Critical
)

// 调度任务结构
type Task struct {
	ID       string
	Priority Priority
	Timeout  time.Duration
	Execute  func(context.Context) error
}

// 添加指标结构定义
type SchedulerMetrics struct {
	Queued int // 等待任务数
	Active int // 活跃任务数
}

// 调度器实现
type Scheduler struct {
	workers    int
	queue      *priorityQueue
	mu         sync.Mutex
	taskChan   chan *Task
	workerSem  *Semaphore
	cancelFunc context.CancelFunc
	wg         sync.WaitGroup
}

// 创建调度器实例
func NewScheduler(workers int) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Scheduler{
		workers:    workers,
		queue:      &priorityQueue{},
		taskChan:   make(chan *Task, 1000),
		workerSem:  NewSemaphore(workers),
		cancelFunc: cancel,
	}
	heap.Init(s.queue)
	go s.dispatch(ctx)
	return s
}

// 提交任务到调度器
func (s *Scheduler) Submit(task *Task) error {
	s.mu.Lock()
	heap.Push(s.queue, task)
	s.mu.Unlock()
	return nil
}

// 核心调度逻辑
func (s *Scheduler) dispatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			s.mu.Lock()
			if s.queue.Len() > 0 {
				task := heap.Pop(s.queue).(*Task)
				s.mu.Unlock()

				s.wg.Add(1)
				go s.executeTask(task)
			} else {
				s.mu.Unlock()
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
}

// 执行任务逻辑
func (s *Scheduler) executeTask(task *Task) {
	defer s.wg.Done()

	ctx, cancel := context.WithTimeout(context.Background(), task.Timeout)
	defer cancel()

	s.workerSem.Acquire(1)
	defer s.workerSem.Release(1)

	select {
	case <-ctx.Done():
		// 处理超时逻辑
		return
	default:
		if err := task.Execute(ctx); err != nil {
			// 错误处理逻辑
		}
	}
}

// 优雅关闭调度器
func (s *Scheduler) Shutdown() {
	s.cancelFunc()
	s.wg.Wait()
}

// 优先队列实现
type priorityQueue []*Task

func (pq priorityQueue) Len() int { return len(pq) }

func (pq priorityQueue) Less(i, j int) bool {
	return pq[i].Priority > pq[j].Priority // 数值越大优先级越高
}

func (pq priorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
}

func (pq *priorityQueue) Push(x interface{}) {
	*pq = append(*pq, x.(*Task))
}

func (pq *priorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	*pq = old[0 : n-1]
	return item
}

func (s *Scheduler) Metrics() SchedulerMetrics {
	s.mu.Lock()
	defer s.mu.Unlock()

	return SchedulerMetrics{
		Queued: s.queue.Len(),
		Active: s.workers - s.workerSem.Available(),
	}
}
