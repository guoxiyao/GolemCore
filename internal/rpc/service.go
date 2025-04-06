// E:\go\GolemCore\internal\rpc\service.go
package rpc

import (
	"bytes"
	"context"
	"encoding/gob"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	_ "log"
	"time"

	"GolemCore/internal/game"
	"GolemCore/internal/rpc/protocol"
	_ "google.golang.org/grpc"
)

type RPCServer struct {
	protocol.UnimplementedConwayServiceServer
	simulator *game.ParallelSimulator
}

func NewRPCServer(simulator *game.ParallelSimulator) *RPCServer {
	return &RPCServer{
		simulator: simulator,
	}
}

// ComputeSync 同步计算接口
func (s *RPCServer) ComputeSync(ctx context.Context, task *protocol.ComputeTask) (*protocol.ComputeResult, error) {
	start := time.Now()

	// 转换协议数据到游戏网格
	grid := deserializeGrid(task.GridData) // 添加返回值接收

	// 执行计算
	s.simulator.SetGrid(grid)
	for i := 0; i < int(task.Generations); i++ {
		s.simulator.Evolve()
	}

	// 获取结果网格
	resultGrid := s.simulator.GetGrid()

	return &protocol.ComputeResult{
		TaskId:     task.TaskId,
		ResultData: serializeGrid(resultGrid), // 传入有效参数
		ExecTime:   time.Since(start).Seconds(),
	}, nil
}

// ComputeStream 流式处理接口
// ComputeStream 流式处理接口（正确实现）
func (s *RPCServer) ComputeStream(stream protocol.ConwayService_ComputeStreamServer) error {
	for {
		// 接收流式请求
		task, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		// 处理任务
		result, err := s.ComputeSync(stream.Context(), task)
		if err != nil {
			return status.Error(codes.DataLoss, err.Error())
		}

		// 发送流式响应
		if sendErr := stream.Send(result); sendErr != nil {
			return status.Error(codes.Internal, sendErr.Error())
		}
	}
}

// 辅助函数
// 序列化网格为二进制格式
func serializeGrid(grid *game.Grid) []byte {
	var buf bytes.Buffer
	enc := gob.NewEncoder(&buf)

	type ExportGrid struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	err := enc.Encode(ExportGrid{
		Width:  grid.Width,
		Height: grid.Height,
		Cells:  grid.Cells,
	})

	if err != nil {
		return []byte{}
	}
	return buf.Bytes()
}

// 反序列化二进制数据为网格
func deserializeGrid(data []byte) *game.Grid {
	var export struct {
		Width  int
		Height int
		Cells  [][]bool
	}

	dec := gob.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(&export); err != nil {
		return game.NewGrid(0, 0)
	}

	return &game.Grid{
		Width:  export.Width,
		Height: export.Height,
		Cells:  export.Cells,
	}
}
