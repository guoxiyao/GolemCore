// E:\go\GolemCore\internal\game\rules.go
package game

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

// countNeighbors 计算周期性边界下的邻居数量
func (u *Universe) countNeighbors(x, y int) int {
	count := 0
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			if dx == 0 && dy == 0 {
				continue
			}

			// 周期性边界计算
			nx := (x + dx + u.width) % u.width
			ny := (y + dy + u.height) % u.height

			if u.cells[ny][nx] {
				count++
			}
		}
	}
	return count
}
