package main

import (
	"GolemCore/internal/conway"
	"flag"
	"log"
	"net"

	"GolemCore/internal/broker"
	"GolemCore/internal/rpc/protocol"

	"google.golang.org/grpc"
)

//任务调度器 负责启动broker服务

func main() {
	addr := flag.String("addr", ":50051", "gRPC服务监听地址")
	strategy := flag.String("lb", "weighted_roundrobin", "负载均衡策略 (weighted_roundrobin/leastconn)")
	workerAddr := flag.String("worker", "localhost:50052", "工作节点地址")
	flag.Parse()

	// 初始化负载均衡器
	var lb broker.LoadBalancer
	switch *strategy {
	case "weighted_roundrobin":
		lb = broker.NewWeightedRoundRobin()
	case "leastconn":
		lb = broker.NewLeastConnLB() // 需要实现该类型
	default:
		log.Fatalf("不支持的负载均衡策略: %s", *strategy)
	}

	// 初始化Broker
	b := broker.NewBroker(lb, 1) // 1 worker by default

	// 添加工作节点
	b.AddWorker(*workerAddr)

	// 启动gRPC服务
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer()
	// 注册 Conway 服务
	conwayService := conway.NewConwayService()
	protocol.RegisterConwayServiceServer(grpcServer, conwayService)
	protocol.RegisterBrokerServiceServer(grpcServer, broker.NewBrokerService(b))
	log.Printf("Broker服务已启动，策略: %s，监听地址: %s，工作节点: %s", *strategy, *addr, *workerAddr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
