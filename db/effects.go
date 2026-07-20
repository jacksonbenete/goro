package db

import (
	"image/color"
	"time"
)

const (
	ReferenceActiveEffectTableEntries = 607
	ReferenceNumericEffectConstants   = 1147
)

const (
	effectFireBolt       = 10019
	effectNapalmBeat     = 32
	effectGroundSample   = 513
	effectCastRing       = 10021
	effectProvoke        = 67
	effectMvp            = 68
	effectEndure         = 11
	effectBeginSpell     = 12
	effectSafetyWall     = 315
	effectColdBolt       = 10014
	effectBashBegin      = 16
	effectHit1           = 0
	effectBashHit        = 1
	effectHit3           = 2
	effectHit4           = 3
	effectHit5           = 4
	effectHit6           = 5
	effectEntry          = 6
	effectExit           = 7
	effectWarp           = 8
	effectEnhance        = 9
	effectArrowShot      = 10060
	effectArrowShower    = 10061
	effectMammonite      = 10
	effectCartRevolution = 170
	effectSight          = 22
	effectSoulStrike     = 15
	effectGlassWall      = 13
	effectHealSP         = 14
	effectMagnumBreak    = 17
	effectQuakeMagnum    = 10022
	effectSteal          = 18
	effectSummonSlave    = 215
	effectPoisonAttack   = 20
	effectDetoxication   = 21
	effectStoneCurse     = 23
	effectFireBall       = 24
	effectFireWall       = 25
	effectIceArrow       = 26
	effectFrostDiver     = 27
	effectFrostDiverHit  = 28
	effectLightningBolt  = 29
	effectThunderStorm   = 30
	effectFireArrow      = 31
	effectTeleportOld    = 34
	effectReadyPortalOld = 35
	effectIncAgility     = 37
	effectDecAgility     = 38
	effectIncAgiDex      = 43
	effectRuwach         = 33
	effectAqua           = 39
	effectSignum         = 40
	effectAngelus        = 41
	effectBlessing       = 42
	effectGloria         = 75
	effectMagnificat     = 76
	effectResurrection   = 77
	effectLexAeterna     = 85
	effectSuffragium     = 88
	effectStormGust      = 89
	effectWeaponPerfect  = 103
	effectMaximizePower  = 104
	effectKyrie          = 112
	effectChristmasCarol = 717
	effectFireHit        = 49
	effectFireSplashHit  = 50
	effectColdHit        = 51
	effectWindHit        = 52
	effectPoisonHit      = 53
	effectBeginSpell2    = 54
	effectBeginSpell3    = 55
	effectBeginSpell4    = 56
	effectBeginSpell5    = 57
	effectBeginSpell6    = 58
	effectBeginSpell7    = 59
	effectLockOnTarget   = 60
	effectWarpZone       = 61
	effectSightTrasher   = 62
	effectArrowShotRO    = 64
	effectInvenom        = 65
	effectSkidTrap       = 69
	effectBrandishSpear  = 70
	effectIceWall        = 74
	effectRecovery       = 78
	effectEarthSpike     = 79
	effectSpearBoomerang = 80
	effectPierce         = 81
	effectTurnUndead     = 82
	effectSanctuary      = 83
	effectImpositio      = 84
	effectAspersio       = 86
	effectLexDivina      = 87
	effectLordVermilion  = 90
	effectBenedictio     = 91
	effectMeteorStorm    = 92
	effectJupitelThunder = 93
	effectJupitelHit     = 94
	effectQuagmire       = 95
	effectFirePillar     = 96
	effectFirePillarBomb = 97
	effectHasteUp        = 98
	effectFlasher        = 99
	effectRemoveTrap     = 100
	effectRepairWeapon   = 101
	effectCrashEarth     = 102
	effectBlastMine      = 105
	effectBlastMineBomb  = 106
	effectClaymore       = 107
	effectFreezingTrap   = 108
	effectGasPush        = 110
	effectSpringTrap     = 111
	effectMagnus         = 113
	effectBlitzBeat      = 115
	effectWaterBall      = 116
	effectWaterBall2     = 117
	effectDetecting      = 119
	effectCloaking       = 120
	effectSonicBlow      = 121
	effectSonicBlowHit   = 122
	effectGrimtooth      = 123
	effectVenomDust      = 124
	effectPoisonReact    = 126
	effectPoisonReact2   = 127
	effectOverthrust     = 128
	effectVenomSplasher  = 129
	effectTwoHandQuicken = 130
	effectAutoCounter    = 131
	effectGrimtoothAtk   = 132
	effectFreeze         = 133
	effectFreezed        = 134
	effectIceCrash       = 135
	effectSlowPoison     = 136
	effectFirePillarOn   = 138
	effectSandman        = 139
	effectRevive         = 140
	effectHeavenDrive    = 142
	effectSonicBlow2     = 143
	effectBrandishSpear2 = 144
	effectShockwave      = 145
	effectShockwaveHit   = 146
	effectEarthHit       = 147
	effectPierceSelf     = 148
	effectBowlingSelf    = 149
	effectSpearStabSelf  = 150
	effectRain           = 161
	effectSnow           = 162
	effectSakura         = 163
	effectBanjjakii      = 165
	effectSmoke          = 44
	effectFirefly        = 45
	effectTorch          = 47
	effectBubble         = 109
	effectCure           = 66
	effectPneuma         = 141
	effectHolyLight      = 152
	effectConcentration  = 153
	effectRefineOK       = 154
	effectRefineFail     = 155
	effectTeleportation  = 304
	effectPharmacyOK     = 305
	effectPharmacyFail   = 306
	effectFirstAid       = 309
	effectHeal           = 312
	effectReadyPortal    = 316
	effectPortal         = 317
	effectHealOffensive  = 320
	effectBaseLevelUp    = 371
	effectJobLevelUp     = 158
	effectPotionRed      = 204
	effectPotionOrange   = 205
	effectPotionYellow   = 206
	effectPotionWhite    = 207
	effectPotionBlue     = 208
	effectPotionGreen    = 209
	effectFood           = 210
	effectFoodBlue       = 211
	effectItemFast       = 218
	effectItemFast2      = 219
	effectItemFast3      = 220
	effectFoodChocolate  = 363
	effectResistPotion   = 491
	effectItemAccel      = 507
	effectFirecracker    = 508
	effectItemSlow       = 519
	effectBoxThunder     = 576
	effectBoxResentment  = 577
	effectBoxDrowsiness  = 579
	effectBoxSunlight    = 580
	effectStatFoodSTR    = 593
	effectStatFoodINT    = 594
	effectStatFoodVIT    = 595
	effectStatFoodAGI    = 596
	effectStatFoodDEX    = 597
	effectStatFoodLUK    = 598
	effectFirecracker1   = 612
	effectFirecracker2   = 682
	effectFirecracker3   = 683
	effectFirecracker4   = 684
	effectFirecracker5   = 685
	effectFirecracker6   = 686
	effectFirecracker7   = 709
	effectEnergyCoat     = 169
	effectThrowItem3     = 308
	effectSprinkleSand   = 310
	effectLoud           = 311
	effectPokJuk         = 297
	effectCloud          = 229
	effectCloud2         = 230
	effectMapPillar      = 231
	effectCloud3         = 233
	effectMaple          = 333
	effectDragonSmoke    = 373
	effectRainbow        = 410
	effectCloud4         = 515
	effectCloud5         = 516
	effectCloud6         = 592
	effectBubbleDrop     = 665
	effectTorchRed       = 690
	effectTorchGreen     = 691
	effectMapGhost       = 692
	effectGlow1          = 693
	effectGlow2          = 694
	effectGlow4          = 695
	effectTorchPurple    = 696
	effectCloud7         = 697
	effectCloud8         = 698
)

const EffectPixelRatio = 1.0 / 35.0

type EffectComponentKind int

const (
	EffectComponentSTR EffectComponentKind = iota + 1
	EffectComponentCylinder
	EffectComponent2D
	EffectComponent3D
	EffectComponentSPR
	EffectComponentFUNC
	EffectComponentQuadHorn
)

type EffectSpec struct {
	Duration         time.Duration
	CameraShake      time.Duration
	CameraShakeDelay time.Duration
	DetachLocalActor bool
	SFX              []string
	SFXDelays        []time.Duration
	SFXRandMin       int
	SFXRandMax       int
	Components       []EffectComponent
}

type EffectComponent struct {
	Kind               EffectComponentKind
	FuncName           string
	Color              color.RGBA
	Duration           time.Duration
	DurationRandMin    time.Duration
	DurationRandMax    time.Duration
	Delay              time.Duration
	DuplicateDelay     time.Duration
	DelayOffsetDelta   time.Duration
	Repeat             bool
	RepeatDelay        time.Duration
	STRFile            string
	STRMinFile         string
	STRRandMin         int
	STRRandMax         int
	AttachedEntity     bool
	TexturePath        string
	TextureName        string
	TextureFile        string
	TextureFiles       []string
	FrameDelay         time.Duration
	SpriteFile         string
	ShadowTexture      bool
	SpriteHead         bool
	SpriteDirection    bool
	SpriteRepeat       bool
	SpriteStopAtEnd    bool
	SpriteFrame        int
	SpriteDelay        time.Duration
	SpriteXOffset      float64
	SpriteYOffset      float64
	FromSrc            bool
	ToSrc              bool
	Arc                float64
	Retreat            float64
	AlphaMax           float64
	AlphaMaxDelta      float64
	Sparkling          bool
	SparkNumber        int
	Fade               bool
	FadeIn             bool
	FadeOut            bool
	Rotate             bool
	RotateWithCamera   bool
	FixedPerspective   bool
	RotateToTarget     bool
	WorldSizedSprite   bool
	Animation          int
	BottomSize         float64
	TopSize            float64
	Height             float64
	PosX               float64
	PosY               float64
	PosZ               float64
	PosXEnd            float64
	PosYEnd            float64
	PosZEnd            float64
	PosXRand           float64
	PosYRand           float64
	PosZRand           float64
	PosXStartRand      float64
	PosYStartRand      float64
	PosZStartRand      float64
	PosXStartMiddle    float64
	PosYStartMiddle    float64
	PosZStartMiddle    float64
	PosXEndRand        float64
	PosYEndRand        float64
	PosZEndRand        float64
	PosXEndMiddle      float64
	PosYEndMiddle      float64
	PosZEndMiddle      float64
	PosXSmooth         bool
	PosYSmooth         bool
	PosZSmooth         bool
	SizeStart          float64
	SizeEnd            float64
	SizeRand           float64
	SizeStartX         float64
	SizeStartY         float64
	SizeEndX           float64
	SizeEndY           float64
	SizeStartXRandMin  float64
	SizeStartXRandMax  float64
	SizeStartYRandMin  float64
	SizeStartYRandMax  float64
	SizeEndXRandMin    float64
	SizeEndXRandMax    float64
	SizeEndYRandMin    float64
	SizeEndYRandMax    float64
	SizeRandX          float64
	SizeRandY          float64
	SizeRandXMiddle    float64
	SizeRandYMiddle    float64
	SizeDelta          float64
	SizeSmooth         bool
	AngleStart         float64
	AngleEnd           float64
	AngleX             float64
	AngleY             float64
	AngleZ             float64
	AngleRandMin       float64
	AngleRandMax       float64
	CirclePattern      bool
	CircleInnerSize    float64
	CircleOuterRandMin float64
	CircleOuterRandMax float64
	OrbitRadiusX       float64
	OrbitRadiusY       float64
	OrbitRadiusZ       float64
	OrbitRotations     float64
	OrbitPhase         float64
	OrbitPhaseDelta    float64
	OrbitClockwise     bool
	TotalCircleSides   int
	CircleSides        int
	Duplicate          int
	AngleZRandom       float64
	BlendMode          int
	BlendAdditive      bool
	Overlay            bool
	QuadHornHeightMin  float64
	QuadHornHeightMax  float64
	QuadHornOffsetXMin float64
	QuadHornOffsetXMax float64
	QuadHornOffsetYMin float64
	QuadHornOffsetYMax float64
	QuadHornOffsetZ    float64
	QuadHornBottomMin  float64
	QuadHornBottomMax  float64
	QuadHornRotateXMin float64
	QuadHornRotateXMax float64
	QuadHornRotateYMin float64
	QuadHornRotateYMax float64
	QuadHornRotateZMin float64
	QuadHornRotateZMax float64
	QuadHornAnimSpeed  time.Duration
	QuadHornAnimOut    bool
}

func effectTableSize(value float64) float64 {
	return value * EffectPixelRatio
}

func strEffectSpec(file, wav string) EffectSpec {
	return strEffectSpecRandom(file, wav, 0, 0)
}

func strEffectSpecRandom(file, wav string, randMin, randMax int) EffectSpec {
	return strEffectSpecRandomAttached(file, wav, randMin, randMax, false, false)
}

func strEffectSpecAttached(file, wav string, head bool) EffectSpec {
	return strEffectSpecRandomAttached(file, wav, 0, 0, true, head)
}

func strEffectSpecAttachedMin(file, minFile, wav string, head bool) EffectSpec {
	spec := strEffectSpecAttached(file, wav, head)
	spec.Components[0].STRMinFile = minFile
	return spec
}

