# GolemCore - 分布式元胞自动机模拟系统

GolemCore 是一个使用 Go 语言实现的分布式元胞自动机模拟系统，基于 John Conway 的经典"生命游戏"规则。该系统采用 RPC 架构，支持大规模并发计算和高效数据交换。

## 特性

- 基于 gRPC 的分布式架构
- 高性能并行计算优化
- 支持大规模矩阵演化计算
- 周期性边界条件支持
- 实时可视化支持

## 项目结构

```
GolemCore/
├── cmd/
│   ├── client/        # 客户端程序
│   └── server/        # 服务器程序
├── internal/
│   ├── broker/        # 消息代理
│   ├── cell/          # 元胞自动机核心逻辑
│   ├── rpc/           # RPC相关实现
│   └── utils/         # 工具函数
├── pkg/
│   ├── config/        # 配置管理
│   └── types/         # 共享类型定义
├── api/               # API定义
├── scripts/           # 部署脚本
└── docs/              # 文档
```

## 快速开始

### 依赖安装

```bash
go get -u google.golang.org/grpc
go get -u github.com/golang/protobuf/protoc-gen-go
```

### 编译

```bash
# 生成 protobuf 代码
protoc --go_out=plugins=grpc:. api/service.proto

# 编译服务器
go build -o bin/server cmd/server/main.go

# 编译客户端
go build -o bin/client cmd/client/main.go
```

### 运行

1. 启动服务器：
```bash
./bin/server
```

2. 启动客户端：
```bash
./bin/client
```

## 配置说明

- 服务器默认监听端口：50051
- 网格默认大小：100x100
- 并行计算分片数：4

## 性能优化

- 使用 goroutine 进行并行计算
- 采用连接池优化 RPC 通信
- 使用读写锁保护共享资源
- 实现周期性边界条件，支持闭环域

## 许可证

MIT License 