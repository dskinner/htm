package htm

type Edge struct {
	Start, End, Mid int
}

func (e Edge) Empty() bool {
	return e.Start == 0 && e.End == 0 && e.Mid == 0
}

// TODO(d) This would grow out of control with each subdivision if topology was being maintained during
// subdivision. Normaly a Match(start, end) is going to receive a vertex indice and then use that to lookup
// any current midpoints. Given (0, 1), and subdividing, there would then be (0, 2) due to the new vertex and
// (0, 1) would need to be ?deleted/replaced/if-exists? along with (1, 0) ?is-that-being-stored-im-brain-dead-2am?
// and then also store (1, 2) so that at-most six references are only ever held given that only triangle polygons
// are supported.
//
// Part of what Match does as-is though is return the vertex indice if a mid point has already been calculated.
// This can occur when a neighboring face has already been subdivided. If that neighbor undergoes multiple subdivisions,
// that does not invalidate the current face and creates breaks in the surface of the mesh (visually unpleasant).
//
// I can't imagine a case for subdividing a mesh where I would want it to look like shit so it might be ideal to force
// neighboring faces to subdivide a minimum number of times to maintain continuity. ltree may help with this.
type Edges struct {
	slice     []Edge
	bootstrap [9444]Edge // memory to hold first slice; Helps avoid allocation for L5 and below.
}

func (ed *Edges) Match(start, end int) (int, int, bool) {
	if start < end {
		start, end = end, start
	}
	ed.grow(start)
	offset := start * 6
	for i, x := range ed.slice[offset : offset+6] {
		if x.Empty() {
			ed.slice[offset+i].Start = start
			ed.slice[offset+i].End = end
			return 0, offset + i, false
		} else if end == x.End {
			return x.Mid, offset + i, true
		}
	}
	panic("fail")
}

func (ed *Edges) grow(n int) {
	n = n*6 + 6

	if n > cap(ed.slice) {
		var slice []Edge
		if ed.slice == nil && n <= len(ed.bootstrap) {
			slice = ed.bootstrap[0:]
		} else {
			slice = make([]Edge, n*2)
			copy(slice, ed.slice)
		}
		ed.slice = slice
	}
}