func strEffectSpecRandomAttached(file, wav string, randMin, randMax int, attached, head bool) EffectSpec {
	spec := EffectSpec{
		Components: []EffectComponent{{
			Kind:           EffectComponentSTR,
			STRFile:        file,
			STRRandMin:     randMin,
			STRRandMax:     randMax,
			AttachedEntity: attached,
			SpriteHead:     head,
		}},
	}
	if wav != "" {
		spec.SFX = []string{wav}
		spec.SFXRandMin = randMin
		spec.SFXRandMax = randMax
	}
	return spec
}

func soundOnlyEffectSpec(paths ...string) EffectSpec {
	return EffectSpec{
		Duration: 500 * time.Millisecond,
		SFX:      paths,
	}
}

func delayedSoundEffectSpec(paths []string, delays []time.Duration) EffectSpec {
	return EffectSpec{
		Duration:  500 * time.Millisecond,
		SFX:       paths,
		SFXDelays: delays,
	}
}

func potionEffectSpec(file string, c color.RGBA) EffectSpec {
	return EffectSpec{
		Duration: 850 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:    EffectComponentSTR,
			Color:   c,
			STRFile: file,
		}},
	}
}

func berserkPotionEffectSpec() EffectSpec {
	spec := strEffectSpecAttached("버서크", "effect\\ac_concentration.wav", false)
	spec.CameraShake = 200 * time.Millisecond
	spec.CameraShakeDelay = 200 * time.Millisecond
	return spec
}

func robrCylinderComponent(textureName string, tint color.RGBA, duration time.Duration, alphaMax float64, animation int, bottomSize, topSize, height float64, attached, rotate, fade, additive bool) EffectComponent {
	blendMode := 0
	if additive {
		blendMode = 2
	}
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		Color:            tint,
		TextureName:      textureName,
		Duration:         duration,
		AlphaMax:         alphaMax,
		Fade:             fade,
		Rotate:           rotate,
		Animation:        animation,
		BottomSize:       bottomSize,
		TopSize:          topSize,
		Height:           height,
		AttachedEntity:   attached,
		TotalCircleSides: 32,
		CircleSides:      32,
		BlendMode:        blendMode,
		BlendAdditive:    additive,
	}
}

func tintedEffectComponent(component EffectComponent, tint color.RGBA) EffectComponent {
	component.Color = tint
	return component
}

func quadHornEffectComponent(textureFile string, duration time.Duration, height, offsetX, offsetY [2]float64, offsetZ float64, bottomSize, rotateY, rotateZ [2]float64, animation int, animationSpeed time.Duration, animationOut bool, blendMode int) EffectComponent {
	return quadHornEffectComponentFull(textureFile, duration, height, offsetX, offsetY, offsetZ, bottomSize, [2]float64{}, rotateY, rotateZ, animation, animationSpeed, animationOut, blendMode)
}

func quadHornEffectComponentFull(textureFile string, duration time.Duration, height, offsetX, offsetY [2]float64, offsetZ float64, bottomSize, rotateX, rotateY, rotateZ [2]float64, animation int, animationSpeed time.Duration, animationOut bool, blendMode int) EffectComponent {
	return EffectComponent{
		Kind:               EffectComponentQuadHorn,
		TextureFile:        textureFile,
		Duration:           duration,
		AttachedEntity:     false,
		Color:              color.RGBA{R: 255, G: 255, B: 255, A: 255},
		BlendMode:          blendMode,
		BlendAdditive:      blendMode == 2,
		Animation:          animation,
		QuadHornHeightMin:  height[0],
		QuadHornHeightMax:  height[1],
		QuadHornOffsetXMin: offsetX[0],
		QuadHornOffsetXMax: offsetX[1],
		QuadHornOffsetYMin: offsetY[0],
		QuadHornOffsetYMax: offsetY[1],
		QuadHornOffsetZ:    offsetZ,
		QuadHornBottomMin:  bottomSize[0],
		QuadHornBottomMax:  bottomSize[1],
		QuadHornRotateXMin: rotateX[0],
		QuadHornRotateXMax: rotateX[1],
		QuadHornRotateYMin: rotateY[0],
		QuadHornRotateYMax: rotateY[1],
		QuadHornRotateZMin: rotateZ[0],
		QuadHornRotateZMax: rotateZ[1],
		QuadHornAnimSpeed:  animationSpeed,
		QuadHornAnimOut:    animationOut,
	}
}

func iceWallEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 5 * time.Minute,
		SFX:      []string{"effect\\wizard_icewall.wav"},
		Components: []EffectComponent{
			quadHornEffectComponent("effect/ice.tga", 5*time.Minute, [2]float64{2.8, 3.3}, [2]float64{0.25, 0.75}, [2]float64{0.25, 0.75}, -0.1, [2]float64{0.3, 0.5}, [2]float64{1, 360}, [2]float64{}, 1, 50*time.Millisecond, false, 8),
			quadHornEffectComponent("effect/ice.tga", 5*time.Minute, [2]float64{2.3, 2.8}, [2]float64{0.25, 0.75}, [2]float64{0.25, 0.75}, -0.1, [2]float64{0.3, 0.5}, [2]float64{1, 360}, [2]float64{}, 1, 50*time.Millisecond, false, 8),
			quadHornEffectComponent("effect/ice.tga", 5*time.Minute, [2]float64{2.5, 2.9}, [2]float64{0.25, 0.75}, [2]float64{0.25, 0.75}, -0.1, [2]float64{0.3, 0.5}, [2]float64{1, 360}, [2]float64{}, 1, 50*time.Millisecond, false, 8),
		},
	}
}

func earthSpikeEffectSpec() EffectSpec {
	return EffectSpec{
		Duration:    5 * time.Second,
		CameraShake: 200 * time.Millisecond,
		SFX:         []string{"effect\\wizard_earthspike.wav"},
		Components: []EffectComponent{
			quadHornEffectComponent("effect/stone.bmp", 5*time.Second, [2]float64{0.95, 1.5}, [2]float64{0.4, 0.6}, [2]float64{0.4, 0.6}, -0.1, [2]float64{0.5, 0.6}, [2]float64{1, 360}, [2]float64{-8, 8}, 3, 120*time.Millisecond, true, 1),
			quadHornEffectComponent("effect/stone.bmp", 5*time.Second, [2]float64{0.2, 0.4}, [2]float64{0, 0}, [2]float64{0.5, 0.5}, -0.1, [2]float64{0.1, 0.2}, [2]float64{1, 360}, [2]float64{-15, 15}, 2, 100*time.Millisecond, true, 1),
			quadHornEffectComponent("effect/stone.bmp", 5*time.Second, [2]float64{0.2, 0.4}, [2]float64{0, 0.5}, [2]float64{0.5, 1.0}, -0.1, [2]float64{0.1, 0.2}, [2]float64{1, 360}, [2]float64{-15, 15}, 2, 100*time.Millisecond, true, 1),
			quadHornEffectComponent("effect/stone.bmp", 5*time.Second, [2]float64{0.2, 0.4}, [2]float64{1.0, 1.2}, [2]float64{0.5, 0.8}, -0.1, [2]float64{0.1, 0.2}, [2]float64{1, 360}, [2]float64{-15, 15}, 2, 100*time.Millisecond, true, 1),
			quadHornEffectComponent("effect/stone.bmp", 5*time.Second, [2]float64{0.2, 0.4}, [2]float64{0.5, 0.7}, [2]float64{0.0, -0.2}, -0.1, [2]float64{0.1, 0.2}, [2]float64{1, 360}, [2]float64{-15, 15}, 2, 100*time.Millisecond, true, 1),
		},
	}
}

func warpZoneEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 2800 * time.Millisecond,
		Components: []EffectComponent{
			robrCylinderComponent("ring_blue", color.RGBA{R: 128, G: 128, B: 255, A: 255}, 2800*time.Millisecond, 0.3, 3, 2, 3.3, 1.1, true, true, true, true),
			robrCylinderComponent("ring_blue", color.RGBA{R: 128, G: 128, B: 255, A: 255}, 2800*time.Millisecond, 0.3, 3, 1.9, 3.2, 1.1, true, true, true, true),
			{
				Kind:           EffectComponent3D,
				TextureFile:    "effect/pok1.tga",
				Duration:       time.Second,
				Duplicate:      3,
				AlphaMax:       1,
				FadeIn:         true,
				FadeOut:        true,
				PosXStartRand:  3,
				PosYStartRand:  3,
				PosZ:           0,
				PosZEndRand:    2,
				PosZEndMiddle:  2,
				SizeStart:      effectTableSize(100),
				SizeEnd:        effectTableSize(100),
				SizeRand:       effectTableSize(17),
				AttachedEntity: true,
				Color:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
		},
	}
}

func sightTrasherEffectSpec() EffectSpec {
	return EffectSpec{
		Duration:   800 * time.Millisecond,
		SFX:        []string{"effect\\wizard_sightrasher.wav"},
		Components: sightTrasherComponents(),
	}
}

func sightTrasherComponents() []EffectComponent {
	directions := [][2]float64{
		{0, -8},
		{5.66, -5.66},
		{8, 0},
		{5.66, 5.66},
		{0, 8},
		{-5.66, 5.66},
		{-8, 0},
		{-5.66, -5.66},
	}
	components := make([]EffectComponent, 0, len(directions)*2)
	for _, direction := range directions {
		components = append(components, sightTrasherComponent(true, direction[0], direction[1]), sightTrasherComponent(false, direction[0], direction[1]))
	}
	return components
}

func sightTrasherComponent(shadow bool, xEnd, yEnd float64) EffectComponent {
	component := EffectComponent{
		Kind:           EffectComponent3D,
		Duration:       800 * time.Millisecond,
		Duplicate:      4,
		DuplicateDelay: 100 * time.Millisecond,
		AlphaMax:       0.5,
		PosXEnd:        xEnd,
		PosYEnd:        yEnd,
		PosZ:           2,
		PosZEnd:        2,
		SizeStart:      effectTableSize(60),
		SizeEnd:        effectTableSize(160),
		SizeDelta:      -60,
		FadeIn:         true,
		FadeOut:        true,
		SpriteFile:     "data\\sprite\\shadow",
		ShadowTexture:  true,
		SpriteRepeat:   true,
		BlendMode:      8,
	}
	if !shadow {
		component.SpriteFile = "sight"
		component.ShadowTexture = false
		component.AlphaMax = 123.0 / 255.0
		component.AlphaMaxDelta = 3.0 / 255.0
		component.SizeStart = effectTableSize(20)
		component.SizeEnd = effectTableSize(260)
	}
	return component
}

func jupitelThunderEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 200 * time.Millisecond,
		SFX:      []string{"effect\\hunter_shockwavetrap.wav"},
		Components: []EffectComponent{
			{
				Kind:          EffectComponent3D,
				TextureFile:   "effect/thunder_center.bmp",
				Duration:      200 * time.Millisecond,
				ToSrc:         true,
				BlendMode:     2,
				BlendAdditive: true,
				Overlay:       true,
				AlphaMax:      0.66,
				SizeStart:     effectTableSize(35),
				SizeEnd:       effectTableSize(35),
			},
			{
				Kind: EffectComponent3D,
				TextureFiles: []string{
					"effect/thunder_ball_a.bmp",
					"effect/thunder_ball_b.bmp",
					"effect/thunder_ball_c.bmp",
					"effect/thunder_ball_d.bmp",
					"effect/thunder_ball_e.bmp",
					"effect/thunder_ball_f.bmp",
				},
				FrameDelay:    10 * time.Millisecond,
				Duration:      200 * time.Millisecond,
				ToSrc:         true,
				BlendMode:     2,
				BlendAdditive: true,
				Overlay:       true,
				SizeStart:     effectTableSize(45),
				SizeEnd:       effectTableSize(45),
			},
		},
	}
}

func jupitelHitEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 300 * time.Millisecond,
		Components: []EffectComponent{
			{
				Kind:           EffectComponent3D,
				TextureFile:    "effect/thunder_pang.bmp",
				Duration:       100 * time.Millisecond,
				SizeStart:      0,
				SizeEnd:        effectTableSize(25),
				BlendMode:      2,
				BlendAdditive:  true,
				RotateToTarget: true,
				FadeOut:        true,
				Overlay:        true,
				AttachedEntity: true,
			},
			{
				Kind: EffectComponent3D,
				TextureFiles: []string{
					"effect/thunder_plazma_blast_a.bmp",
					"effect/thunder_plazma_blast_b.bmp",
					"effect/thunder_ball_d.bmp",
					"effect/thunder_ball_e.bmp",
					"effect/thunder_ball_f.bmp",
				},
				FrameDelay:     10 * time.Millisecond,
				Duration:       300 * time.Millisecond,
				SizeStart:      effectTableSize(75),
				SizeEnd:        effectTableSize(75),
				BlendMode:      2,
				BlendAdditive:  true,
				Overlay:        true,
				AttachedEntity: true,
			},
		},
	}
}

func repairWeaponEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 1820 * time.Millisecond,
		SFX:      []string{"effect\\black_weapon_repair_a.wav", "effect\\black_weapon_repair_a.wav"},
		SFXDelays: []time.Duration{
			480 * time.Millisecond,
			1320 * time.Millisecond,
		},
		Components: []EffectComponent{{
			Kind:           EffectComponentSTR,
			STRFile:        "repairweapon",
			AttachedEntity: true,
		}},
	}
}

