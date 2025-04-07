package main

import (
	"flag"
	"log"
	"time"

	"GolemCore/internal/client"
	"GolemCore/internal/rpc/protocol"
)

func main() {
	brokerAddr := flag.String("broker", "localhost:50051", "Broker服务地址")
	generations := flag.Int("gen", 100, "演化代数")
	gridFile := flag.String("grid", "pattern.gol", "初始网格文件")
	flag.Parse()

	// 初始化客户端
	cfg := client.Config{
		BrokerAddress:  *brokerAddr,
		RequestTimeout: 10 * time.Second,
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

	// 启动结果监听
	cli.StreamResults()

	// 构建任务
	task := &protocol.ComputeTask{
		TaskId:      generateTaskID(),
		GridData:    loadGridFile(*gridFile),
		Generations: int32(*generations),
		Timestamp:   time.Now().Unix(),
	}

	// 提交任务并获取ID
	taskID, err := cli.SubmitTask(task)
	if err != nil {
		log.Fatal("任务提交失败: ", err)
	}
	log.Printf("任务已提交，ID: %s", taskID)

	// 获取任务结果
	result, err := cli.GetResult(taskID)
	if err != nil {
		log.Fatal("获取结果失败: ", err)
	}

	// 处理结果
	log.Printf("任务完成，耗时: %.2f秒", result.ExecTime)
	if err := cli.VisualizeResult(result); err != nil {
		log.Printf("结果可视化失败: %v", err)
	}
}

func generateTaskID() string {
	return time.Now().Format("20060102-150405")
}

func loadGridFile(path string) []byte {
	// 示例数据（实际应从文件加载）
	return []byte{
		0x00, 0x01, 0x00,
		0x00, 0x01, 0x00,
		0x00, 0x01, 0x00,
	}
}

/*
package main

import (
	"flag"
	"log"
	"time"

	"GolemCore/internal/client"
	"GolemCore/internal/rpc/protocol"
)

func main() {
	brokerAddr := flag.String("broker", "localhost:50051", "Broker服务地址")
	generations := flag.Int("gen", 100, "演化代数")
	gridFile := flag.String("grid", "pattern.gol", "初始网格文件")
	flag.Parse()

	// 初始化客户端
	cfg := client.Config{
		BrokerAddress:  *brokerAddr,
		RequestTimeout: 10 * time.Second,
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

	// 启动结果监听
	cli.StreamResults()

	// 提交任务
	task := &protocol.ComputeTask{
		TaskId:      generateTaskID(),
		GridData:    loadGridFile(*gridFile),
		Generations: int32(*generations),
		Timestamp:   time.Now().Unix(),
	}

	_, err = cli.SubmitTask(task)
	if err != nil {
		log.Fatal("任务提交失败: ", err)
	}
	taskID, err := cli.SubmitTask(task)
	if err != nil {
		log.Fatal("任务提交失败: ", err)
	}
	log.Printf("任务已提交，ID: %s", taskID)

	// 处理结果
	log.Printf("任务完成，耗时: %.2f秒", result.ExecTime)
	if err := cli.VisualizeResult(result); err != nil {
		log.Printf("结果可视化失败: %v", err)
	}
}

func generateTaskID() string {
	return time.Now().Format("20060102-150405")
}

func loadGridFile(path string) []byte {
	// 实现网格文件加载（示例代码）
	// 实际应从文件读取二进制数据
	return []byte{
		0x00, 0x01, 0x00,
		0x00, 0x01, 0x00,
		0x00, 0x01, 0x00,
	}
}
*/
