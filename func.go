package htm

import "github.com/azul3d/engine/lmath"

// Intersections returns a slice of node indexes that are inside completely or partially.
//
// TODO(d) iter children to pass on, possibly via a separate method or make docs more clear.
// Currently, the use of this is assuming that an index returned will not also return its
// children for performance reasons.
func Intersections(h *HTM, idx int, t Tester, mt *[]int) {
	switch t.Test(h.VerticesAt(idx)) {
	case Inside:
		*mt = append(*mt, idx)
	case Partial:
		if h.EmptyAt(idx) {
			*mt = append(*mt, idx)
		} else {
			a, b, c, d := h.ChildrenAt(idx)
			Intersections(h, a, t, mt)
			Intersections(h, b, t, mt)
			Intersections(h, c, t, mt)
			Intersections(h, d, t, mt)
		}
	case Outside:
		if !h.EmptyAt(idx) {
			a, b, c, d := h.ChildrenAt(idx)
			Intersections(h, a, t, mt)
			Intersections(h, b, t, mt)
			Intersections(h, c, t, mt)
			Intersections(h, d, t, mt)
		}
	}
}

// Vec3Inside tests if vector is contained within bounds of triangle.
func Vec3Inside(h *HTM, idx int, v lmath.Vec3) bool {
	v0, v1, v2 := h.VerticesAt(idx)
	a := v0.Cross(v1).Dot(v)
	b := v1.Cross(v2).Dot(v)
	c := v2.Cross(v0).Dot(v)
	return a > 0 && b > 0 && c > 0
}

// LookupByCart recurses nodes by subdivisions and tests if vector is inside
// triangle. Locates the single, smallest subdivision that matches.
func LookupByCart(h *HTM, idx int, v lmath.Vec3, i *int) {
	if h.Trees[idx].Empty() {
		return
	}
	if Vec3Inside(h, idx, v) {
		if h.EmptyAt(idx) {
			*i = idx
		} else {
			a, b, c, d := h.ChildrenAt(idx)
			LookupByCart(h, a, v, i)
			LookupByCart(h, b, v, i)
			LookupByCart(h, c, v, i)
			LookupByCart(h, d, v, i)
		}
	}
}

// TODO(d) this seems really quite arbitrarily simple. Take two vertices, add them together, normalize.
// Since the initial structure is an octahedron, this simply works out to make a sphere, but the process
// seems like it would work fine for arbitrary refinement.
//
// The important part would be my notes on Edges struct, and maintaining continuity with neighbor faces by
// triggering minimum subdivisions there with an easy way to traverse and locate neighbors.
//
// Or automagically updating their indices if such a thing could also keep the master indices that makes its
// way to the gpu up-to-date without having to reiter over edges to regenerate, or is that such a bad thing?
func SubDivide(h *HTM, idx int, level int) {
	if h.LevelAt(idx) >= level {
		return
	}

	if !h.EmptyAt(idx) {
		a, b, c, d := h.ChildrenAt(idx)
		SubDivide(h, a, level)
		SubDivide(h, b, level)
		SubDivide(h, c, level)
		SubDivide(h, d, level)
		return
	}

	// here we get our face to be subdivided
	i0, i1, i2 := h.IndicesAt(idx)
	v0, v1, v2 := h.VerticesAt(idx)

	// check each edge to see if it has already been subdivided due to a neighboring face
	// subdivision that has already performed the calculation.
	getMid := func(i0, i1 int, v0, v1 lmath.Vec3) int {
		eidx, _ := h.Edges.Init(i0, i1)
		if h.Edges.slice[eidx].Mid == 0 {
			w0, _ := v0.Add(v1).Normalized()
			h.Vertices = append(h.Vertices, w0)
			e0 := len(h.Vertices) - 1
			h.Edges.slice[eidx].Mid = e0
		}
		return h.Edges.slice[eidx].Mid
	}

	e0 := getMid(i1, i2, v1, v2)
	e1 := getMid(i0, i2, v0, v2)
	e2 := getMid(i0, i1, v0, v1)

	i := len(h.Trees)
	a, b, c, d := i, i+1, i+2, i+3
	l := h.LevelAt(idx) + 1

	h.Trees = append(h.Trees,
		Tree{Index: a, Level: l, Indices: [3]int{i0, e2, e1}, Parent: idx}, // v0, w2, w1
		Tree{Index: b, Level: l, Indices: [3]int{i1, e0, e2}, Parent: idx}, // v1, w0, w2
		Tree{Index: c, Level: l, Indices: [3]int{i2, e1, e0}, Parent: idx}, // v2, w1, w0
		Tree{Index: d, Level: l, Indices: [3]int{e0, e1, e2}, Parent: idx}) // w0, w1, w2

	h.Trees[idx].Children = [4]int{a, b, c, d}

	h.Edges.Init(e2, e1)
	h.Edges.Init(i0, e1)
	h.Edges.Init(i0, e2)
	h.Edges.Init(e0, e2)
	h.Edges.Init(i1, e2)
	h.Edges.Init(i1, e0)
	h.Edges.Init(e1, e0)
	h.Edges.Init(i2, e0)
	h.Edges.Init(i2, e1)
	h.Edges.Init(e1, e2)
	h.Edges.Init(e0, e2)
	h.Edges.Init(e0, e1)

	SubDivide(h, a, level)
	SubDivide(h, b, level)
	SubDivide(h, c, level)
	SubDivide(h, d, level)
}

