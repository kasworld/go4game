package snakebase

import (
	// "fmt"
	"github.com/kasworld/go4game"
	// "runtime"
	//"time"
	"log"
	"math"
)

type Collider interface {
	CollideTo(Collider) bool
	Center() go4game.Vector3D
	Vol() *go4game.HyperRect
}

type Sphere struct {
	center go4game.Vector3D
	Radius float64
}

func BoxIntersectsSphere(cube go4game.HyperRect, c go4game.Vector3D, r float64) bool {
	r2 := r * r
	dmin := 0.0
	for i := 0; i < 3; i++ {
		if c[i] < cube.Min[i] {
			dmin += math.Pow(c[i]-cube.Min[i], 2)
		} else if c[i] > cube.Max[i] {
			dmin += math.Pow(c[i]-cube.Max[i], 2)
		}
	}
	return dmin <= r2
}

func CollCube2Sphere(cc *Cube, sp *Sphere) bool {
	return BoxIntersectsSphere(cc.HyperRect, sp.center, sp.Radius)
}

func (c *Sphere) Center() go4game.Vector3D {
	return c.center
}
func (c *Sphere) CollideTo(d Collider) bool {
	switch t := d.(type) {
	default:
		log.Printf("unknown type %T", t)
	case *Sphere:
		cs := d.(*Sphere)
		return c.Center().Sqd(cs.Center()) <= math.Pow(c.Radius+cs.Radius, 2)
	case *Cube:
		cc := d.(*Cube)
		return CollCube2Sphere(cc, c)
	}
	return false
}
func (o *Sphere) Vol() *go4game.HyperRect {
	return go4game.NewHyperRectByCR(o.center, o.Radius)
}

type Cube struct {
	go4game.HyperRect
}

func (c *Cube) Center() go4game.Vector3D {
	return c.HyperRect.Center()
}
func (c *Cube) CollideTo(d Collider) bool {
	switch t := d.(type) {
	default:
		log.Printf("unknown type %T", t)
	case *Sphere:
		sp := d.(*Sphere)
		return CollCube2Sphere(c, sp)
	case *Cube:
		cc := d.(*Cube)
		return c.IsOverlap(&cc.HyperRect)
	}
	return false
}
func (o *Cube) Vol() *go4game.HyperRect {
	return &o.HyperRect
}

func test_Collider() {
	s := Sphere{
		Radius: 100,
	}
	c := Cube{
		HyperRect: go4game.HyperRect{
			go4game.Vector3D{101, 0, 0},
			go4game.Vector3D{200, 200, 200},
		},
	}
	print(c.CollideTo(&c))
	print(s.CollideTo(&s))
	print(s.CollideTo(&c))
	println(c.CollideTo(&s))
}
