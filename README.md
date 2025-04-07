# GolemCore

GolemCore 是一个基于 Go 语言实现的分布式生命游戏计算系统。它采用主从架构，通过 Broker 节点协调任务分发，Worker 节点执行计算，Client 节点提交任务和获取结果。

## 系统架构

系统包含三个主要组件：

1. **Broker（代理节点）**
   - 负责任务调度和分发
   - 管理 Worker 节点注册和心跳
   - 实现负载均衡策略
   - 提供 gRPC 服务接口

2. **Worker（工作节点）**
   - 执行具体的生命游戏计算
   - 自动注册到 Broker
   - 定期发送心跳
   - 处理计算任务

3. **Client（客户端）**
   - 提交计算任务
   - 获取计算结果
   - 支持同步和异步操作

## 功能特性

- 分布式计算：支持多节点并行计算
- 负载均衡：支持加权轮询和最少连接策略
- 自动容错：节点故障检测和自动恢复
- 任务追踪：实时监控任务状态
- 性能优化：高效的数据序列化和传输

## 核心逻辑说明

### 1. 生命游戏计算
生命游戏基于以下规则：
- 任何活细胞如果周围活细胞数小于2个，则死亡（模拟人口稀少）
- 任何活细胞如果周围有2个或3个活细胞，则继续存活
- 任何活细胞如果周围活细胞数大于3个，则死亡（模拟人口过剩）
- 任何死细胞如果周围正好有3个活细胞，则复活（模拟繁殖）

### 2. 分布式计算流程
1. **任务提交**
   - Client 将初始网格和演化代数发送给 Broker
   - Broker 生成唯一任务ID并记录任务状态

2. **任务分发**
   - Broker 根据负载均衡策略选择 Worker
   - 将任务分发给选中的 Worker
   - 更新任务状态为 PROCESSING

3. **计算执行**
   - Worker 接收任务后开始计算
   - 每代演化都应用生命游戏规则
   - 计算完成后将结果返回给 Broker

4. **结果返回**
   - Broker 接收计算结果
   - 更新任务状态为 COMPLETED
   - 将结果返回给 Client

### 3. 节点管理
1. **Worker 注册**
   - Worker 启动时向 Broker 注册
   - 提供节点地址和计算容量
   - Broker 将节点加入可用节点列表

2. **心跳机制**
   - Worker 每15秒发送一次心跳
   - Broker 记录节点最后活动时间
   - 超过1分钟无心跳的节点被标记为失效

3. **负载均衡**
   - 支持两种策略：
     - 加权轮询：根据节点权重分配任务
     - 最少连接：选择当前负载最轻的节点

### 4. 数据序列化
1. **网格数据**
   - 使用 gob 编码序列化网格
   - 包含网格宽度、高度和细胞状态
   - 优化传输效率

2. **任务数据**
   - 包含任务ID、网格数据和演化代数
   - 使用 Protocol Buffers 定义数据结构
   - 支持跨语言通信

### 5. 错误处理
1. **节点故障**
   - 自动检测失效节点
   - 将任务重新分配给其他节点
   - 更新任务状态

2. **网络错误**
   - 实现重试机制
   - 设置超时时间
   - 记录错误日志

3. **任务失败**
   - 标记失败任务
   - 提供错误信息
   - 支持任务重试

## 代码示例说明

### 1. 生命游戏核心实现
```go
// internal/game/grid.go
type Grid struct {
    Width    int
    Height   int
    Cells    [][]bool
    Boundary BoundaryType
}

// 计算下一代
func (g *Grid) NextGeneration() {
    next := make([][]bool, g.Height)
    for y := 0; y < g.Height; y++ {
        next[y] = make([]bool, g.Width)
        for x := 0; x < g.Width; x++ {
            neighbors := g.countNeighbors(x, y)
            next[y][x] = g.Cells[y][x]
            // 应用生命游戏规则
            if g.Cells[y][x] {
                if neighbors < 2 || neighbors > 3 {
                    next[y][x] = false
                }
            } else {
                if neighbors == 3 {
                    next[y][x] = true
                }
            }
        }
    }
    g.Cells = next
}
```

