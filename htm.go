// Package htm implements a hierarchical triangular mesh suitable for graphic display and querying
// as defined here: http://www.noao.edu/noao/staff/yao/sdss_papers/kunszt.pdf
package htm

import (
	"errors"
	"fmt"

	"azul3d.org/lmath.v1"
)

// HTM defines the initial octahedron and allows subdivision nodes.
type HTM struct {
	Vertices *[]lmath.Vec3

	S0, S1, S2, S3 *Tree
	N0, N1, N2, N3 *Tree

	edges *Edges
}

// New returns an HTM initialized with an initial octahedron.
func New() *HTM {
	verts := []lmath.Vec3{
		{0, 0, 1},
		{1, 0, 0},
		{0, 1, 0},
		{-1, 0, 0},
		{0, -1, 0},
		{0, 0, -1},
	}
	edges := NewEdges()

	// fn := func(e *Edge) {
	// 	if em := edges.match(e); em != nil {
	// 		return em
	// 	} else {
	// 		edges.insert(e)
	// 		return e
	// 	}
	// }

	return &HTM{
		Vertices: &verts,

		S0: NewTree(1, &verts, 1, 5, 2, edges),
		S1: NewTree(1, &verts, 2, 5, 3, edges),
		S2: NewTree(1, &verts, 3, 5, 4, edges),
		S3: NewTree(1, &verts, 4, 5, 1, edges),
		N0: NewTree(1, &verts, 1, 0, 4, edges),
		N1: NewTree(1, &verts, 4, 0, 3, edges),
		N2: NewTree(1, &verts, 3, 0, 2, edges),
		N3: NewTree(1, &verts, 2, 0, 1, edges),

		edges: edges,
	}
}

// SubDivide starts a recursion along all root nodes.
func (h *HTM) SubDivide(level int) {
	h.S0.SubDivide(level)
	h.S1.SubDivide(level)
	h.S2.SubDivide(level)
	h.S3.SubDivide(level)
	h.N0.SubDivide(level)
	h.N1.SubDivide(level)
	h.N2.SubDivide(level)
	h.N3.SubDivide(level)
}

func (h *HTM) SubDivide2(level int) {
	h.S0.SubDivide2(level)
	h.S1.SubDivide2(level)
	h.S2.SubDivide2(level)
	h.S3.SubDivide2(level)
	h.N0.SubDivide2(level)
	h.N1.SubDivide2(level)
	h.N2.SubDivide2(level)
	h.N3.SubDivide2(level)
}

// Indices returns a flattened slice of all indices suitable for vertex lookup in native opengl calls.
func (h *HTM) Indices() []uint32 {
	var indices []uint32
	h.S0.CollectIndices(&indices)
	h.S1.CollectIndices(&indices)
	h.S2.CollectIndices(&indices)
	h.S3.CollectIndices(&indices)
	h.N0.CollectIndices(&indices)
	h.N1.CollectIndices(&indices)
	h.N2.CollectIndices(&indices)
	h.N3.CollectIndices(&indices)
	return indices
}

func (h *HTM) TexCoords() []float32 {
	return TexCoords(*h.Vertices)
}

// LookupByCart looks up which triangle a given object belongs to by it's given cartesian coordinates.
func (h *HTM) LookupByCart(v lmath.Vec3) (*Tree, error) {
	ch := make(chan *Tree)
	go func() {
		h.S0.LookupByCart(v, ch)
		h.S1.LookupByCart(v, ch)
		h.S2.LookupByCart(v, ch)
		h.S3.LookupByCart(v, ch)
		h.N0.LookupByCart(v, ch)
		h.N1.LookupByCart(v, ch)
		h.N2.LookupByCart(v, ch)
		h.N3.LookupByCart(v, ch)
		close(ch)
	}()
	t, ok := <-ch
	if ok {
		return t, nil
	}
	return nil, errors.New(fmt.Sprintf("Failed to lookup triangle by given cartesian coordinates: %v", v))
}

func (h *HTM) Intersections(cn *Constraint) <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		h.S0.Intersections(cn, ch)
		h.S1.Intersections(cn, ch)
		h.S2.Intersections(cn, ch)
		h.S3.Intersections(cn, ch)
		h.N0.Intersections(cn, ch)
		h.N1.Intersections(cn, ch)
		h.N2.Intersections(cn, ch)
		h.N3.Intersections(cn, ch)
		close(ch)
	}()
	return ch
}

func (h *HTM) Iter() <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		h.S0.Iter(ch)
		h.S1.Iter(ch)
		h.S2.Iter(ch)
		h.S3.Iter(ch)
		h.N0.Iter(ch)
		h.N1.Iter(ch)
		h.N2.Iter(ch)
		h.N3.Iter(ch)
		close(ch)
	}()
	return ch
}
