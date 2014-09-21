package htm

type Edge struct {
	Start, End, Mid int
}

func (e Edge) Empty() bool {
	return e.Start == 0 && e.End == 0 && e.Mid == 0
}

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
