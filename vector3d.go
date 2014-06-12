package go4game

import (
	//"log"
	"fmt"
	"math"
	"math/rand"
)

type Vector3D [3]float64

func (v Vector3D) String() string {
	return fmt.Sprintf("[%5.2f,%5.2f,%5.2f]", v[0], v[1], v[2])
}

var V3DZero = Vector3D{0, 0, 0}
var V3DUnitX = Vector3D{1, 0, 0}
var V3DUnitY = Vector3D{0, 1, 0}
var V3DUnitZ = Vector3D{0, 0, 1}

// func (p Vector3D) Copy() Vector3D {
// 	return Vector3D{p[0], p[1], p[2]}
// }
func (p Vector3D) Eq(other Vector3D) bool {
	return p == other
	//return p[0] == other[0] && p[1] == other[1] && p[2] == other[2]
}
func (p Vector3D) Ne(other Vector3D) bool {
	return !p.Eq(other)
}
func (p Vector3D) IsZero() bool {
	return p.Eq(V3DZero)
}
func (p Vector3D) Add(other Vector3D) Vector3D {
	return Vector3D{p[0] + other[0], p[1] + other[1], p[2] + other[2]}
}
func (p Vector3D) Neg() Vector3D {
	return Vector3D{-p[0], -p[1], -p[2]}
}
func (p Vector3D) Sub(other Vector3D) Vector3D {
	return Vector3D{p[0] - other[0], p[1] - other[1], p[2] - other[2]}
}
func (p Vector3D) Mul(other Vector3D) Vector3D {
	return Vector3D{p[0] * other[0], p[1] * other[1], p[2] * other[2]}
}
func (p Vector3D) Imul(other float64) Vector3D {
	return Vector3D{p[0] * other, p[1] * other, p[2] * other}
}
func (p Vector3D) Idiv(other float64) Vector3D {
	return Vector3D{p[0] / other, p[1] / other, p[2] / other}
}
func (p Vector3D) Abs() float64 {
	return math.Sqrt(p[0]*p[0] + p[1]*p[1] + p[2]*p[2])
}
func (p Vector3D) Sqd(q Vector3D) float64 {
	var sum float64
	for dim, pCoord := range p {
		d := pCoord - q[dim]
		sum += d * d
	}
	return sum
}

func (p Vector3D) LenTo(other Vector3D) float64 {
	return math.Sqrt(p.Sqd(other))
}

func (p *Vector3D) Normalize() {
	d := p.Abs()
	if d > 0 {
		p[0] /= d
		p[1] /= d
		p[2] /= d
	}
}
func (p Vector3D) Normalized() Vector3D {
	d := p.Abs()
	if d > 0 {
		return p.Idiv(d)
	}
	return p
}
func (p Vector3D) NormalizedTo(l float64) Vector3D {
	d := p.Abs() / l
	if d != 0 {
		return p.Idiv(d)
	}
	return p
}
func (p Vector3D) Dot(other Vector3D) float64 {
	return p[0]*other[0] + p[1]*other[1] + p[2]*other[2]
}
func (p Vector3D) Cross(other Vector3D) Vector3D {
	return Vector3D{
		p[1]*other[2] - p[2]*other[1],
		-p[0]*other[2] + p[2]*other[0],
		p[0]*other[1] - p[1]*other[0],
	}
}

// reflect plane( == normal vector )
func (p Vector3D) Reflect(normal Vector3D) Vector3D {
	d := 2 * (p[0]*normal[0] + p[1]*normal[1] + p[2]*normal[2])
	return Vector3D{p[0] - d*normal[0], p[1] - d*normal[1], p[2] - d*normal[2]}
}
func (p Vector3D) RotateAround(axis Vector3D, theta float64) Vector3D {
	// Return the vector rotated around axis through angle theta. Right hand rule applies
	// Adapted from equations published by Glenn Murray.
	// http://inside.mines.edu/~gmurray/ArbitraryAxisRotation/ArbitraryAxisRotation.html
	x, y, z := p[0], p[1], p[2]
	u, v, w := axis[0], axis[1], axis[2]

	// Extracted common factors for simplicity and efficiency
	r2 := u*u + v*v + w*w
	r := math.Sqrt(r2)
	ct := math.Cos(theta)
	st := math.Sin(theta) / r
	dt := (u*x + v*y + w*z) * (1 - ct) / r2
	return Vector3D{
		(u*dt + x*ct + (-w*y+v*z)*st),
		(v*dt + y*ct + (w*x-u*z)*st),
		(w*dt + z*ct + (-v*x+u*y)*st),
	}
}
func (p Vector3D) Angle(other Vector3D) float64 {
	// Return the angle to the vector other
	return math.Acos(p.Dot(other) / (p.Abs() * other.Abs()))
}
func (p Vector3D) Project(other Vector3D) Vector3D {
	// Return one vector projected on the vector other
	n := other.Normalized()
	return n.Imul(p.Dot(n))
}

// for aim ahead target with projectile
// return time dur
func (srcpos Vector3D) CalcAimAheadDur(dstpos Vector3D, dstmv Vector3D, bulletspeed float64) float64 {
	totargetvt := dstpos.Sub(srcpos)
	a := dstmv.Dot(dstmv) - bulletspeed*bulletspeed
	b := 2 * dstmv.Dot(totargetvt)
	c := totargetvt.Dot(totargetvt)
	p := -b / (2 * a)
	q := math.Sqrt((b*b)-4*a*c) / (2 * a)
	t1 := p - q
	t2 := p + q

	var rtn float64
	if t1 > t2 && t2 > 0 {
		rtn = t2
	} else {
		rtn = t1
	}
	if rtn < 0 || math.IsNaN(rtn) {
		return math.Inf(1)
	}
	return rtn
}