func crashEarthEffectSpec() EffectSpec {
	spec := strEffectSpec("crashearth", "effect\\black_hammerfall.wav")
	spec.CameraShake = 650 * time.Millisecond
	spec.CameraShakeDelay = 350 * time.Millisecond
	return spec
}

func waterBallEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 500 * time.Millisecond,
		SFX:      []string{"effect\\wizard_waterball_chulung.wav"},
		Components: []EffectComponent{{
			Kind: EffectComponent3D,
			TextureFiles: []string{
				"effect/water_out_a.bmp",
				"effect/water_out_b.bmp",
				"effect/water_out_c.bmp",
			},
			FrameDelay:       10 * time.Millisecond,
			Duration:         500 * time.Millisecond,
			FadeOut:          true,
			PosXRand:         1.5,
			PosZRand:         1.5,
			PosY:             0,
			PosYEnd:          3,
			PosYSmooth:       true,
			SizeStart:        effectTableSize(30.5),
			SizeEnd:          effectTableSize(30.5),
			RotateWithCamera: true,
			BlendMode:        2,
			BlendAdditive:    true,
			AttachedEntity:   true,
		}},
	}
}

func waterBall2EffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 1450 * time.Millisecond,
		SFX:      []string{"effect\\wizard_waterball_chulung.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "data\\sprite\\이팩트\\waterball",
			Duration:       500 * time.Millisecond,
			Duplicate:      20,
			DuplicateDelay: 50 * time.Millisecond,
			FromSrc:        true,
			RotateToTarget: true,
			FadeOut:        true,
			SizeStart:      effectTableSize(50),
			SizeEnd:        effectTableSize(50),
			PosZ:           5,
			PosZEnd:        0.0001,
			Arc:            7.5,
			Retreat:        5,
		}},
	}
}

func sonicBlowEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 400 * time.Millisecond,
		SFX:      []string{"effect\\ef_stonecurse.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			TextureFile:    "effect/ring2.bmp",
			Duration:       400 * time.Millisecond,
			AlphaMax:       1,
			FadeOut:        true,
			SizeStart:      effectTableSize(100),
			SizeEnd:        effectTableSize(300),
			BlendMode:      2,
			BlendAdditive:  true,
			AttachedEntity: true,
		}},
	}
}

func sonicBlowHitEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 500 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:           EffectComponentFUNC,
			FuncName:       "SonicBlowHitSpin",
			AttachedEntity: true,
		}},
	}
}

func grimtoothAttackEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 15 * time.Second,
		Components: []EffectComponent{
			quadHornEffectComponentFull("effect/stone.bmp", 15*time.Second, [2]float64{2.5, 2.5}, [2]float64{0, 0}, [2]float64{0.4, 0.4}, -0.2, [2]float64{0.15, 0.15}, [2]float64{-15, -15}, [2]float64{0, 0}, [2]float64{0, 0}, 3, 120*time.Millisecond, true, 1),
			quadHornEffectComponentFull("effect/stone.bmp", 15*time.Second, [2]float64{2.5, 2.5}, [2]float64{0, 0}, [2]float64{0.5, 0.5}, 0, [2]float64{0.15, 0.15}, [2]float64{5, 5}, [2]float64{45, 45}, [2]float64{-15, -15}, 3, 120*time.Millisecond, true, 1),
			quadHornEffectComponentFull("effect/stone.bmp", 15*time.Second, [2]float64{2.5, 2.5}, [2]float64{0, 0}, [2]float64{0.5, 0.5}, 0, [2]float64{0.15, 0.15}, [2]float64{5, 5}, [2]float64{45, 45}, [2]float64{15, 15}, 3, 120*time.Millisecond, true, 1),
		},
	}
}

func firePillarOnEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 6 * time.Second,
		Components: []EffectComponent{
			firePillarOnCylinder(1.0, 2.0, 3),
			firePillarOnCylinder(0.7, 1.5, 5),
			firePillarOnCylinder(0.5, 1.0, 7),
		},
	}
}

func firePillarOnCylinder(bottomSize, topSize, height float64) EffectComponent {
	component := robrCylinderComponent("magic_red", color.RGBA{}, 5*time.Second, 1, 0, bottomSize, topSize, height, false, true, false, false)
	component.Delay = time.Second
	return component
}

func heavenDriveEffectSpec() EffectSpec {
	return EffectSpec{
		Duration:    time.Second,
		CameraShake: 200 * time.Millisecond,
		SFX:         []string{"effect\\wizard_earthspike.wav"},
		Components:  heavenDriveComponents(),
	}
}

func heavenDriveComponents() []EffectComponent {
	components := make([]EffectComponent, 0, 25)
	for x := -2; x <= 2; x++ {
		for y := -2; y <= 2; y++ {
			component := quadHornEffectComponent("effect/stone.bmp", time.Second, [2]float64{0.75, 1.2}, [2]float64{0.2, 0.2}, [2]float64{0.2, 0.2}, -0.1, [2]float64{0.4, 0.7}, [2]float64{1, 360}, [2]float64{-8, 8}, 3, 250*time.Millisecond, true, 1)
			component.PosX = float64(x)
			component.PosY = float64(y)
			components = append(components, component)
		}
	}
	return components
}

func teleportCylinderComponent(bottomSize, topSize, height float64) EffectComponent {
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		Color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		TextureName:      "ring_blue",
		Duration:         1500 * time.Millisecond,
		AlphaMax:         0.5,
		Fade:             true,
		Rotate:           true,
		Animation:        5,
		BottomSize:       bottomSize,
		TopSize:          topSize,
		Height:           height,
		TotalCircleSides: 32,
		CircleSides:      32,
		BlendAdditive:    true,
	}
}

func readyPortalCylinderComponent() EffectComponent {
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		Color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		TextureName:      "ring_blue",
		Duration:         500 * time.Millisecond,
		Repeat:           true,
		RepeatDelay:      -300 * time.Millisecond,
		AlphaMax:         0.4,
		FadeOut:          true,
		Rotate:           true,
		Animation:        4,
		BottomSize:       2.4,
		TopSize:          3.9,
		Height:           0.1,
		PosZ:             0.1,
		TotalCircleSides: 32,
		CircleSides:      32,
		BlendAdditive:    true,
	}
}

func portalCylinderComponent(bottomSize, topSize, height, posZ float64, textureName string, alphaMax float64) EffectComponent {
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		Color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
		TextureName:      textureName,
		Duration:         25000 * time.Millisecond,
		AlphaMax:         alphaMax,
		Fade:             true,
		Rotate:           true,
		Animation:        0,
		BottomSize:       bottomSize,
		TopSize:          topSize,
		Height:           height,
		PosZ:             posZ,
		TotalCircleSides: 32,
		CircleSides:      32,
		BlendAdditive:    true,
	}
}

func healCylinderComponent(bottomSize, topSize, height float64) EffectComponent {
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		Color:            color.RGBA{R: 178, G: 255, B: 178, A: 255},
		TextureName:      "ring_white",
		Duration:         1500 * time.Millisecond,
		AlphaMax:         0.2,
		Fade:             true,
		Rotate:           true,
		Animation:        1,
		BottomSize:       bottomSize,
		TopSize:          topSize,
		Height:           height,
		TotalCircleSides: 32,
		CircleSides:      32,
		BlendAdditive:    true,
	}
}

func healOffensiveCylinderComponent(bottomSize, topSize, height float64) EffectComponent {
	component := healCylinderComponent(bottomSize, topSize, height)
	component.Duration = time.Second
	component.Color = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	component.BlendAdditive = true
	return component
}

func healParticleComponent(alpha float64, duration, delay, duplicateDelay time.Duration, duplicate int, posXRand, posYRand, posZStartRand, posZStartMiddle, posZEnd, posZEndRand, posZEndMiddle float64, sparkling bool) EffectComponent {
	component := EffectComponent{
		Kind:            EffectComponent3D,
		Color:           color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextureFile:     "effect/pok3.tga",
		Duration:        duration,
		Delay:           delay,
		DuplicateDelay:  duplicateDelay,
		AlphaMax:        alpha,
		Sparkling:       sparkling,
		FadeIn:          true,
		FadeOut:         true,
		PosXRand:        posXRand,
		PosYRand:        posYRand,
		PosZStartRand:   posZStartRand,
		PosZStartMiddle: posZStartMiddle,
		PosZEnd:         posZEnd,
		PosZEndRand:     posZEndRand,
		PosZEndMiddle:   posZEndMiddle,
		SizeStart:       9 * EffectPixelRatio,
		SizeEnd:         9 * EffectPixelRatio,
		SizeRand:        2 * EffectPixelRatio,
		Duplicate:       duplicate,
		BlendAdditive:   true,
	}
	if sparkling {
		component.SparkNumber = 2
	}
	return component
}

func coloredTorchEffectSpec(tint color.RGBA) EffectSpec {
	return EffectSpec{
		Duration: 24 * time.Hour,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "torch_01",
			SpriteRepeat:   true,
			Duration:       600 * time.Millisecond,
			SpriteDelay:    100 * time.Millisecond,
			PosX:           0.1,
			PosZ:           0.8,
			SizeStart:      effectTableSize(100),
			SizeEnd:        effectTableSize(100),
			AngleStart:     270,
			AngleEnd:       270,
			RotateToTarget: true,
			AlphaMax:       1,
			BlendAdditive:  true,
			Color:          tint,
		}},
	}
}

func mapGlowEffectSpec(tint color.RGBA, radius float64) EffectSpec {
	return EffectSpec{
		Duration: 24 * time.Hour,
		Components: []EffectComponent{{
			Kind:             EffectComponentCylinder,
			TextureName:      "alpha_center",
			Duration:         1600 * time.Millisecond,
			AlphaMax:         0.36,
			Fade:             true,
			Rotate:           true,
			BottomSize:       radius,
			TopSize:          radius,
			Height:           0.06,
			TotalCircleSides: 32,
			CircleSides:      32,
			BlendAdditive:    true,
			Color:            tint,
		}},
	}
}

func mapFallingParticleEffectSpec(textureFile string, tint color.RGBA, count int) EffectSpec {
	return EffectSpec{
		Duration: 1400 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			TextureFile:    textureFile,
			Duration:       1400 * time.Millisecond,
			Duplicate:      count,
			DuplicateDelay: 18 * time.Millisecond,
			AlphaMax:       0.78,
			FadeIn:         true,
			FadeOut:        true,
			BlendAdditive:  true,
			PosXRand:       4,
			PosYRand:       4,
			PosZStartRand:  1.5,
			PosZ:           12,
			PosZEnd:        0.2,
			PosXEndRand:    1,
			PosYEndRand:    1,
			PosZSmooth:     true,
			SizeStart:      effectTableSize(8),
			SizeEnd:        effectTableSize(5),
			Color:          tint,
		}},
	}
}

func mapRainbowBand(tint color.RGBA, zOffset float64) EffectComponent {
	return EffectComponent{
		Kind:          EffectComponent3D,
		TextureFile:   "effect/alpha_center.tga",
		Duration:      2200 * time.Millisecond,
		AlphaMax:      0.3,
		Fade:          true,
		BlendAdditive: true,
		PosZ:          7 + zOffset,
		SizeStartX:    effectTableSize(900),
		SizeStartY:    effectTableSize(45),
		SizeEndX:      effectTableSize(900),
		SizeEndY:      effectTableSize(45),
		AngleStart:    -18,
		Color:         tint,
	}
}

func weatherRainEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 900 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:          EffectComponent3D,
			TextureFile:   "effect/alpha_center.tga",
			Duration:      900 * time.Millisecond,
			Duplicate:     85,
			AlphaMax:      0.5,
			FadeIn:        true,
			FadeOut:       true,
			BlendAdditive: true,
			PosXRand:      28,
			PosYRand:      28,
			PosZ:          18,
			PosZEnd:       0,
			PosXEnd:       -3.5,
			PosYEnd:       2.0,
			PosZSmooth:    false,
			SizeStartX:    effectTableSize(2),
			SizeStartY:    effectTableSize(52),
			SizeEndX:      effectTableSize(2),
			SizeEndY:      effectTableSize(52),
			AngleStart:    -25,
			Color:         color.RGBA{R: 185, G: 215, B: 255, A: 255},
		}},
	}
}

func weatherSnowEffectSpec() EffectSpec {
	return EffectSpec{
		Duration: 8 * time.Second,
		Components: []EffectComponent{{
			Kind:          EffectComponent3D,
			TextureFile:   "effect/pok3.tga",
			Duration:      8 * time.Second,
			Duplicate:     180,
			AlphaMax:      0.85,
			FadeIn:        true,
			FadeOut:       true,
			BlendAdditive: true,
			PosXRand:      60,
			PosYRand:      60,
			PosZStartRand: 2,
			PosZ:          20,
			PosZEnd:       -4,
			PosZSmooth:    false,
			SizeStart:     effectTableSize(7),
			SizeEnd:       effectTableSize(7),
			Color:         color.RGBA{R: 245, G: 250, B: 255, A: 255},
		}},
	}
}

