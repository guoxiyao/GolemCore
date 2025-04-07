package conway

import (
	"GolemCore/internal/game"
	"sync"
)

// ParallelSimulator 并行演化模拟器
type ParallelSimulator struct {
	grid    *game.Grid
	workers int
	buffer  *game.Grid // 使用双缓冲避免竞争
}

// NewParallelSimulator 创建并行模拟器
func NewParallelSimulator(width, height, workers int) *ParallelSimulator {
	return &ParallelSimulator{
		grid:    game.NewGrid(width, height, game.Periodic),
		workers: workers,
		buffer:  game.NewGrid(width, height, game.Periodic),
	}
}

// Evolve 执行并行演化
func (s *ParallelSimulator) Evolve() {
	var wg sync.WaitGroup

	// 根据worker数量划分垂直区块
	chunkHeight := s.grid.Height / s.workers
	for w := 0; w < s.workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			startY := workerID * chunkHeight
			endY := startY + chunkHeight
			if workerID == s.workers-1 {
				endY = s.grid.Height // 最后一个worker处理剩余行
			}

			for y := startY; y < endY; y++ {
				for x := 0; x < s.grid.Width; x++ {
					neighbors := game.CountNeighbors(s.grid, x, y)
					s.buffer.Set(x, y, applyRules(s.grid.Get(x, y), neighbors))
				}
			}
		}(w)
	}

	wg.Wait()

	// 交换缓冲区
	s.grid, s.buffer = s.buffer, s.grid
}

// SetGrid 设置网格状态
func (s *ParallelSimulator) SetGrid(g *game.Grid) {
	s.grid = g.Clone()
	s.buffer = game.NewGrid(g.Width, g.Height, g.Boundary)
}

// GetGrid 获取当前网格状态
func (s *ParallelSimulator) GetGrid() *game.Grid {
	return s.grid.Clone()
}

// applyRules 应用Conway生命游戏规则
func applyRules(current bool, neighbors int) bool {
	switch {
	case current && (neighbors == 2 || neighbors == 3):
		return true
	case !current && neighbors == 3:
		return true
	default:
		return false
	}
}
