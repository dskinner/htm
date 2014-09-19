package htm

type Edge struct {
	Start, End, Mid int
}

func NewEdge(s, e, m int) Edge {
	if s < e {
		s, e = e, s
	}
	return Edge{s, e, m}
}

func (e Edge) Empty() bool {
	return e.Start == 0 && e.End == 0 && e.Mid == 0
}

type Edges struct {
	slice []Edge
}

func (ed *Edges) New(a, b int) (Edge, bool) {
	e := NewEdge(a, b, -1)
	ed.Grow(e)
	if em := ed.Match(e); !em.Empty() {
		return em, true
	} else {
		return e, false
	}
}

func (ed *Edges) Insert(e Edge) {
	offset := e.Start * 6
	for i, x := range ed.slice[offset : offset+6] {
		if x.Empty() {
			ed.slice[offset+i] = e
			return
		}
	}
}

func (ed *Edges) Match(e Edge) Edge {
	offset := e.Start * 6
	for _, x := range ed.slice[offset : offset+6] {
		if e.End == x.End {
			return x
		}
	}
	return Edge{}
}

func (ed *Edges) Grow(e Edge) {
	n := e.Start*6 + 6
	if cap(ed.slice) < n {
		ed2 := make([]Edge, n*2)
		copy(ed2, ed.slice)
		ed.slice = ed2
	}
}