// for serialize
func (v Vector3D) NewInt32Vector() [3]int32 {
	return [3]int32{int32(v[0]), int32(v[1]), int32(v[2])}
}

func FromInt32Vector(s [3]int32) Vector3D {
	return Vector3D{float64(s[0]), float64(s[1]), float64(s[2])}
}

func RandVector3D(st, end float64) Vector3D {
	return Vector3D{
		rand.Float64()*(end-st) + st,
		rand.Float64()*(end-st) + st,
		rand.Float64()*(end-st) + st,
	}
}

func RandVector(st, end Vector3D) Vector3D {
	return Vector3D{
		rand.Float64()*(end[0]-st[0]) + st[0],
		rand.Float64()*(end[1]-st[1]) + st[1],
		rand.Float64()*(end[2]-st[2]) + st[2],
	}
}

func (center Vector3D) To8Direct(v2 Vector3D) int {
	rtn := 0
	for i := 0; i < 3; i++ {
		if center[i] > v2[i] {
			rtn += 1 << uint(i)
		}
	}
	return rtn
}

func (h *HyperRect) MakeCubeBy8Driect(center Vector3D, direct8 int) *HyperRect {
	rtn := Vector3D{}
	for i := 0; i < 3; i++ {
		if direct8&(1<<uint(i)) != 0 {
			rtn[i] = h.Min[i]
		} else {
			rtn[i] = h.Max[i]
		}
	}
	return NewHyperRect(center, rtn)
}

type HyperRect struct {
	Min, Max Vector3D
}

func (h *HyperRect) Center() Vector3D {
	return h.Min.Add(h.Max).Idiv(2)
}

func (h *HyperRect) DiagLen() float64 {
	return h.Min.LenTo(h.Max)
}

func (h *HyperRect) SizeVector() Vector3D {
	return h.Max.Sub(h.Min)
}

func (h *HyperRect) IsContact(c Vector3D, r float64) bool {
	hc := h.Center()
	hl := h.DiagLen()
	return hl/2+r >= hc.LenTo(c)
}

func NewHyperRectByCR(c Vector3D, r float64) *HyperRect {
	return &HyperRect{
		Vector3D{c[0] - r, c[1] - r, c[2] - r},
		Vector3D{c[0] + r, c[1] + r, c[2] + r},
	}
}

func (h *HyperRect) RandVector() Vector3D {
	return Vector3D{
		rand.Float64()*(h.Max[0]-h.Min[0]) + h.Min[0],
		rand.Float64()*(h.Max[1]-h.Min[1]) + h.Min[1],
		rand.Float64()*(h.Max[2]-h.Min[2]) + h.Min[2],
	}
}

func (h *HyperRect) Move(v Vector3D) *HyperRect {
	return &HyperRect{
		Min: h.Min.Add(v),
		Max: h.Max.Add(v),
	}
}

func (h *HyperRect) IMul(i float64) *HyperRect {
	hs := h.SizeVector().Imul(i / 2)
	hc := h.Center()
	return &HyperRect{
		Min: hc.Sub(hs),
		Max: hc.Add(hs),
	}
}

// make normalized hyperrect , if not need use HyperRect{Min: , Max:}
func NewHyperRect(v1 Vector3D, v2 Vector3D) *HyperRect {
	rtn := HyperRect{
		Min: Vector3D{},
		Max: Vector3D{},
	}
	for i := 0; i < 3; i++ {
		if v1[i] > v2[i] {
			rtn.Max[i] = v1[i]
			rtn.Min[i] = v2[i]
		} else {
			rtn.Max[i] = v2[i]
			rtn.Min[i] = v1[i]
		}
	}
	return &rtn
}

func (h1 *HyperRect) IsOverlap(h2 *HyperRect) bool {
	return !((h1.Min[0] > h2.Max[0] || h1.Max[0] < h2.Min[0]) ||
		(h1.Min[1] > h2.Max[1] || h1.Max[1] < h2.Min[1]) ||
		(h1.Min[2] > h2.Max[2] || h1.Max[2] < h2.Min[2]))

	// for i := 0; i < 3; i++ {
	// 	if !between(h1.Min[i], h1.Max[i], h2.Min[i]) && !between(h1.Min[i], h1.Max[i], h2.Max[i]) {
	// 		return false
	// 	}
	// }
	// return true
}

func (p Vector3D) IsIn(hr *HyperRect) bool {
	return hr.Min[0] <= p[0] && p[0] <= hr.Max[0] &&
		hr.Min[1] <= p[1] && p[1] <= hr.Max[1] &&
		hr.Min[2] <= p[2] && p[2] <= hr.Max[2]
	// for i := 0; i < 3; i++ {
	// 	if hr.Min[i] > p[i] || hr.Max[i] < p[i] {
	// 		return false
	// 	}
	// }
	// return true
}

func (p *Vector3D) MakeIn(hr *HyperRect) int {
	changed := 0
	var i uint
	for i = 0; i < 3; i++ {
		if p[i] > hr.Max[i] {
			p[i] = hr.Max[i]
			changed += 1 << (i*2 + 1)
		}
		if p[i] < hr.Min[i] {
			p[i] = hr.Min[i]
			changed += 1 << (i * 2)
		}
	}
	return changed
}
