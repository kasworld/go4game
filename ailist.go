package go4game

import ()

func NewAIRandom() AIActor {
	return NewAIAdv("Random",
		[...]int{
			ActionAccel:         3,
			ActionBullet:        1,
			ActionSuperBullet:   1,
			ActionHommingBullet: 1,
			ActionBurstBullet:   1,
		})
}
func NewAICloud() AIActor {
	return NewAIAdv("Cloud",
		[...]int{
			ActionAccel:         2,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 1,
			ActionBurstBullet:   0,
		})
}
func NewAINothing() AIActor {
	return NewAIAdv("Nothing",
		[...]int{
			ActionAccel:         0,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 0,
			ActionBurstBullet:   0,
		})
}
func NewAINoMove() AIActor {
	return NewAIAdv("NoMove",
		[...]int{
			ActionAccel:         1,
			ActionBullet:        0,
			ActionSuperBullet:   0,
			ActionHommingBullet: 0,
			ActionBurstBullet:   0,
		})
}

func NewAI2() AIActor {
	return NewAIAdv("2",
		[...]int{
			ActionAccel:         4,
			ActionBullet:        4,
			ActionSuperBullet:   4,
			ActionHommingBullet: 4,
			ActionBurstBullet:   4,
		})
}
func NewAI3() AIActor {
	return NewAIAdv("3",
		[...]int{
			ActionAccel:         4,
			ActionBullet:        4,
			ActionSuperBullet:   4,
			ActionHommingBullet: 4,
			ActionBurstBullet:   4,
		})
}
func NewAI4() AIActor {
	return NewAIAdv("4",
		[...]int{
			ActionAccel:         4,
			ActionBullet:        4,
			ActionSuperBullet:   4,
			ActionHommingBullet: 4,
			ActionBurstBullet:   4,
		})
}
func NewAI5() AIActor {
	return NewAIAdv("5",
		[...]int{
			ActionAccel:         4,
			ActionBullet:        4,
			ActionSuperBullet:   4,
			ActionHommingBullet: 4,
			ActionBurstBullet:   4,
		})
}
