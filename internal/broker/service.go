package broker

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc/codes"

	"GolemCore/internal/rpc/protocol"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpcstatus "google.golang.org/grpc/status"
)

// broker的gRPC服务实现
type BrokerService struct {
	protocol.UnimplementedBrokerServiceServer // 需确保proto文件已生成该类型
	broker                                    *Broker
	mu                                        sync.RWMutex
}

func NewBrokerService(b *Broker) *BrokerService {
	return &BrokerService{broker: b}
}

// 提交任务
func (s *BrokerService) SubmitTask(
	ctx context.Context,
	task *protocol.ComputeTask,
) (*protocol.TaskResponse, error) {
	// 先添加任务
	s.broker.mu.Lock()
	s.broker.taskQueue.PushBack(task)
	s.broker.taskStatusCache[task.TaskId] = protocol.TaskStatusResponse_PENDING
	s.broker.cond.Signal() // 通知等待的StreamResults
	s.broker.mu.Unlock()

	// 启动任务处理
	go s.processTask(task)

	return &protocol.TaskResponse{
		TaskId:    task.TaskId,
		Timestamp: time.Now().Unix(),
	}, nil
}

// 注册工作节点
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

// 任务处理逻辑
func (s *BrokerService) processTask(task *protocol.ComputeTask) {
	log.Printf("开始处理任务: %s", task.TaskId)
	s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_PROCESSING)

	target, err := s.broker.DispatchTask(task)
	if err != nil {
		log.Printf("任务分发失败: %s, 错误: %v", task.TaskId, err)
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return
	}
	log.Printf("任务 %s 分发到节点: %s", task.TaskId, target)

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Printf("尝试连接到工作节点: %s", target)
	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Printf("连接工作节点失败: %s, 错误: %v", target, err)
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return
	}
	defer conn.Close()
	log.Printf("成功连接到工作节点: %s", target)

	client := protocol.NewConwayServiceClient(conn)
	// 增加计算超时时间到 5 分钟
	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	log.Printf("开始执行计算: %s, 网格数据长度: %d", task.TaskId, len(task.GridData))
	res, err := client.ComputeSync(ctx, task)
	if err != nil {
		log.Printf("计算执行失败: %s, 错误: %v", task.TaskId, err)
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return
	}

	// 存储任务结果
	s.broker.StoreTaskResult(task.TaskId, res)

	// 更新任务状态为 COMPLETED
	s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_COMPLETED)
	log.Printf("任务完成: %s, 耗时: %.2f 秒", task.TaskId, res.ExecTime)
}

// 实现 StreamResults 流式结果推送
func (s *BrokerService) StreamResults(
	req *protocol.StreamRequest,
	stream protocol.BrokerService_StreamResultsServer,
) error {
	log.Printf("开始处理客户端 %s 的结果流", req.ClientId)

	for {
		select {
		case <-stream.Context().Done():
			log.Printf("客户端 %s 断开连接", req.ClientId)
			return nil
		default:
			// 获取该客户端的所有任务
			tasks := s.broker.GetClientTasks(req.ClientId)
			if len(tasks) == 0 {
				log.Printf("客户端 %s 没有任务", req.ClientId)
				return grpcstatus.Error(codes.NotFound, "no tasks found for client")
			}

			// 检查每个任务的状态
			for _, task := range tasks {
				status, err := s.GetTaskStatus(stream.Context(), &protocol.TaskStatusRequest{TaskId: task.TaskId})
				if err != nil {
					log.Printf("获取任务 %s 状态失败: %v", task.TaskId, err)
					continue
				}

				// 如果任务已完成，返回结果
				if status.Status == protocol.TaskStatusResponse_COMPLETED {
					result, err := s.broker.GetTaskResult(task.TaskId)
					if err != nil {
						log.Printf("获取任务 %s 结果失败: %v", task.TaskId, err)
						continue
					}

					if err := stream.Send(result); err != nil {
						log.Printf("发送任务 %s 结果失败: %v", task.TaskId, err)
						return err
					}

					log.Printf("成功发送任务 %s 的结果", task.TaskId)
				} else if status.Status == protocol.TaskStatusResponse_FAILED {
					log.Printf("任务 %s 执行失败", task.TaskId)
					// 不要立即返回错误，继续处理其他任务
					continue
				}
			}

			// 等待一段时间后重试
			time.Sleep(1 * time.Second)
		}
	}
}

// 获取任务状态
func (s *BrokerService) GetTaskStatus(
	ctx context.Context,
	req *protocol.TaskStatusRequest,
) (*protocol.TaskStatusResponse, error) {
	s.broker.mu.RLock()
	defer s.broker.mu.RUnlock()

	status, exists := s.broker.taskStatusCache[req.TaskId]
	if !exists {
		return nil, grpcstatus.Error(codes.NotFound, "任务不存在")
	}
	return &protocol.TaskStatusResponse{Status: status}, nil
}

// BrokerService 处理任务逻辑
func (s *BrokerService) processTaskAndGetResult(task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
	// 正确处理 error 类型
	target, err := s.broker.lbStrategy.SelectNode(task)
	if err != nil {
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("选择节点失败: %v", err))
	}

	// 设置连接超时
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		target,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return nil, grpcstatus.Error(codes.Unavailable, fmt.Sprintf("连接Worker失败: %v", err))
	}
	defer conn.Close()

	client := protocol.NewConwayServiceClient(conn)
	ctx, cancel = context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.ComputeSync(ctx, task)
	if err != nil {
		s.broker.UpdateTaskStatus(task.TaskId, protocol.TaskStatusResponse_FAILED)
		return nil, grpcstatus.Error(codes.Internal, fmt.Sprintf("计算失败: %v", err))
	}

	return result, nil
}

// 修复 RequeueTask 方法
func (b *Broker) RequeueTask(task *protocol.ComputeTask) {
	if task == nil {
		log.Println("警告：忽略空任务重入队列")
		return
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	b.taskQueue.PushFront(task)
}

// UpdateWorkerActivity 更新工作节点的最后活动时间
func (s *BrokerService) UpdateWorkerActivity(
	ctx context.Context,
	req *protocol.WorkerRegistration,
) (*protocol.RegistrationResponse, error) {
	s.broker.mu.Lock()
	defer s.broker.mu.Unlock()

	addr := req.Address
	if worker, exists := s.broker.workers[addr]; exists {
		worker.LastActive = time.Now()
		log.Printf("更新节点 %s 的活动时间", addr)
		return &protocol.RegistrationResponse{
			Status: protocol.RegistrationResponse_SUCCESS,
		}, nil
	}

	return &protocol.RegistrationResponse{
		Status: protocol.RegistrationResponse_FAILED,
	}, nil
}
