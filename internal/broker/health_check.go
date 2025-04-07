package broker

import (
	"log"
	"time"
)

type HealthChecker struct {
	broker   *Broker
	interval time.Duration
}

func NewHealthChecker(b *Broker, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		broker:   b,
		interval: interval,
	}
}

func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	for range ticker.C {
		hc.checkWorkers()
	}
}

func (hc *HealthChecker) checkWorkers() {
	hc.broker.mu.RLock()
	defer hc.broker.mu.RUnlock()

	// 记录需要移除的节点
	var nodesToRemove []string

	for addr, worker := range hc.broker.workers {
		// 如果节点超过2个检查间隔没有活动，标记为移除
		if time.Since(worker.LastActive) > 2*hc.interval {
			log.Printf("节点 %s 超过 %v 没有活动，将被移除", addr, 2*hc.interval)
			nodesToRemove = append(nodesToRemove, addr)
		}
	}

	// 释放读锁，以便可以获取写锁
	hc.broker.mu.RUnlock()

	// 移除失效节点
	for _, addr := range nodesToRemove {
		hc.broker.RemoveWorker(addr)
		log.Printf("已移除失效节点: %s", addr)
	}

	// 重新获取读锁
	hc.broker.mu.RLock()
}
