package htm

import (
	"errors"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"os"
	"sort"

	"azul3d.org/lmath.v1"
)

// max returns largest value contained in vertices.
func max(h *HTM) float64 {
	var n float64
	for _, v0 := range h.Vertices {
		if n < v0.X {
			n = v0.X
		}
		if n < v0.Y {
			n = v0.Y
		}
		if n < v0.Z {
			n = v0.Z
		}
	}
	return n
}

// min returns smallest value contained in vertices.
func min(h *HTM) float64 {
	var n float64
	for _, v0 := range h.Vertices {
		if n > v0.X {
			n = v0.X
		}
		if n > v0.Y {
			n = v0.Y
		}
		if n > v0.Z {
			n = v0.Z
		}
	}
	return n
}

func maxAbs(h *HTM) float64 {
	a, b := math.Abs(min(h)), max(h)
	if a > b {
		return a
	}
	return b
}

func norm(x float64, max float64) float64 {
	return (x + max) / (max * 2)
}

type vec3Slice []lmath.Vec3

func (p vec3Slice) Len() int           { return len(p) }
func (p vec3Slice) Less(i, j int) bool { return p[i].Z < p[j].Z }
func (p vec3Slice) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// sortVec3s sorts by Z back to front.
func sortVec3s(p vec3Slice) { sort.Sort(p) }

type ColorSet func(m *image.RGBA, x, y int, z float64)

func Image(h *HTM, size image.Point, cset ColorSet) *image.RGBA {
	r := image.Rect(0, 0, size.X, size.Y)
	m := image.NewRGBA(r)

	max := maxAbs(h)
	p := append([]lmath.Vec3(nil), h.VerticesNotEmpty()...)
	sortVec3s(p)
	for _, v0 := range p {
		x := int(norm(v0.X, max) * float64(size.X))
		y := int(norm(v0.Y, max) * float64(size.Y))
		z := norm(v0.Z, max)
		cset(m, x, y, z)
	}

	return m
}

func ImageConstraint(h *HTM, cn Tester, size image.Point, cset ColorSet) *image.RGBA {
	r := image.Rect(0, 0, size.X, size.Y)
	m := image.NewRGBA(r)

	max := maxAbs(h)
	var p []lmath.Vec3
	for idx := range Iter(h, h.Intersections(cn)...) {
		v0, v1, v2 := h.VerticesAt(idx)
		p = append(p, v0, v1, v2)
	}
	sortVec3s(p)
	for _, v0 := range p {
		x := int(norm(v0.X, max) * float64(size.X))
		y := int(norm(v0.Y, max) * float64(size.Y))
		z := norm(v0.Z, max)
		cset(m, x, y, z)
	}

	return m
}

func MergeImages(dst *image.RGBA, srcs ...*image.RGBA) {
	for _, src := range srcs {
		draw.Draw(dst, src.Bounds(), src, image.ZP, draw.Over)
	}
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