### 2. 任务分发流程
```go
// internal/broker/broker.go
func (b *Broker) SubmitTask(ctx context.Context, task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
    // 1. 生成任务ID
    taskID := generateTaskID(task.ClientId)
    
    // 2. 选择Worker节点
    worker, err := b.selectWorker()
    if err != nil {
        return nil, err
    }
    
    // 3. 分发任务
    result, err := worker.ExecuteTask(task)
    if err != nil {
        return nil, err
    }
    
    return result, nil
}
```

### 3. Worker注册和心跳
```go
// cmd/worker/main.go
func registerToBroker(brokerAddr string, workerAddr string, cap int) {
    // 1. 连接到Broker
    conn, err := grpc.Dial(brokerAddr, grpc.WithInsecure())
    if err != nil {
        log.Fatalf("连接失败: %v", err)
    }
    
    // 2. 发送注册请求
    client := protocol.NewBrokerServiceClient(conn)
    resp, err := client.RegisterWorker(ctx, &protocol.WorkerRegistration{
        Address:  workerAddr,
        Capacity: int32(cap),
    })
}

// 定期发送心跳
func sendHeartbeat() {
    ticker := time.NewTicker(15 * time.Second)
    for range ticker.C {
        client.SendHeartbeat(ctx, &protocol.Heartbeat{
            WorkerId: workerID,
        })
    }
}
```

### 4. 客户端任务提交
```go
// internal/client/client.go
func (c *Client) SubmitTask(task *protocol.ComputeTask) (string, error) {
    // 1. 连接到Broker
    conn, err := grpc.Dial(c.config.BrokerAddress, grpc.WithInsecure())
    if err != nil {
        return "", err
    }
    
    // 2. 提交任务
    client := protocol.NewBrokerServiceClient(conn)
    resp, err := client.SubmitTask(ctx, task)
    if err != nil {
        return "", err
    }
    
    return resp.TaskId, nil
}
```

### 5. 数据序列化
```go
// internal/rpc/protocol/conway.proto
message ComputeTask {
    string task_id = 1;
    bytes grid_data = 2;
    int32 generations = 3;
}

message ComputeResult {
    string task_id = 1;
    bytes result_data = 2;
    double exec_time = 3;
}

// internal/broker/broker.go
func serializeGrid(grid *game.Grid) []byte {
    var buf bytes.Buffer
    enc := gob.NewEncoder(&buf)
    err := enc.Encode(grid)
    if err != nil {
        return nil
    }
    return buf.Bytes()
}
```

这些代码示例展示了系统的核心实现：
1. 生命游戏的核心计算逻辑
2. Broker 如何分发任务
3. Worker 如何注册和维持心跳
4. Client 如何提交任务
5. 数据如何序列化和传输

## 快速开始

### 环境要求

- Go 1.16 或更高版本
- 支持 gRPC 的网络环境

### 启动说明

系统需要按顺序启动三个组件：

1. **启动 Broker**
```bash
go run cmd/broker/main.go
```
成功启动后显示：
```
Broker服务已启动，策略: weighted_roundrobin，监听地址: :50051，工作节点: localhost:50052
```

2. **启动 Worker**
```bash
go run cmd/worker/main.go
```
成功启动后显示：
```
成功注册到Broker，ID: 0.0.0.0:50052
计算节点已启动，容量: 100
```

3. **启动 Client**
```bash
go run cmd/client/main.go
```
成功启动后会自动提交任务并显示结果。

### 启动参数

#### Broker 参数
- `-addr`: gRPC服务监听地址（默认: ":50051"）
- `-lb`: 负载均衡策略（可选: "weighted_roundrobin"/"leastconn"，默认: "weighted_roundrobin"）
- `-worker`: 工作节点地址（默认: "localhost:50052"）

#### Worker 参数
- `-addr`: 计算节点服务地址（默认: "0.0.0.0:50052"）
- `-broker`: Broker服务地址（默认: "localhost:50051"）
- `-cap`: 最大并行任务数（默认: 100）

#### Client 参数
- `-broker`: Broker服务地址（默认: "localhost:50051"）
- `-gen`: 演化代数（默认: 1000）
- `-grid`: 初始网格文件（默认: "pattern.gol"）
- `-client`: 客户端ID（默认: "client1"）

