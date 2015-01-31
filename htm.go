// Package htm implements a hierarchical triangular mesh suitable for graphic display and querying
// as defined here: http://www.noao.edu/noao/staff/yao/sdss_papers/kunszt.pdf
package htm

import (
	"errors"
	"fmt"

	"azul3d.org/lmath.v1"
)

type Tree struct {
	Index    int
	Level    int
	Indices  [3]int
	Children [4]int
}

type HTM struct {
	*Edges

	Vertices []lmath.Vec3
	Trees    []Tree
}

// New returns an HTM with the first eight nodes that create an octahedron initialized.
func New() *HTM {
	return &HTM{
		Edges: &Edges{},
		Vertices: []lmath.Vec3{
			{0, 0, 1},
			{1, 0, 0},
			{0, 1, 0},
			{-1, 0, 0},
			{0, -1, 0},
			{0, 0, -1},
		},
		Trees: []Tree{
			{Index: 0, Level: 1, Indices: [3]int{1, 5, 2}}, // S0
			{Index: 1, Level: 1, Indices: [3]int{2, 5, 3}}, // S1
			{Index: 2, Level: 1, Indices: [3]int{3, 5, 4}}, // S2
			{Index: 3, Level: 1, Indices: [3]int{4, 5, 1}}, // S3
			{Index: 4, Level: 1, Indices: [3]int{1, 0, 4}}, // N0
			{Index: 5, Level: 1, Indices: [3]int{4, 0, 3}}, // N1
			{Index: 6, Level: 1, Indices: [3]int{3, 0, 2}}, // N2
			{Index: 7, Level: 1, Indices: [3]int{2, 0, 1}}, // N3
		},
	}
}

// Indices returns a slice of all vertex indices of the lowest subdivisions.
func (h *HTM) Indices() []uint32 {
	indices := make([]uint32, 0, len(h.Trees))
	for _, t := range h.Trees {
		if t.Children[0] == 0 {
			indices = append(indices, uint32(t.Indices[0]), uint32(t.Indices[1]), uint32(t.Indices[2]))
		}
	}
	return indices
}

// IndicesAt returns a node's indices.
func (h *HTM) IndicesAt(idx int) (i0, i1, i2 int) {
	return h.Trees[idx].Indices[0], h.Trees[idx].Indices[1], h.Trees[idx].Indices[2]
}

// VerticesAt looks up a node's vertices from its indices.
func (h *HTM) VerticesAt(idx int) (v0, v1, v2 lmath.Vec3) {
	i0, i1, i2 := h.IndicesAt(idx)
	return h.Vertices[i0], h.Vertices[i1], h.Vertices[i2]
}

// VerticesFor looks up a node's vertices by the given node.
func (h *HTM) VerticesFor(t Tree) (v0, v1, v2 lmath.Vec3) {
	return h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
}

// LevelAt returns a node's subdivision level. The eight root nodes are level one.
func (h *HTM) LevelAt(idx int) int { return h.Trees[idx].Level }

// TODO(d) better name
func (h *HTM) EmptyAt(idx int) bool { return h.Trees[idx].Children[0] == 0 }

// TODO(d) better name
func (h *HTM) ChildrenAt(idx int) (a, b, c, d int) {
	return h.Trees[idx].Children[0], h.Trees[idx].Children[1], h.Trees[idx].Children[2], h.Trees[idx].Children[3]
}

// TexCoords is a convenience method.
func (h *HTM) TexCoords() []float32 {
	return TexCoords(h.Vertices)
}

// TexCoordsPlanar is a convenience method.
func (h *HTM) TexCoordsPlanar() []float32 {
	return TexCoordsPlanar(h.Vertices)
}

// SubDivide starts a recursive subdivision along all eight root nodes.
func (h *HTM) SubDivide(level int) {
	SubDivide(h, 0, level)
	SubDivide(h, 1, level)
	SubDivide(h, 2, level)
	SubDivide(h, 3, level)
	SubDivide(h, 4, level)
	SubDivide(h, 5, level)
	SubDivide(h, 6, level)
	SubDivide(h, 7, level)
}

// LookupByCart looks up which triangle a given object belongs to by it's given cartesian coordinates.
func (h *HTM) LookupByCart(v lmath.Vec3) (Tree, error) {
	i := -1

	// Only one of these will recurse within first call.
	LookupByCart(h, 0, v, &i)
	LookupByCart(h, 1, v, &i)
	LookupByCart(h, 2, v, &i)
	LookupByCart(h, 3, v, &i)
	LookupByCart(h, 4, v, &i)
	LookupByCart(h, 5, v, &i)
	LookupByCart(h, 6, v, &i)
	LookupByCart(h, 7, v, &i)

	if i != -1 {
		return h.Trees[i], nil
	}

	return Tree{}, errors.New(fmt.Sprintf("Failed to lookup triangle by given cartesian coordinates: %v", v))
}

func (h *HTM) Intersections(cn *Constraint) []int {
	var mt []int
	Intersections(h, 0, cn, &mt)
	Intersections(h, 1, cn, &mt)
	Intersections(h, 2, cn, &mt)
	Intersections(h, 3, cn, &mt)
	Intersections(h, 4, cn, &mt)
	Intersections(h, 5, cn, &mt)
	Intersections(h, 6, cn, &mt)
	Intersections(h, 7, cn, &mt)
	return mt
}
