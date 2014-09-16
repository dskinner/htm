package htm

import (
	// "fmt"
	"testing"

	"azul3d.org/lmath.v1"
)

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

func TestDupX(t *testing.T) {
	h0, h1 := New(), New()

	h0.SubDivide(7)
	h1.SubDivide2(7)

	t.Log(len(*h0.Vertices), len(*h1.Vertices))

	// for _, e := range h1.edges.E {
	// 	t.Log(e)
	// }
}

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

// func TestNewHTM(t *testing.T) {
// 	h := New()
// 	if h.S0 == nil || h.S1 == nil || h.S2 == nil || h.S3 == nil {
// 		t.Fatal("Southern hemisphere not initialized.")
// 	}
// 	if h.N0 == nil || h.N1 == nil || h.N2 == nil || h.N3 == nil {
// 		t.Fatal("Northern hemisphere not initialized.")
// 	}
// 	if len(*h.Vertices) == 0 {
// 		t.Fatal("HTM vertices not initialized.")
// 	}
// }

// func TestHTMSubDivide2(t *testing.T) {
// 	h := New()
// 	h.SubDivide(2)
// 	if len(*h.Vertices) != 30 {
// 		t.Fatalf("Expected 30 vertices but got %v.", len(*h.Vertices))
// 	}
// }

// func TestHTMIndices(t *testing.T) {
// 	h := New()
// 	h.SubDivide(2)
// 	n := h.Indices()
// 	if len(n) != 96 {
// 		t.Fatalf("Expected 96 indices but got %v.", len(n))
// 	}
// }

// func TestTexCoords(t *testing.T) {
// 	h := New()
// 	h.SubDivide(2)
// 	tc := h.TexCoords()
// 	if (len(tc) % 2) != 0 {
// 		t.Fatal("Uneven UV mapping.")
// 	}
// }

func TestLookupByCart(t *testing.T) {
	h := New()
	h.SubDivide(7)
	tr, err := h.LookupByCart(lmath.Vec3{0.9, 0.1, 0.1})
	if err != nil {
		t.Fatal(err)
	}
	if tr == nil {
		t.Fatal("Tree should not be nil.")
	}
}

func testNoDups(t *testing.T) {
	h := New()
	h.SubDivide(5)
	for _, v1 := range *h.Vertices {
		for _, v2 := range *h.Vertices {
			if v1.Equals(v2) {
				t.Fail()
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
		h.SubDivide2(7)
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
		for t := range h.Intersections(cn) {
			_ = t
		}
	}
}