func weatherLeafEffectSpec(spriteFile string, tint color.RGBA) EffectSpec {
	return EffectSpec{
		Duration: 5200 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			SpriteFile:       spriteFile,
			SpriteRepeat:     true,
			Duration:         5200 * time.Millisecond,
			Duplicate:        36,
			AlphaMax:         0.92,
			FadeIn:           true,
			FadeOut:          true,
			PosXRand:         30,
			PosYRand:         30,
			PosZ:             20,
			PosZEnd:          0,
			PosXEnd:          -4,
			PosYEnd:          2,
			PosXEndRand:      3,
			PosYEndRand:      3,
			PosXSmooth:       true,
			PosYSmooth:       true,
			SizeStart:        effectTableSize(80),
			SizeEnd:          effectTableSize(80),
			AngleRandMin:     0,
			AngleRandMax:     360,
			RotateWithCamera: true,
			Color:            tint,
		}},
	}
}

func weatherCloudEffectSpec(textures []string, tint color.RGBA, alpha, radius, z float64, duration time.Duration) EffectSpec {
	return EffectSpec{
		Duration: duration,
		Components: []EffectComponent{{
			Kind:          EffectComponent3D,
			TextureFiles:  textures,
			Duration:      duration,
			Duplicate:     28,
			AlphaMax:      alpha,
			FadeIn:        true,
			FadeOut:       true,
			BlendAdditive: false,
			PosXRand:      radius,
			PosYRand:      radius,
			PosXEnd:       2.4,
			PosYEnd:       0.8,
			PosZ:          z,
			SizeStartX:    effectTableSize(1500),
			SizeStartY:    effectTableSize(900),
			SizeEndX:      effectTableSize(1500),
			SizeEndY:      effectTableSize(900),
			AngleRandMin:  -10,
			AngleRandMax:  10,
			Color:         tint,
		}},
	}
}

