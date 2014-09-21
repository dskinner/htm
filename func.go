package htm

import (
	"azul3d.org/lmath.v1"
)

// Intersections returns a slice of node indexes that are inside completely or partially.
//
// TODO(d) iter children to pass on, possibly via a separate method or make docs more clear.
// Currently, the use of this is assuming that an index returned will not also return its
// children for performance reasons.
func Intersections(h *HTM, idx int, cn *Constraint, mt *[]int) {
	switch cn.Test(h.VerticesAt(idx)) {
	case Inside:
		*mt = append(*mt, idx)
	case Partial:
		if h.EmptyAt(idx) {
			*mt = append(*mt, idx)
		} else {
			a, b, c, d := h.ChildrenAt(idx)
			Intersections(h, a, cn, mt)
			Intersections(h, b, cn, mt)
			Intersections(h, c, cn, mt)
			Intersections(h, d, cn, mt)
		}
	case Outside:
		if !h.EmptyAt(idx) {
			a, b, c, d := h.ChildrenAt(idx)
			Intersections(h, a, cn, mt)
			Intersections(h, b, cn, mt)
			Intersections(h, c, cn, mt)
			Intersections(h, d, cn, mt)
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

	i0, i1, i2 := h.IndicesAt(idx)
	v0, v1, v2 := h.VerticesAt(idx)

	e0, off, ok := h.Edges.Match(i1, i2)
	if !ok {
		w0, _ := v1.Add(v2).Normalized()
		h.Vertices = append(h.Vertices, w0)
		e0 = len(h.Vertices) - 1
		h.Edges.slice[off].Mid = e0
	}

	e1, off, ok := h.Edges.Match(i0, i2)
	if !ok {
		w1, _ := v0.Add(v2).Normalized()
		h.Vertices = append(h.Vertices, w1)
		e1 = len(h.Vertices) - 1
		h.Edges.slice[off].Mid = e1
	}

	e2, off, ok := h.Edges.Match(i0, i1)
	if !ok {
		w2, _ := v0.Add(v1).Normalized()
		h.Vertices = append(h.Vertices, w2)
		e2 = len(h.Vertices) - 1
		h.Edges.slice[off].Mid = e2
	}

	i := len(h.Trees)
	a, b, c, d := i, i+1, i+2, i+3
	l := h.LevelAt(idx) + 1

	h.Trees = append(h.Trees,
		Tree{Index: a, Level: l, Indices: [3]int{i0, e2, e1}}, // v0, w2, w1
		Tree{Index: b, Level: l, Indices: [3]int{i1, e0, e2}}, // v1, w0, w2
		Tree{Index: c, Level: l, Indices: [3]int{i2, e1, e0}}, // v2, w1, w0
		Tree{Index: d, Level: l, Indices: [3]int{e0, e1, e2}}) // w0, w1, w2

	h.Trees[idx].Children = [4]int{a, b, c, d}

	SubDivide(h, a, level)
	SubDivide(h, b, level)
	SubDivide(h, c, level)
	SubDivide(h, d, level)
}
