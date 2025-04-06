package grid

import (
	"math"
	"sync"
)

// 网格分区元数据
type Partition struct {
	StartX int
	StartY int
	SizeX  int
	SizeY  int
	Data   [][]float64
	Mutex  sync.RWMutex
}

// 网格分割器
type Partitioner struct {
	Overlap    int // 分区重叠区域大小
	Grid       [][]float64
	Partitions []*Partition
}

// 创建新的分区器
func NewPartitioner(grid [][]float64, overlap int) *Partitioner {
	return &Partitioner{
		Overlap: overlap,
		Grid:    grid,
	}
}

// 水平分割网格
func (p *Partitioner) SplitHorizontal(parts int) {
	height := len(p.Grid)
	partHeight := height / parts

	for i := 0; i < parts; i++ {
		start := i * partHeight
		end := start + partHeight
		if i == parts-1 {
			end = height
		}

		// 添加重叠区域
		start = max(0, start-p.Overlap)
		end = min(height, end+p.Overlap)

		p.Partitions = append(p.Partitions, &Partition{
			StartX: 0,
			StartY: start,
			SizeX:  len(p.Grid[0]),
			SizeY:  end - start,
			Data:   p.Grid[start:end],
		})
	}
}

// 合并分区结果
func (p *Partitioner) Merge() [][]float64 {
	result := make([][]float64, len(p.Grid))
	for i := range result {
		result[i] = make([]float64, len(p.Grid[0]))
	}

	var wg sync.WaitGroup
	for _, part := range p.Partitions {
		wg.Add(1)
		go func(part *Partition) {
			defer wg.Done()
			part.Mutex.RLock()
			defer part.Mutex.RUnlock()

			for y := 0; y < part.SizeY; y++ {
				globalY := part.StartY + y
				if globalY >= len(p.Grid) {
					continue
				}
				copy(result[globalY], part.Data[y])
			}
		}(part)
	}
	wg.Wait()
	return result
}

// 同步分区边界
func (p *Partitioner) SyncBoundaries(boundary BoundaryType) {
	handler := NewBoundaryHandler(boundary)
	var wg sync.WaitGroup

	for _, part := range p.Partitions {
		wg.Add(1)
		go func(part *Partition) {
			defer wg.Done()
			part.Mutex.Lock()
			defer part.Mutex.Unlock()

			// 同步左右边界
			for y := 0; y < part.SizeY; y++ {
				globalY := part.StartY + y
				for dx := -p.Overlap; dx <= p.Overlap; dx++ {
					if dx == 0 {
						continue
					}
					x := part.StartX + dx
					if x < 0 || x >= part.SizeX {
						adjX, adjY, _ := handler.Handle(x, globalY, part.SizeX, len(p.Grid))
						part.Data[y][x] = p.Grid[adjY][adjX]
					}
				}
			}
		}(part)
	}
	wg.Wait()
}

func max(a, b int) int {
	return int(math.Max(float64(a), float64(b)))
}

func min(a, b int) int {
	return int(math.Min(float64(a), float64(b)))
}