// EffectSpecs is adapted from robr's DB/Effects/EffectTable.js. Do not add
// guessed local visual behavior here; either import it from robr or leave the
// effect unsupported until the reference behavior is understood.
var EffectSpecs = map[int]EffectSpec{
	effectRain:        weatherRainEffectSpec(),
	effectSnow:        weatherSnowEffectSpec(),
	effectSakura:      weatherLeafEffectSpec("data/sprite/이팩트/sakura01", color.RGBA{R: 255, G: 210, B: 225, A: 255}),
	effectMaple:       weatherLeafEffectSpec("data/sprite/이팩트/단풍", color.RGBA{R: 255, G: 170, B: 80, A: 255}),
	effectCloud:       weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.18, 35, -10, 20*time.Second),
	effectCloud2:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.58, 35, 4, 20*time.Second),
	effectCloud3:      weatherCloudEffectSpec([]string{"effect/fog1.tga", "effect/fog2.tga", "effect/fog3.tga"}, color.RGBA{R: 120, G: 110, B: 100, A: 255}, 0.78, 45, -10, 34*time.Second),
	effectCloud4:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.58, 35, -10, 20*time.Second),
	effectCloud5:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 225, G: 212, B: 194, A: 255}, 0.70, 50, 4, 8*time.Second),
	effectCloud6:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.58, 35, 4, 28*time.Second),
	effectCloud7:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 51, G: 79, B: 161, A: 255}, 0.55, 35, 4, 20*time.Second),
	effectCloud8:      weatherCloudEffectSpec([]string{"effect/cloud4.tga", "effect/cloud1.tga", "effect/cloud2.tga"}, color.RGBA{R: 255, G: 140, B: 51, A: 255}, 0.62, 35, 4, 20*time.Second),
	effectBeginSpell:  castAuraEffectSpec("ring_yellow", color.RGBA{R: 255, G: 245, B: 120, A: 255}, 0.8, 4, 5, false),
	effectBeginSpell2: elementalCastAuraEffectSpec("ring_blue", color.RGBA{R: 128, G: 128, B: 255, A: 255}, 0.6),
	effectBeginSpell3: elementalCastAuraEffectSpec("ring_red", color.RGBA{R: 255, G: 100, B: 100, A: 255}, 0.7),
	effectBeginSpell4: elementalCastAuraEffectSpec("ring_white", color.RGBA{R: 150, G: 255, B: 150, A: 255}, 0.6),
	effectBeginSpell5: elementalCastAuraEffectSpec("ring_yellow", color.RGBA{R: 255, G: 245, B: 120, A: 255}, 0.8),
	effectBeginSpell6: castAuraEffectSpec("ring_white", color.RGBA{R: 255, G: 255, B: 255, A: 255}, 0.8, 4, 5, true),
	effectBeginSpell7: elementalCastAuraEffectSpec("ring_purple", color.RGBA{R: 200, G: 160, B: 255, A: 255}, 0.7),
	effectLockOnTarget: {
		Components: []EffectComponent{{
			Kind:           EffectComponentFUNC,
			FuncName:       "LockOnTarget",
			TextureFile:    "effect/lockon128.tga",
			AttachedEntity: true,
		}},
	},
	effectSmoke: {
		Duration: 10 * time.Second,
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			SpriteFile:       "굴뚝연기",
			SpriteRepeat:     true,
			Duration:         10 * time.Second,
			Duplicate:        10,
			DuplicateDelay:   time.Second,
			AlphaMax:         0.8,
			FadeOut:          true,
			BlendAdditive:    true,
			PosZ:             0,
			PosZEnd:          20,
			PosXEndRand:      3,
			PosXSmooth:       true,
			SizeStart:        effectTableSize(70),
			SizeEnd:          effectTableSize(300),
			SizeSmooth:       true,
			AngleStart:       -90,
			AngleEnd:         0,
			Rotate:           true,
			RotateWithCamera: true,
			Color:            color.RGBA{R: 255, G: 255, B: 255, A: 255},
		}},
	},
	effectFirefly: {
		Duration: 3 * time.Second,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "data/sprite/이팩트/particle1",
			SpriteRepeat:   true,
			Duration:       3 * time.Second,
			Duplicate:      3,
			DuplicateDelay: 420 * time.Millisecond,
			PosXRand:       0.18,
			PosYRand:       0.18,
			PosZ:           0.25,
			PosZEnd:        1.0,
			PosZStartRand:  0.2,
			PosZEndRand:    0.25,
			SizeStart:      effectTableSize(70),
			SizeEnd:        effectTableSize(110),
			AlphaMax:       0.22,
			FadeIn:         true,
			FadeOut:        true,
			Color:          color.RGBA{R: 230, G: 250, B: 255, A: 255},
		}},
	},
	effectTorch: {
		Duration: 24 * time.Hour,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "torch_01",
			SpriteRepeat:   true,
			Duration:       600 * time.Millisecond,
			SpriteDelay:    100 * time.Millisecond,
			PosX:           0.1,
			PosZ:           0.8,
			SizeStart:      effectTableSize(100),
			SizeEnd:        effectTableSize(100),
			AngleStart:     270,
			AngleEnd:       270,
			RotateToTarget: true,
			AlphaMax:       1,
			Color:          color.RGBA{R: 255, G: 255, B: 255, A: 255},
		}},
	},
	effectTorchRed:    coloredTorchEffectSpec(color.RGBA{R: 255, G: 110, B: 75, A: 255}),
	effectTorchGreen:  coloredTorchEffectSpec(color.RGBA{R: 90, G: 255, B: 135, A: 255}),
	effectTorchPurple: coloredTorchEffectSpec(color.RGBA{R: 190, G: 120, B: 255, A: 255}),
	effectBubble:      strEffectSpecRandom("bubble%d", "", 1, 4),
	effectDragonSmoke: {
		Duration: 10 * time.Second,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "poisonhit",
			SpriteRepeat:   true,
			Duration:       1100 * time.Millisecond,
			Duplicate:      9,
			DuplicateDelay: 280 * time.Millisecond,
			AlphaMax:       0.48,
			FadeIn:         true,
			FadeOut:        true,
			PosZ:           0.2,
			PosZEnd:        2.6,
			PosXEndRand:    0.8,
			PosYEndRand:    0.8,
			PosXSmooth:     true,
			PosYSmooth:     true,
			PosZSmooth:     true,
			SizeStart:      effectTableSize(90),
			SizeEnd:        effectTableSize(260),
			SizeSmooth:     true,
			BlendAdditive:  true,
			Color:          color.RGBA{R: 210, G: 210, B: 210, A: 255},
		}},
	},
	effectBanjjakii: {
		Duration: time.Second,
		Components: []EffectComponent{{
			Kind:          EffectComponent3D,
			SpriteFile:    "크리스마스",
			SpriteRepeat:  true,
			Duration:      time.Second,
			AlphaMax:      1,
			FadeIn:        true,
			FadeOut:       true,
			BlendAdditive: true,
			PosZ:          4,
			SizeStart:     effectTableSize(140),
			SizeEnd:       effectTableSize(230),
			SizeSmooth:    true,
			Color:         color.RGBA{R: 255, G: 240, B: 180, A: 255},
		}},
	},
	effectMapPillar: {
		Duration: 24 * time.Hour,
		Components: []EffectComponent{{
			Kind:             EffectComponentCylinder,
			TextureName:      "alpha_center",
			Duration:         1600 * time.Millisecond,
			Repeat:           true,
			AlphaMax:         0.42,
			Fade:             true,
			Rotate:           true,
			Animation:        1,
			BottomSize:       0.55,
			TopSize:          1.2,
			Height:           8,
			TotalCircleSides: 32,
			CircleSides:      32,
			BlendAdditive:    true,
			Color:            color.RGBA{R: 205, G: 235, B: 255, A: 255},
		}},
	},
	effectMapGhost: {
		Duration: 24 * time.Hour,
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			SpriteFile:     "particle2",
			SpriteRepeat:   true,
			Duration:       1400 * time.Millisecond,
			Duplicate:      7,
			DuplicateDelay: 190 * time.Millisecond,
			AlphaMax:       0.5,
			FadeIn:         true,
			FadeOut:        true,
			BlendAdditive:  true,
			PosZ:           0.2,
			PosZEnd:        2,
			PosXRand:       0.35,
			PosYRand:       0.35,
			SizeStart:      effectTableSize(70),
			SizeEnd:        effectTableSize(120),
			Color:          color.RGBA{R: 180, G: 230, B: 255, A: 255},
		}},
	},
	effectGlow1:      mapGlowEffectSpec(color.RGBA{R: 255, G: 245, B: 190, A: 255}, 1.2),
	effectGlow2:      mapGlowEffectSpec(color.RGBA{R: 150, G: 210, B: 255, A: 255}, 1.45),
	effectGlow4:      mapGlowEffectSpec(color.RGBA{R: 220, G: 175, B: 255, A: 255}, 1.8),
	effectBubbleDrop: mapFallingParticleEffectSpec("effect/pok3.tga", color.RGBA{R: 95, G: 180, B: 255, A: 255}, 30),
	effectRainbow: {
		Duration: 24 * time.Hour,
		Components: []EffectComponent{
			mapRainbowBand(color.RGBA{R: 255, G: 90, B: 90, A: 255}, 0),
			mapRainbowBand(color.RGBA{R: 255, G: 210, B: 90, A: 255}, 0.18),
			mapRainbowBand(color.RGBA{R: 90, G: 255, B: 120, A: 255}, 0.36),
			mapRainbowBand(color.RGBA{R: 90, G: 170, B: 255, A: 255}, 0.54),
			mapRainbowBand(color.RGBA{R: 185, G: 110, B: 255, A: 255}, 0.72),
		},
	},
	effectPneuma: strEffectSpecRandom("pneuma%d", "", 1, 3),
	effectCastRing: {
		Duration: 900 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponentFUNC,
			FuncName:         "CastRing",
			TextureName:      "ring_yellow",
			AlphaMax:         0.9,
			Fade:             true,
			Rotate:           true,
			Animation:        1,
			BottomSize:       0.8,
			TopSize:          2.45,
			Height:           2.8,
			PosZ:             0.08,
			TotalCircleSides: 20,
			CircleSides:      20,
			Color:            color.RGBA{R: 255, G: 245, B: 150, A: 255},
		}},
	},
	effectGroundSample: {
		Duration: 900 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:        EffectComponentFUNC,
			FuncName:    "MagicTarget",
			TextureFile: "effect/magic_target.tga",
			AlphaMax:    0.9,
			PosZ:        0.08,
			SizeStart:   1,
			SizeEnd:     1,
		}},
	},
	effectHit1: {
		Duration: 300 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:        EffectComponent3D,
			TextureFile: "effect/pok3.tga",
			Duration:    300 * time.Millisecond,
			Duplicate:   4,
			AlphaMax:    0.8,
			FadeIn:      true,
			FadeOut:     true,
			Sparkling:   true,
			PosZ:        1,
			PosXEndRand: 2,
			PosYEndRand: 2,
			PosZEndRand: 2,
			SizeStart:   effectTableSize(10),
			SizeEnd:     effectTableSize(10),
			SizeRand:    effectTableSize(20),
			SizeSmooth:  true,
		}},
	},
	effectBashHit: {
		Duration:   350 * time.Millisecond,
		SFX:        []string{"effect\\ef_hit2.wav"},
		Components: bashHitComponents(),
	},
	effectHit3: {
		Duration: 150 * time.Millisecond,
		SFX:      []string{"effect\\ef_hit3.wav"},
		Components: []EffectComponent{
			hitCylinderComponent(0.37, 1),
			hitCylinderComponent(0.37, 0.37),
		},
	},
	effectHit4: {
		Duration: 150 * time.Millisecond,
		SFX:      []string{"effect\\ef_hit4.wav"},
		Components: []EffectComponent{
			hitCylinderComponent(0.15, 1),
		},
	},
	effectHit5: {
		Duration: 400 * time.Millisecond,
		SFX:      []string{"effect\\ef_hit5.wav"},
		Components: []EffectComponent{
			hitSlashComponent(EffectComponent3D, effectTableSize(15), effectTableSize(200), 90, 0, false),
			hitSlashComponent(EffectComponent3D, effectTableSize(15), effectTableSize(200), 180, 90, false),
		},
	},
	effectHit6: {
		Duration: 400 * time.Millisecond,
		SFX:      []string{"effect\\ef_hit6.wav"},
		Components: []EffectComponent{
			hitSlashComponent(EffectComponent2D, effectTableSize(10), effectTableSize(150), 90, 0, true),
			hitSlashComponent(EffectComponent2D, effectTableSize(10), effectTableSize(150), 180, 90, true),
		},
	},
	effectEntry: {
		Duration: 500 * time.Millisecond,
		SFX:      []string{"effect\\ef_portal.wav"},
		Components: []EffectComponent{
			robrCylinderComponent("ring_blue", color.RGBA{}, 500*time.Millisecond, 0.62, 1, 0.9, 0.9, 7.5, false, true, true, false),
			robrCylinderComponent("ring_blue", color.RGBA{}, 500*time.Millisecond, 0.62, 1, 0.85, 0.85, 8, true, true, true, false),
			robrCylinderComponent("ring_blue", color.RGBA{}, 500*time.Millisecond, 0.8, 1, 0.9, 1.5, 1, true, true, true, false),
		},
	},
	effectExit: {
		Duration: 2000 * time.Millisecond,
		SFX:      []string{"_heal_effect.wav"},
		Components: []EffectComponent{
			robrCylinderComponent("alpha_down", color.RGBA{}, 2000*time.Millisecond, 0.2, 1, 0.95, 0.95, 10, true, false, true, true),
			healParticleComponent(0.8, 1000*time.Millisecond, 400*time.Millisecond, 80*time.Millisecond, 6, 1.5, 1.5, 0, 0, 0, 3, 6, true),
			healParticleComponent(0.8, 900*time.Millisecond, 200*time.Millisecond, 200*time.Millisecond, 3, 1, 1, 1, 0, 6, 0, 0, true),
		},
	},
	effectWarp: {
		Duration: 1000 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponentCylinder,
			TextureName:      "ring_yellow",
			Duration:         1000 * time.Millisecond,
			Duplicate:        4,
			DuplicateDelay:   300 * time.Millisecond,
			AlphaMax:         0.8,
			Fade:             true,
			Animation:        4,
			BottomSize:       10,
			TopSize:          13,
			PosZ:             0.1,
			AttachedEntity:   true,
			TotalCircleSides: 32,
			CircleSides:      32,
		}},
	},
	effectEnhance: {
		Duration: 2000 * time.Millisecond,
		Components: []EffectComponent{
			robrCylinderComponent("alpha_down", color.RGBA{}, 2000*time.Millisecond, 0.2, 1, 0.95, 0.95, 10, true, false, true, true),
			incAgilityParticleComponent(0.8, 500*time.Millisecond, 7),
			incAgilityParticleComponent(0.8, 400*time.Millisecond, 3),
		},
	},
	effectArrowShot: {
		Duration: 140 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			SpriteFile:       "data/sprite/npc/skel_archer_arrow",
			ToSrc:            true,
			RotateToTarget:   true,
			RotateWithCamera: true,
			Duration:         140 * time.Millisecond,
			AlphaMax:         1,
			FadeIn:           true,
			FadeOut:          true,
			PosZ:             1,
			SizeStart:        effectTableSize(100),
			SizeEnd:          effectTableSize(100),
			AngleStart:       180,
		}},
	},
	effectArrowShower: {
		Duration: 140 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			SpriteFile:       "data/sprite/npc/skel_archer_arrow",
			ToSrc:            true,
			RotateToTarget:   true,
			RotateWithCamera: true,
			Duration:         140 * time.Millisecond,
			Duplicate:        10,
			DuplicateDelay:   0,
			AlphaMax:         1,
			FadeIn:           true,
			FadeOut:          true,
			PosZ:             1,
			PosXEndRand:      1.5,
			PosYEndRand:      1.5,
			SizeStart:        effectTableSize(100),
			SizeEnd:          effectTableSize(100),
			AngleStart:       180,
		}},
	},
	effectSafetyWall: {
		Duration: 50 * time.Second,
		SFX:      []string{"effect\\ef_glasswall.wav"},
		Components: []EffectComponent{
			{
				Kind:    EffectComponentSTR,
				STRFile: "safetywall",
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_down",
				Duration:         50 * time.Second,
				AlphaMax:         0.4,
				Fade:             true,
				Rotate:           true,
				Animation:        0,
				BottomSize:       0.6,
				TopSize:          0.6,
				Height:           7,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            color.RGBA{R: 128, G: 25, B: 128, A: 255},
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_down",
				Duration:         50 * time.Second,
				Delay:            50 * time.Millisecond,
				AlphaMax:         0.4,
				Fade:             true,
				Rotate:           true,
				Animation:        0,
				BottomSize:       0.65,
				TopSize:          0.65,
				Height:           6,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            color.RGBA{R: 128, G: 25, B: 128, A: 255},
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_down",
				Duration:         50 * time.Second,
				Delay:            100 * time.Millisecond,
				AlphaMax:         0.4,
				Fade:             true,
				Rotate:           true,
				Animation:        0,
				BottomSize:       0.7,
				TopSize:          0.7,
				Height:           7,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            color.RGBA{R: 128, G: 25, B: 128, A: 255},
			},
		},
	},
	effectMammonite:      strEffectSpecAttachedMin("maemor", "memor_min", "effect\\ef_coin2.wav", false),
	effectCartRevolution: strEffectSpecAttached("cartrevolution", "effect\\ef_magnumbreak.wav", false),
	effectLoud:           strEffectSpecAttached("loud", "effect\\고성방가.wav", false),
	effectGlassWall:      strEffectSpec("effect/safetywall", "effect\\ef_glasswall.wav"),
	effectHealSP: {
		Duration: 2000 * time.Millisecond,
		SFX:      []string{"_heal_effect.wav"},
		Components: []EffectComponent{
			robrCylinderComponent("ring_blue", color.RGBA{R: 25, G: 128, B: 255, A: 255}, 2000*time.Millisecond, 0.2, 1, 0.95, 0.95, 10, true, true, true, true),
			tintedEffectComponent(healParticleComponent(0.8, 1000*time.Millisecond, 400*time.Millisecond, 80*time.Millisecond, 6, 1.5, 1.5, 0, 0, 0, 3, 6, true), color.RGBA{R: 25, G: 128, B: 255, A: 255}),
			tintedEffectComponent(healParticleComponent(0.8, 900*time.Millisecond, 200*time.Millisecond, 200*time.Millisecond, 3, 1, 1, 1, 0, 6, 0, 0, true), color.RGBA{R: 25, G: 128, B: 255, A: 255}),
		},
	},
	effectSight: {
		Duration: 12200 * time.Millisecond,
		SFX:      []string{"effect\\ef_sight.wav"},
		Components: []EffectComponent{
			{
				Kind:            EffectComponent3D,
				SpriteFile:      "data\\sprite\\shadow",
				ShadowTexture:   true,
				SpriteRepeat:    true,
				Duration:        12200 * time.Millisecond,
				AlphaMax:        0.5,
				PosX:            -2,
				PosZ:            4,
				SizeStart:       effectTableSize(30),
				SizeEnd:         effectTableSize(30),
				SizeDelta:       effectTableSize(10),
				OrbitRadiusX:    3,
				OrbitRadiusY:    3,
				OrbitRotations:  10,
				OrbitPhase:      0.9,
				OrbitPhaseDelta: -0.1,
				OrbitClockwise:  true,
				Duplicate:       10,
			},
			{
				Kind:            EffectComponent3D,
				SpriteFile:      "sight",
				SpriteRepeat:    true,
				Duration:        12200 * time.Millisecond,
				AlphaMax:        123.0 / 255.0,
				AlphaMaxDelta:   3.0 / 255.0,
				PosX:            -2,
				PosZ:            4,
				SizeStart:       effectTableSize(60),
				SizeEnd:         effectTableSize(60),
				SizeDelta:       effectTableSize(20),
				OrbitRadiusX:    3,
				OrbitRadiusY:    3,
				OrbitRotations:  10,
				OrbitPhase:      0.9,
				OrbitPhaseDelta: -0.1,
				OrbitClockwise:  true,
				Duplicate:       10,
			},
		},
	},
	effectNapalmBeat: {
		Duration: 700 * time.Millisecond,
		SFX:      []string{"effect\\ef_napalmbeat.wav"},
	},
	effectPokJuk: {
		Duration: 4 * time.Second,
		Components: []EffectComponent{
			{
				Kind:          EffectComponent3D,
				TextureFile:   "effect/pok3.tga",
				Duration:      900 * time.Millisecond,
				AlphaMax:      0.9,
				FadeIn:        true,
				FadeOut:       true,
				BlendAdditive: true,
				PosZ:          0,
				PosZEnd:       8,
				PosXEndRand:   1.5,
				PosYEndRand:   1.5,
				PosZSmooth:    true,
				SizeStart:     effectTableSize(8),
				SizeEnd:       effectTableSize(12),
				Color:         color.RGBA{R: 255, G: 255, B: 255, A: 255},
			},
			{
				Kind:          EffectComponent3D,
				TextureFile:   "effect/pok3.tga",
				Duration:      900 * time.Millisecond,
				Delay:         850 * time.Millisecond,
				Duplicate:     45,
				AlphaMax:      0.9,
				FadeOut:       true,
				BlendAdditive: true,
				PosZ:          8,
				PosXEndRand:   4.5,
				PosYEndRand:   4.5,
				PosZEndRand:   4.5,
				SizeStart:     effectTableSize(14),
				SizeEnd:       effectTableSize(4),
				SizeRand:      effectTableSize(4),
				Color:         color.RGBA{R: 120, G: 180, B: 255, A: 255},
			},
			{
				Kind:          EffectComponent3D,
				TextureFile:   "effect/pok3.tga",
				Duration:      900 * time.Millisecond,
				Delay:         900 * time.Millisecond,
				Duplicate:     45,
				AlphaMax:      0.85,
				FadeOut:       true,
				BlendAdditive: true,
				PosZ:          8,
				PosXEndRand:   4.0,
				PosYEndRand:   4.0,
				PosZEndRand:   4.0,
				SizeStart:     effectTableSize(12),
				SizeEnd:       effectTableSize(3),
				SizeRand:      effectTableSize(3),
				Color:         color.RGBA{R: 255, G: 120, B: 160, A: 255},
			},
			{
				Kind:          EffectComponent3D,
				TextureFile:   "effect/pok3.tga",
				Duration:      750 * time.Millisecond,
				Delay:         950 * time.Millisecond,
				Duplicate:     30,
				AlphaMax:      0.75,
				FadeOut:       true,
				BlendAdditive: true,
				PosZ:          8,
				PosXEndRand:   3.2,
				PosYEndRand:   3.2,
				PosZEndRand:   3.2,
				SizeStart:     effectTableSize(10),
				SizeEnd:       effectTableSize(2),
				SizeRand:      effectTableSize(2),
				Color:         color.RGBA{R: 255, G: 255, B: 130, A: 255},
			},
		},
	},
	effectSoulStrike: {
		Duration: 450 * time.Millisecond,
		SFX:      []string{"effect\\ef_soulstrike.wav"},
		Components: []EffectComponent{
			{
				Kind:            EffectComponent3D,
				Color:           color.RGBA{R: 255, G: 255, B: 255, A: 255},
				TextureFile:     "effect/pok3.tga",
				Duration:        200 * time.Millisecond,
				Delay:           250 * time.Millisecond,
				DuplicateDelay:  150 * time.Millisecond,
				ToSrc:           true,
				AlphaMax:        1,
				FadeIn:          true,
				FadeOut:         true,
				PosZEnd:         1,
				PosZStartRand:   5,
				PosZStartMiddle: 6,
				PosZSmooth:      true,
				SizeStart:       50 * EffectPixelRatio,
				SizeEnd:         50 * EffectPixelRatio,
			},
			{
				Kind:           EffectComponent3D,
				SpriteFile:     "data/sprite/이팩트/particle1",
				SpriteRepeat:   true,
				Duration:       250 * time.Millisecond,
				Duplicate:      5,
				DuplicateDelay: 20 * time.Millisecond,
				ToSrc:          true,
				RotateToTarget: true,
				Arc:            4,
				Retreat:        3,
				PosZ:           3,
				SizeStart:      100 * EffectPixelRatio,
				SizeEnd:        effectTableSize(500),
			},
		},
	},
	effectEndure: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_endure.wav"},
		Components: []EffectComponent{{
			Kind:        EffectComponent3D,
			TextureFile: "effect/endure.tga",
			Duration:    1000 * time.Millisecond,
			AlphaMax:    1,
			FadeIn:      true,
			FadeOut:     true,
			PosZ:        2,
			SizeStart:   200 * EffectPixelRatio,
			SizeEnd:     70 * EffectPixelRatio,
			SizeSmooth:  true,
		}},
	},
	effectBashBegin: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_bash.wav"},
		Components: []EffectComponent{
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_down",
				Duration:         1000 * time.Millisecond,
				AlphaMax:         0.6,
				Fade:             true,
				Rotate:           true,
				FixedPerspective: true,
				Animation:        2,
				BottomSize:       0.1,
				TopSize:          2.0,
				PosZ:             1.5,
				TotalCircleSides: 20,
				CircleSides:      20,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_center",
				Duration:         1000 * time.Millisecond,
				AlphaMax:         0.6,
				Fade:             true,
				Rotate:           true,
				FixedPerspective: true,
				Animation:        2,
				BottomSize:       0.01,
				TopSize:          2.5,
				PosZ:             1.5,
				TotalCircleSides: 30,
				CircleSides:      1,
				Duplicate:        10,
				AngleZRandom:     360,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "alpha_center",
				Duration:         1000 * time.Millisecond,
				AlphaMax:         0.6,
				Fade:             true,
				Rotate:           true,
				FixedPerspective: true,
				Animation:        2,
				BottomSize:       0.01,
				TopSize:          4.0,
				PosZ:             1.5,
				TotalCircleSides: 30,
				CircleSides:      1,
				Duplicate:        8,
				AngleZRandom:     360,
			},
		},
	},
	effectProvoke: {
		Duration: 900 * time.Millisecond,
		SFX:      []string{"effect\\swordman_provoke.wav"},
		Components: []EffectComponent{{
			Kind:    EffectComponentSTR,
			Color:   color.RGBA{R: 255, G: 70, B: 42, A: 255},
			STRFile: "provoke",
		}},
	},
	effectMagnumBreak: {
		Duration: 300 * time.Millisecond,
		SFX:      []string{"effect\\ef_magnumbreak.wav"},
		Components: []EffectComponent{
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "ring_yellow",
				Duration:         300 * time.Millisecond,
				AlphaMax:         0.7,
				Fade:             true,
				Rotate:           true,
				Animation:        4,
				BottomSize:       4,
				TopSize:          6,
				Height:           1,
				TotalCircleSides: 32,
				CircleSides:      32,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "대폭발",
				Duration:         300 * time.Millisecond,
				AlphaMax:         0.6,
				Fade:             true,
				Rotate:           true,
				Animation:        4,
				BottomSize:       4,
				TopSize:          1,
				Height:           4,
				TotalCircleSides: 32,
				CircleSides:      32,
			},
		},
	},
	effectQuakeMagnum: {
		Duration:    50 * time.Millisecond,
		CameraShake: 50 * time.Millisecond,
	},
	effectSteal: {
		Duration: 500 * time.Millisecond,
		SFX:      []string{"effect\\ef_steal.wav"},
		Components: []EffectComponent{{
			Kind:          EffectComponent3D,
			Color:         color.RGBA{R: 255, G: 255, B: 216, A: 255},
			TextureFile:   "effect/pok1.tga",
			Duration:      500 * time.Millisecond,
			AlphaMax:      1,
			FadeOut:       true,
			PosXEndRand:   3.5,
			PosYEndRand:   3.5,
			PosZEndRand:   1,
			PosZEndMiddle: 3,
			SizeStart:     200 * EffectPixelRatio,
			SizeEnd:       10 * EffectPixelRatio,
			Duplicate:     7,
		}},
	},
	effectThrowItem3: {
		Duration: 200 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			TextureFile:      "유저인터페이스/item/돌.bmp",
			Duration:         200 * time.Millisecond,
			AlphaMax:         1,
			FadeIn:           true,
			FadeOut:          true,
			ToSrc:            true,
			RotateToTarget:   true,
			RotateWithCamera: true,
			Rotate:           true,
			AngleStart:       180,
			AngleEnd:         360,
			PosZ:             1,
			SizeStart:        effectTableSize(20),
			SizeEnd:          effectTableSize(20),
		}},
	},
	effectSummonSlave: {
		Components: []EffectComponent{{
			Kind:           EffectComponentSPR,
			SpriteFile:     "smoke",
			AttachedEntity: true,
		}},
	},
	effectPoisonAttack: {
		Duration: 2800 * time.Millisecond,
		SFX:      []string{"effect\\ef_detoxication.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			Color:          color.RGBA{R: 255, G: 178, B: 255, A: 255},
			TextureFile:    "effect/pok1.tga",
			Duration:       1000 * time.Millisecond,
			AlphaMax:       1,
			FadeIn:         true,
			FadeOut:        true,
			PosXRand:       1,
			PosYRand:       1,
			PosZEnd:        5,
			SizeStart:      100 * EffectPixelRatio,
			SizeEnd:        100 * EffectPixelRatio,
			SizeRand:       20 * EffectPixelRatio,
			Duplicate:      10,
			DuplicateDelay: 200 * time.Millisecond,
		}},
	},
	effectDetoxication: {
		Duration: 2800 * time.Millisecond,
		SFX:      []string{"effect\\ef_detoxication.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			Color:          color.RGBA{R: 178, G: 255, B: 255, A: 255},
			TextureFile:    "effect/pok1.tga",
			Duration:       1000 * time.Millisecond,
			AlphaMax:       1,
			FadeIn:         true,
			FadeOut:        true,
			PosXRand:       1,
			PosYRand:       1,
			PosZEnd:        5,
			SizeStart:      100 * EffectPixelRatio,
			SizeEnd:        100 * EffectPixelRatio,
			SizeRand:       20 * EffectPixelRatio,
			Duplicate:      10,
			DuplicateDelay: 200 * time.Millisecond,
		}},
	},
	effectStoneCurse: strEffectSpecAttached("stonecurse", "", false),
	effectIceArrow: {
		Duration:   500 * time.Millisecond,
		SFX:        []string{"effect\\ef_icearrow%d.wav"},
		SFXRandMin: 1,
		SFXRandMax: 3,
	},
	effectColdBolt: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_icearrow1.wav", "effect\\ef_icearrow2.wav", "effect\\ef_icearrow3.wav"},
		Components: []EffectComponent{
			{
				Kind:            EffectComponent3D,
				TextureFile:     "effect/icearrow.tga",
				Duration:        500 * time.Millisecond,
				AlphaMax:        1,
				PosXStartMiddle: 5,
				PosXStartRand:   1,
				PosYStartMiddle: 2,
				PosYStartRand:   1,
				PosZ:            20,
				PosXEnd:         0.0001,
				PosYEnd:         0.0001,
				PosZEnd:         0.0001,
				SizeStart:       50 * EffectPixelRatio,
				SizeEnd:         50 * EffectPixelRatio,
				AngleStart:      112.5,
				AngleEnd:        112.5,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      "ring_blue",
				Duration:         1000 * time.Millisecond,
				Delay:            500 * time.Millisecond,
				AlphaMax:         0.7,
				Fade:             true,
				Rotate:           true,
				Animation:        4,
				BottomSize:       3,
				TopSize:          5,
				Height:           0.1,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            color.RGBA{R: 128, G: 170, B: 255, A: 255},
			},
		},
	},
	effectFireBolt: {
		Duration: 500 * time.Millisecond,
		SFX:      []string{"effect\\ef_firearrow1.wav", "effect\\ef_firearrow2.wav", "effect\\ef_firearrow3.wav"},
		Components: []EffectComponent{{
			Kind: EffectComponent3D,
			TextureFiles: []string{
				"effect/불화살1.tga",
				"effect/불화살2.tga",
				"effect/불화살3.tga",
				"effect/불화살4.tga",
				"effect/불화살5.tga",
				"effect/불화살6.tga",
			},
			Duration:        500 * time.Millisecond,
			AlphaMax:        1,
			BlendAdditive:   true,
			PosXStartMiddle: 5,
			PosXStartRand:   1,
			PosYStartMiddle: 2,
			PosYStartRand:   1,
			PosZ:            20,
			PosXEnd:         0.0001,
			PosYEnd:         0.0001,
			PosZEnd:         0.0001,
			SizeStartX:      100 * EffectPixelRatio,
			SizeEndX:        100 * EffectPixelRatio,
			SizeStartY:      50 * EffectPixelRatio,
			SizeEndY:        50 * EffectPixelRatio,
			AngleStart:      112.5,
			AngleEnd:        112.5,
		}},
	},
	effectFireBall: {
		Duration: 410 * time.Millisecond,
		SFX:      []string{"effect\\ef_fireball.wav"},
		Components: []EffectComponent{{
			Kind:             EffectComponent3D,
			SpriteFile:       "fireball",
			SpriteRepeat:     true,
			ToSrc:            true,
			RotateToTarget:   true,
			Duration:         250 * time.Millisecond,
			Delay:            160 * time.Millisecond,
			DelayOffsetDelta: -40 * time.Millisecond,
			AlphaMax:         0.2,
			AlphaMaxDelta:    0.2,
			RotateWithCamera: true,
			PosZ:             2,
			SizeStart:        200 * EffectPixelRatio,
			SizeEnd:          200 * EffectPixelRatio,
			Duplicate:        5,
		}},
	},
	effectFireWall: strEffectSpecRandom("firewall%d", "effect\\ef_firewall.wav", 1, 2),
	effectFrostDiver: {
		Duration: 450 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:        EffectComponent3D,
			TextureFile: "effect/ice.tga",
			Duration:    450 * time.Millisecond,
			ToSrc:       true,
			AlphaMax:    0.75,
			FadeOut:     true,
			PosZ:        0.7,
			PosZEnd:     0.9,
			SizeStart:   90 * EffectPixelRatio,
			SizeEnd:     130 * EffectPixelRatio,
			AngleStart:  0,
			AngleEnd:    -180,
		}},
	},
	effectFrostDiverHit: strEffectSpecAttached("freeze", "effect\\ef_frostdiver2.wav", false),
	effectLightningBolt: {
		Components: []EffectComponent{
			{
				Kind:           EffectComponentSTR,
				STRFile:        "lightning",
				AttachedEntity: true,
			},
			{
				Kind:           EffectComponentSTR,
				STRFile:        "windhit%d",
				STRRandMin:     1,
				STRRandMax:     3,
				AttachedEntity: true,
			},
		},
	},
	effectThunderStorm: strEffectSpec("thunderstorm", "effect\\magician_thunderstorm.wav"),
	effectFireArrow:    soundOnlyEffectSpec("effect\\ef_firearrow1.wav"),
	effectTeleportOld: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_teleportation.wav"},
		Components: []EffectComponent{
			robrCylinderComponent("ring_blue", color.RGBA{R: 255, G: 255, B: 255, A: 255}, 1000*time.Millisecond, 0.5, 5, 0.8, 0.7, 35, true, true, true, false),
		},
	},
	effectReadyPortalOld: {
		Duration: 25000 * time.Millisecond,
		SFX:      []string{"effect\\ef_readyportal.wav"},
		Components: []EffectComponent{
			robrCylinderComponent("alpha_down", color.RGBA{R: 178, G: 178, B: 255, A: 255}, 25000*time.Millisecond, 0.6, 0, 0.6, 0.6, 15, true, true, true, false),
		},
	},
	effectRuwach: {
		Duration: 12200 * time.Millisecond,
		SFX:      []string{"effect\\ef_ruwach.wav"},
		Components: []EffectComponent{
			{
				Kind:            EffectComponent3D,
				SpriteFile:      "data\\sprite\\shadow",
				ShadowTexture:   true,
				Duration:        12200 * time.Millisecond,
				AlphaMax:        0.5,
				PosX:            -2,
				SizeStart:       effectTableSize(0.4),
				SizeEnd:         effectTableSize(0.4),
				SizeDelta:       0.15,
				Duplicate:       8,
				OrbitRadiusX:    3,
				OrbitRadiusY:    3,
				OrbitRotations:  8,
				OrbitPhase:      0.7,
				OrbitPhaseDelta: -0.1,
				OrbitClockwise:  true,
			},
			{
				Kind:            EffectComponent3D,
				Color:           color.RGBA{G: 255, B: 255, A: 255},
				SpriteFile:      "particle2",
				Duration:        12200 * time.Millisecond,
				AlphaMax:        0.07,
				AlphaMaxDelta:   0.07,
				BlendAdditive:   true,
				PosX:            -2,
				PosZ:            2,
				SizeStart:       effectTableSize(80),
				SizeEnd:         effectTableSize(80),
				SizeDelta:       30,
				Duplicate:       10,
				OrbitRadiusX:    3,
				OrbitRadiusY:    3,
				OrbitRotations:  8,
				OrbitPhase:      0.7,
				OrbitPhaseDelta: -0.1,
				OrbitClockwise:  true,
			},
		},
	},
	effectIncAgility: {
		Duration: 1500 * time.Millisecond,
		SFX:      []string{"effect\\ef_incagility.wav"},
		Components: []EffectComponent{
			incAgilityParticleComponent(1, 500*time.Millisecond, 7),
			incAgilityParticleComponent(0.75, 400*time.Millisecond, 3),
			incAgilityParticleComponent(1, 0, 10),
			{
				Kind:        EffectComponent3D,
				Color:       color.RGBA{R: 255, G: 255, B: 255, A: 255},
				TextureFile: "effect/agi_up.bmp",
				Duration:    1000 * time.Millisecond,
				AlphaMax:    1,
				FadeIn:      true,
				FadeOut:     true,
				PosZ:        0.4,
				PosZEnd:     3,
				SizeStart:   100 * EffectPixelRatio,
				SizeEnd:     100 * EffectPixelRatio,
				SizeStartY:  45 * EffectPixelRatio,
				SizeEndY:    45 * EffectPixelRatio,
				SizeSmooth:  true,
				Overlay:     true,
			},
		},
	},
	effectDecAgility: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_decagility.wav"},
		Components: []EffectComponent{
			decAgilityParticleComponent(),
			{
				Kind:        EffectComponent3D,
				TextureFile: "effect/slow.bmp",
				Duration:    1000 * time.Millisecond,
				AlphaMax:    1,
				FadeIn:      true,
				FadeOut:     true,
				PosZ:        2.8,
				PosZEnd:     0.4,
				SizeStart:   effectTableSize(100),
				SizeEnd:     effectTableSize(100),
				SizeStartY:  effectTableSize(45),
				SizeEndY:    effectTableSize(45),
				SizeSmooth:  true,
			},
		},
	},
	effectIncAgiDex: {
		Duration: 1000 * time.Millisecond,
		SFX:      []string{"effect\\ef_incagidex.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponent3D,
			TextureFile:    "effect/dex_agi_up.bmp",
			Duration:       1000 * time.Millisecond,
			AlphaMax:       1,
			FadeIn:         true,
			FadeOut:        true,
			PosZ:           0.4,
			PosZEnd:        3,
			SizeStart:      100 * EffectPixelRatio,
			SizeEnd:        100 * EffectPixelRatio,
			SizeStartY:     45 * EffectPixelRatio,
			SizeEndY:       45 * EffectPixelRatio,
			SizeSmooth:     true,
			AttachedEntity: true,
			Overlay:        true,
		}},
	},
	effectAqua: {
		SFX: []string{"effect\\ef_aqua.wav"},
		Components: []EffectComponent{{
			Kind:             EffectComponentSPR,
			SpriteFile:       "성수뜨기",
			SpriteHead:       true,
			WorldSizedSprite: true,
		}},
	},
	effectSignum:  strEffectSpecAttached("cross", "effect\\ef_signum.wav", false),
	effectAngelus: strEffectSpecAttachedMin("angelus", "jong_mini", "effect\\ef_angelus.wav", true),
	effectGloria:  strEffectSpecAttachedMin("gloria", "gloria_min", "effect\\priest_gloria.wav", false),
	effectBlessing: {
		Duration: 2500 * time.Millisecond,
		SFX:      []string{"effect\\ef_blessing.wav"},
		Components: []EffectComponent{
			{
				Kind:             EffectComponentSPR,
				SpriteFile:       "축복",
				Duration:         1500 * time.Millisecond,
				SpriteDelay:      30 * time.Millisecond,
				SpriteRepeat:     true,
				SpriteHead:       true,
				SpriteYOffset:    -120,
				WorldSizedSprite: true,
			},
			{
				Kind:            EffectComponent3D,
				SpriteFile:      "particle6",
				Duration:        1200 * time.Millisecond,
				Delay:           300 * time.Millisecond,
				Duplicate:       6,
				DuplicateDelay:  0,
				AlphaMax:        1,
				Sparkling:       true,
				SparkNumber:     2,
				FadeIn:          true,
				FadeOut:         true,
				PosXRand:        1.2,
				PosYRand:        1,
				PosZStartRand:   2,
				PosZStartMiddle: 5.5,
				PosZEndRand:     0.5,
				PosZEndMiddle:   1,
				SizeStart:       50 * EffectPixelRatio,
				SizeEnd:         50 * EffectPixelRatio,
			},
			{
				Kind:            EffectComponent3D,
				SpriteFile:      "particle6",
				Duration:        1200 * time.Millisecond,
				Delay:           400 * time.Millisecond,
				Duplicate:       6,
				DuplicateDelay:  0,
				AlphaMax:        1,
				FadeIn:          true,
				FadeOut:         true,
				PosXRand:        1.4,
				PosYRand:        1.1,
				PosZStartRand:   2,
				PosZStartMiddle: 5.5,
				PosZEndRand:     0.5,
				PosZEndMiddle:   1,
				SizeStart:       50 * EffectPixelRatio,
				SizeEnd:         50 * EffectPixelRatio,
			},
			{
				Kind:          EffectComponent3D,
				Color:         color.RGBA{R: 25, G: 191, B: 255, A: 255},
				TextureFile:   "effect/pok2.tga",
				Duration:      2500 * time.Millisecond,
				AlphaMax:      0.3,
				FadeIn:        true,
				FadeOut:       true,
				SizeStart:     140 * EffectPixelRatio,
				SizeEnd:       140 * EffectPixelRatio,
				BlendAdditive: true,
			},
		},
	},
	effectFireHit: strEffectSpecRandomAttached("firehit%d", "effect\\ef_firehit.wav", 1, 3, true, false),
	effectFireSplashHit: {
		Duration: 500 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:        EffectComponent2D,
			TextureFile: "effect/firering.tga",
			Duration:    500 * time.Millisecond,
			FadeOut:     true,
			PosZ:        1,
			SizeStart:   10 * EffectPixelRatio,
			SizeEnd:     300 * EffectPixelRatio,
			AngleStart:  0,
			AngleEnd:    -360,
		}},
	},
	effectColdHit: soundOnlyEffectSpec("_hit_fist3.wav", "_hit_fist4.wav"),
	effectWindHit: strEffectSpecRandomAttached("windhit%d", "_hit_fist%d.wav", 1, 3, true, false),
	effectPoisonHit: {
		SFX: []string{"effect\\ef_poisonattack.wav"},
		Components: []EffectComponent{{
			Kind:       EffectComponentSPR,
			SpriteFile: "poisonhit",
		}},
	},
	effectWarpZone:       warpZoneEffectSpec(),
	effectSightTrasher:   sightTrasherEffectSpec(),
	effectArrowShotRO:    strEffectSpecAttached("arrowshot", "", false),
	effectInvenom:        strEffectSpecAttached("invenom", "effect\\thief_invenom.wav", false),
	effectSkidTrap:       strEffectSpec("skidtrap", "effect\\hunter_skidtrap.wav"),
	effectBrandishSpear:  strEffectSpec("brandish", "effect\\knight_brandish_spear.wav"),
	effectCure:           strEffectSpecAttachedMin("cure", "cure_min", "effect\\acolyte_cure.wav", false),
	effectMvp:            strEffectSpecAttached("mvp", "effect\\st_mvp.wav", false),
	effectIceWall:        iceWallEffectSpec(),
	effectMagnificat:     strEffectSpecAttachedMin("magnificat", "magnificat_min", "effect\\priest_magnificat.wav", false),
	effectResurrection:   strEffectSpecAttachedMin("resurrection", "resurrection_min", "effect\\priest_resurrection.wav", false),
	effectRecovery:       strEffectSpecAttached("recovery", "effect\\priest_recovery.wav", false),
	effectEarthSpike:     earthSpikeEffectSpec(),
	effectSpearBoomerang: soundOnlyEffectSpec("effect\\ef_fireball.wav"),
	effectPierce:         soundOnlyEffectSpec("effect\\ef_bash.wav"),
	effectTurnUndead:     soundOnlyEffectSpec("effect\\ef_bash.wav"),
	effectSanctuary:      strEffectSpecAttached("sanctuary", "effect\\priest_sanctuary.wav", false),
	effectImpositio:      strEffectSpecAttached("impositio", "effect\\priest_impositio.wav", false),
	effectLexAeterna:     strEffectSpecAttachedMin("lexaeterna", "lexaeterna_min", "effect\\priest_lexaeterna.wav", false),
	effectAspersio:       strEffectSpecAttached("aspersio", "effect\\priest_aspersio.wav", false),
	effectLexDivina:      strEffectSpecAttached("lexdivina", "effect\\priest_lexdivina.wav", false),
	effectSuffragium:     strEffectSpecAttachedMin("suffragium", "suffragium_min", "effect\\priest_suffragium.wav", false),
	effectStormGust:      strEffectSpecAttachedMin("stormgust", "storm_min", "effect\\wizard_stormgust.wav", false),
	effectLordVermilion:  strEffectSpecAttached("lord", "effect\\wizard_fire_ivy.wav", false),
	effectBenedictio:     strEffectSpecAttached("benedictio", "effect\\priest_benedictio.wav", false),
	effectMeteorStorm: {
		CameraShake:      650 * time.Millisecond,
		CameraShakeDelay: 600 * time.Millisecond,
		SFX:              []string{"effect\\wizard_meteor.wav"},
		SFXRandMin:       1,
		SFXRandMax:       4,
		Components: []EffectComponent{{
			Kind:           EffectComponentSTR,
			STRFile:        "meteor%d",
			STRRandMin:     1,
			STRRandMax:     4,
			AttachedEntity: true,
		}},
	},
	effectJupitelThunder: jupitelThunderEffectSpec(),
	effectJupitelHit:     jupitelHitEffectSpec(),
	effectQuagmire:       strEffectSpec("quagmire", "effect\\wizard_quagmire.wav"),
	effectFirePillar:     strEffectSpec("firepillar", "effect\\wizard_fire_pillar_a.wav"),
	effectFirePillarBomb: strEffectSpec("firepillarbomb", "effect\\wizard_fire_pillar_b.wav"),
	effectHasteUp:        soundOnlyEffectSpec("effect\\black_adrenalinerush_b.wav"),
	effectFlasher:        soundOnlyEffectSpec("effect\\hunter_flasher.wav"),
	effectRemoveTrap:     soundOnlyEffectSpec("effect\\hunter_removetrap.wav"),
	effectRepairWeapon:   repairWeaponEffectSpec(),
	effectCrashEarth:     crashEarthEffectSpec(),
	effectBlastMine:      soundOnlyEffectSpec("effect\\hun_anklesnare.wav"),
	effectBlastMineBomb:  strEffectSpec("blastmine", "effect\\hunter_blastmine.wav"),
	effectClaymore:       strEffectSpec("claymore", "effect\\hunter_claymoretrap.wav"),
	effectFreezingTrap:   strEffectSpec("freezing", "effect\\hunter_freezingtrap.wav"),
	effectGasPush:        strEffectSpec("gaspush", ""),
	effectSpringTrap:     strEffectSpec("spring", "effect\\hunter_springtrap.wav"),
	effectWeaponPerfect:  strEffectSpecAttachedMin("weaponperfection", "weaponperfection_min", "effect\\black_weapon_perfection.wav", false),
	effectMaximizePower:  strEffectSpecAttachedMin("maximizepower", "maximize_min", "", false),
	effectKyrie:          strEffectSpecAttachedMin("kyrie", "kyrie_min", "effect\\priest_kyrie_eleison_a.wav", false),
	effectMagnus:         strEffectSpec("magnus", "effect\\priest_magnus.wav"),
	effectBlitzBeat:      soundOnlyEffectSpec("effect\\hunter_blitzbeat.wav"),
	effectWaterBall:      waterBallEffectSpec(),
	effectWaterBall2:     waterBall2EffectSpec(),
	effectDetecting:      soundOnlyEffectSpec("effect\\hunter_detecting.wav"),
	effectCloaking:       soundOnlyEffectSpec("effect\\assasin_cloaking.wav"),
	effectSonicBlow:      sonicBlowEffectSpec(),
	effectSonicBlowHit:   sonicBlowHitEffectSpec(),
	effectGrimtooth:      soundOnlyEffectSpec("effect\\ef_frostdiver.wav"),
	effectVenomDust:      strEffectSpec("venomdust", "effect\\assasin_poisonreact.wav"),
	effectPoisonReact:    strEffectSpecAttached("poisonreact_1st", "effect\\assasin_poisonreact.wav", false),
	effectPoisonReact2:   strEffectSpecAttached("poisonreact", "effect\\assasin_poisonreact.wav", false),
	effectOverthrust:     soundOnlyEffectSpec("effect\\black_overthrust.wav"),
	effectVenomSplasher:  strEffectSpecAttached("venomsplasher", "effect\\assasin_venomsplasher.wav", false),
	effectTwoHandQuicken: strEffectSpecAttached("twohand", "effect\\knight_twohandquicken.wav", true),
	effectAutoCounter:    strEffectSpecAttached("autocounter", "effect\\knight_autocounter.wav", false),
	effectGrimtoothAtk:   grimtoothAttackEffectSpec(),
	effectFreeze:         strEffectSpecAttached("freeze", "", false),
	effectFreezed:        strEffectSpecAttached("freezed", "", false),
	effectIceCrash:       strEffectSpecAttached("icecrash", "", false),
	effectSlowPoison:     strEffectSpec("slowp", "effect\\priest_slowpoison.wav"),
	effectFirePillarOn:   firePillarOnEffectSpec(),
	effectSandman:        strEffectSpec("sandman", "effect\\hunter_sandman.wav"),
	effectRevive:         soundOnlyEffectSpec("effect\\priest_resurrection.wav"),
	effectHeavenDrive:    heavenDriveEffectSpec(),
	effectSonicBlow2:     strEffectSpecAttached("sonicblow", "", false),
	effectBrandishSpear2: strEffectSpecAttached("brandish2", "effect\\knight_brandish_spear.wav", false),
	effectShockwave: {
		SFX: []string{"effect\\hunter_shockwavetrap.wav"},
		Components: []EffectComponent{{
			Kind:           EffectComponentSPR,
			SpriteFile:     "shockwave",
			AttachedEntity: true,
		}},
	},
	effectShockwaveHit:  strEffectSpecAttached("shockwavehit", "", false),
	effectEarthHit:      strEffectSpecAttached("earthhit", "", false),
	effectPierceSelf:    strEffectSpecAttached("pierce", "", false),
	effectBowlingSelf:   strEffectSpecAttached("bowling", "_enemy_hit_normal1.wav", true),
	effectSpearStabSelf: strEffectSpecAttached("spearstab", "_enemy_hit_normal1.wav", false),
	effectHolyLight:     strEffectSpecAttached("holyhit", "", false),
	effectConcentration: strEffectSpecAttached("concentration", "effect\\ac_concentration.wav", false),
	effectRefineOK:      strEffectSpecAttached("bs_refinesuccess", "effect\\bs_refinesuccess.wav", false),
	effectRefineFail:    strEffectSpecAttached("bs_refinefailed", "effect\\bs_refinefailed.wav", false),
	effectEnergyCoat:    strEffectSpecAttached("energycoat", "", false),
	effectFirstAid: {
		Duration: time.Second,
		SFX:      []string{"_heal_effect.wav"},
		Components: []EffectComponent{{
			Kind:          EffectComponent2D,
			TextureFile:   "effect/pikapika2.bmp",
			Duration:      time.Second,
			AlphaMax:      0.2,
			BlendAdditive: true,
			FadeOut:       true,
			PosZ:          2,
			SizeStart:     100 * EffectPixelRatio,
			SizeEnd:       100 * EffectPixelRatio,
		}},
	},
	effectTeleportation: {
		Duration:         1500 * time.Millisecond,
		DetachLocalActor: true,
		SFX:              []string{"effect\\ef_teleportation.wav"},
		Components: []EffectComponent{
			teleportCylinderComponent(0.3, 0.3, 35),
			teleportCylinderComponent(0.6, 0.8, 25),
			teleportCylinderComponent(0.8, 1.0, 13),
			teleportCylinderComponent(1.0, 1.3, 5),
		},
	},
	effectReadyPortal: {
		Duration: 500 * time.Millisecond,
		SFX:      []string{"effect\\ef_readyportal.wav"},
		Components: []EffectComponent{
			readyPortalCylinderComponent(),
		},
	},
	effectPortal: {
		Duration: 25000 * time.Millisecond,
		SFX:      []string{"effect\\ef_readyportal.wav", "effect\\ef_portal.wav"},
		Components: []EffectComponent{
			readyPortalCylinderComponent(),
			portalCylinderComponent(0.6, 0.6, 15, 0, "ring_blue", 0.3),
			portalCylinderComponent(0.8, 0.8, 13, 0, "ring_blue", 0.3),
			{
				Kind:             EffectComponentCylinder,
				Color:            color.RGBA{R: 153, G: 153, B: 255, A: 255},
				TextureName:      "alpha1",
				Duration:         25000 * time.Millisecond,
				AlphaMax:         0.5,
				Fade:             true,
				Rotate:           true,
				Animation:        0,
				BottomSize:       1,
				TopSize:          1,
				Height:           1,
				PosZ:             2,
				TotalCircleSides: 20,
				CircleSides:      10,
				BlendAdditive:    true,
			},
		},
	},
	effectPharmacyOK:   strEffectSpecAttached("p_success", "effect\\p_success.wav", false),
	effectPharmacyFail: strEffectSpecAttached("p_failed", "effect\\p_failed.wav", false),
	effectHeal: {
		Duration: 1840 * time.Millisecond,
		SFX:      []string{"_heal_effect.wav"},
		Components: []EffectComponent{
			healCylinderComponent(0.95, 0.95, 8),
			healCylinderComponent(1.0, 1.0, 8),
			healParticleComponent(0.6, 1300*time.Millisecond, 400*time.Millisecond, 10*time.Millisecond, 15, 1.5, 1.5, 0, 0, 0, 2, 6, false),
			healParticleComponent(0.6, 1100*time.Millisecond, 200*time.Millisecond, 50*time.Millisecond, 7, 1, 1, 1, 0, 5, 0, 0, false),
		},
	},
	effectHealOffensive: {
		Duration: 1490 * time.Millisecond,
		SFX:      []string{"_heal_effect.wav"},
		Components: []EffectComponent{
			healOffensiveCylinderComponent(0.95, 0.95, 10),
			healOffensiveCylinderComponent(1.0, 1.0, 9),
			healParticleComponent(0.8, 1000*time.Millisecond, 400*time.Millisecond, 10*time.Millisecond, 10, 1.5, 1.5, 0, 0, 0, 3, 6, true),
			healParticleComponent(0.8, 900*time.Millisecond, 200*time.Millisecond, 50*time.Millisecond, 5, 1, 1, 1, 0, 6, 0, 0, true),
		},
	},
	effectBaseLevelUp: {
		Duration: 1300 * time.Millisecond,
		SFX:      []string{"levelup.wav"},
		Components: []EffectComponent{{
			Kind:    EffectComponentSTR,
			STRFile: "angel",
		}},
	},
	effectJobLevelUp: {
		Components: []EffectComponent{{
			Kind:           EffectComponentSTR,
			STRFile:        "joblvup",
			AttachedEntity: true,
		}},
	},
	effectPotionRed:    potionEffectSpec("빨간포션", color.RGBA{R: 255, G: 82, B: 70, A: 255}),
	effectPotionOrange: potionEffectSpec("주홍포션", color.RGBA{R: 255, G: 145, B: 58, A: 255}),
	effectPotionYellow: potionEffectSpec("노란포션", color.RGBA{R: 255, G: 226, B: 76, A: 255}),
	effectPotionWhite:  potionEffectSpec("하얀포션", color.RGBA{R: 245, G: 245, B: 255, A: 255}),
	effectPotionBlue:   bluePotionEffectSpec(),
	effectPotionGreen:  potionEffectSpec("초록포션", color.RGBA{R: 78, G: 225, B: 98, A: 255}),
	effectItemFast:     strEffectSpecAttached("집중", "effect\\ac_concentration.wav", false),
	effectItemFast2:    strEffectSpecAttached("각성", "effect\\ac_concentration.wav", false),
	effectItemFast3:    berserkPotionEffectSpec(),
	effectFood: {
		Duration: 850 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:    EffectComponentSTR,
			Color:   color.RGBA{R: 255, G: 182, B: 86, A: 255},
			STRFile: "fruit",
		}},
	},
	effectFoodBlue: {
		Duration: 850 * time.Millisecond,
		Components: []EffectComponent{{
			Kind:    EffectComponentSTR,
			Color:   color.RGBA{R: 132, G: 112, B: 255, A: 255},
			STRFile: "fruit",
		}},
	},
	effectChristmasCarol: strEffectSpecAttachedMin("angelus", "jong_mini", "effect\\wewish.wav", false),
}

