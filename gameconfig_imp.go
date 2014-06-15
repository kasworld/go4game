package go4game

// -------------------

const WorldSize = 500
const WorldSizeY = 0

var defaultConfig = GameConfig{
	AICountPerWorld: 6,
	SetTerrain:      false,
	SetFood:         true,
	StartWorldCount: 1,
	AINames: []string{
		// "Nothing-0-0-0-0-0",
		// "NoMove-1-0-0-0-0",
		// "Home-2-0-0-0-0",
		// "Cloud-2-0-0-1-0",
		// "Random-3-1-1-1-1",
		"Adv-4-3-3-2-2",
		"Adv-4-4-4-2-2",
		"Adv-5-5-5-2-2",
		"Adv-5-3-3-2-2",
		"Adv-5-4-4-2-2",
		"Adv-4-5-5-2-2",
	},
	APIncFrame:           10,
	ShieldCount:          8,
	MaxTcpClientPerWorld: 32,
	MaxWsClientPerWorld:  32,
	RemoveEmptyWorld:     false,
	TcpClientEncode:      "gob",
	TcpListen:            "0.0.0.0:6666",
	WsListen:             "0.0.0.0:8080",
	FramePerSec:          60.0,
	KillScore:            1,
	WorldCube: &HyperRect{
		Vector3D{-WorldSize, -WorldSizeY, -WorldSize},
		Vector3D{WorldSize, WorldSizeY, WorldSize}},
	MoveLimit: [GameObjEnd]float64{
		GameObjMain:          100,
		GameObjShield:        200,
		GameObjBullet:        300,
		GameObjHommingBullet: 200,
		GameObjSuperBullet:   600,
		GameObjDeco:          600,
		GameObjMark:          100},
	Radius: [GameObjEnd]float64{
		GameObjMain:          10,
		GameObjShield:        5,
		GameObjBullet:        5,
		GameObjHommingBullet: 7,
		GameObjSuperBullet:   15,
		GameObjDeco:          3,
		GameObjMark:          3,
		GameObjHard:          3,
		GameObjFood:          3},
	IsInteract: [GameObjEnd][GameObjEnd]bool{
		GameObjMain: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjShield: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        true,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjHommingBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjSuperBullet: [GameObjEnd]bool{
			GameObjMain:          true,
			GameObjShield:        true,
			GameObjBullet:        false,
			GameObjHommingBullet: true,
			GameObjSuperBullet:   true,
			GameObjHard:          true},
		GameObjFood: [GameObjEnd]bool{
			GameObjMain: true},
	},
	AP: [ActionEnd]int{
		ActionAccel:         1,
		ActionBullet:        10,
		ActionSuperBullet:   80,
		ActionHommingBullet: 100,
		ActionBurstBullet:   10,
	},
}

var GameConst = defaultConfig
