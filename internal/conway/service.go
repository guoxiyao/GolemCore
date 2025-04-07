package conway

import (
	"GolemCore/internal/rpc/protocol"
	"context"
	"errors"
	"log"
	"time"
)

type ConwayService struct {
	protocol.UnimplementedConwayServiceServer
}

// NewConwayService 构造函数
func NewConwayService() *ConwayService {
	return &ConwayService{}
}

// ComputeSync 同步计算
func (s *ConwayService) ComputeSync(
	ctx context.Context,
	task *protocol.ComputeTask,
) (*protocol.ComputeResult, error) {
	// 记录开始时间
	startTime := time.Now()
	log.Printf("开始处理任务 %s，演化代数: %d", task.TaskId, task.Generations)

	// 检查输入数据
	if task.GridData == nil {
		log.Printf("任务 %s 的网格数据为空", task.TaskId)
		return nil, errors.New("grid data is nil")
	}
	if len(task.GridData) == 0 {
		log.Printf("任务 %s 的网格数据长度为0", task.TaskId)
		return nil, errors.New("grid data is empty")
	}

	// 解析输入网格
	grid := parseGrid(task.GridData)
	if grid == nil {
		log.Printf("任务 %s 的网格数据解析失败", task.TaskId)
		return nil, errors.New("invalid grid data")
	}

	// 执行指定代数的演化
	for i := 0; i < int(task.Generations); i++ {
		// 检查上下文是否被取消
		select {
		case <-ctx.Done():
			log.Printf("任务 %s 被取消", task.TaskId)
			return nil, ctx.Err()
		default:
			grid = evolve(grid)
			// 每10代打印一次进度
			if (i+1)%10 == 0 {
				log.Printf("任务 %s 已完成 %d/%d 代演化", task.TaskId, i+1, task.Generations)
			}
		}
	}

	// 计算执行时间
	execTimeMs := time.Since(startTime).Milliseconds()
	execTime := float64(execTimeMs) / 1000.0
	log.Printf("任务 %s 完成，总耗时: %.3f 秒", task.TaskId, execTime)

	// 返回结果
	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: serializeGrid(grid),
		ExecTime:   execTime,
	}, nil
}

// parseGrid 解析网格数据
func parseGrid(data []byte) [][]bool {
	// 检查数据长度是否为9（3x3网格）
	if len(data) != 9 {
		log.Printf("无效的网格数据长度: %d，期望长度: 9", len(data))
		return nil
	}

	// 创建3x3网格
	grid := make([][]bool, 3)
	for i := range grid {
		grid[i] = make([]bool, 3)
		for j := range grid[i] {
			// 将字节值转换为布尔值
			// 任何非零值都视为活细胞
			grid[i][j] = data[i*3+j] != 0
		}
	}

	// 打印网格状态以便调试
	log.Printf("解析后的网格状态:")
	for i := range grid {
		row := ""
		for j := range grid[i] {
			if grid[i][j] {
				row += "1 "
			} else {
				row += "0 "
			}
		}
		log.Printf("%s", row)
	}

	return grid
}

// evolve 执行一代演化
func evolve(grid [][]bool) [][]bool {
	rows := len(grid)
	cols := len(grid[0])
	newGrid := make([][]bool, rows)
	for i := range newGrid {
		newGrid[i] = make([]bool, cols)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			neighbors := countNeighbors(grid, i, j)
			if grid[i][j] {
				// 活细胞的规则
				newGrid[i][j] = neighbors == 2 || neighbors == 3
			} else {
				// 死细胞的规则
				newGrid[i][j] = neighbors == 3
			}
		}
	}

	return newGrid
}

// countNeighbors 计算邻居数量
func countNeighbors(grid [][]bool, row, col int) int {
	count := 0
	rows := len(grid)
	cols := len(grid[0])

	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			if i == 0 && j == 0 {
				continue
			}
			r := (row + i + rows) % rows
			c := (col + j + cols) % cols
			if grid[r][c] {
				count++
			}
		}
	}
	return count
}

// serializeGrid 序列化网格数据
func serializeGrid(grid [][]bool) []byte {
	rows := len(grid)
	cols := len(grid[0])
	data := make([]byte, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			if grid[i][j] {
				data[i*cols+j] = 1
			} else {
				data[i*cols+j] = 0
			}
		}
	}
	return data
}
