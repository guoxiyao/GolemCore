// E:\go\GolemCore\internal\rpc\async_rpc.go
package rpc

import (
	"context"
	"sync"
	"time"

	"GolemCore/internal/rpc/protocol"
)

type AsyncDispatcher struct {
	pool       *ClientPool
	taskChan   chan *protocol.ComputeTask
	resultChan chan *protocol.ComputeResult
	wg         sync.WaitGroup
}

func NewAsyncDispatcher(pool *ClientPool, workers int) *AsyncDispatcher {
	d := &AsyncDispatcher{
		pool:       pool,
		taskChan:   make(chan *protocol.ComputeTask, 1000),
		resultChan: make(chan *protocol.ComputeResult, 1000),
	}

	// 启动工作协程
	for i := 0; i < workers; i++ {
		go d.worker()
	}
	return d
}

func (d *AsyncDispatcher) Submit(task *protocol.ComputeTask) <-chan *protocol.ComputeResult {
	resultChan := make(chan *protocol.ComputeResult, 1)
	d.wg.Add(1)

	go func() {
		defer d.wg.Done()
		d.taskChan <- task
		// 结果处理逻辑
	}()

	return resultChan
}

func (d *AsyncDispatcher) worker() {
	for task := range d.taskChan {
		client, err := d.pool.Get()
		if err != nil {
			// 错误处理逻辑
			continue
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		res, err := client.ComputeSync(ctx, task)
		cancel()

		if err != nil {
			// 重试或错误处理
		} else {
			d.resultChan <- res
		}

		d.pool.Put(client)
	}
}

func (d *AsyncDispatcher) Shutdown() {
	close(d.taskChan)
	d.wg.Wait()
	close(d.resultChan)
}
