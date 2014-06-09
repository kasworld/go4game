package go4game

import (
	//"log"
	//"time"
	//"math"
	"math/rand"
	//"sort"
)

func NewAIRandom() AIActor {
	return &AIBase{
		act: [ActionEnd]int{
			ActionAccel:         1,
			ActionBullet:        1,
			ActionSuperBullet:   1,
			ActionHommingBullet: 1,
			ActionBurstBullet:   1,
		},
	}
}
func NewAICloud() AIActor {
	return &AIBase{
		act: [ActionEnd]int{
			ActionAccel:         1,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 1,
			ActionBurstBullet:   0,
		},
	}
}
func NewAINothing() AIActor {
	return &AIBase{
		act: [ActionEnd]int{
			ActionAccel:         0,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 0,
			ActionBurstBullet:   0,
		},
	}
}
func NewAINoMove() AIActor {
	return &AIBase{
		act: [ActionEnd]int{
			ActionAccel:         2,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 0,
			ActionBurstBullet:   0,
		},
	}
}
