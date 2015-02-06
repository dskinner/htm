package htm

// Edge holdes indices of three vertices representing a line.
type Edge struct {
	Start, End, Mid int
}

// Empty returns true if edge is zero value.
func (e Edge) Empty() bool {
	return e.Start == 0 && e.End == 0 && e.Mid == 0
}

func (e Edge) Equals(x Edge) bool {
	return e.Start == x.Start && e.End == x.End && e.Mid == x.Mid
}

// Edges is a container for Edge where the end of an edge has at most six neighbors, and so edges
// are stored sorted by their HTM indices in ordered groups of six.
//
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

// Init locates an Edge with given start and end indices, or initializes one otherwise. The order of start and
// end do not matter as long as they represent a line with a potential midpoint in the data structure. If a match is found,
// the midpoint indice of the edge is returned and can be used to look up a vertex in an HTM. The edge index is
// also returned whether matched or initialized. Finally, a bool if the match is sucessful is returned.
func (ed *Edges) Init(start, end int) (int, bool) {
	if start < end {
		start, end = end, start
	}
	ed.grow(start)
	offset := start * 6
	for i, x := range ed.slice[offset : offset+6] {
		idx := offset + i
		if x.Empty() {
			ed.slice[idx].Start = start
			ed.slice[idx].End = end
			return idx, true
		}
		if x.End == end {
			return idx, false
		}
	}
	panic("edge init failure")
}

func (ed *Edges) traceDeleteMid(mid int, mids *[]int) {
	offset := mid * 6
	for _, x := range ed.slice[offset : offset+6] {
		if x.Empty() {
			return
		} else if x.Mid != 0 {
			*mids = append(*mids, x.Mid)
			ed.traceDeleteMid(x.Mid, mids)
			ed.zeroStart(x.Mid)
		}
		// TODO(d) maybe could zero out edge here instead of calling zeroStart above
	}
}

func (ed *Edges) Merge(start, end int) ([]int, bool) {
	if start < end {
		start, end = end, start
	}
	ed.grow(start)
	offset := start * 6
	for i, x := range ed.slice[offset : offset+6] {
		if x.End == end && x.Mid != 0 {
			mid := x.Mid
			x.Mid = 0
			ed.slice[offset+i] = x

			mids := []int{mid}
			ed.traceDeleteMid(mid, &mids)
			ed.zeroStart(mid)
			return mids, true
		}
	}
	return nil, false
}

// zeroStart will replace any edge where Edge.Start == start with empty struct.
func (ed *Edges) zeroStart(start int) {
	ed.grow(start)
	offset := start * 6
	for i := range ed.slice[offset : offset+6] {
		ed.slice[offset+i] = Edge{}
	}
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

func (ed *Edges) nonempty() []Edge {
	var n []Edge
	for _, x := range ed.slice {
		if !x.Empty() {
			n = append(n, x)
		}
	}
	return n
}