## 项目结构

```
GolemCore/
├── cmd/                    # 命令行入口
│   ├── broker/            # Broker 主程序
│   ├── client/            # Client 主程序
│   └── worker/            # Worker 主程序
├── internal/              # 内部包
│   ├── broker/           # Broker 核心逻辑
│   ├── client/           # Client 核心逻辑
│   ├── conway/           # 生命游戏实现
│   ├── game/             # 游戏核心逻辑
│   └── rpc/              # RPC 相关实现

```


## 注意事项

1. 确保按顺序启动组件：Broker -> Worker -> Client
2. 如果端口被占用，可以修改相应的 `-addr` 参数
3. Worker 会自动注册到 Broker 并定期发送心跳
4. Client 完成任务后会自动退出
5. 确保网络环境允许 gRPC 通信

## 性能优化

- 使用高效的序列化方式
- 实现任务批处理
- 优化内存使用
- 减少网络传输

## 未来计划

### 1. 实时可视化实现
- [ ] **Web界面可视化**
  - 使用HTML5 Canvas或CSS Grid显示网格
  - WebSocket实时更新细胞状态
  - 支持暂停/继续/重置操作
  - 显示演化统计信息

- [ ] **系统监控面板**
  - Worker节点状态监控
  - 任务队列可视化
  - 性能指标实时展示
  - 系统资源使用统计

- [ ] **可视化控制功能**
  - 手动编辑初始网格
  - 调整演化速度
  - 保存/加载网格状态
  - 预设模式选择

### 2. 系统优化
- [ ] 添加更多负载均衡策略
- [ ] 实现任务优先级
- [ ] 优化错误处理
- [ ] 添加日志系统

### 3. 功能扩展
- [ ] 支持更多元胞自动机规则
- [ ] 添加REST API接口
- [ ] 实现集群管理功能
- [ ] 支持Docker部署

## 许可证

MIT License

## 运行示例

系统启动后，各组件的运行日志如下：

### 1. Broker 运行日志
```
2025/04/07 23:31:05 Broker服务已启动，策略: weighted_roundrobin, 监听地址: :50051, 工作节点: localhost:50052
2025/04/07 23:31:15 开始处理客户端 client1 的结果流
2025/04/07 23:31:15 开始处理任务: client1-20250407-233115
2025/04/07 23:31:15 任务 client1-20250407-233115 状态更新为: PROCESSING
2025/04/07 23:31:15 可用工作节点: map[0.0.0.0:50052:0xc00012c300 localhost:50052:0xc000099740]
2025/04/07 23:31:15 同步后的节点列表: [localhost:50052 0.0.0.0:50052]
2025/04/07 23:31:15 当前节点列表: [localhost:50052 0.0.0.0:50052]
2025/04/07 23:31:15 当前权重: map[0.0.0.0:50052:10 localhost:50052:10]
2025/04/07 23:31:15 选择节点: 0.0.0.0:50052 (权重: 10)
2025/04/07 23:31:15 任务 client1-20250407-233115 分发到节点: 0.0.0.0:50052
2025/04/07 23:31:15 任务 client1-20250407-233115 状态更新为: COMPLETED
2025/04/07 23:31:15 任务完成: client1-20250407-233115, 耗时: 0.00 秒
```

### 2. Worker 运行日志
```
2025/04/07 23:31:10 成功注册到Broker, ID: 0.0.0.0:50052
2025/04/07 23:31:10 计算节点已启动, 容量: 100
2025/04/07 23:31:15 开始处理任务 client1-20250407-233115, 演化代数: 1000
2025/04/07 23:31:15 解析后的网格状态:
2025/04/07 23:31:15 0 1 0
2025/04/07 23:31:15 0 1 0
2025/04/07 23:31:15 0 1 0
2025/04/07 23:31:15 任务 client1-20250407-233115 已完成 1000/1000 代演化
2025/04/07 23:31:15 任务 client1-20250407-233115 完成, 总耗时: 0.000 秒
```

### 3. Client 运行日志
```
2025/04/07 23:31:15 任务已提交, ID: client1-20250407-233115
2025/04/07 23:31:17 任务完成, 耗时: 0.000秒
```


