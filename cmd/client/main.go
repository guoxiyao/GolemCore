package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"GolemCore/internal/client"
	"GolemCore/internal/rpc/protocol"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//客户端 负责启动客户端服务 提交任务和获取结果

func main() {
	brokerAddr := flag.String("broker", "localhost:50051", "Broker服务地址")
	generations := flag.Int("gen", 1000, "演化代数")
	gridFile := flag.String("grid", "pattern.gol", "初始网格文件")
	clientID := flag.String("client", "client1", "客户端ID")
	flag.Parse()

	// 初始化客户端
	cfg := client.Config{
		BrokerAddress:  *brokerAddr,
		RequestTimeout: 30 * time.Second,
		ClientID:       *clientID,
	}
	cli, err := client.NewClient(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer cli.Shutdown()

	// 连接到Broker
	if err := cli.Connect(); err != nil {
		log.Fatal("连接Broker失败: ", err)
	}

	// 构建任务
	task := &protocol.ComputeTask{
		TaskId:      generateTaskID(*clientID),
		GridData:    loadGridFile(*gridFile),
		Generations: int32(*generations),
		Timestamp:   time.Now().Unix(),
	}

	// 启动结果监听（在提交任务之前）
	go cli.StreamResults()

	// 提交任务
	taskID, err := cli.SubmitTask(task)
	if err != nil {
		log.Fatal("任务提交失败: ", err)
	}
	log.Printf("任务已提交，ID: %s", taskID)

	// 等待一段时间，让任务开始处理
	time.Sleep(2 * time.Second)

	// 等待任务完成
	for {
		status, err := cli.GetTaskStatus(taskID)
		if err != nil {
			log.Printf("获取任务状态失败: %v", err)
			// 不要立即退出，继续重试
			time.Sleep(1 * time.Second)
			continue
		}

		switch status {
		case protocol.TaskStatusResponse_COMPLETED:
			// 获取任务结果
			result, err := cli.GetResult(taskID)
			if err != nil {
				log.Fatal("获取结果失败: ", err)
			}
			log.Printf("任务完成，耗时: %.3f秒", result.ExecTime)
			if err := cli.VisualizeResult(result); err != nil {
				log.Printf("结果可视化失败: %v", err)
			}
			return
		case protocol.TaskStatusResponse_FAILED:
			log.Fatal("任务执行失败")
		case protocol.TaskStatusResponse_PROCESSING:
			log.Printf("任务正在处理中...")
		case protocol.TaskStatusResponse_PENDING:
			log.Printf("任务等待处理中...")
		}

		time.Sleep(1 * time.Second)
	}
}

// 生成任务ID
func generateTaskID(clientID string) string {
	return fmt.Sprintf("%s-%s", clientID, time.Now().Format("20060102-150405"))
}

// 加载网格文件
func loadGridFile(path string) []byte {
	// 创建一个3x3的网格，中间一列是活细胞
	return []byte{
		0, 1, 0,
		0, 1, 0,
		0, 1, 0,
	}
}

// 注册到Broker
func registerToBroker(addr string, cap int) {
	// 连接到Broker
	// 问题原因：之前没有实现注册逻辑，导致工作节点无法被Broker识别
	// 修改原因：工作节点需要主动注册到Broker，以便Broker知道它的存在
	conn, err := grpc.Dial("localhost:50051",
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
	// 问题原因：工作节点没有向Broker提供必要的信息（地址和容量）
	// 修改原因：Broker需要知道工作节点的地址和容量来正确分配任务
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.RegisterWorker(ctx, &protocol.WorkerRegistration{
		Address:  addr,
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
