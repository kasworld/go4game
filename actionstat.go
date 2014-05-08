package go4game

import (
	"fmt"
	"time"
)

type ActionStatElement struct {
	Count int64
	Time  time.Time
}

type ActionStat struct {
	Total ActionStatElement
	Laps  []ActionStatElement
}

func IntDurToStr(count int64, dur time.Duration) string {
	return fmt.Sprintf("[%v/%5.1f/s]", count, float64(count)/dur.Seconds())
}

func (a ActionStat) String() string {
	defer a.UpdateLap()
	dur := time.Now().Sub(a.Total.Time)
	lastlap := a.Laps[len(a.Laps)-1]
	lapcount, lapdur := a.Total.Count-lastlap.Count, time.Now().Sub(lastlap.Time)
	//a.Total.Sub(a.Laps[len(a.Laps)-1])
	return fmt.Sprintf("(total:%v lap:%v)",
		IntDurToStr(a.Total.Count, dur),
		IntDurToStr(lapcount, lapdur))
}

func NewActionStat() *ActionStat {
	r := &ActionStat{
		Total: ActionStatElement{0, time.Now()},
		Laps:  make([]ActionStatElement, 1),
	}
	r.UpdateLap()
	return r
}

func (a *ActionStat) NewLap() {
	a.Laps = append(a.Laps, ActionStatElement{a.Total.Count, time.Now()})
}

func (a *ActionStat) UpdateLap() {
	a.Laps[len(a.Laps)-1] = ActionStatElement{a.Total.Count, time.Now()}
}

func (a *ActionStat) Inc() {
	a.Total.Count++
}
