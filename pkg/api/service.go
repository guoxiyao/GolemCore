package api

import (
	"GolemCore/internal/rpc/protocol"
	"encoding/json"
	"net/http"
	"time"

	"GolemCore/pkg/concurrency"
	"GolemCore/pkg/engine"
	"github.com/gorilla/mux"
)

type APIService struct {
	router  *mux.Router
	engine  *engine.Engine
	sched   *concurrency.Scheduler
	metrics MetricsCollector
}

type MetricsCollector interface {
	GetSchedulerMetrics() concurrency.SchedulerMetrics
	GetEngineStatus() engine.EngineStatus
}

// API请求响应格式
type TaskResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

// 任务提交请求体
type SubmitTaskRequest struct {
	GridData    []byte `json:"grid_data"`
	Generations int    `json:"generations"`
}

// 工作节点注册请求体
type RegisterWorkerRequest struct {
	Address  string `json:"address"`
	Capacity int    `json:"capacity"`
}

func NewAPIService(e *engine.Engine, s *concurrency.Scheduler, m MetricsCollector) *APIService {
	srv := &APIService{
		router:  mux.NewRouter(),
		engine:  e,
		sched:   s,
		metrics: m,
	}
	srv.registerRoutes()
	return srv
}

func (s *APIService) registerRoutes() {
	s.router.HandleFunc("/tasks", s.submitTask).Methods("POST")
	s.router.HandleFunc("/workers", s.registerWorker).Methods("POST")
	s.router.HandleFunc("/metrics", s.getMetrics).Methods("GET")
	s.router.Use(loggingMiddleware)
}

// 任务提交处理器
func (s *APIService) submitTask(w http.ResponseWriter, r *http.Request) {
	var req SubmitTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	taskID, err := s.engine.SubmitTask(convertToProtocolTask(req))
	if err != nil {
		respondWithError(w, http.StatusServiceUnavailable, "任务提交失败")
		return
	}

	respondWithJSON(w, http.StatusAccepted, TaskResponse{
		TaskID:    taskID,
		Status:    "已接受",
		Timestamp: time.Now().UTC(),
	})
}

// 工作节点注册处理器
func (s *APIService) registerWorker(w http.ResponseWriter, r *http.Request) {
	var req RegisterWorkerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "无效的请求格式")
		return
	}

	if err := s.engine.RegisterWorker(req.Address, int32(req.Capacity)); err != nil {
		respondWithError(w, http.StatusConflict, "节点注册失败")
		return
	}

	respondWithJSON(w, http.StatusCreated, TaskResponse{
		TaskID:    "worker_" + req.Address,
		Status:    "已注册",
		Timestamp: time.Now().UTC(),
	})
}

// 系统指标查询
func (s *APIService) getMetrics(w http.ResponseWriter, r *http.Request) {
	metrics := struct {
		Scheduler concurrency.SchedulerMetrics `json:"scheduler"`
		Engine    engine.EngineStatus          `json:"engine_status"`
	}{
		Scheduler: s.metrics.GetSchedulerMetrics(),
		Engine:    s.metrics.GetEngineStatus(),
	}

	respondWithJSON(w, http.StatusOK, metrics)
}

// 启动HTTP服务
func (s *APIService) Start(addr string) error {
	return http.ListenAndServe(addr, s.router)
}

// 辅助函数
func convertToProtocolTask(req SubmitTaskRequest) *protocol.ComputeTask {
	return &protocol.ComputeTask{
		GridData:    req.GridData,
		Generations: int32(req.Generations),
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 请求日志记录逻辑
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
