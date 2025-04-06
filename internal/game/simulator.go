// E:\go\GolemCore\internal\game\simulator.go
package game

import (
	"sync"
)

// ParallelSimulator 并行演化模拟器
type ParallelSimulator struct {
	universe *Universe
	workers  int
	grid     *Grid
}

// NewParallelSimulator 创建并行模拟器
func NewParallelSimulator(u *Universe, workers int) *ParallelSimulator {
	return &ParallelSimulator{
		universe: u,
		workers:  workers,
	}
}

// Evolve 执行并行演化
func (s *ParallelSimulator) Evolve() {
	u := s.universe
	var wg sync.WaitGroup

	// 根据worker数量划分垂直区块
	chunkHeight := u.height / s.workers
	for w := 0; w < s.workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			startY := workerID * chunkHeight
			endY := startY + chunkHeight
			if workerID == s.workers-1 {
				endY = u.height // 最后一个worker处理剩余行
			}

			for y := startY; y < endY; y++ {
				for x := 0; x < u.width; x++ {
					neighbors := u.countNeighbors(x, y)
					u.buffer[y][x] = applyRules(u.cells[y][x], neighbors)
				}
			}
		}(w)
	}

	wg.Wait()
	u.NextGeneration()
}
func (s *ParallelSimulator) SetGrid(g *Grid) {
	s.grid = g.Clone() // 使用深拷贝保证线程安全
}

func (s *ParallelSimulator) GetGrid() *Grid {
	return s.grid.Clone()
}

// 添加克隆方法
func (g *Grid) Clone() *Grid {
	newGrid := NewGrid(g.Width, g.Height)
	for y := range g.Cells {
		copy(newGrid.Cells[y], g.Cells[y])
	}
	return newGrid
}