func bashHitComponents() []EffectComponent {
	angleRanges := [][2]float64{
		{0, 35},
		{50, 85},
		{100, 135},
		{150, 185},
		{200, 235},
		{255, 290},
		{300, 335},
		{340, 360},
	}
	components := make([]EffectComponent, 0, len(angleRanges))
	for index, angleRange := range angleRanges {
		textureFile := "effect/lens1.tga"
		if index%2 == 1 {
			textureFile = "effect/lens2.tga"
		}
		components = append(components, EffectComponent{
			Kind:               EffectComponent2D,
			TextureFile:        textureFile,
			Duration:           250 * time.Millisecond,
			DurationRandMin:    200 * time.Millisecond,
			DurationRandMax:    350 * time.Millisecond,
			AlphaMax:           12,
			Fade:               true,
			FadeOut:            true,
			SizeStartXRandMin:  25 * EffectPixelRatio,
			SizeStartXRandMax:  40 * EffectPixelRatio,
			SizeStartY:         10 * EffectPixelRatio,
			SizeEndX:           1 * EffectPixelRatio,
			SizeEndYRandMin:    250 * EffectPixelRatio,
			SizeEndYRandMax:    300 * EffectPixelRatio,
			AngleRandMin:       angleRange[0],
			AngleRandMax:       angleRange[1],
			CirclePattern:      true,
			CircleInnerSize:    2.2,
			CircleOuterRandMin: 5,
			CircleOuterRandMax: 6,
			Overlay:            true,
		})
	}
	return components
}

