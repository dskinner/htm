package ltree

import (
	"fmt"
	"math"
	"testing"
)

func printWord(t *testing.T, pad string, title string, key uint32) {
	level, p := Decode(key)
	norm, size := Cell(key)
	t.Logf("%s%s:\n", pad, title)
	t.Logf("%s  parent: %b\n", pad, Parent(key))
	t.Logf("%s  node: %b\n", pad, key)
	t.Logf("%s  level: %v\n", pad, level)
	t.Logf("%s  pos: %+v\n", pad, p)
	t.Logf("%s  norm: %+v\n", pad, norm)
	t.Logf("%s  size: %v\n", pad, size)
	t.Logf("%s  UL: %v\n", pad, IsUpperLeft(key))
}

func recurse(t *testing.T, parent uint32) {
	for i, child := range Children(parent) {
		plvl, _ := Decode(parent)
		clvl, _ := Decode(child)
		if plvl+1 != clvl {
			t.Fail()
		}
		if Parent(child) != parent {
			t.Fail()
		}
		if IsUpperLeft(child) && i != 0 {
			t.Fail()
		}
		if clvl < 10 {
			recurse(t, child)
		}
	}
}

func recurseAppend(parent uint32, nodes *[]uint32, lvl int, maxLvl int) {
	clvl := lvl + 1
	for _, child := range Children(parent) {
		if clvl < maxLvl {
			recurseAppend(child, nodes, clvl, maxLvl)
		} else {
			*nodes = append(*nodes, child)
		}
	}
}

func TestTree(t *testing.T) {
	var root uint32 = 0

	rlvl, _ := Decode(root)
	if rlvl != 0 {
		t.Fail()
	}

	recurse(t, root)
}

func TestLength(t *testing.T) {
	maxLvl := 10
	maxLen := int(math.Pow(4, float64(maxLvl)))

	var nodes []uint32
	recurseAppend(0, &nodes, 0, maxLvl)
	if len(nodes) != maxLen {
		t.Fatalf("Expected %v but got %v\n", maxLen, len(nodes))
	}
}

func TestPrint(t *testing.T) {
	var x uint32 = 0
	// 1110 0010
	// x |= 1 << 1
	// x |= 1 << 5
	// x |= 1 << 6
	// x |= 1 << 7

	for i, a := range Children(x) {
		printWord(t, "  ", fmt.Sprintf("x i %v", i), a)
		for j, b := range Children(a) {
			printWord(t, "    ", fmt.Sprintf("x i,j %v,%v", i, j), b)
		}
	}
}
