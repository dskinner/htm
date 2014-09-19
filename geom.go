package htm

import (
	"azul3d.org/lmath.v1"
)

type Sign int

const (
	Negative Sign = iota
	Zero
	Positive
	Mixed
)

type Coverage int

const (
	Inside Coverage = iota
	Partial
	Outside
)

// Constraint is a circular area, given by the plane slicing it off the sphere.
type Constraint struct {
	P lmath.Vec3
	D float64
}

func (c *Constraint) Test(v0, v1, v2 lmath.Vec3) Coverage {
	a0 := c.P.Dot(v0) > c.D
	a1 := c.P.Dot(v1) > c.D
	a2 := c.P.Dot(v2) > c.D

	if a0 && a1 && a2 {
		return Inside
	} else if a0 || a1 || a2 {
		return Partial
	} else {
		// TODO(d) finish test as this is not definitive.
		return Outside
	}
}

// Convex is a combination of constraints (logical AND of constraints).
type Convex []*Constraint

func (c Convex) Test(v0, v1, v2 lmath.Vec3) bool {
	for _, cn := range c {
		if cn.Test(v0, v1, v2) == Outside {
			return false
		}
	}
	return true
}

func (c Convex) Sign() Sign {
	return Zero
}

// Domain is several convexes (logical OR of convexes).
type Domain []*Convex

func (d Domain) Test(v0, v1, v2 lmath.Vec3) bool {
	for _, cx := range d {
		if cx.Test(v0, v1, v2) {
			return true
		}
	}
	return false
}
