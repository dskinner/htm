// Package ltree provides parallelizable functions for a linear quad tree.
package ltree

import "math"

// Undilate deinterleaves word using shift-or algorithm.
func Undilate(x uint64) uint64 {
	x = (x | (x >> 1)) & 0x33333333
	x = (x | (x >> 2)) & 0x0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF
	return (x & 0x0000FFFF)
}

// Decode retrieves column major position and level from word.
func Decode(key uint64) (x, y, level uint64) {
	x = Undilate((key >> 4) & 0x05555555)
	y = Undilate((key >> 5) & 0x55555555)
	level = key & 0xF
	return
}

// Children generates nodes from a quadtree encoded word.
func Children(key uint64) (uint64, uint64, uint64, uint64) {
	key = ((key + 1) & 0xF) | ((key & 0xFFFFFFF0) << 2)
	return key, key | 0x10, key | 0x20, key | 0x30
}

// Parent generates node from quadtree encoded word.
func Parent(key uint64) uint64 {
	return ((key - 1) & 0xF) | ((key >> 2) & 0x3FFFFFF0)
}

// IsUpperLeft determines if node represents the upper-left child of its parent.
func IsUpperLeft(key uint64) bool {
	return ((key & 0x30) == 0x00)
}

// IsUpperRight determines if node represents the upper-right child of its parent.
func IsUpperRight(key uint64) bool {
	return ((key & 0x30) == 0x10)
}

// IsLowerLeft determines if node represents the lower-left child of its parent.
func IsLowerLeft(key uint64) bool {
	return ((key & 0x30) == 0x20)
}

// IsLowerRight determines if node represents the lower-right child of its parent.
func IsLowerRight(key uint64) bool {
	return ((key & 0x30) == 0x30)
}

// Cell retrieves normalized coordinates and size.
func Cell(key uint64) (nx, ny, size float64) {
	x, y, level := Decode(key)
	size = 1 / float64(uint64(1<<level))
	nx = float64(x) * size
	ny = float64(y) * size
	return
}

// Cap calculates the required capacity to hold all nodes of a given level.
func Cap(lvl int) int {
	return int(math.Pow(4, float64(lvl)))
}

// Split recursively collects children at the given level into nodes pointer.
func Split(key uint64, lvl int, nodes *[]uint64) {
	if key&0xF == uint64(lvl) {
		*nodes = append(*nodes, key)
	} else {
		a, b, c, d := Children(key)
		Split(a, lvl, nodes)
		Split(b, lvl, nodes)
		Split(c, lvl, nodes)
		Split(d, lvl, nodes)
	}
}

// ProjectMercator converts normalized coordinates to mercator projection, just for fun.
func ProjectMercator(nx, ny float64, radius float64) (x, y, z float64) {
	nx = math.Pi / 4 * (2*nx - 1)
	ny = math.Pi / 4 * (4*ny + 1)
	x = radius * math.Cos(ny) * math.Cos(nx)
	y = radius * math.Cos(ny) * math.Sin(nx)
	z = radius * math.Sin(ny)
	return
}
