package htm

import (
	// "fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"azul3d.org/lmath.v1"

	"github.com/sixthgear/noise"
)

func RenderNoise(h *HTM) {
	for i, v0 := range h.Vertices {
		//v, _ := v0.Normalized()
		offset := noise.OctaveNoise3d(v0.X, v0.Y, v0.Z, 1, 0.9, 1.5) + 1
		// offset /= 10
		h.Vertices[i] = v0.Add(v0.MulScalar(offset)).MulScalar(0.4)
	}
}

func Image(h *HTM, size int) *image.RGBA {
	r := image.Rect(0, 0, size, size)
	m := image.NewRGBA(r)

	var max float64
	for _, t := range h.Trees {
		v0, v1, v2 := h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
		if v0.Z > max {
			max = v0.Z
		}
		if v1.Z > max {
			max = v1.Z
		}
		if v2.Z > max {
			max = v2.Z
		}
	}
	max = (max + 1) / 2
	dt := 255 / max

	// for x := 0; x < size; x++ {
	// 	for y := 0; y < size; y++ {
	// 		m.Set(x, y, color.RGBA{255, 255, 255, 255})
	// 	}
	// }

	fn := func(v0 lmath.Vec3) {
		// if v0.Z < 0 {
		// 	return
		// }
		x := int((v0.X + 1) / 2 * float64(size))
		y := int((v0.Y + 1) / 2 * float64(size))
		// z := (v0.Z + 1) / 2 * 255
		m.Set(int(x), int(y), color.RGBA{uint8((v0.Z + 1) / 2 * dt), 0, 0, 255})
	}

	for _, t := range h.Trees {
		v0, v1, v2 := h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
		fn(v0)
		fn(v1)
		fn(v2)
	}

	return m
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

func Iter(h *HTM, pos int) <-chan int {
	ch := make(chan int)
	go func() {
		iter(h, pos, ch)
		close(ch)
	}()
	return ch
}

func TestImageL9Intersect(t *testing.T) {
	h := New()
	h.SubDivide(9)

	size := 640
	m := Image(h, size)
	clr := func(v0 lmath.Vec3) {
		if v0.Z < 0 {
			return
		}
		x := int((v0.X + 1) / 2 * float64(size))
		y := int((v0.Y + 1) / 2 * float64(size))
		z := (v0.Z + 1) / 2 * 255
		c := m.At(int(x), int(y))
		r, _, _, _ := c.RGBA()
		m.Set(int(x), int(y), color.RGBA{uint8(r), uint8(z), 0, 255})
	}

	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.5}
	for _, idx := range h.Intersections(cn) {
		for p := range Iter(h, idx) {
			t := h.Trees[p]
			v0, v1, v2 := h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
			clr(v0)
			clr(v1)
			clr(v2)
		}
	}

	if m == nil {
		t.Fatal("received nil image")
	}
	out, err := os.Create("test.htm.L9.constraint.001.0.5.png")
	if err != nil {
		t.Fatal("Failed to create new file heightmap_image.png")
	}
	defer out.Close()
	if err := png.Encode(out, m); err != nil {
		t.Fatal("Failed to encode htm.png")
	}
}

func TestImageL9Noise(t *testing.T) {
	h := New()
	h.SubDivide(9)

	RenderNoise(h)

	size := 640
	m := Image(h, size)

	_ = color.Black

	clr := func(v0 lmath.Vec3) {
		if v0.Z < 0 {
			return
		}
		x := int((v0.X + 1) / 2 * float64(size))
		y := int((v0.Y + 1) / 2 * float64(size))
		z := (v0.Z + 1) / 2 * 255
		c := m.At(int(x), int(y))
		r, _, _, _ := c.RGBA()
		m.Set(int(x), int(y), color.RGBA{uint8(r), uint8(z), 0, 255})
	}

	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.5}
	for pos := range h.Intersections(cn) {
		for p := range Iter(h, pos) {
			t := h.Trees[p]
			v0, v1, v2 := h.Vertices[t.Indices[0]], h.Vertices[t.Indices[1]], h.Vertices[t.Indices[2]]
			clr(v0)
			clr(v1)
			clr(v2)
		}
	}

	if m == nil {
		t.Fatal("received nil image")
	}
	out, err := os.Create("htm.png")
	if err != nil {
		t.Fatal("Failed to create new file heightmap_image.png")
	}
	defer out.Close()
	if err := png.Encode(out, m); err != nil {
		t.Fatal("Failed to encode htm.png")
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

func TestHTMSubDivide2(t *testing.T) {
	h := New()
	h.SubDivide(2)
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
