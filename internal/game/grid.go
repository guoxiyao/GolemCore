// E:\go\GolemCore\internal\game\grid.go
package game

import (
	"sync"
)

// 边界类型
type BoundaryType int

const (
	Periodic BoundaryType = iota
	Fixed
	Reflective
)

// 网格
type Grid struct {
	Width    int
	Height   int
	Cells    [][]bool //每个细胞的状态 是否存活
	Boundary BoundaryType
	mu       sync.RWMutex
}

// 获取指定位置的细胞状态
func (g *Grid) Get(x, y int) bool {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// 处理边界
	x, y = g.handleBoundary(x, y)
	return g.Cells[y][x]
}

// 设置指定位置的细胞状态
func (g *Grid) Set(x, y int, state bool) {
	g.mu.Lock()
	defer g.mu.Unlock()

	// 处理边界
	x, y = g.handleBoundary(x, y)
	g.Cells[y][x] = state
}

// 创建一个新的网格
func NewGrid(width, height int, boundary BoundaryType) *Grid {
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}
	return &Grid{
		Width:    width,
		Height:   height,
		Cells:    cells,
		Boundary: boundary,
	}
}

// Clone 创建网格的深拷贝
func (g *Grid) Clone() *Grid {
	g.mu.RLock()
	defer g.mu.RUnlock()

	newGrid := NewGrid(g.Width, g.Height, g.Boundary)
	for y := range g.Cells {
		copy(newGrid.Cells[y], g.Cells[y])
	}
	return newGrid
}

// CountNeighbors 计算指定位置的邻居数量
func CountNeighbors(grid *Grid, x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			nx, ny := grid.handleBoundary(x+dx, y+dy)
			if grid.Get(nx, ny) {
				count++
			}
		}
	}
	return count
}

// handleBoundary 处理边界条件
func (g *Grid) handleBoundary(x, y int) (int, int) {
	switch g.Boundary {
	case Periodic:
		return (x + g.Width) % g.Width, (y + g.Height) % g.Height
	case Fixed:
		if x < 0 || x >= g.Width || y < 0 || y >= g.Height {
			return -1, -1
		}
		return x, y
	case Reflective:
		return reflectIndex(x, g.Width), reflectIndex(y, g.Height)
	default:
		return (x + g.Width) % g.Width, (y + g.Height) % g.Height
	}
}

// reflectIndex 反射边界计算
func reflectIndex(idx, size int) int {
	if idx < 0 {
		return -idx - 1
	}
	if idx >= size {
		return 2*size - idx - 1
	}
	return idx
}
