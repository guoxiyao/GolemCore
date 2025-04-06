// E:\go\GolemCore\internal\game\grid.go
package game

type Grid struct {
	Width  int
	Height int
	Cells  [][]bool
}

// 线程安全访问方法
func (g *Grid) Get(x, y int) bool {
	return g.Cells[y][x]
}

func (g *Grid) Set(x, y int, state bool) {
	g.Cells[y][x] = state
}

func NewGrid(width, height int) *Grid {
	cells := make([][]bool, height)
	for i := range cells {
		cells[i] = make([]bool, width)
	}
	return &Grid{
		Width:  width,
		Height: height,
		Cells:  cells,
	}
}
