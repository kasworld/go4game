package go4game

type ActionPoint struct {
	point int
	as    [ActionEnd]ActionStat
}

func NewActionPoint() *ActionPoint {
	r := ActionPoint{}
	for i := ActionAccel; i < ActionEnd; i++ {
		r.as[i] = *NewActionStat()
	}
	return &r
}

func (ap *ActionPoint) Add(val int) {
	ap.point += val
}

func (ap *ActionPoint) Use(apt ClientActionType, count int) bool {
	if ap.CanUse(apt, count) {
		ap.point -= GameConst.AP[apt]
		ap.as[apt].Inc()
		return true
	}
	return false
}

func (ap *ActionPoint) CanUse(apt ClientActionType, count int) bool {
	return ap.point >= GameConst.AP[apt]*count
}
