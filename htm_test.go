package htm

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sort"
	"testing"

	"azul3d.org/lmath.v1"

	"github.com/sixthgear/noise"
)

func RenderNoise(h *HTM) {
	for i, v0 := range h.Vertices {
		offset := noise.OctaveNoise3d(v0.X, v0.Y, v0.Z, 10, 0.8, 1.8) + 1
		h.Vertices[i] = v0.Add(v0.MulScalar(offset))
	}
}

var benchImage *image.RGBA

func BenchmarkImage(b *testing.B) {
	b.StopTimer()
	h := New()
	h.SubDivide(7)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		benchImage = Image(h, image.Pt(640, 640))
	}
}

func norm(x float64, max float64) float64 {
	return (x + max) / (max * 2)
}

func TestNorm(t *testing.T) {
	for i := 1; i < 10; i++ {
		n := norm(0, float64(i))
		fmt.Println(i, n)
		if n != 0.5 {
			t.Fatalf("%v != 0.5", n)
		}
	}
}

type Option func(*image.RGBA)

func Background(c color.RGBA) Option {
	return func(m *image.RGBA) {
		size := m.Bounds().Size()
		for x := 0; x < size.X; x++ {
			for y := 0; y < size.Y; y++ {
				m.Set(x, y, c)
			}
		}
	}
}

type Vec3Slice []lmath.Vec3

func (p Vec3Slice) Len() int           { return len(p) }
func (p Vec3Slice) Less(i, j int) bool { return p[i].Z < p[j].Z }
func (p Vec3Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

func SortVec3s(p Vec3Slice) { sort.Sort(p) }

func Image(h *HTM, pt image.Point, options ...Option) *image.RGBA {
	r := image.Rect(0, 0, pt.X, pt.Y)
	m := image.NewRGBA(r)

	for _, opt := range options {
		opt(m)
	}

	max := h.Max()
	p := append([]lmath.Vec3(nil), h.Vertices...)
	SortVec3s(p)
	for _, v0 := range p {
		x := int(norm(v0.X, max) * float64(pt.X))
		y := int(norm(v0.Y, max) * float64(pt.Y))
		z := uint8(norm(v0.Z, max) * 255)
		m.Set(x, y, color.RGBA{z, 0, 0, 255})
	}

	return m
}

func WriteImage(fn string, m *image.RGBA) error {
	if m == nil {
		return errors.New("received nil image")
	}
	out, err := os.Create(fn)
	if err != nil {
		return fmt.Errorf("Failed to create new file %s: %s", fn, err)
	}
	defer out.Close()
	if err := png.Encode(out, m); err != nil {
		return fmt.Errorf("Failed to encode %s: %s", fn, err)
	}
	return nil
}

// func TestNewTree(t *testing.T) {
// 	tree := NewTree("S0", nil, 0, 1, 2)
// 	if tree.Name != "S0" {
// 		t.Fatal("Tree name not initialized.")
// 	}
// 	if len(tree.indices) != 3 {
// 		t.Fatal("Tree indices not of correct length.")
// 	}
// 	if tree.indices[0] != 0 && tree.indices[1] != 1 && tree.indices[2] != 2 {
// 		t.Fatal("Tree indicies not initialized.")
// 	}
// }

// func TestTreeSubDivide(t *testing.T) {
// 	verts := []lmath.Vec3{
// 		{1, 0, 0},
// 		{0, 1, 0},
// 		{0, 0, 1},
// 	}
// 	l1 := len(verts)
// 	tree := NewTree(1, &verts, 0, 1, 2)
// 	tree.SubDivide(2)
// 	if len(verts) == l1 {
// 		t.Fatal("Vertices not updated.")
// 	}
// 	if len(verts) != 6 {
// 		t.Fatal("Vertices not of correct length.")
// 	}

// 	cmp := func(a float64, b string) bool { return fmt.Sprintf("%.3f", a) == b }

// 	if !cmp(verts[3].X, "0.000") || !cmp(verts[3].Y, "0.707") || !cmp(verts[3].Z, "0.707") {
// 		t.Fatal("First subdivision of tree not correct.")
// 	}
// 	if !cmp(verts[4].X, "0.707") || !cmp(verts[4].Y, "0.000") || !cmp(verts[4].Z, "0.707") {
// 		t.Fatal("Second subdivision of tree not correct.")
// 	}
// 	if !cmp(verts[5].X, "0.707") || !cmp(verts[5].Y, "0.707") || !cmp(verts[5].Z, "0.000") {
// 		t.Fatal("Third subdivision of tree not correct.")
// 	}
// 	if tree.indices[0] != 0 || tree.indices[1] != 1 || tree.indices[2] != 2 {
// 		t.Fatal("Tree indices not initialized.")
// 	}
// }

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

func TestImageL9Intersect(t *testing.T) {
	h := New()
	h.SubDivide(9)

	size := 640
	max := h.Max()
	m := Image(h, image.Pt(size, size))

	clr := func(v0 lmath.Vec3) {
		x := int(norm(v0.X, max) * float64(size))
		y := int(norm(v0.Y, max) * float64(size))
		z := norm(v0.Z, max) * 255
		c := m.At(x, y)
		r, _, _, _ := c.RGBA()
		m.Set(x, y, color.RGBA{uint8(r), uint8(z), 100, 255})
	}

	var sl []lmath.Vec3
	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.75}
	for p := range Iter(h, h.Intersections(cn)...) {
		v0, v1, v2 := h.VerticesAt(p)
		sl = append(sl, v0, v1, v2)
	}
	SortVec3s(sl)
	for _, v0 := range sl {
		clr(v0)
	}

	if err := WriteImage("test.htm.L9.constraint.001.0.5.png", m); err != nil {
		t.Fatal(err)
	}
}

