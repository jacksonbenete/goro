package gamemode

import (
	"image/color"
	"time"
)

const (
	robrowserActiveEffectTableEntries = 607
	robrowserNumericEffectConstants   = 1147
)

var worldEffectSpecs = map[int]worldEffectSpec{
	effectBashHit: {
		duration: 280 * time.Millisecond,
		sfx:      []string{"effect\\ef_hit2.wav"},
		components: []worldEffectComponent{{
			kind:  effectPrimitiveBashHit,
			color: color.RGBA{R: 255, G: 248, B: 220, A: 255},
		}},
	},
	effectEndure: {
		duration: 1000 * time.Millisecond,
		sfx:      []string{"effect\\ef_endure.wav"},
		components: []worldEffectComponent{{
			kind:        effectPrimitiveBillboard,
			textureFile: "effect\\endure.tga",
			duration:    1000 * time.Millisecond,
			alphaMax:    1,
			fadeIn:      true,
			fadeOut:     true,
			posZ:        2,
			sizeStart:   2.0,
			sizeEnd:     0.7,
			sizeSmooth:  true,
		}},
	},
	effectBashBegin: {
		duration: 1000 * time.Millisecond,
		sfx:      []string{"effect\\ef_bash.wav"},
		components: []worldEffectComponent{
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "alpha_down",
				duration:         1000 * time.Millisecond,
				alphaMax:         0.6,
				fade:             true,
				rotate:           true,
				fixedPerspective: true,
				animation:        2,
				bottomSize:       0.1,
				topSize:          2.0,
				posZ:             1.5,
				totalCircleSides: 20,
				circleSides:      20,
			},
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "alpha_center",
				duration:         1000 * time.Millisecond,
				alphaMax:         0.6,
				fade:             true,
				rotate:           true,
				fixedPerspective: true,
				animation:        2,
				bottomSize:       0.01,
				topSize:          2.5,
				posZ:             1.5,
				totalCircleSides: 30,
				circleSides:      1,
				duplicate:        10,
				angleZRandom:     360,
			},
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "alpha_center",
				duration:         1000 * time.Millisecond,
				alphaMax:         0.6,
				fade:             true,
				rotate:           true,
				fixedPerspective: true,
				animation:        2,
				bottomSize:       0.01,
				topSize:          4.0,
				posZ:             1.5,
				totalCircleSides: 30,
				circleSides:      1,
				duplicate:        8,
				angleZRandom:     360,
			},
		},
	},
	effectProvoke: {
		duration: 900 * time.Millisecond,
		sfx:      []string{"effect\\swordman_provoke.wav"},
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			color:   color.RGBA{R: 255, G: 70, B: 42, A: 255},
			strFile: "provoke",
		}},
	},
	effectMagnumBreak: {
		duration: 300 * time.Millisecond,
		sfx:      []string{"effect\\ef_magnumbreak.wav"},
		components: []worldEffectComponent{
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "ring_yellow",
				duration:         300 * time.Millisecond,
				alphaMax:         0.7,
				fade:             true,
				rotate:           true,
				animation:        4,
				bottomSize:       4,
				topSize:          6,
				height:           1,
				totalCircleSides: 32,
				circleSides:      32,
			},
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "\xb4\xeb\xc6\xf8\xb9\xdf",
				duration:         300 * time.Millisecond,
				alphaMax:         0.6,
				fade:             true,
				rotate:           true,
				animation:        4,
				bottomSize:       4,
				topSize:          1,
				height:           4,
				totalCircleSides: 32,
				circleSides:      32,
			},
		},
	},
	effectTeleportation: {
		duration: 1500 * time.Millisecond,
		sfx:      []string{"effect\\ef_teleportation.wav"},
		components: []worldEffectComponent{
			teleportCylinderComponent(0.3, 0.3, 35),
			teleportCylinderComponent(0.6, 0.8, 25),
			teleportCylinderComponent(0.8, 1.0, 13),
			teleportCylinderComponent(1.0, 1.3, 5),
		},
	},
	effectPortal: {
		duration: 25000 * time.Millisecond,
		sfx:      []string{"effect\\ef_readyportal.wav"},
		components: []worldEffectComponent{
			{
				kind:             effectPrimitiveCylinder,
				textureName:      "ring_blue",
				duration:         500 * time.Millisecond,
				alphaMax:         0.4,
				fadeOut:          true,
				rotate:           true,
				animation:        4,
				bottomSize:       2.4,
				topSize:          3.9,
				height:           0.1,
				posZ:             0.1,
				totalCircleSides: 32,
				circleSides:      32,
			},
			portalCylinderComponent(0.6, 0.6, 15, 0, "ring_blue", 0.3),
			portalCylinderComponent(0.8, 0.8, 13, 0, "ring_blue", 0.3),
			portalCylinderComponent(1.0, 1.0, 1, 2, "alpha1", 0.5),
		},
	},
	effectBaseLevelUp: {
		duration: 1300 * time.Millisecond,
		sfx:      []string{"levelup.wav"},
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			strFile: "angel",
		}},
	},
	effectJobLevelUp: {
		duration: 1300 * time.Millisecond,
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			strFile: "joblvup",
		}},
	},
	effectPotionRed:    potionEffectSpec("\xbb\xa1\xb0\xa3\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 82, B: 70, A: 255}),
	effectPotionOrange: potionEffectSpec("\xc1\xd6\xc8\xab\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 145, B: 58, A: 255}),
	effectPotionYellow: potionEffectSpec("\xb3\xeb\xb6\xf5\xc6\xf7\xbc\xc7", color.RGBA{R: 255, G: 226, B: 76, A: 255}),
	effectPotionWhite:  potionEffectSpec("\xc7\xcf\xbe\xe1\xc6\xf7\xbc\xc7", color.RGBA{R: 245, G: 245, B: 255, A: 255}),
	effectPotionBlue:   bluePotionEffectSpec(),
	effectPotionGreen:  potionEffectSpec("\xc3\xca\xb7\xcf\xc6\xf7\xbc\xc7", color.RGBA{R: 78, G: 225, B: 98, A: 255}),
	effectFood: {
		duration: 850 * time.Millisecond,
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			color:   color.RGBA{R: 255, G: 182, B: 86, A: 255},
			strFile: "fruit",
		}},
	},
	effectFoodBlue: {
		duration: 850 * time.Millisecond,
		components: []worldEffectComponent{{
			kind:    effectPrimitiveSTR,
			color:   color.RGBA{R: 132, G: 112, B: 255, A: 255},
			strFile: "fruit",
		}},
	},
}

type effectCoverage struct {
	Implemented     int
	RobrowserActive int
	RobrowserAll    int
	ActivePercent   float64
	AllPercent      float64
}

func effectCoverageSnapshot() effectCoverage {
	implemented := len(worldEffectSpecs)
	return effectCoverage{
		Implemented:     implemented,
		RobrowserActive: robrowserActiveEffectTableEntries,
		RobrowserAll:    robrowserNumericEffectConstants,
		ActivePercent:   float64(implemented) / float64(robrowserActiveEffectTableEntries) * 100,
		AllPercent:      float64(implemented) / float64(robrowserNumericEffectConstants) * 100,
	}
}

func bluePotionEffectSpec() worldEffectSpec {
	spec := potionEffectSpec("\xc6\xc4\xb6\xf5\xc6\xf7\xbc\xc7", color.RGBA{R: 92, G: 150, B: 255, A: 255})
	spec.sfx = []string{"effect\\\xc8\xed\xb1\xe2.wav"}
	return spec
}
