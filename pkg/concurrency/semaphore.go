package concurrency

import (
	"context"
	"sync"
	"time"
)

// 带权重的信号量实现
type Semaphore struct {
	capacity int
	used     int
	waiters  []chan struct{}
	mutex    sync.Mutex
}

// 创建信号量实例
func NewSemaphore(capacity int) *Semaphore {
	return &Semaphore{
		capacity: capacity,
		waiters:  make([]chan struct{}, 0),
	}
}

// 获取信号量资源（带超时）
func (s *Semaphore) Acquire(weight int) error {
	return s.AcquireWithTimeout(weight, 0) // 0表示无限等待
}

// 带超时的资源获取
func (s *Semaphore) AcquireWithTimeout(weight int, timeout time.Duration) error {
	s.mutex.Lock()
	if s.capacity-s.used >= weight && len(s.waiters) == 0 {
		s.used += weight
		s.mutex.Unlock()
		return nil
	}

	ch := make(chan struct{})
	s.waiters = append(s.waiters, ch)
	s.mutex.Unlock()

	if timeout == 0 {
		<-ch
		return nil
	}

	select {
	case <-ch:
		return nil
	case <-time.After(timeout):
		s.mutex.Lock()
		defer s.mutex.Unlock()
		// 从等待队列中移除
		for i, c := range s.waiters {
			if c == ch {
				s.waiters = append(s.waiters[:i], s.waiters[i+1:]...)
				break
			}
		}
		return context.DeadlineExceeded
	}
}

// 释放信号量资源
func (s *Semaphore) Release(weight int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.used -= weight
	if s.used < 0 {
		s.used = 0
	}

	// 唤醒等待的goroutine
	for len(s.waiters) > 0 {
		ch := s.waiters[0]
		s.waiters = s.waiters[1:]

		select {
		case ch <- struct{}{}:
			return
		default:
			// 通道已关闭，继续处理下一个
		}
	}
}

// 动态调整容量
func (s *Semaphore) Resize(newCapacity int) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if newCapacity < s.used {
		s.used = newCapacity
	}
	s.capacity = newCapacity

	// 唤醒可能因容量变化而可执行的等待者
	for len(s.waiters) > 0 {
		if s.capacity-s.used <= 0 {
			break
		}
		ch := s.waiters[0]
		s.waiters = s.waiters[1:]
		close(ch)
	}
}

// 添加Available方法实现
func (s *Semaphore) Available() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return s.capacity - s.used
}
