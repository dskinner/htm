package htm

import (
	"errors"
	"fmt"

	"azul3d.org/gfx.v1"
	"azul3d.org/lmath.v1"
)

type HTM struct {
	Vertices []lmath.Vec3
	Indices  [][]int
	Nodes    map[string]int
}

func NewHTM() *HTM {
	return &HTM{
		Vertices: []lmath.Vec3{
			{0, 0, 1},
			{1, 0, 0},
			{0, 1, 0},
			{-1, 0, 0},
			{0, -1, 0},
			{0, 0, -1},
		},
		Indices: [][]int{
			{1, 5, 2}, // S0
			{2, 5, 3}, // S1
			{3, 5, 4}, // S2
			{4, 5, 1}, // S3

			{1, 0, 4}, // N0
			{4, 0, 3}, // N1
			{3, 0, 2}, // N2
			{2, 0, 1}, // N3
		},
		Nodes: map[string]int{
			"S0": 0,
			"S1": 1,
			"S2": 2,
			"S3": 3,
			"N0": 4,
			"N1": 5,
			"N2": 6,
			"N3": 7,
		},
	}
}

func (h *HTM) LookupNode(x string) ([]int, error) {
	if i, ok := h.Nodes[x]; ok {
		return h.Indices[i], nil
	}
	return nil, errors.New(fmt.Sprintf("Node name '%s' does not exist.", x))
}

func (h *HTM) ConnectMidpoints(i0, i1 int) lmath.Vec3 {
	v0, v1 := h.Vertices[i0], h.Vertices[i1]
	w0, _ := v0.Add(v1).Normalized()
	return w0
}

func (h *HTM) IterLevels(x int) error {
	for i := 2; i <= x; i++ {
		for k := range h.Nodes {
			if len(k) == i {
				err := h.NextLevel(k)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (h *HTM) NextLevel(x string) error {
	n, err := h.LookupNode(x)
	if err != nil {
		return err
	}

	v0, v1, v2 := n[0], n[1], n[2]

	w0 := h.ConnectMidpoints(v1, v2)
	w1 := h.ConnectMidpoints(v0, v2)
	w2 := h.ConnectMidpoints(v0, v1)

	h.Vertices = append(h.Vertices, w0, w1, w2)

	l := len(h.Vertices)
	ind := [][]int{
		{v0, l - 1, l - 2},    // v0, w2, w1
		{v1, l - 3, l - 1},    // v1, w0, w2
		{v2, l - 2, l - 3},    // v2, w1, w0
		{l - 3, l - 2, l - 1}, // w0, w1, w2
	}

	h.Nodes[x+"0"] = len(h.Indices)
	h.Nodes[x+"1"] = len(h.Indices) + 1
	h.Nodes[x+"2"] = len(h.Indices) + 2
	h.Nodes[x+"3"] = len(h.Indices) + 3

	h.Indices = append(h.Indices, ind...)

	return nil
}

func (h *HTM) ConvertVertices() []gfx.Vec3 {
	var vertices []gfx.Vec3
	for _, a := range h.Vertices {
		vertices = append(vertices, gfx.ConvertVec3(a))
	}
	return vertices
}

func (h *HTM) ConvertIndices() []uint32 {
	var indices []uint32
	for _, a := range h.Indices {
		for _, b := range a {
			indices = append(indices, uint32(b))
		}
	}
	return indices
}

func (h *HTM) NewObject() *gfx.Object {
	o := gfx.NewObject()
	m := gfx.NewMesh()
	m.Vertices = h.ConvertVertices()
	m.Indices = h.ConvertIndices()
	o.Meshes = []*gfx.Mesh{m}
	return o
}

func (h *HTM) PPrint() {
	for _, v := range h.Vertices {
		fmt.Println(v)
	}
	for _, n := range h.Indices {
		fmt.Println(n)
	}
}
