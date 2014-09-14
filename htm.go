// package htm implements a hierarchical triangular mesh suitable for graphic display and querying.
package htm

import (
	"errors"
	"fmt"
	"time"

	"azul3d.org/lmath.v1"
)

// Tree represents a node in an HTM struct that can contain indices and be subdivided.
type Tree struct {
	Name string

	indices  [3]int
	vertices *[]lmath.Vec3

	T0, T1, T2, T3 *Tree
}

// NewTree returns an initialized node by the given name with the given index values.
func NewTree(name string, verts *[]lmath.Vec3, i0, i1, i2 int) *Tree {
	return &Tree{
		Name:     name,
		indices:  [3]int{i0, i1, i2},
		vertices: verts,
	}
}

func (t *Tree) V0() lmath.Vec3 { return (*t.vertices)[t.indices[0]] }
func (t *Tree) V1() lmath.Vec3 { return (*t.vertices)[t.indices[1]] }
func (t *Tree) V2() lmath.Vec3 { return (*t.vertices)[t.indices[2]] }

// SubDivide calculates the midpoints of the node's triangle and produces four derivative triangles.
func (t *Tree) SubDivide(level int) {
	if len(t.Name) > level {
		return
	}

	i0, i1, i2 := t.indices[0], t.indices[1], t.indices[2]
	v0, v1, v2 := (*t.vertices)[i0], (*t.vertices)[i1], (*t.vertices)[i2]

	w0, _ := v1.Add(v2).Normalized()
	w1, _ := v0.Add(v2).Normalized()
	w2, _ := v0.Add(v1).Normalized()

	*t.vertices = append(*t.vertices, w0, w1, w2)

	l := len(*t.vertices)

	t.T0 = NewTree(t.Name+"0", t.vertices, i0, l-1, l-2)  // v0, w2, w1
	t.T1 = NewTree(t.Name+"1", t.vertices, i1, l-3, l-1)  // v1, w0, w2
	t.T2 = NewTree(t.Name+"2", t.vertices, i2, l-2, l-3)  // v2, w1, w0
	t.T3 = NewTree(t.Name+"3", t.vertices, l-3, l-2, l-1) // w0, w1, w2

	t.T0.SubDivide(level)
	t.T1.SubDivide(level)
	t.T2.SubDivide(level)
	t.T3.SubDivide(level)
}

// CollectIndices appends the current node's indices to the slice pointer unless it should recurse.
func (t *Tree) CollectIndices(indices *[]uint32) {
	if t.T0 == nil {
		*indices = append(*indices, uint32(t.indices[0]), uint32(t.indices[1]), uint32(t.indices[2]))
	} else {
		t.T0.CollectIndices(indices)
		t.T1.CollectIndices(indices)
		t.T2.CollectIndices(indices)
		t.T3.CollectIndices(indices)
	}
}

// Vertices returns a subset of the HTM's vertices that is not intended for
// use with this tree's indices.
func (t *Tree) Vertices() []lmath.Vec3 {
	var indices []uint32
	t.CollectIndices(&indices)

	var vertices []lmath.Vec3
	for _, i := range indices {
		vertices = append(vertices, (*t.vertices)[i])
	}
	return vertices
}

