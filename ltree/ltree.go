// Package ltree provides parallelizable functions for a linear quad tree.
package ltree

type Vec2 struct {
	X, Y uint32
}

type Vec2f struct {
	X, Y float32
}

// Undilate deinterleaves word using shift-or algorithm.
func Undilate(x uint32) uint32 {
	x = (x | (x >> 1)) & 0x33333333
	x = (x | (x >> 2)) & 0x0F0F0F0F
	x = (x | (x >> 4)) & 0x00FF00FF
	x = (x | (x >> 8)) & 0x0000FFFF
	return (x & 0x0000FFFF)
}

// Decode retrieves column major position and level from word.
func Decode(key uint32) (level uint32, p Vec2) {
	level = key & 0xF
	p.X = Undilate((key >> 4) & 0x05555555)
	p.Y = Undilate((key >> 5) & 0x55555555)
	return
}

// Children generates nodes from a quadtree encoded word.
func Children(key uint32) [4]uint32 {
	var c [4]uint32
	key = ((key + 1) & 0xF) | ((key & 0xFFFFFFF0) << 2)
	c[0] = key
	c[1] = key | 0x10
	c[2] = key | 0x20
	c[3] = key | 0x30
	return c
}

// Parent generates node from quadtree encoded word.
func Parent(key uint32) uint32 {
	return ((key - 1) & 0xF) | ((key >> 2) & 0x3FFFFFF0)
}

// IsUpperLeft determines if node represents the upper-left child of its parent.
func IsUpperLeft(key uint32) bool {
	return ((key & 0x30) == 0x00)
}

// Cell retrieves normalized coordinates and size.
func Cell(key uint32) (p Vec2f, size float32) {
	level, pos := Decode(key)
	size = 1 / float32(uint32(1<<level))
	p.X = float32(pos.X) * size
	p.Y = float32(pos.Y) * size
	return
}
