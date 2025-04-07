package main

import (
	"GolemCore/internal/conway"
	"context"
	"flag"
	"log"
	"net"
	"time"

	"GolemCore/internal/rpc/protocol"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//计算节点 负责启动计算服务
//负责启动worker服务

func main() {
	addr := flag.String("addr", "0.0.0.0:50052", "计算节点服务地址")
	brokerAddr := flag.String("broker", "localhost:50051", "Broker服务地址")
	capacity := flag.Int("cap", 100, "最大并行任务数")
	flag.Parse()

	// 注册到Broker
	registerToBroker(*brokerAddr, *addr, *capacity)

	// 启动心跳
	go sendHeartbeat(*brokerAddr, *addr)

	// 启动计算服务
	lis, err := net.Listen("tcp", *addr)
	if err != nil {
		log.Fatalf("监听失败: %v", err)
	}

	grpcServer := grpc.NewServer()
	conwayService := conway.NewConwayService()
	protocol.RegisterConwayServiceServer(grpcServer, conwayService)
	log.Printf("计算节点已启动，容量: %d", *capacity)

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}

func registerToBroker(brokerAddr string, workerAddr string, cap int) {
	// 连接到Broker
	conn, err := grpc.Dial(brokerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		log.Fatalf("连接Broker失败: %v", err)
	}
	defer conn.Close()

	// 创建Broker客户端
	client := protocol.NewBrokerServiceClient(conn)

	// 注册工作节点
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RegisterWorker(ctx, &protocol.WorkerRegistration{
		Address:  workerAddr,
		Capacity: int32(cap),
	})
	if err != nil {
		log.Fatalf("注册失败: %v", err)
	}

	if resp.Status != protocol.RegistrationResponse_SUCCESS {
		log.Fatalf("注册失败，状态: %v", resp.Status)
	}

	log.Printf("成功注册到Broker，ID: %s", resp.AssignedId)
}

// 定期发送心跳
func sendHeartbeat(brokerAddr string, workerAddr string) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		// 连接到Broker
		conn, err := grpc.Dial(brokerAddr,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithBlock(),
		)
		if err != nil {
			log.Printf("连接Broker失败: %v", err)
			continue
		}

		// 创建Broker客户端
		client := protocol.NewBrokerServiceClient(conn)

		// 发送心跳
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		_, err = client.RegisterWorker(ctx, &protocol.WorkerRegistration{
			Address: workerAddr,
		})
		cancel()
		conn.Close()

		if err != nil {
			log.Printf("发送心跳失败: %v", err)
		} else {
			log.Printf("心跳发送成功")
		}
	}
}
