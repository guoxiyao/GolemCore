package main

import (
	"flag"
	"log"
	"net"

	"GolemCore/internal/rpc/protocol"
	"GolemCore/internal/server"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50052", "计算节点服务地址")
	capacity := flag.Int("cap", 100, "最大并行任务数")
	flag.Parse()

	// 注册到Broker
	registerToBroker(*addr, *capacity)

	// 启动计算服务
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer()
	protocol.RegisterConwayServiceServer(grpcServer, server.NewServer(*capacity))
	log.Printf("计算节点已启动，容量: %d", *capacity)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}

func registerToBroker(addr string, cap int) {
	// 实现注册逻辑，例如：
	// conn, _ := grpc.Dial(brokerAddr)
	// client := protocol.NewBrokerServiceClient(conn)
	// client.RegisterWorker(...)
}