// HTM defines the initial octahedron and allows subdivision nodes.
type HTM struct {
	Vertices *[]lmath.Vec3

	S0, S1, S2, S3 *Tree
	N0, N1, N2, N3 *Tree
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
	return &HTM{
		Vertices: &verts,

		S0: NewTree("S0", &verts, 1, 5, 2),
		S1: NewTree("S1", &verts, 2, 5, 3),
		S2: NewTree("S2", &verts, 3, 5, 4),
		S3: NewTree("S3", &verts, 4, 5, 1),
		N0: NewTree("N0", &verts, 1, 0, 4),
		N1: NewTree("N1", &verts, 4, 0, 3),
		N2: NewTree("N2", &verts, 3, 0, 2),
		N3: NewTree("N3", &verts, 2, 0, 1),
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
	sch := Walker(v, h.S0, h.S1, h.S2, h.S3)
	nch := Walker(v, h.N0, h.N1, h.N2, h.N3)

	timeout := time.After(1 * time.Second)

	for sch != nil || nch != nil {
		select {
		case t, ok := <-sch:
			if ok {
				return t, nil
			} else {
				sch = nil
			}
		case t, ok := <-nch:
			if ok {
				return t, nil
			} else {
				nch = nil
			}
		case <-timeout:
			return nil, errors.New("Timed out while walking trees.")
		}
	}

	return nil, errors.New(fmt.Sprintf("Failed to lookup triangle by given cartesian coordinates: %v", v))
}

func Walk(t *Tree, v lmath.Vec3, ch chan *Tree) {
	if t == nil {
		panic("nil tree not allowed during walk.")
	}
	if !PointInside(t, v) {
		return
	}
	if t.T0 == nil {
		ch <- t
	} else {
		Walk(t.T0, v, ch)
		Walk(t.T1, v, ch)
		Walk(t.T2, v, ch)
		Walk(t.T3, v, ch)
	}
}

func Walker(v lmath.Vec3, trees ...*Tree) <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		for _, t := range trees {
			Walk(t, v, ch)
		}
		close(ch)
	}()
	return ch
}

func Walk2(t *Tree, ch chan *Tree) {
	if t == nil {
		panic("nil tree not allowed during walk.")
	}
	if t.T0 == nil {
		ch <- t
		return
	}

	// TODO(d) alternate walk that returns all trees
	// ch <- t.T0
	// ch <- t.T1
	// ch <- t.T2
	// ch <- t.T3

	Walk2(t.T0, ch)
	Walk2(t.T1, ch)
	Walk2(t.T2, ch)
	Walk2(t.T3, ch)
}

func Walker2(trees ...*Tree) <-chan *Tree {
	ch := make(chan *Tree)
	go func() {
		for _, t := range trees {
			Walk2(t, ch)
		}
		close(ch)
	}()
	return ch
}

func PointInside(t *Tree, v lmath.Vec3) bool {
	i0, i1, i2 := t.indices[0], t.indices[1], t.indices[2]
	v0, v1, v2 := (*t.vertices)[i0], (*t.vertices)[i1], (*t.vertices)[i2]
	a := v0.Cross(v1).Dot(v)
	b := v1.Cross(v2).Dot(v)
	c := v2.Cross(v0).Dot(v)
	return a > 0 && b > 0 && c > 0
}

type Sign int

const (
	Negative Sign = iota
	Zero
	Positive
	Mixed
)

type Coverage int

const (
	Inside Coverage = iota
	Partial
	Outside
)

// Constraint is a circular area, given by the plane slicing it off the sphere.
type Constraint struct {
	P lmath.Vec3
	D float64
}

func (c *Constraint) Test(t *Tree) Coverage {
	a0 := c.P.Dot(t.V0()) > c.D
	a1 := c.P.Dot(t.V1()) > c.D
	a2 := c.P.Dot(t.V2()) > c.D

	if a0 && a1 && a2 {
		return Inside
	} else if a0 || a1 || a2 {
		return Partial
	} else {
		// TODO(d) P center, LookupByCart needed to determine final fate.
		return Outside
	}
}

// Convex is a combination of constraints (logical AND of constraints).
type Convex []*Constraint

func (c Convex) Test(t *Tree) bool {
	for _, cn := range c {
		if cn.Test(t) == Outside {
			return false
		}
	}
	return true
}

func (c Convex) Sign() Sign {
	return Zero
}

// Domain is several convexes (logical OR of convexes).
type Domain []*Convex

func (d Domain) Test(t *Tree) bool {
	for _, cx := range d {
		if cx.Test(t) {
			return true
		}
	}
	return false
}
