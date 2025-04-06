package grid

import (
	"errors"
)

// 边界类型枚举
type BoundaryType int

const (
	Periodic BoundaryType = iota
	Fixed
	Reflective
)

// 边界处理器接口
type BoundaryHandler interface {
	Handle(i, j int, sizeX, sizeY int) (int, int, error)
}

// 周期性边界处理
type periodicBoundary struct{}

func (b *periodicBoundary) Handle(i, j, sizeX, sizeY int) (int, int, error) {
	return (i + sizeX) % sizeX, (j + sizeY) % sizeY, nil
}

// 固定值边界处理
type fixedBoundary struct {
	value float64
}

func (b *fixedBoundary) Handle(_, _ int, _, _ int) (int, int, error) {
	return -1, -1, errors.New("fixed boundary condition")
}

// 反射边界处理
type reflectiveBoundary struct{}

func (b *reflectiveBoundary) Handle(i, j, sizeX, sizeY int) (int, int, error) {
	i = reflectIndex(i, sizeX)
	j = reflectIndex(j, sizeY)
	return i, j, nil
}

func reflectIndex(idx, size int) int {
	if idx < 0 {
		return -idx - 1
	}
	if idx >= size {
		return 2*size - idx - 1
	}
	return idx
}

// 边界工厂方法
func NewBoundaryHandler(bType BoundaryType, params ...interface{}) BoundaryHandler {
	switch bType {
	case Periodic:
		return &periodicBoundary{}
	case Fixed:
		value := 0.0
		if len(params) > 0 {
			if v, ok := params[0].(float64); ok {
				value = v
			}
		}
		return &fixedBoundary{value: value}
	case Reflective:
		return &reflectiveBoundary{}
	default:
		return &periodicBoundary{}
	}
}
