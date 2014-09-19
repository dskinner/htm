// Package htm implements a hierarchical triangular mesh suitable for graphic display and querying
// as defined here: http://www.noao.edu/noao/staff/yao/sdss_papers/kunszt.pdf
package htm

import (
	"errors"
	"fmt"
	"image"
	"image/color"

	"azul3d.org/lmath.v1"
)

type Tree struct {
	Index    int
	Level    int
	Indices  [3]int
	Children [4]int
}

type HTM struct {
	*Edges

	Vertices []lmath.Vec3
	Trees    []Tree
}

func New() *HTM {
	return &HTM{
		Edges: &Edges{},
		Vertices: []lmath.Vec3{
			{0, 0, 1},
			{1, 0, 0},
			{0, 1, 0},
			{-1, 0, 0},
			{0, -1, 0},
			{0, 0, -1},
		},
		Trees: []Tree{
			{Index: 0, Level: 1, Indices: [3]int{1, 5, 2}}, // S0
			{Index: 1, Level: 1, Indices: [3]int{2, 5, 3}}, // S1
			{Index: 2, Level: 1, Indices: [3]int{3, 5, 4}}, // S2
			{Index: 3, Level: 1, Indices: [3]int{4, 5, 1}}, // S3
			{Index: 4, Level: 1, Indices: [3]int{1, 0, 4}}, // N0
			{Index: 5, Level: 1, Indices: [3]int{4, 0, 3}}, // N1
			{Index: 6, Level: 1, Indices: [3]int{3, 0, 2}}, // N2
			{Index: 7, Level: 1, Indices: [3]int{2, 0, 1}}, // N3
		},
	}
}

// Indices returns a flattened slice of all indices suitable for vertex lookup in native opengl calls.
func (h *HTM) Indices() []uint32 {
	var indices []uint32
	CollectIndices(h, 0, &indices)
	CollectIndices(h, 1, &indices)
	CollectIndices(h, 2, &indices)
	CollectIndices(h, 3, &indices)
	CollectIndices(h, 4, &indices)
	CollectIndices(h, 5, &indices)
	CollectIndices(h, 6, &indices)
	CollectIndices(h, 7, &indices)
	return indices
}

func (h *HTM) TexCoords() []float32 {
	return TexCoords(h.Vertices)
}

func (h *HTM) SubDivide(level int) {
	SubDivide(h, 0, level)
	SubDivide(h, 1, level)
	SubDivide(h, 2, level)
	SubDivide(h, 3, level)
	SubDivide(h, 4, level)
	SubDivide(h, 5, level)
	SubDivide(h, 6, level)
	SubDivide(h, 7, level)
}

// LookupByCart looks up which triangle a given object belongs to by it's given cartesian coordinates.
func (h *HTM) LookupByCart(v lmath.Vec3) (Tree, error) {
	ch := make(chan int)
	go func() {
		LookupByCart(h, 0, v, ch)
		LookupByCart(h, 1, v, ch)
		LookupByCart(h, 2, v, ch)
		LookupByCart(h, 3, v, ch)
		LookupByCart(h, 4, v, ch)
		LookupByCart(h, 5, v, ch)
		LookupByCart(h, 6, v, ch)
		LookupByCart(h, 7, v, ch)
		close(ch)
	}()
	pos, ok := <-ch
	if ok {
		return h.Trees[pos], nil
	}
	return Tree{}, errors.New(fmt.Sprintf("Failed to lookup triangle by given cartesian coordinates: %v", v))
}

func (h *HTM) Intersections(cn *Constraint) <-chan int {
	ch := make(chan int)
	go func() {
		Intersections(h, 0, cn, ch)
		Intersections(h, 1, cn, ch)
		Intersections(h, 2, cn, ch)
		Intersections(h, 3, cn, ch)
		Intersections(h, 4, cn, ch)
		Intersections(h, 5, cn, ch)
		Intersections(h, 6, cn, ch)
		Intersections(h, 7, cn, ch)
		close(ch)
	}()
	return ch
}

func Intersections(h *HTM, pos int, cn *Constraint, ch chan int) {
	t := h.Trees[pos]
	i0, i1, i2 := t.Indices[0], t.Indices[1], t.Indices[2]
	v0, v1, v2 := h.Vertices[i0], h.Vertices[i1], h.Vertices[i2]

	switch cn.Test(v0, v1, v2) {
	case Inside:
		ch <- pos
	case Partial:
		if t.Children[0] == 0 {
			ch <- pos
		} else {
			Intersections(h, t.Children[0], cn, ch)
			Intersections(h, t.Children[1], cn, ch)
			Intersections(h, t.Children[2], cn, ch)
			Intersections(h, t.Children[3], cn, ch)
		}
	case Outside:
		if t.Children[0] != 0 {
			Intersections(h, t.Children[0], cn, ch)
			Intersections(h, t.Children[1], cn, ch)
			Intersections(h, t.Children[2], cn, ch)
			Intersections(h, t.Children[3], cn, ch)
		}
	}
}

