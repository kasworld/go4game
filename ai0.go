package go4game

import ()

func NewAIRandom() AIActor {
	return NewAIBase([...]int{
		ActionAccel:         1,
		ActionBullet:        1,
		ActionSuperBullet:   1,
		ActionHommingBullet: 1,
		ActionBurstBullet:   1,
	})
}
func NewAICloud() AIActor {
	return NewAIBase([...]int{
		ActionAccel:         1,
		ActionBullet:        0,
		ActionSuperBullet:   0,
		ActionHommingBullet: 1,
		ActionBurstBullet:   0,
	})
}
func NewAINothing() AIActor {
	return NewAIBase([...]int{
		ActionAccel:         0,
		ActionBullet:        0,
		ActionSuperBullet:   0,
		ActionHommingBullet: 0,
		ActionBurstBullet:   0,
	})
}
func NewAINoMove() AIActor {
	return NewAIBase([...]int{
		ActionAccel:         2,
		ActionBullet:        0,
		ActionSuperBullet:   0,
		ActionHommingBullet: 0,
		ActionBurstBullet:   0,
	})
}

func NewAI5() AIActor {
	return NewAIAdv([...]int{
		ActionAccel:         4,
		ActionBullet:        4,
		ActionSuperBullet:   4,
		ActionHommingBullet: 4,
		ActionBurstBullet:   4,
	})
}
