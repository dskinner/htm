package htm

import (
	"math"

	"azul3d.org/lmath.v1"
)

// TexCoords returns a slice of UV coordinates for texture mapping.
// TODO(d) seam does not wrap correctly.
// TODO(d) allow user to declare which axis is up.
func TexCoords(verts []lmath.Vec3) []float32 {
	var tc []float32
	for _, v0 := range verts {
		u := 0.5 + math.Atan2(v0.Y, v0.X)/(math.Pi*2)
		v := 0.5 - math.Asin(v0.Z)/math.Pi
		tc = append(tc, float32(u), float32(v))
	}
	return tc
}

// TexCoordsPlanar returns a slice of UV coordinates mapped against
// a 2-dimensional plane.
// TODO(d) allow user to declare up axis.
func TexCoordsPlanar(verts []lmath.Vec3) []float32 {
	var xlo, xhi, zlo, zhi float64
	for _, v0 := range verts {
		if v0.X < xlo {
			xlo = v0.X
		}
		if v0.X > xhi {
			xhi = v0.X
		}
		if v0.Z < zlo {
			zlo = v0.Z
		}
		if v0.Z > zhi {
			zhi = v0.Z
		}
	}
	xrng := (xlo - xhi) * -1
	xoffset := 0 - xlo

	zrng := (zlo - zhi) * -1
	zoffset := 0 - zlo

	//
	var tc []float32
	for _, v0 := range verts {
		u := (v0.X + xoffset) / xrng
		v := (v0.Z + zoffset) / zrng
		tc = append(tc, float32(u), float32(v))
	}
	return tc
}
