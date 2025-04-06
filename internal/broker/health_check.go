// health_check.go (补充文件)
package broker

import "time"

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

	for addr, worker := range hc.broker.workers {
		if time.Since(worker.LastActive) > 2*hc.interval {
			go hc.broker.RemoveWorker(addr)
		}
	}
}
