package htm

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"path"
	"sort"
	"strings"
	"testing"

	"azul3d.org/lmath.v1"

	"github.com/sixthgear/noise"
)

func colorRed(m *image.RGBA, x, y int, z float64) {
	m.Set(x, y, color.RGBA{uint8(z * 255), 0, 0, 255})
}

func newRGBA(size image.Point, color color.RGBA) *image.RGBA {
	m := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(m, m.Bounds(), &image.Uniform{color}, image.ZP, draw.Src)
	return m
}

func saveImage(t *testing.T, h *HTM, name string) {
	size := image.Pt(640, 640)
	m := newRGBA(size, color.RGBA{0, 0, 0, 255})
	MergeImages(m, Image(h, size, colorRed))
	if err := WriteImage(path.Join("samples", name), m); err != nil {
		t.Log(err)
	}
}

func compareVec3s(a, b []lmath.Vec3) error {
	if len(a) != len(b) {
		return fmt.Errorf("lengths don't match %v != %v\n", len(a), len(b))
	}
	var errs []string
	for i, v := range a {
		if !v.Equals(b[i]) {
			errs = append(errs, fmt.Sprintf("value at %v expected %+v but have %+v.", i, v, b[i]))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func compareEdges(a, b []Edge) error {
	if len(a) != len(b) {
		return fmt.Errorf("lengths don't match %v != %v\n", len(a), len(b))
	}
	var errs []string
	for i, v := range a {
		if !v.Equals(b[i]) {
			errs = append(errs, fmt.Sprintf("value at %v expected %+v but have %+v.", i, v, b[i]))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func compareTrees(a, b []Tree) error {
	if len(a) != len(b) {
		return fmt.Errorf("lengths don't match %v != %v\n", len(a), len(b))
	}
	var errs []string
	for i, v := range a {
		if !v.Equals(b[i]) {
			errs = append(errs, fmt.Sprintf("value at %v expected %+v but have %+v.", i, v, b[i]))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func compareUint32s(a, b []uint32) error {
	if len(a) != len(b) {
		return fmt.Errorf("lengths don't match %v != %v\n", len(a), len(b))
	}
	var errs []string
	for i, v := range a {
		if v != b[i] {
			errs = append(errs, fmt.Sprintf("value at %v expected %v but have %v.", i, v, b[i]))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf(strings.Join(errs, "\n"))
	}
	return nil
}

func compareHTMs(a, b *HTM) error {
	if err := compareEdges(a.Edges.nonempty(), b.Edges.nonempty()); err != nil {
		fmt.Println("### A ###")
		for i, edge := range a.Edges.slice {
			if !edge.Empty() {
				fmt.Printf("%v: %+v\n", i, edge)
			}
		}
		fmt.Println("### B ###")
		for i, edge := range b.Edges.slice {
			if !edge.Empty() {
				fmt.Printf("%v: %+v\n", i, edge)
			}
		}

		fmt.Println(fmt.Errorf("edges %s", err))
	}

	if err := compareVec3s(a.VerticesNotEmpty(), b.VerticesNotEmpty()); err != nil {
		fmt.Println("### A ###")
		for i, vec := range a.Vertices {
			if !vec.Equals(lmath.Vec3Zero) {
				fmt.Printf("%v: %+v\n", i, vec)
			}
		}
		fmt.Println("### B ###")
		for i, vec := range b.Vertices {
			if !vec.Equals(lmath.Vec3Zero) {
				fmt.Printf("%v: %+v\n", i, vec)
			}
		}

		return fmt.Errorf("vertices %s", err)
	}

	if err := compareTrees(a.TreesNotEmpty(), b.TreesNotEmpty()); err != nil {
		return fmt.Errorf("trees %s", err)
	}

	if err := compareUint32s(a.Indices(), b.Indices()); err != nil {
		return fmt.Errorf("indices %s", err)
	}

	return nil
}

func validateHTM(h *HTM) error {
	var errs []string

	// no duplicate edges
	for i0, e0 := range h.Edges.slice {
		if e0.Empty() {
			continue
		}
		for i1, e1 := range h.Edges.slice {
			if e1.Empty() {
				continue
			}
			if i0 != i1 && e0.Start == e1.Start && e0.End == e1.End {
				errs = append(errs, fmt.Sprintf("duplicate edge at %v and %v: %+v", i0, i1, e0))
			}
		}
	}

	// no duplicate vertices
	for i0, v0 := range h.Vertices {
		if v0.Equals(lmath.Vec3Zero) {
			continue
		}
		for i1, v1 := range h.Vertices {
			if v1.Equals(lmath.Vec3Zero) {
				continue
			}
			if i0 != i1 && v0.Equals(v1) {
				errs = append(errs, fmt.Sprintf("duplicate vertex at %v and %v", i0, i1))
			}
		}
	}

	// all indices and vertices accounted for
	indices := make(map[int]bool)
	for _, index := range h.Indices() {
		indices[int(index)] = false
	}
	for i, v := range h.Vertices {
		if v.Equals(lmath.Vec3Zero) {
			continue
		}
		if _, ok := indices[i]; !ok {
			errs = append(errs, fmt.Sprintf("vertex %+v not accounted for in indices at %v", v, i))
		} else {
			indices[i] = true
		}
	}
	for i, ok := range indices {
		if !ok {
			var tree Tree
			for _, tr := range h.Trees {
				if tr.Indices[0] == i || tr.Indices[1] == i || tr.Indices[2] == i {
					tree = tr
					break
				}
			}
			errs = append(errs, fmt.Sprintf("index %v does not point to a valid vertex, from tree %+v", i, tree))
		}
	}

	// all indices and edge indices account for
	indices = make(map[int]bool)
	for _, index := range h.Indices() {
		indices[int(index)] = false
	}
	for i, e := range h.Edges.slice {
		if e.Empty() {
			continue
		}
		if _, ok := indices[e.Start]; !ok {
			errs = append(errs, fmt.Sprintf("edge.Start for %+v not accounted for in indices at %v", e, i))
		} else {
			indices[e.Start] = true
		}
		if _, ok := indices[e.End]; !ok {
			errs = append(errs, fmt.Sprintf("edge.End for %+v not accounted for in indices at %v", e, i))
		} else {
			indices[e.End] = true
		}
		if _, ok := indices[e.Mid]; !ok {
			errs = append(errs, fmt.Sprintf("edge.Mid for %+v not accounted for in indices at %v", e, i))
		} else {
			indices[e.Mid] = true
		}
	}
	for i, ok := range indices {
		if !ok {
			errs = append(errs, fmt.Sprintf("index %v (%+v) is not referenced in any edge", i, h.Vertices[i]))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("\n" + strings.Join(errs, "\n"))
	}

	return nil
}

func TestConstraintSubDivisionCullSubDivision(t *testing.T) {
	l0, l1 := 1, 2
	h0, h1 := New(), New()
	defer saveImage(t, h1, "test.constraint.subdivision.cull.subdivision.png")
	h0.SubDivide(l0)
	h1.SubDivide(l0)

	vec := lmath.Vec3{0.001, 0.001, 0.999}
	tr, err := h1.LookupByCart(vec)
	if err != nil {
		t.Fatal(err)
	}

	SubDivide(h0, tr.Index, l1)
	SubDivide(h1, tr.Index, l1)

	h1.CullToLevel(l0)
	h1.Compact()
	tr, err = h1.LookupByCart(vec)
	if err != nil {
		t.Fatal(err)
	}
	SubDivide(h1, tr.Index, l1)

	if err := compareHTMs(h0, h1); err != nil {
		t.Fatal(err)
	}
}

func TestRepeatedCull(t *testing.T) {
	lfrom, lto := 2, 5
	h0, h1 := New(), New()
	defer saveImage(t, h1, "test.repeated.cull.png")
	h0.SubDivide(lfrom)
	h1.SubDivide(lfrom)

	tr, err := h1.LookupByCart(lmath.Vec3{0.001, 0.999, 0.001})
	if err != nil {
		t.Fatal(err)
	}

	SubDivide(h1, tr.Index, lto)
	Cull(h1, tr.Index)
	SubDivide(h1, tr.Index, lto)
	Cull(h1, tr.Index)

	if err := validateHTM(h1); err != nil {
		t.Fatal(err)
	}

	if err := compareHTMs(h0, h1); err != nil {
		t.Fatal(err)
	}
}

func TestConstraintSubDivision(t *testing.T) {
	l0, l1 := 1, 3
	h := New()
	defer saveImage(t, h, "test.constraint.subdivision.png")
	h.SubDivide(l0)

	cn := &Constraint{lmath.Vec3{0.001, 0.001, 0.999}, 0.85}
	for idx := range h.Intersections(cn) {
		SubDivide(h, idx, l1)
	}

	if err := validateHTM(h); err != nil {
		t.Fatal(err)
	}
}

func TestConvexCull(t *testing.T) {
	l0, l1 := 3, 5
	h := New()
	defer saveImage(t, h, "test.convex.cull.png")
	h.SubDivide(l0)

	d := 0.85
	pos := lmath.Vec3{0.001, 0.001, 0.999}
	lastPos := lmath.Vec3{0.001, 0.444, 0.999}

	cn0 := &Constraint{pos, d}
	cn1 := &Constraint{pos.MulScalar(-1), -d}
	cn2 := &Constraint{lastPos.MulScalar(-1), -d}
	cv := Convex{cn1, cn2}

	for _, idx := range h.Intersections(cn0) {
		SubDivide(h, idx, l1)
	}
	for _, idx := range h.Intersections(cv) {
		CullToLevel(h, idx, l0)
	}

	if err := validateHTM(h); err != nil {
		t.Fatal(err)
	}
}

func TestConstraintCull(t *testing.T) {
	l0, l1 := 3, 7
	h0, h1 := New(), New()
	h0.SubDivide(l0)
	h1.SubDivide(l0)

	cn := &Constraint{lmath.Vec3{0.001, 0.001, 0.999}, 0.85}
	for idx := range h1.Intersections(cn) {
		SubDivide(h1, idx, l1)
	}

	if err := validateHTM(h1); err != nil {
		t.Fatal(err)
	}

	h1.CullToLevel(l0)

	if err := validateHTM(h1); err != nil {
		t.Fatal(err)
	}
	if err := compareHTMs(h0, h1); err != nil {
		t.Fatal(err)
	}
}

func TestMotionCull(t *testing.T) {
	l0, l1 := 3, 7
	h := New()
	defer saveImage(t, h, "test.motion.cull.png")

	h.SubDivide(l0)

	// simulate motion
	d := 0.99
	pos := lmath.Vec3{0.001, 0.001, 0.999}
	lastPos := lmath.Vec3{0, 0, 0}

	for pos.Y = 0.001; pos.Y < 0.644; pos.Y += 0.001 {
		if !lastPos.Equals(lmath.Vec3Zero) {
			cn0 := &Constraint{pos, d}

			cn1 := &Constraint{pos.MulScalar(-1), -d}
			cn2 := &Constraint{lastPos.MulScalar(-1), -d}
			cv0 := Convex{cn1, cn2}

			for _, idx := range h.Intersections(cn0) {
				SubDivide(h, idx, l1)
			}
			for _, idx := range h.Intersections(cv0) {
				CullToLevel(h, idx, l0)
			}
		}
		lastPos = pos
	}

	if err := validateHTM(h); err != nil {
		t.Fatal(err)
	}
}

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

func TestImageL7SubdivideL9Intersect(t *testing.T) {
	h := New()
	h.SubDivide(7)

	x := 640
	size := image.Pt(x, x)

	m0 := image.NewRGBA(image.Rect(0, 0, size.X, size.Y))
	draw.Draw(m0, m0.Bounds(), &image.Uniform{color.RGBA{0, 0, 0, 255}}, image.ZP, draw.Src)

	m1 := Image(h, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{uint8(z * 255), 0, 0, 255})
	})

	cn0 := &Constraint{lmath.Vec3{0, 1, 0}, -0.01}
	cn1 := &Constraint{lmath.Vec3{0, -1, 0}, -0.01}
	cv := Convex{cn0, cn1}
	for _, idx := range h.Intersections(cv) {
		SubDivide(h, idx, 9)
	}
	m2 := ImageConstraint(h, cn0, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{0, 0, uint8(z * 255), 0})
	})
	m3 := ImageConstraint(h, cn1, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{0, 50, uint8(z * 100), 100})
	})
	m4 := ImageConstraint(h, cv, size, func(m *image.RGBA, x, y int, z float64) {
		m.Set(x, y, color.RGBA{0, uint8(z * 255), 0, 0})
	})

	MergeImages(m0, m1, m2, m3, m4)

	if err := WriteImage("samples/test.htm.L7.subdivide.L9.constraint.001.0.75.png", m0); err != nil {
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

	size := image.Pt(640, 640)

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

	if err := WriteImage("samples/test.image.l9.noise.png", m0); err != nil {
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

func testUnique(t *testing.T) {
	h := New()
	h.SubDivide(8)
	s := &set{}
	for _, v := range h.Vertices {
		s.put(v.X)
		s.put(v.Y)
		s.put(v.Z)
	}
	sort.Float64s(s.Data)
	// t.Log("Unique")
	// for _, v := range s.Data {
	// 	t.Logf("%v", v)
	// }
	t.Logf("Unique: %v\n", len(s.Data))
}

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

	if err := validateHTM(h); err != nil {
		t.Fatal(err)
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
