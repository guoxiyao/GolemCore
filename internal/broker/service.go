package broker

import (
	"context"
	"log"
	"time"

	"GolemCore/internal/rpc/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type BrokerService struct {
	protocol.UnimplementedConwayServiceServer // 需确保proto文件已生成该类型
	broker                                    *Broker
}

func NewBrokerService(b *Broker) *BrokerService {
	return &BrokerService{broker: b}
}

func (s *BrokerService) SubmitTask(
	ctx context.Context,
	task *protocol.ComputeTask,
) (*protocol.TaskResponse, error) {
	s.broker.mu.Lock()
	s.broker.taskQueue.PushBack(task)
	s.broker.mu.Unlock()

	go s.processTask(task)

	return &protocol.TaskResponse{
		TaskId:    task.TaskId,
		Timestamp: time.Now().Unix(),
	}, nil
}

// 修复后的RegisterWorker方法
func (s *BrokerService) RegisterWorker(
	ctx context.Context,
	req *protocol.WorkerRegistration, // 添加请求参数
) (*protocol.RegistrationResponse, error) {
	s.broker.AddWorker(req.Address) // 使用请求中的地址

	return &protocol.RegistrationResponse{
		Status:            protocol.RegistrationResponse_SUCCESS,
		AssignedId:        req.Address, // 匹配proto定义字段名
		HeartbeatInterval: 30,
	}, nil
}

// 完善后的任务处理逻辑
func (s *BrokerService) processTask(task *protocol.ComputeTask) {
	target, err := s.broker.DispatchTask(task)
	if err != nil {
		log.Printf("Task dispatch failed: %v", err)
		return
	}

	conn, err := grpc.Dial(
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()), // 替代弃用的WithInsecure
	)
	if err != nil {
		log.Printf("Connection failed: %v", err)
		return
	}
	defer conn.Close()

	client := protocol.NewConwayServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	res, err := client.ComputeSync(ctx, task)
	if err != nil {
		log.Printf("RPC call failed: %v", err)
		return
	}

	log.Printf("耗时: %.2f 秒", res.ExecTime)
}

// 添加ConwayService要求的ComputeSync方法实现
func (s *BrokerService) ComputeSync(
	ctx context.Context,
	task *protocol.ComputeTask,
) (*protocol.ComputeResult, error) {
	// 这里实现你的同步计算逻辑
	// 示例实现（需要根据实际逻辑修改）：
	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: []byte{}, // 填充实际结果
		ExecTime:   0.0,      // 填充执行时间
	}, nil
}
