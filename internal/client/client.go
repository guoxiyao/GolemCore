package client

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"

	"GolemCore/internal/rpc/protocol"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client 核心客户端结构 负责提交任务和获取结果
type Client struct {
	config       Config//客户端配置
	conn         *grpc.ClientConn//gRPC连接
	broker       protocol.BrokerServiceClient//broker客户端
	resultChan   chan *protocol.ComputeResult//结果通道
	shutdown     chan struct{}//关闭通道
	wg           sync.WaitGroup//等待组
	mu           sync.RWMutex//读写锁
	pendingTasks map[string]context.CancelFunc // 任务ID -> 取消函数
	//待处理任务映射
}

// Config 客户端配置
type Config struct {
	BrokerAddress    string        `yaml:"broker_address"`
	MaxRetries       int           `yaml:"max_retries"`
	RequestTimeout   time.Duration `yaml:"request_timeout"`
	HeartbeatTimeout time.Duration `yaml:"heartbeat_timeout"`
	ClientID         string        `yaml:"client_id"`
}

// NewClient 创建新的客户端实例
func NewClient(cfg Config) (*Client, error) {
	if cfg.BrokerAddress == "" {
		return nil, errors.New("broker address required")
	}

	// 设置默认值
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 15 * time.Second
	}
	if cfg.ClientID == "" {
		cfg.ClientID = "default-client"
	}

	return &Client{
		config:       cfg,
		resultChan:   make(chan *protocol.ComputeResult, 100),
		shutdown:     make(chan struct{}),
		pendingTasks: make(map[string]context.CancelFunc),
	}, nil
}

// Connect 建立到Broker的gRPC连接
func (c *Client) Connect() error {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.RequestTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, c.config.BrokerAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("connection failed: %w", err)
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.conn = conn
	c.broker = protocol.NewBrokerServiceClient(conn)
	return nil
}

// SubmitTask 提交计算任务（异步）
func (c *Client) SubmitTask(task *protocol.ComputeTask) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.RequestTimeout)
	defer cancel()

	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		resp, err := c.broker.SubmitTask(ctx, task)
		if err == nil {
			c.trackTask(task.TaskId, cancel)
			return resp.TaskId, nil
		}

		lastErr = err
		log.Printf("SubmitTask attempt %d failed: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
	}

	return "", fmt.Errorf("submit failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// StreamResults 启动结果流监听
func (c *Client) StreamResults() {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()

		stream, err := c.broker.StreamResults(context.Background(), &protocol.StreamRequest{
			ClientId: c.config.ClientID,
		})
		if err != nil {
			log.Printf("stream setup failed: %v", err)
			return
		}

		for {
			select {
			case <-c.shutdown:
				return
			default:
				res, err := stream.Recv()
				if err != nil {
					log.Printf("stream error: %v", err)
					return
				}
				c.resultChan <- res
			}
		}
	}()
}

// GetResult 获取计算结果（阻塞模式）
func (c *Client) GetResult(taskID string) (*protocol.ComputeResult, error) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			return nil, errors.New("result wait timeout")
		case res := <-c.resultChan:
			if res.TaskId == taskID {
				return res, nil
			}
		case <-ticker.C:
			// 主动查询任务状态
			status, err := c.GetTaskStatus(taskID)
			if err != nil {
				return nil, fmt.Errorf("task failed: %w", err)
			}
			switch status {
			case protocol.TaskStatusResponse_FAILED:
				return nil, errors.New("task failed")
			case protocol.TaskStatusResponse_COMPLETED:
				// 如果任务已完成但结果还未收到，继续等待
				continue
			case protocol.TaskStatusResponse_PENDING, protocol.TaskStatusResponse_PROCESSING:
				// 任务仍在处理中，继续等待
				continue
			case protocol.TaskStatusResponse_UNKNOWN:
				return nil, errors.New("task status unknown")
			}
		}
	}
}

// GetTaskStatus 查询任务状态
func (c *Client) GetTaskStatus(taskID string) (protocol.TaskStatusResponse_TaskStatus, error) {
	var lastErr error
	for i := 0; i < c.config.MaxRetries; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		resp, err := c.broker.GetTaskStatus(ctx, &protocol.TaskStatusRequest{
			TaskId: taskID,
		})
		if err == nil {
			return resp.Status, nil
		}

		lastErr = err
		log.Printf("GetTaskStatus attempt %d failed: %v", i+1, err)
		time.Sleep(time.Duration(i+1) * time.Second) // 指数退避
	}

	return protocol.TaskStatusResponse_UNKNOWN, fmt.Errorf("get status failed after %d retries: %w", c.config.MaxRetries, lastErr)
}

// Shutdown 优雅关闭客户端
func (c *Client) Shutdown() {
	close(c.shutdown)
	c.wg.Wait()

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.conn != nil {
		c.conn.Close()
	}
}

// 任务生命周期管理
func (c *Client) trackTask(taskID string, cancel context.CancelFunc) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.pendingTasks[taskID] = cancel
}

// CancelTask 取消进行中的任务
func (c *Client) CancelTask(taskID string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if cancel, exists := c.pendingTasks[taskID]; exists {
		cancel()
		delete(c.pendingTasks, taskID)
		return nil
	}
	return errors.New("task not found")
}

// 可视化集成示例
func (c *Client) VisualizeResult(res *protocol.ComputeResult) error {
	// 调用可视化模块接口
	// viz.RenderGrid(res.ResultData)
	return nil
}
