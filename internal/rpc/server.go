package rpc

import (
	"GolemCore/internal/broker"
	"GolemCore/internal/game"
	"GolemCore/internal/rpc/protocol"
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"io"
	"log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RPCServer RPC服务器
type RPCServer struct {
	protocol.UnimplementedConwayServiceServer
	broker *broker.Broker
	router *mux.Router
}

// NewRPCServer 创建RPC服务器
func NewRPCServer(broker *broker.Broker) *RPCServer {
	srv := &RPCServer{
		broker: broker,
		router: mux.NewRouter(),
	}
	srv.registerRoutes()
	return srv
}

// registerRoutes 注册HTTP路由
func (s *RPCServer) registerRoutes() {
	s.router.HandleFunc("/tasks", s.submitTask).Methods("POST")
	s.router.HandleFunc("/workers", s.registerWorker).Methods("POST")
	s.router.HandleFunc("/metrics", s.getMetrics).Methods("GET")
	s.router.Use(loggingMiddleware)
}

// Start 启动服务器
func (s *RPCServer) Start(grpcAddr, httpAddr string) error {
	// 启动gRPC服务器
	lis, err := net.Listen("tcp", grpcAddr)
	if err != nil {
		return err
	}

	grpcServer := grpc.NewServer()
	protocol.RegisterConwayServiceServer(grpcServer, s)

	go func() {
		log.Printf("gRPC服务器启动在 %s", grpcAddr)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC服务器错误: %v", err)
		}
	}()

	// 启动HTTP服务器
	go func() {
		log.Printf("HTTP服务器启动在 %s", httpAddr)
		if err := http.ListenAndServe(httpAddr, s.router); err != nil {
			log.Fatalf("HTTP服务器错误: %v", err)
		}
	}()

	return nil
}

// ComputeSync 同步计算
func (s *RPCServer) ComputeSync(ctx context.Context, task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
	start := time.Now()

	// 执行计算
	result, err := s.broker.SubmitTask(ctx, task)
	if err != nil {
		return nil, err
	}

	// 计算执行时间
	elapsed := time.Since(start)
	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: result.ResultData,
		ExecTime:   elapsed.Seconds(),
	}, nil
}

// ComputeStream 流式处理接口
func (s *RPCServer) ComputeStream(stream protocol.ConwayService_ComputeStreamServer) error {
	for {
		// 接收流式请求
		task, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// 处理任务
		result, err := s.ComputeSync(stream.Context(), task)
		if err != nil {
			return status.Error(codes.DataLoss, err.Error())
		}

		// 发送流式响应
		if sendErr := stream.Send(result); sendErr != nil {
			return status.Error(codes.Internal, sendErr.Error())
		}
	}
}

// HTTP处理函数
func (s *RPCServer) submitTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GridData    []byte `json:"grid_data"`
		Generations int    `json:"generations"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	task := &protocol.ComputeTask{
		GridData:    req.GridData,
		Generations: int32(req.Generations),
	}

	result, err := s.ComputeSync(r.Context(), task)
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "任务处理失败")
		return
	}

	respondWithJSON(w, http.StatusOK, struct {
		TaskID    string    `json:"task_id"`
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
	}{
		TaskID:    result.TaskId,
		Status:    "已完成",
		Timestamp: time.Now().UTC(),
	})
}

func (s *RPCServer) registerWorker(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Address  string `json:"address"`
		Capacity int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	respondWithJSON(w, http.StatusCreated, struct {
		TaskID    string    `json:"task_id"`
		Status    string    `json:"status"`
		Timestamp time.Time `json:"timestamp"`
	}{
		TaskID:    "worker_" + req.Address,
		Status:    "已注册",
		Timestamp: time.Now().UTC(),
	})
}

func (s *RPCServer) getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := struct {
		Tasks   int `json:"tasks"`
		Workers int `json:"workers"`
		Active  int `json:"active"`
		Queued  int `json:"queued"`
	}{
		Tasks:   0, // TODO: 实现指标收集
		Workers: 0,
		Active:  0,
		Queued:  0,
	}

	respondWithJSON(w, http.StatusOK, metrics)
}

// 辅助函数
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO: 实现请求日志记录
		next.ServeHTTP(w, r)
	})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

// 序列化网格为二进制格式
func serializeGrid(grid *game.Grid) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	type ExportGrid struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	err := enc.Encode(ExportGrid{
		Width:  grid.Width,
		Height: grid.Height,
		Cells:  grid.Cells,
	})

	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

// 反序列化二进制数据为网格
func deserializeGrid(data []byte) *game.Grid {
	var export struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&export); err != nil {
		return game.NewGrid(0, 0, game.Periodic)
	}

	return &game.Grid{
		Width:    export.Width,
		Height:   export.Height,
		Cells:    export.Cells,
		Boundary: game.Periodic,
	}
}
