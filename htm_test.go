package htm

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"sort"
	"testing"

	"azul3d.org/lmath.v1"

	"github.com/sixthgear/noise"
)

func TestImageL9Intersect(t *testing.T) {
	h := New()
	h.SubDivide(9)

	sq := 640
	size := image.Pt(sq, sq)

	m0 := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(m0, m0.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	m1 := Image(h, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{uint8(z * 255), 0, 0, 255})
	})

	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.75}
	m2 := ImageConstraint(h, cn, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{0, uint8(z * 255), 0, 0})
	})

	MergeImages(m0, m1, m2)

	if err := WriteImage("samples/test.htm.L9.constraint.001.0.75.png", m0); err != nil {
		t.Fatal(err)
	}
}

func TestImageL9Noise(t *testing.T) {
	h := New()
	h.SubDivide(9)

	for i, v0 := range h.Vertices {
		offset := noise.OctaveNoise3d(v0.X, v0.Y, v0.Z, 5, 0.8, 1.3) + 1
		h.Vertices[i] = v0.Add(v0.MulScalar(offset))
	}

	z := 640
	size := image.Pt(z, z)

	m0 := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(m0, m0.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	m1 := Image(h, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{uint8(z * 255), 0, 0, 255})
	})

	cn := &Constraint{lmath.Vec3{0, 0, 1}, 0.5}
	m2 := ImageConstraint(h, cn, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{25, uint8(z * 110), 0, 40})
	})

	MergeImages(m0, m1, m2)

	if err := WriteImage("samples/test.htm.L9.noise.png", m0); err != nil {
		t.Fatal(err)
	}
}

func testL11Info(t *testing.T) {
	h := New()
	h.SubDivide(11)
	t.Logf("\n\n   Trees: %v\nVertices:  %v\n Indices: %v\n   Edges: %v\n\n", len(h.Trees), len(h.Vertices), len(h.Indices()), len(h.Edges.slice))
}

func TestNew(t *testing.T) {
	h := New()
	if len(h.Trees) != 8 {
		t.Fatal("Trees not initialized correctly.")
	}
	if len(h.Vertices) == 0 {
		t.Fatal("HTM vertices not initialized.")
	}
}

type set struct {
	Data []float64
}

func (s *set) put(x float64) {
	x = math.Abs(x)
	for _, v := range s.Data {
		if lmath.Equal(v, x) {
			return
		}
	}
	s.Data = append(s.Data, x)
}

func TestUnique(t *testing.T) {
	h := New()
	h.SubDivide(5)
	s := &set{}
	for _, v := range h.Vertices {
		s.put(v.X)
		s.put(v.Y)
		s.put(v.Z)
	}
	sort.Float64s(s.Data)
	t.Log("Unique")
	for _, v := range s.Data {
		t.Logf("%v", v)
	}
	t.Logf("Unique: %v\n", len(s.Data))
}

// TODO(d) this test is adapted from an older test and incomplete. Needs to account
// for the full tree in a table driven test and also account for edges.
func TestSubDivide2(t *testing.T) {
	h := New()
	h.SubDivide(2)
	if len(h.Vertices) != 18 {
		t.Fatalf("Expected 18 vertices but got %v.", len(h.Vertices))
	}

	cmp := func(a float64, b string) bool { return fmt.Sprintf("%.3f", a) == b }
	check := func(msg string, expects [3]string, v lmath.Vec3) {
		if !cmp(v.X, expects[0]) {
			t.Fatal(msg, "failed for x, expected", expects[0], "but have", v.X)
		}
		if !cmp(v.Y, expects[1]) {
			t.Fatal(msg, "failed for y, expected", expects[1], "but have", v.Y)
		}
		if !cmp(v.Z, expects[2]) {
			t.Fatal(msg, "failed for z, expected", expects[2], "but have", v.Z)
		}
	}

	check("first subdivision", [3]string{"0.000", "0.707", "-0.707"}, h.Vertices[6])
	check("second subdivision", [3]string{"0.707", "0.707", "0.000"}, h.Vertices[7])
	check("third subdivision", [3]string{"0.707", "0.000", "-0.707"}, h.Vertices[8])
}

func TestIndices(t *testing.T) {
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

var benchImage *image.RGBA

func BenchmarkImage(b *testing.B) {
	b.StopTimer()
	h := New()
	h.SubDivide(7)
	b.StartTimer()
	for n := 0; n < b.N; n++ {
		benchImage = Image(h, image.Pt(640, 640), func(m *image.RGBA, x, y int, z float64) { m.Set(x, y, color.RGBA{uint8(z * 255), 0, 0, 255}) })
	}
}
