package server

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"GolemCore/internal/rpc/protocol"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

type Server struct {
	protocol.UnimplementedBrokerServiceServer
	protocol.UnimplementedConwayServiceServer

	mu          sync.RWMutex
	tasks       map[string]*protocol.ComputeTask
	workers     map[string]*protocol.WorkerRegistration
	taskQueue   chan *protocol.ComputeTask
	statusCache map[string]protocol.TaskStatusResponse_TaskStatus
}

func NewServer(workerCount int) *Server {
	return &Server{
		tasks:       make(map[string]*protocol.ComputeTask),
		workers:     make(map[string]*protocol.WorkerRegistration),
		taskQueue:   make(chan *protocol.ComputeTask, 1000),
		statusCache: make(map[string]protocol.TaskStatusResponse_TaskStatus),
	}
}

// 实现BrokerService接口
func (s *Server) SubmitTask(ctx context.Context, task *protocol.ComputeTask) (*protocol.TaskResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if task.TaskId == "" {
		return nil, grpcstatus.Error(codes.InvalidArgument, "task_id required")
	}

	s.tasks[task.TaskId] = task
	s.statusCache[task.TaskId] = protocol.TaskStatusResponse_PENDING
	s.taskQueue <- task

	return &protocol.TaskResponse{
		TaskId:    task.TaskId,
		Timestamp: time.Now().Unix(),
	}, nil
}

// 实现流式结果推送
func (s *Server) StreamResults(req *protocol.StreamRequest, stream protocol.BrokerService_StreamResultsServer) error {
	for {
		select {
		case <-stream.Context().Done():
			return nil
		case task := <-s.taskQueue:
			// 分发任务到worker的逻辑
			result := &protocol.ComputeResult{
				TaskId:     task.TaskId,
				ResultData: processTask(task), // 实际计算逻辑
				ExecTime:   time.Since(time.Unix(task.Timestamp, 0)).Seconds(),
			}
			if err := stream.Send(result); err != nil {
				log.Printf("Stream send error: %v", err)
				return err
			}
			s.updateTaskStatus(task.TaskId, protocol.TaskStatusResponse_COMPLETED)
		}
	}
}

// 实现状态查询接口
func (s *Server) GetTaskStatus(ctx context.Context, req *protocol.TaskStatusRequest) (*protocol.TaskStatusResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.statusCache[req.TaskId]
	if !exists {
		return nil, grpcstatus.Error(codes.InvalidArgument, "task_id required")
	}
	//错误处理出现问题 应该是包引入冲突 给包设置别名就可以
	//如果不依赖grpc Status包 可以自定义错误类型+中间件拦截
	//比如
	/*
				// 自定义错误类型
				type GRPCError struct {
				    Code    codes.Code
				    Message string
				}

				func (e *GRPCError) Error() string {
				    return e.Message
				}

				// 辅助函数生成错误
				func NewGRPCError(code codes.Code, message string) error {
				    return &GRPCError{Code: code, Message: message}
				}
			// 修改 GetTaskStatus 方法
			func (s *Server) GetTaskStatus(ctx context.Context, req *protocol.TaskStatusRequest) (*protocol.TaskStatusResponse, error) {
			    // 参数校验
			    if req.TaskId == "" {
			        return nil, NewGRPCError(codes.InvalidArgument, "task_id required")
			    }

			    s.mu.RLock()
			    defer s.mu.RUnlock()

			    // 查询任务状态
			    taskStatus, exists := s.statusCache[req.TaskId]
			    if !exists {
			        return nil, NewGRPCError(codes.NotFound, "task not found")
			    }

			    return &protocol.TaskStatusResponse{
			        Status: taskStatus,
			    }, nil
			}
		// 错误拦截中间件
		func errorInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		    resp, err = handler(ctx, req)
		    if err != nil {
		        if e, ok := err.(*GRPCError); ok {
		            // 转换自定义错误为 gRPC 状态
		            return nil, grpcstatus.Error(e.Code, e.Message)
		        }
		        // 默认错误处理
		        return nil, grpcstatus.Error(codes.Unknown, err.Error())
		    }
		    return resp, nil
		}

		// 启用中间件
		func StartGRPCServer(addr string) *grpc.Server {
		    grpcServer := grpc.NewServer(
		        grpc.UnaryInterceptor(errorInterceptor),
		    )
		    // ...其他代码不变
		}
	*/

	return &protocol.TaskStatusResponse{
		Status: status,
	}, nil
}

// 实现ConwayService计算接口
func (s *Server) ComputeSync(ctx context.Context, task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: processTask(task),
		ExecTime:   time.Since(time.Unix(task.Timestamp, 0)).Seconds(),
	}, nil
}

// 辅助方法
func (s *Server) updateTaskStatus(taskID string, status protocol.TaskStatusResponse_TaskStatus) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statusCache[taskID] = status
}

func processTask(task *protocol.ComputeTask) []byte {
	// 实际元胞自动机计算逻辑
	return []byte{0x01, 0x00, 0x01} // 示例数据
}

// 启动gRPC服务
func StartGRPCServer(addr string) *grpc.Server {
	lis, err := net.Listen("tcp", addr) // 添加监听器
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	srv := NewServer(4)
	grpcServer := grpc.NewServer()
	protocol.RegisterBrokerServiceServer(grpcServer, srv)
	protocol.RegisterConwayServiceServer(grpcServer, srv)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()
	return grpcServer
}
