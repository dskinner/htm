package htm

import (
	"github.com/azul3d/engine/lmath"
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

type Tester interface {
	Test(v0, v1, v2 lmath.Vec3) Coverage
}

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

func (c Convex) Test(v0, v1, v2 lmath.Vec3) Coverage {
	r := Inside
	for _, cn := range c {
		cv := cn.Test(v0, v1, v2)
		if cv == Outside {
			return Outside
		} else if cv == Partial {
			r = Partial
		}
	}
	return r
}

func (c Convex) Sign() Sign {
	// TODO(d) ...
	return Zero
}

// Domain is several convexes (logical OR of convexes).
type Domain []*Convex

func (d Domain) Test(v0, v1, v2 lmath.Vec3) Coverage {
	r := Outside
	for _, cv := range d {
		t := cv.Test(v0, v1, v2)
		if t == Inside {
			return Inside
		} else if t == Partial {
			r = Partial
		}
	}
	return r
}
