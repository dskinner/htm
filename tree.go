package htm

import (
	"azul3d.org/lmath.v1"
)

// Tree represents a node in an HTM struct that can contain indices and be subdivided.
type Tree struct {
	// Name  string
	Level int

	indices  [3]int
	vertices *[]lmath.Vec3
	edges    *Edges

	T0, T1, T2, T3 *Tree
}

// NewTree returns an initialized node by the given name with the given index values.
func NewTree(level int, verts *[]lmath.Vec3, i0, i1, i2 int, edges *Edges) *Tree {
	return &Tree{
		Level:    level,
		indices:  [3]int{i0, i1, i2},
		vertices: verts,
		edges:    edges,
	}
}

func (t *Tree) I0() int { return t.indices[0] }
func (t *Tree) I1() int { return t.indices[1] }
func (t *Tree) I2() int { return t.indices[2] }

func (t *Tree) V0() lmath.Vec3 { return (*t.vertices)[t.indices[0]] }
func (t *Tree) V1() lmath.Vec3 { return (*t.vertices)[t.indices[1]] }
func (t *Tree) V2() lmath.Vec3 { return (*t.vertices)[t.indices[2]] }

type Edges struct {
	E []*Edge
}

func NewEdges() *Edges {
	e := make([]*Edge, 500000)
	return &Edges{e}
}

func (es *Edges) insert(e *Edge) {
	j := e.Start * 6
	for i := 0; i < 6; i, j = i+1, j+1 {
		if es.E[j] == nil {
			es.E[j] = e
			return
		}
	}
}

func (es *Edges) match(e *Edge) *Edge {
	for i := e.Start * 6; es.E[i] != nil; i++ {
		if e.End == es.E[i].End {
			return es.E[i]
		}
	}
	return nil
}

func (es *Edges) grow(e *Edge) {
	n := e.Start*6 + 6
	if cap(es.E) < n {
		eb := make([]*Edge, n)
		copy(eb, es.E)
		es.E = eb
	}
}

type Edge struct {
	Start, End, Mid int
}

func NewEdge(s, e, m int) *Edge {
	if s < e {
		s, e = e, s
	}
	return &Edge{s, e, m}
}

func NewAutoEdge(a, b int, t *Tree) *Edge {
	e := NewEdge(a, b, -1)
	t.edges.grow(e)
	if em := t.edges.match(e); em != nil {
		return em
	} else {
		t.edges.insert(e)
		return e
	}
}

// SubDivide calculates the midpoints of the node's triangle and produces four derivative triangles.
func (t *Tree) SubDivide(level int) {
	if t.Level >= level {
		return
	}

	if t.T0 != nil {
		t.T0.SubDivide(level)
		t.T1.SubDivide(level)
		t.T2.SubDivide(level)
		t.T3.SubDivide(level)
		return
	}

	i0, i1, i2 := t.indices[0], t.indices[1], t.indices[2]
	v0, v1, v2 := (*t.vertices)[i0], (*t.vertices)[i1], (*t.vertices)[i2]

	w0, _ := v1.Add(v2).Normalized()
	w1, _ := v0.Add(v2).Normalized()
	w2, _ := v0.Add(v1).Normalized()

	*t.vertices = append(*t.vertices, w0, w1, w2)

	l := len(*t.vertices)

	t.T0 = NewTree(t.Level+1, t.vertices, i0, l-1, l-2, nil)  // v0, w2, w1
	t.T1 = NewTree(t.Level+1, t.vertices, i1, l-3, l-1, nil)  // v1, w0, w2
	t.T2 = NewTree(t.Level+1, t.vertices, i2, l-2, l-3, nil)  // v2, w1, w0
	t.T3 = NewTree(t.Level+1, t.vertices, l-3, l-2, l-1, nil) // w0, w1, w2

	t.T0.SubDivide(level)
	t.T1.SubDivide(level)
	t.T2.SubDivide(level)
	t.T3.SubDivide(level)
}

