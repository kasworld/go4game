package go4game

type ActionType int

type ActionPoint struct {
	point int
	as    []*ActionStat
}

func NewActionPoint(typecount int) *ActionPoint {
	r := ActionPoint{
		as: make([]*ActionStat, typecount),
	}
	for i := 0; i < typecount; i++ {
		r.as[i] = NewActionStat()
	}
	return &r
}

func (ap *ActionPoint) Add(val int) {
	ap.point += val
}

func (ap *ActionPoint) Use(apt ActionType, value int) bool {
	if ap.CanUse(value) {
		ap.point -= value
		ap.as[apt].Inc()
		return true
	}
	return false
}

func (ap *ActionPoint) CanUse(value int) bool {
	return ap.point >= value
}
