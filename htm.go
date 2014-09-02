package htm

import (
	"azul3d.org/lmath.v1"
)

type Tree struct {
	Name     string
	Indices  [3]int
	Vertices *[]lmath.Vec3

	T0 *Tree
	T1 *Tree
	T2 *Tree
	T3 *Tree
}

func NewTree(name string, verts *[]lmath.Vec3, i0, i1, i2 int) *Tree {
	return &Tree{
		Name:     name,
		Indices:  [3]int{i0, i1, i2},
		Vertices: verts,
	}
}

func (t *Tree) SubDivide(level int) {
	if len(t.Name) > level {
		return
	}

	i0, i1, i2 := t.Indices[0], t.Indices[1], t.Indices[2]
	v0, v1, v2 := (*t.Vertices)[i0], (*t.Vertices)[i1], (*t.Vertices)[i2]

	w0, _ := v1.Add(v2).Normalized()
	w1, _ := v0.Add(v2).Normalized()
	w2, _ := v0.Add(v1).Normalized()

	*t.Vertices = append(*t.Vertices, w0, w1, w2)

	l := len(*t.Vertices)

	t.T0 = NewTree(t.Name+"0", t.Vertices, i0, l-1, l-2)  // v0, w2, w1
	t.T1 = NewTree(t.Name+"1", t.Vertices, i1, l-3, l-1)  // v1, w0, w2
	t.T2 = NewTree(t.Name+"2", t.Vertices, i2, l-2, l-3)  // v2, w1, w0
	t.T3 = NewTree(t.Name+"3", t.Vertices, l-3, l-2, l-1) // w0, w1, w2

	t.T0.SubDivide(level)
	t.T1.SubDivide(level)
	t.T2.SubDivide(level)
	t.T3.SubDivide(level)
}

func (t *Tree) CollectIndices(indices *[]uint32) {
	if t.T0 == nil {
		*indices = append(*indices, uint32(t.Indices[0]), uint32(t.Indices[1]), uint32(t.Indices[2]))
	} else {
		t.T0.CollectIndices(indices)
		t.T1.CollectIndices(indices)
		t.T2.CollectIndices(indices)
		t.T3.CollectIndices(indices)
	}
}

type HTM struct {
	Vertices *[]lmath.Vec3

	S0, S1, S2, S3 *Tree
	N0, N1, N2, N3 *Tree
}

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
