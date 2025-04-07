package main

import (
	"flag"
	"log"
	"net"

	"GolemCore/internal/broker"
	"GolemCore/internal/rpc/protocol"
	"google.golang.org/grpc"
)

func main() {
	addr := flag.String("addr", ":50051", "gRPC服务监听地址")
	strategy := flag.String("lb", "weighted_roundrobin", "负载均衡策略 (weighted_roundrobin/leastconn)")
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
	b := broker.NewBroker(lb)

	// 启动gRPC服务
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer()
	protocol.RegisterBrokerServiceServer(grpcServer, broker.NewBrokerService(b))
	log.Printf("Broker服务已启动，策略: %s，监听地址: %s", *strategy, *addr)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