func lowestLevel(h *HTM, idx int, lvl *int) {
	if h.Trees[idx].Children[0] == 0 {
		*lvl = h.Trees[idx].Level
	} else {
		lowestLevel(h, h.Trees[idx].Children[0], lvl)
	}
}

func merge(h *HTM, i0, i1 int) {
	if mids, ok := h.Edges.Merge(i0, i1); ok {
		for _, mid := range mids {
			h.Vertices[mid] = lmath.Vec3{0, 0, 0}
		}
	}
}

func CullToLevel(h *HTM, idx int, lvl int) {
	for _, tidx := range IterLevel(h, idx, lvl) {
		t := h.Trees[tidx]
		if t.Level == lvl && t.Children[0] != 0 {
			Cull(h, tidx)
		}
	}
}

func Cull(h *HTM, idx int) {

	var llvl int
	lowestLevel(h, idx, &llvl)

	emptyChildren := func(tr Tree) {
		if tr.Children[0] != 0 {
			h.Trees[tr.Children[0]] = Tree{}
		}
		if tr.Children[1] != 0 {
			h.Trees[tr.Children[1]] = Tree{}
		}
		if tr.Children[2] != 0 {
			h.Trees[tr.Children[2]] = Tree{}
		}
		if tr.Children[3] != 0 {
			h.Trees[tr.Children[3]] = Tree{}
		}
		tr.Children = [4]int{0, 0, 0, 0}
		h.Trees[tr.Index] = tr
	}

	t := h.Trees[idx]
	for curLvl := llvl; curLvl > t.Level; curLvl-- {
		trees := IterLevel(h, t.Index, curLvl)

		for _, tidx := range trees {
			tr := h.Trees[tidx]
			i0, i1, i2 := tr.Indices[0], tr.Indices[1], tr.Indices[2]
			merge(h, i1, i2)
			merge(h, i0, i2)
			merge(h, i0, i1)
			emptyChildren(tr)
		}
	}

	i0, i1, i2 := t.Indices[0], t.Indices[1], t.Indices[2]
	merge(h, i1, i2)
	merge(h, i0, i2)
	merge(h, i0, i1)
	emptyChildren(t)
}

func iter(h *HTM, pos int, ch chan int) {
	t := h.Trees[pos]
	if t.Children[0] == 0 {
		ch <- pos
	} else {
		iter(h, t.Children[0], ch)
		iter(h, t.Children[1], ch)
		iter(h, t.Children[2], ch)
		iter(h, t.Children[3], ch)
	}
}

// Iter accepts a series of node indices and returns a channel that receives node indices
// of the smallest subdivisions.
func Iter(h *HTM, positions ...int) <-chan int {
	ch := make(chan int)
	go func() {
		for _, pos := range positions {
			iter(h, pos, ch)
		}
		close(ch)
	}()
	return ch
}

func iterLevel(h *HTM, idx int, lvl int, indices *[]int) {
	t := h.Trees[idx]
	if t.Level == lvl {
		*indices = append(*indices, idx)
	} else if t.Children[0] != 0 {
		iterLevel(h, t.Children[0], lvl, indices)
		iterLevel(h, t.Children[1], lvl, indices)
		iterLevel(h, t.Children[2], lvl, indices)
		iterLevel(h, t.Children[3], lvl, indices)
	}
}

func IterLevel(h *HTM, idx int, lvl int) []int {
	var indices []int
	iterLevel(h, idx, lvl, &indices)
	return indices
}
