// E:\go\GolemCore\internal\game\universe.go
package game

import (
	"sync"
)

// Universe 管理元胞宇宙状态
type Universe struct {
	mu     sync.RWMutex // 读写锁保证并发安全
	width  int          // 网格宽度
	height int          // 网格高度
	cells  [][]bool     // 当前状态网格
	buffer [][]bool     // 双缓冲临时网格
	epoch  int          // 当前演化代数
}

// NewUniverse 创建新的宇宙实例
func NewUniverse(width, height int) *Universe {
	cells := make([][]bool, height)
	buffer := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
		buffer[i] = make([]bool, width)
	}
	return &Universe{
		width:  width,
		height: height,
		cells:  cells,
		buffer: buffer,
	}
}

// NextGeneration 切换到下一代状态
func (u *Universe) NextGeneration() {
	u.mu.Lock()
	defer u.mu.Unlock()
	u.cells, u.buffer = u.buffer, u.cells // 交换缓冲区
	u.epoch++
}

// GetCells 获取当前状态快照（线程安全）
func (u *Universe) GetCells() [][]bool {
	u.mu.RLock()
	defer u.mu.RUnlock()

	snapshot := make([][]bool, u.height)
	for i := range u.cells {
		snapshot[i] = make([]bool, u.width)
		copy(snapshot[i], u.cells[i])
	}
	return snapshot
}

// SetCell 安全设置细胞状态
func (u *Universe) SetCell(x, y int, state bool) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if x >= 0 && x < u.width && y >= 0 && y < u.height {
		u.cells[y][x] = state
	}
}
