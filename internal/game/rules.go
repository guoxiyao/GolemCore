package game

//游戏规则

// applyRules 应用Conway生命游戏规则
// 根据当前细胞状态和邻居数量，返回新的细胞状态
// 如果当前细胞存活，且邻居数量为2或3，则返回true，否则返回false 存活细胞
// 如果当前细胞死亡，且邻居数量为3，则返回true，否则返回false 复活细胞
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
