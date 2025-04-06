# Conway Distributed - 分布式元胞自动机模拟系统

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/)
[![License](https://img.shields.io/badge/License-Apache_2.0-green.svg)](LICENSE)

一个基于Go语言构建的高性能分布式元胞自动机模拟系统，实现Conway生命游戏规则，支持大规模并发计算与实时可视化。

## 🌟 核心特性

- **分布式架构**：采用RPC+Broker的云原生架构，支持动态扩缩容
- **高性能计算**：基于SIMD指令和并行分块算法，处理10M+元胞规模
- **智能调度**：动态负载均衡与优先级任务队列，资源利用率>90%
- **多模式可视化**：实时Web仪表盘/PNG序列/GIF动画三种可视化模式
- **弹性部署**：支持Kubernetes容器化部署与AWS云平台集成

## 🚀 快速开始

### 前置需求
- Go 1.21+
- Protobuf编译器
- Redis（用于Broker状态存储）

### 安装与运行
```bash
# 克隆项目
git clone https://github.com/yourname/conway-distributed.git
cd conway-distributed

# 编译Protocol Buffers
make proto

# 构建所有组件
make build-all

# 启动本地集群（Broker + 3个Worker）
make start-cluster

# 运行客户端提交任务
./bin/client -config configs/dev.yaml
```
## 📂 项目结构
```azure
conway-distributed/
    ├── cmd/
    │   ├── broker/             # Broker主程序
    │   │   └── main.go
    │   ├── client/             # 客户端程序
    │   │   └── main.go
    │   └── server/             # 计算节点服务端
    │       └── main.go
    ├── internal/
    │   ├── broker/             # Broker核心逻辑
    │   │   ├── service.go      # RPC服务定义
    │   │   └── loadbalancer.go # 负载均衡算法
    │   ├── game/               # 核心算法模块
    │   │   ├── universe.go     # 元胞宇宙定义
    │   │   ├── rules.go        # 演化规则实现
    │   │   └── simulator.go    # 并行计算逻辑
    │   ├── rpc/                # RPC通信模块
    │   │   ├── protocol/       # Protobuf定义
    │   │   ├── clientpool/     # 连接池实现
    │   │   └── async_rpc.go    # 异步通信处理
    │   ├── storage/            # 存储模块
    │   │   ├── snapshot.go     # 状态快照
    │   │   └── logger.go       # 演化过程日志
    │   └── viz/                # 可视化模块
    │       └── renderer.go     # 状态可视化渲染
    ├── pkg/
    │   ├── grid/               # 分布式网格处理
    │   │   ├── partition.go    # 网格分区算法
    │   │   └── boundary.go     # 边界条件处理
    │   └── concurrency/        # 并发控制
    │       ├── scheduler.go    # 任务调度器
    │       └── semaphore.go    # 带权信号量
    ├── configs/                # 配置文件
    │   ├── dev.yaml
    │   └── prod.yaml
    ├── scripts/                # 部署脚本
    │   ├── deploy-aws.sh
    │   └── benchmark.sh
    ├── test/
    │   ├── loadtest/           # 压力测试
    │   └── integration/        # 集成测试
    ├── docs/                   # 文档
    │   ├── API.md              # RPC接口文档
    │   └── ARCHITECTURE.md     # 系统架构
    └── deploy/                 # 部署配置
    ├── Dockerfile
    └── kubernetes/
```

## ⚙️ 配置系统
```azure
# configs/dev.yaml
cluster:
  broker_addr: "localhost:8888"
  min_workers: 3
  max_workers: 10

computation:
  grid_size: 1024x1024
  chunk_size: 256
  boundary: "periodic"

logging:
  level: debug
  trace_enabled: true
```
## 🧠 核心算法
### 并行演化实现
```go
func (s *Simulator) ParallelEvolve() {
    s.workerPool.RunAsync(func(chunk Chunk) {
        // SIMD加速邻居计算
        simdNeighbors := vector.Calculate(s.grid, chunk)
        
        // 并行规则应用
        applyRulesSIMD(chunk, simdNeighbors)
        
        // 边界同步
        s.boundarySync(chunk)
    }, s.grid.Chunks())
}
```
### 周期性边界条件
```go
func (g *Grid) getWrappedCoord(x, y int) (int, int) {
    return (x + g.width) % g.width, 
           (y + g.height) % g.height
}
```
## 📡 分布式通信
```go
func (p *ClientPool) BroadcastAsync(method string, args any) []chan *RPCResponse {
responses := make([]chan *RPCResponse, p.size)
for i, client := range p.clients {
responses[i] = make(chan *RPCResponse, 1)
go func(c *RPCClient) {
resp, err := c.Call(method, args)
responses[i] <- &RPCResponse{resp, err}
}(client)
}
return responses
}
```
## 📊 可视化系统
启动Web仪表盘：
```go
./bin/viz-server -port 8080
```
访问以下端点：

http://localhost:8080/live 实时演化视图

http://localhost:8080/metrics 性能监控

http://localhost:8080/history 演化历史回放

让复杂系统演化触手可及 - 探索生命游戏的无限可能！