func (t *Tree) SubDivide2(level int) {
	if t.Level >= level {
		return
	}

	if t.T0 != nil {
		t.T0.SubDivide2(level)
		t.T1.SubDivide2(level)
		t.T2.SubDivide2(level)
		t.T3.SubDivide2(level)
		return
	}

	i0, i1, i2 := t.indices[0], t.indices[1], t.indices[2]
	v0, v1, v2 := (*t.vertices)[i0], (*t.vertices)[i1], (*t.vertices)[i2]

	e0 := NewAutoEdge(i1, i2, t)

	if e0.Mid == -1 {
		w0, _ := v1.Add(v2).Normalized()
		*t.vertices = append(*t.vertices, w0)
		e0.Mid = len(*t.vertices) - 1
	}

	e1 := NewAutoEdge(i0, i2, t)

	if e1.Mid == -1 {
		w1, _ := v0.Add(v2).Normalized()
		*t.vertices = append(*t.vertices, w1)
		e1.Mid = len(*t.vertices) - 1
	}

	e2 := NewAutoEdge(i0, i1, t)

	if e2.Mid == -1 {
		w2, _ := v0.Add(v1).Normalized()
		*t.vertices = append(*t.vertices, w2)
		e2.Mid = len(*t.vertices) - 1
	}

	t.T0 = NewTree(t.Level+1, t.vertices, i0, e2.Mid, e1.Mid, t.edges)     // v0, w2, w1
	t.T1 = NewTree(t.Level+1, t.vertices, i1, e0.Mid, e2.Mid, t.edges)     // v1, w0, w2
	t.T2 = NewTree(t.Level+1, t.vertices, i2, e1.Mid, e0.Mid, t.edges)     // v2, w1, w0
	t.T3 = NewTree(t.Level+1, t.vertices, e0.Mid, e1.Mid, e2.Mid, t.edges) // w0, w1, w2

	t.T0.SubDivide2(level)
	t.T1.SubDivide2(level)
	t.T2.SubDivide2(level)
	t.T3.SubDivide2(level)
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

func (t *Tree) PointInside(v lmath.Vec3) bool {
	i0, i1, i2 := t.indices[0], t.indices[1], t.indices[2]
	v0, v1, v2 := (*t.vertices)[i0], (*t.vertices)[i1], (*t.vertices)[i2]
	a := v0.Cross(v1).Dot(v)
	b := v1.Cross(v2).Dot(v)
	c := v2.Cross(v0).Dot(v)
	return a > 0 && b > 0 && c > 0
}

func (t *Tree) LookupByCart(v lmath.Vec3, ch chan *Tree) {
	if !t.PointInside(v) {
		return
	}
	if t.T0 == nil {
		ch <- t
	} else {
		t.T0.LookupByCart(v, ch)
		t.T1.LookupByCart(v, ch)
		t.T2.LookupByCart(v, ch)
		t.T3.LookupByCart(v, ch)
	}
}

func (t *Tree) Intersections(cn *Constraint, ch chan *Tree) {
	switch cn.Test(t) {
	case Inside:
		ch <- t
	case Partial:
		if t.T0 == nil {
			ch <- t
		} else {
			t.T0.Intersections(cn, ch)
			t.T1.Intersections(cn, ch)
			t.T2.Intersections(cn, ch)
			t.T3.Intersections(cn, ch)
		}
	case Outside:
		if t.T0 != nil {
			t.T0.Intersections(cn, ch)
			t.T1.Intersections(cn, ch)
			t.T2.Intersections(cn, ch)
			t.T3.Intersections(cn, ch)
		}
	}
}

func (t *Tree) Iter(ch chan *Tree) {
	if t.T0 == nil {
		ch <- t
		return
	}
	t.T0.Iter(ch)
	t.T1.Iter(ch)
	t.T2.Iter(ch)
	t.T3.Iter(ch)
}

func (t *Tree) IterAll(ch chan *Tree) {
	if t.T0 == nil {
		return
	}
	ch <- t.T0
	ch <- t.T1
	ch <- t.T2
	ch <- t.T3
	t.T0.IterAll(ch)
	t.T1.IterAll(ch)
	t.T2.IterAll(ch)
	t.T3.IterAll(ch)
}