func hitCylinderComponent(bottomSize, topSize float64) EffectComponent {
	return EffectComponent{
		Kind:             EffectComponentCylinder,
		TextureName:      "lens2",
		Duration:         150 * time.Millisecond,
		AlphaMax:         0.8,
		Fade:             true,
		RotateWithCamera: true,
		Animation:        1,
		BottomSize:       bottomSize,
		TopSize:          topSize,
		Height:           4,
		PosZ:             1,
		AngleX:           -90,
		AttachedEntity:   true,
		TotalCircleSides: 24,
		CircleSides:      24,
	}
}

func hitSlashComponent(kind EffectComponentKind, sizeX, sizeEndY, angleStart, angleEnd float64, overlay bool) EffectComponent {
	return EffectComponent{
		Kind:        kind,
		TextureFile: "effect/lens2.tga",
		Duration:    400 * time.Millisecond,
		AlphaMax:    1,
		FadeOut:     true,
		Rotate:      true,
		PosZ:        1,
		SizeStartX:  sizeX,
		SizeEndX:    sizeX,
		SizeStartY:  effectTableSize(10),
		SizeEndY:    sizeEndY,
		AngleStart:  angleStart,
		AngleEnd:    angleEnd,
		Overlay:     overlay,
	}
}

func castAuraEffectSpec(texture string, tint color.RGBA, alphaMax, height, topSize float64, rotate bool) EffectSpec {
	return EffectSpec{
		Duration: 900 * time.Millisecond,
		SFX:      []string{"effect\\ef_beginspell.wav"},
		Components: []EffectComponent{{
			Kind:             EffectComponentCylinder,
			TextureName:      texture,
			AlphaMax:         alphaMax,
			Fade:             true,
			Rotate:           rotate,
			Animation:        2,
			BottomSize:       1,
			TopSize:          topSize,
			Height:           height,
			TotalCircleSides: 32,
			CircleSides:      32,
			Color:            tint,
		}},
	}
}