func (h *HTM) Image(size int) image.Image {
	r := image.Rect(0, 0, size, size)
	m := image.NewRGBA(r)

	for x := 0; x < size; x++ {
		for y := 0; y < size; y++ {
			m.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}

	fn := func(v0 lmath.Vec3) {
		if v0.Z < 0 {
			return
		}
		x := int((v0.X + 1) / 2 * float64(size))
		y := int((v0.Y + 1) / 2 * float64(size))
		z := (v0.Z + 1) / 2 * 255
		m.Set(int(x), int(y), color.RGBA{uint8(z), 0, 0, 255})
	}

	for _, t := range h.Trees {
		v0, v1, v2 := h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
		fn(v0)
		fn(v1)
		fn(v2)
	}

	return m
}

func PointInside(h *HTM, pos int, v lmath.Vec3) bool {
	t := h.Trees[pos]
	i0, i1, i2 := t.Indices[0], t.Indices[1], t.Indices[2]
	v0, v1, v2 := h.Vertices[i0], h.Vertices[i1], h.Vertices[i2]
	a := v0.Cross(v1).Dot(v)
	b := v1.Cross(v2).Dot(v)
	c := v2.Cross(v0).Dot(v)
	return a > 0 && b > 0 && c > 0
}

func LookupByCart(h *HTM, pos int, v lmath.Vec3, ch chan int) {
	if !PointInside(h, pos, v) {
		return
	}

	t := h.Trees[pos]
	if t.Children[0] == 0 {
		ch <- pos
	} else {
		LookupByCart(h, t.Children[0], v, ch)
		LookupByCart(h, t.Children[1], v, ch)
		LookupByCart(h, t.Children[2], v, ch)
		LookupByCart(h, t.Children[3], v, ch)
	}
}

// CollectIndices appends the current node's indices to the slice pointer unless it should recurse.
func CollectIndices(h *HTM, pos int, indices *[]uint32) {
	t := h.Trees[pos]

	if t.Children[0] == 0 {
		*indices = append(*indices, uint32(t.Indices[0]), uint32(t.Indices[1]), uint32(t.Indices[2]))
	} else {
		CollectIndices(h, t.Children[0], indices)
		CollectIndices(h, t.Children[1], indices)
		CollectIndices(h, t.Children[2], indices)
		CollectIndices(h, t.Children[3], indices)
	}
}

func SubDivide(h *HTM, pos int, level int) {
	t := h.Trees[pos]

	if t.Level >= level {
		return
	}

	if t.Children[0] != 0 {
		SubDivide(h, t.Children[0], level)
		SubDivide(h, t.Children[1], level)
		SubDivide(h, t.Children[2], level)
		SubDivide(h, t.Children[3], level)
		return
	}

	i0, i1, i2 := t.Indices[0], t.Indices[1], t.Indices[2]
	v0, v1, v2 := h.Vertices[i0], h.Vertices[i1], h.Vertices[i2]

	e0, ok := h.Edges.New(i1, i2)
	if !ok {
		w0, _ := v1.Add(v2).Normalized()
		h.Vertices = append(h.Vertices, w0)
		e0.Mid = len(h.Vertices) - 1
		h.Edges.Insert(e0)
	}

	e1, ok := h.Edges.New(i0, i2)
	if !ok {
		w1, _ := v0.Add(v2).Normalized()
		h.Vertices = append(h.Vertices, w1)
		e1.Mid = len(h.Vertices) - 1
		h.Edges.Insert(e1)
	}

	e2, ok := h.Edges.New(i0, i1)
	if !ok {
		w2, _ := v0.Add(v1).Normalized()
		h.Vertices = append(h.Vertices, w2)
		e2.Mid = len(h.Vertices) - 1
		h.Edges.Insert(e2)
	}

	i := len(h.Trees)

	h.Trees = append(h.Trees,
		Tree{Index: i, Level: t.Level + 1, Indices: [3]int{i0, e2.Mid, e1.Mid}},         // v0, w2, w1
		Tree{Index: i + 1, Level: t.Level + 1, Indices: [3]int{i1, e0.Mid, e2.Mid}},     // v1, w0, w2
		Tree{Index: i + 2, Level: t.Level + 1, Indices: [3]int{i2, e1.Mid, e0.Mid}},     // v2, w1, w0
		Tree{Index: i + 3, Level: t.Level + 1, Indices: [3]int{e0.Mid, e1.Mid, e2.Mid}}) // w0, w1, w2

	t.Children = [4]int{i, i + 1, i + 2, i + 3}
	h.Trees[pos] = t

	SubDivide(h, t.Children[0], level)
	SubDivide(h, t.Children[1], level)
	SubDivide(h, t.Children[2], level)
	SubDivide(h, t.Children[3], level)
}
