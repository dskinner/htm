package htm

import (
	"testing"
)

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