func TestImageL9Noise(t *testing.T) {
	h := New()
	h.SubDivide(9)

	RenderNoise(h)

	size := 560
	m := Image(h, image.Pt(size, size), Background(color.RGBA{0, 0, 0, 255}))

	max := h.Max()

	clr := func(v0 lmath.Vec3) {
		x := int(norm(v0.X, max) * float64(size))
		y := int(norm(v0.Y, max) * float64(size))
		z := norm(v0.Z, max) * 255
		c := m.At(x, y)
		r, _, _, _ := c.RGBA()
		m.Set(x, y, color.RGBA{uint8(r), uint8(z), 0, 255})
	}

	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.5}

	var sl []lmath.Vec3
	for p := range Iter(h, h.Intersections(cn)...) {
		v0, v1, v2 := h.VerticesAt(p)
		sl = append(sl, v0, v1, v2)
	}
	SortVec3s(sl)
	for _, v0 := range sl {
		clr(v0)
	}

	if err := WriteImage("htm.png", m); err != nil {
		t.Fatal(err)
	}
}

func testFoo(t *testing.T) {
	h := New()
	h.SubDivide(11)
	t.Logf("\n\n   Trees: %v\nVertices:  %v\n Indices: %v\n   Edges: %v\n\n", len(h.Trees), len(h.Vertices), len(h.Indices()), len(h.Edges.slice))
}

func TestNewHTM(t *testing.T) {
	h := New()
	if len(h.Trees) != 8 {
		t.Fatal("Trees not initialized correctly.")
	}
	if len(h.Vertices) == 0 {
		t.Fatal("HTM vertices not initialized.")
	}
}

type Set struct {
	Data []float64
}

func (s *Set) Put(x float64) {
	x = math.Abs(x)
	for _, v := range s.Data {
		if lmath.Equal(v, x) {
			return
		}
	}
	s.Data = append(s.Data, x)
}

func TestHTMSubDivide2(t *testing.T) {
	h := New()
	h.SubDivide(7)
	s := &Set{}
	for _, v := range h.Vertices {
		// t.Logf("%+v\n", v)
		s.Put(v.X)
		s.Put(v.Y)
		s.Put(v.Z)
	}
	sort.Float64s(s.Data)
	t.Log("Unique")
	for _, v := range s.Data {
		t.Logf("%v", v)
	}
	t.Logf("Unique: %v\n", len(s.Data))
	if len(h.Vertices) != 18 {
		t.Fatalf("Expected 18 vertices but got %v.", len(h.Vertices))
	}
}

func TestHTMIndices(t *testing.T) {
	h := New()
	h.SubDivide(2)
	n := h.Indices()
	if len(n) != 96 {
		t.Fatalf("Expected 96 indices but got %v.", len(n))
	}
}

func TestTexCoords(t *testing.T) {
	h := New()
	h.SubDivide(2)
	tc := h.TexCoords()
	if (len(tc) % 2) != 0 {
		t.Fatal("Uneven UV mapping.")
	}
}

func TestLookupByCart(t *testing.T) {
	h := New()
	h.SubDivide(7)
	_, err := h.LookupByCart(lmath.Vec3{0.9, 0.1, 0.1})
	if err != nil {
		t.Fatal(err)
	}
}

func TestNoDups(t *testing.T) {
	h := New()
	h.SubDivide(5)
	for _, v1 := range h.Vertices {
		x := 0
		for _, v2 := range h.Vertices {
			if v1.Equals(v2) {
				if x == 0 {
					x++ // allow for one dup, this one
				} else {
					t.Fail()
				}
			}
		}
	}
}

func BenchmarkL5(b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := New()
		h.SubDivide(5)
	}
}

func BenchmarkL7(b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := New()
		h.SubDivide(7)
	}
}

func BenchmarkL9(b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := New()
		h.SubDivide(9)
	}
}

func BenchmarkL11(b *testing.B) {
	for n := 0; n < b.N; n++ {
		h := New()
		h.SubDivide(11)
	}
}

func BenchmarkLookupByCartL7(b *testing.B) {
	b.StopTimer()
	h := New()
	h.SubDivide(7)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_, err := h.LookupByCart(lmath.Vec3{0.9, 0.1, 0.1})
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkConstraintIterL7(b *testing.B) {
	b.StopTimer()
	h := New()
	h.SubDivide(7)
	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.5}
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		for _, t := range h.Intersections(cn) {
			_ = t
		}
	}
}

func BenchmarkIndices(b *testing.B) {
	b.StopTimer()
	h := New()
	h.SubDivide(7)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		_ = h.Indices()
	}
}