func elementalCastAuraEffectSpec(texture string, tint color.RGBA, alphaMax float64) EffectSpec {
	return EffectSpec{
		Duration: 900 * time.Millisecond,
		SFX:      []string{"effect\\ef_beginspell.wav"},
		Components: []EffectComponent{
			{
				Kind:             EffectComponentCylinder,
				TextureName:      texture,
				AlphaMax:         0.3,
				Fade:             true,
				Rotate:           true,
				Animation:        1,
				BottomSize:       1,
				TopSize:          1,
				Height:           30,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            tint,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      texture,
				AlphaMax:         alphaMax,
				Fade:             true,
				Rotate:           true,
				Animation:        1,
				BottomSize:       1,
				TopSize:          1.3,
				Height:           2,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            tint,
			},
			{
				Kind:             EffectComponentCylinder,
				TextureName:      texture,
				AlphaMax:         alphaMax,
				Fade:             true,
				Rotate:           true,
				Animation:        2,
				BottomSize:       1,
				TopSize:          4,
				Height:           3,
				TotalCircleSides: 32,
				CircleSides:      32,
				Color:            tint,
			},
		},
	}
}

func bluePotionEffectSpec() EffectSpec {
	spec := potionEffectSpec("파란포션", color.RGBA{R: 92, G: 150, B: 255, A: 255})
	spec.SFX = []string{"effect\\흡기.wav"}
	return spec
}

func incAgilityParticleComponent(alpha float64, delay time.Duration, duplicate int) EffectComponent {
	return EffectComponent{
		Kind:            EffectComponent3D,
		Color:           color.RGBA{R: 255, G: 255, B: 255, A: 255},
		TextureFile:     "effect/ac_center2.tga",
		Duration:        1000 * time.Millisecond,
		Delay:           delay,
		DuplicateDelay:  200 * time.Millisecond,
		AlphaMax:        alpha,
		FadeOut:         true,
		PosXRand:        1.5,
		PosYRand:        1,
		PosZStartRand:   1,
		PosZStartMiddle: 1,
		PosZEndRand:     1,
		PosZEndMiddle:   6,
		SizeStartX:      2.5 * EffectPixelRatio,
		SizeEndX:        2.5 * EffectPixelRatio,
		SizeRandY:       15 * EffectPixelRatio,
		SizeRandYMiddle: 45 * EffectPixelRatio,
		Duplicate:       duplicate,
	}
}

func decAgilityParticleComponent() EffectComponent {
	return EffectComponent{
		Kind:            EffectComponent3D,
		TextureFile:     "effect/ac_center2.tga",
		Duration:        1000 * time.Millisecond,
		DuplicateDelay:  200 * time.Millisecond,
		AlphaMax:        1,
		FadeOut:         true,
		PosXRand:        1.5,
		PosYRand:        1,
		PosZStartRand:   1,
		PosZStartMiddle: 6,
		PosZEndRand:     1,
		PosZEndMiddle:   1,
		SizeStartX:      effectTableSize(2.5),
		SizeEndX:        effectTableSize(2.5),
		SizeRandY:       effectTableSize(15),
		SizeRandYMiddle: effectTableSize(45),
		Duplicate:       20,
	}
}
