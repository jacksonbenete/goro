package gamemode

import (
	"os"
	"slices"
	"testing"
	"time"
)

func TestParseRobrowserEffectTableSubsetParsesSTRWithoutImplicitSound(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	158: [
		{
			//EF_JOBLVUP Job Level Up
			type: 'STR',
			file: 'joblvup',
			attachedEntity: true,
			head: true
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[158]
	if !ok {
		t.Fatal("effect 158 was not parsed")
	}
	if len(spec.sfx) != 0 {
		t.Fatalf("effect 158 sfx = %v, want none", spec.sfx)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 158 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponentSTR || component.strFile != "joblvup" {
		t.Fatalf("effect 158 component = %#v, want STR joblvup", component)
	}
	if !component.attachedEntity || !component.spriteHead {
		t.Fatalf("effect 158 STR attachment flags = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParsesCylinderAndWav(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	304: [
		{
			type: 'CYLINDER',
			textureName: 'ring_blue',
			duration: 1500,
			alphaMax: 0.5,
			fade: true,
			rotate: true,
			animation: 5,
			bottomSize: 0.3,
			topSize: 0.3,
			height: 35,
			wav: 'effect/ef_teleportation'
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[304]
	if !ok {
		t.Fatal("effect 304 was not parsed")
	}
	if !slices.Equal(spec.sfx, []string{`effect\ef_teleportation.wav`}) {
		t.Fatalf("effect 304 sfx = %v", spec.sfx)
	}
	if spec.duration != 1500*time.Millisecond {
		t.Fatalf("effect 304 duration = %s, want 1500ms", spec.duration)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 304 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponentCylinder || component.textureName != "ring_blue" {
		t.Fatalf("effect 304 component = %#v, want ring_blue cylinder", component)
	}
	if component.alphaMax != 0.5 || !component.fade || !component.rotate || component.animation != 5 {
		t.Fatalf("effect 304 component flags = %#v", component)
	}
	if component.bottomSize != 0.3 || component.topSize != 0.3 || component.height != 35 {
		t.Fatalf("effect 304 component dimensions = %#v", component)
	}
	if component.totalCircleSides != 32 || component.circleSides != 32 {
		t.Fatalf("effect 304 circle sides = %d/%d, want 32/32", component.totalCircleSides, component.circleSides)
	}
}

func TestParseRobrowserEffectTableSubsetParsesSTRRand(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	49: [
		{
			type: 'STR',
			file: 'firehit%d',
			wav: 'effect/ef_firehit',
			rand: [1, 3]
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[49]
	if !ok {
		t.Fatal("effect 49 was not parsed")
	}
	if !slices.Equal(spec.sfx, []string{`effect\ef_firehit.wav`}) {
		t.Fatalf("effect 49 sfx = %v", spec.sfx)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 49 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponentSTR || component.strFile != "firehit%d" || component.strRandMin != 1 || component.strRandMax != 3 {
		t.Fatalf("effect 49 component = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParsesSTRMinFile(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	41: [
		{
			type: 'STR',
			file: 'angelus',
			min: 'jong_mini',
			wav: 'effect/ef_angelus',
			head: true,
			attachedEntity: true
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[41]
	if !ok {
		t.Fatal("effect 41 was not parsed")
	}
	if !slices.Equal(spec.sfx, []string{`effect\ef_angelus.wav`}) {
		t.Fatalf("effect 41 sfx = %v", spec.sfx)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 41 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponentSTR || component.strFile != "angelus" || component.strMinFile != "jong_mini" || !component.spriteHead || !component.attachedEntity {
		t.Fatalf("effect 41 component = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParses2D(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	50: [
		{
			type: '2D',
			duration: 500,
			file: 'effect/firering.tga',
			sizeStart: 10,
			sizeEnd: 300,
			angle: 0,
			toAngle: -360,
			fadeOut: true,
			posz: 1
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[50]
	if !ok {
		t.Fatal("effect 50 was not parsed")
	}
	if spec.duration != 500*time.Millisecond {
		t.Fatalf("effect 50 duration = %s, want 500ms", spec.duration)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 50 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponent2D || component.textureFile != "effect/firering.tga" {
		t.Fatalf("effect 50 component = %#v", component)
	}
	if component.sizeStart != 10*roBrowserEffectPixelRatio || component.sizeEnd != 300*roBrowserEffectPixelRatio {
		t.Fatalf("effect 50 size = %.3f..%.3f", component.sizeStart, component.sizeEnd)
	}
	if component.angleStart != 0 || component.angleEnd != -360 || !component.fadeOut || component.posZ != 1 {
		t.Fatalf("effect 50 transform = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParses3D(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	18: [
		{
			type: '3D',
			file: 'effect/pok1.tga',
			duration: 500,
			duplicate: 7,
			timeBetweenDupli: 20,
			delayOffset: 30,
			delayLate: 40,
			alphaMax: 1,
			fadeOut: true,
			red: 1,
			green: 1,
			blue: 0.85,
			alphaMaxDelta: -0.25,
			posxStartRand: 1.5,
			posxStartRandMiddle: 5,
			posyStartRand: 2.5,
			posyStartRandMiddle: -1,
			posxEndRand: 3.5,
			posxEndRandMiddle: -2,
			posyEndRand: 3.5,
			posyEndRandMiddle: 4,
			poszEndRand: 1,
			poszEndRandMiddle: 3,
			posxSmooth: true,
			posySmooth: true,
			poszSmooth: true,
			sizeEnd: 10,
			sizeStart: 200,
			sizeRand: 20,
			sizeDelta: -60,
			angle: 90,
			toAngle: -270,
			rotate: true,
			rotateWithCamera: true,
			blendMode: 2,
			overlay: true
		},
		{ wav: 'effect/ef_steal' }
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[18]
	if !ok {
		t.Fatal("effect 18 was not parsed")
	}
	if spec.duration != 690*time.Millisecond {
		t.Fatalf("effect 18 duration = %s, want 690ms", spec.duration)
	}
	if !slices.Equal(spec.sfx, []string{`effect\ef_steal.wav`}) {
		t.Fatalf("effect 18 sfx = %v", spec.sfx)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 18 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponent3D || component.textureFile != "effect/pok1.tga" {
		t.Fatalf("effect 18 component = %#v", component)
	}
	if component.delay != 70*time.Millisecond || component.duplicateDelay != 20*time.Millisecond || component.duplicate != 7 {
		t.Fatalf("effect 18 timing = %#v", component)
	}
	if component.color.R != 255 || component.color.G != 255 || component.color.B != 216 || component.color.A != 255 {
		t.Fatalf("effect 18 color = %#v", component.color)
	}
	if component.sizeStart != 200*roBrowserEffectPixelRatio || component.sizeEnd != 10*roBrowserEffectPixelRatio || component.sizeRand != 20*roBrowserEffectPixelRatio {
		t.Fatalf("effect 18 size = %#v", component)
	}
	if component.sizeDelta != -60 {
		t.Fatalf("effect 18 size delta = %#v", component)
	}
	if component.posXStartRand != 1.5 || component.posXStartMiddle != 5 || component.posYStartRand != 2.5 || component.posYStartMiddle != -1 {
		t.Fatalf("effect 18 start position = %#v", component)
	}
	if component.posXEndRand != 3.5 || component.posXEndMiddle != -2 || component.posYEndRand != 3.5 || component.posYEndMiddle != 4 || component.posZEndRand != 1 || component.posZEndMiddle != 3 {
		t.Fatalf("effect 18 position = %#v", component)
	}
	if !component.posXSmooth || !component.posYSmooth || !component.posZSmooth {
		t.Fatalf("effect 18 smooth flags = %#v", component)
	}
	if component.alphaMaxDelta != -0.25 || !component.rotate || !component.rotateWithCamera || component.angleStart != 90 || component.angleEnd != -270 || !component.blendAdditive || !component.overlay {
		t.Fatalf("effect 18 flags = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParsesSpriteBacked3D(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	15: [
		{
			type: '3D',
			duration: 250,
			duplicate: 5,
			timeBetweenDupli: 20,
			absoluteSpriteName: 'data/sprite/\xc0\xcc\xc6\xd1\xc6\xae/particle1',
			playSprite: true,
			toSrc: true,
			rotateToTarget: true,
			sizeStart: 100,
			sizeEnd: 500,
			zOffsetStart: 3,
			arc: 4,
			retreat: 3
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[15]
	if !ok {
		t.Fatal("effect 15 was not parsed")
	}
	if spec.duration != 330*time.Millisecond {
		t.Fatalf("effect 15 duration = %s, want 330ms", spec.duration)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 15 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	wantSprite := fieldString(map[string]string{"sprite": `'data/sprite/\xc0\xcc\xc6\xd1\xc6\xae/particle1'`}, "sprite")
	if component.kind != effectComponent3D || component.spriteFile != wantSprite {
		t.Fatalf("effect 15 sprite component = %#v", component)
	}
	if !component.spriteRepeat || component.duration != 250*time.Millisecond || component.duplicate != 5 || component.duplicateDelay != 20*time.Millisecond {
		t.Fatalf("effect 15 timing = %#v", component)
	}
	if !component.toSrc || !component.rotateToTarget || component.arc != 4 || component.retreat != 3 {
		t.Fatalf("effect 15 trajectory = %#v", component)
	}
	if component.posZ != 3 || component.sizeStart != 100*roBrowserEffectPixelRatio || component.sizeEnd != roBrowserEffectSize(500) {
		t.Fatalf("effect 15 transform = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParsesSPR(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	42: [
		{
			type: 'SPR',
			file: '\xc3\xe0\xba\xb9',
			duration: 1500,
			delayFrame: 30,
			frame: 0,
			repeat: true,
			head: true,
			yOffset: -120
		},
		{ wav: 'effect/ef_blessing' }
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	spec, ok := specs[42]
	if !ok {
		t.Fatal("effect 42 was not parsed")
	}
	if spec.duration != 1500*time.Millisecond {
		t.Fatalf("effect 42 duration = %s, want 1500ms", spec.duration)
	}
	if !slices.Equal(spec.sfx, []string{`effect\ef_blessing.wav`}) {
		t.Fatalf("effect 42 sfx = %v", spec.sfx)
	}
	if len(spec.components) != 1 {
		t.Fatalf("effect 42 component count = %d, want 1", len(spec.components))
	}
	component := spec.components[0]
	if component.kind != effectComponentSPR || component.spriteFile != "\xC3\xE0\xBA\xB9" {
		t.Fatalf("effect 42 component = %#v", component)
	}
	if component.duration != 1500*time.Millisecond || component.spriteDelay != 30*time.Millisecond {
		t.Fatalf("effect 42 timing = %#v", component)
	}
	if !component.spriteRepeat || !component.spriteHead || component.spriteYOffset != -120 || !component.worldSizedSprite {
		t.Fatalf("effect 42 sprite flags = %#v", component)
	}
}

func TestParseRobrowserEffectTableSubsetParsesFUNC(t *testing.T) {
	specs, err := parseRobrowserEffectTableSubset(`
export default {
	513: [
		{
			type: 'FUNC',
			attachedEntity: false,
			func: function (Params) {
				this.add(new MagicTarget(Params.Init.skillId), Params);
			}
		}
	],
	60: [
		{
			type: 'FUNC',
			attachedEntity: true,
			func: LockOnTarget
		}
	],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	ground, ok := specs[effectGroundSample]
	if !ok {
		t.Fatal("effect 513 was not parsed")
	}
	if len(ground.components) != 1 {
		t.Fatalf("effect 513 component count = %d, want 1", len(ground.components))
	}
	component := ground.components[0]
	if component.kind != effectComponentFUNC || component.funcAdapter != effectFuncGroundSample || component.funcName != "MagicTarget" || component.attachedEntity {
		t.Fatalf("effect 513 component = %#v", component)
	}
	lockOn := specs[60].components[0]
	if lockOn.kind != effectComponentFUNC || lockOn.funcAdapter != effectFuncUnknown || lockOn.funcName != "LockOnTarget" || !lockOn.attachedEntity {
		t.Fatalf("effect 60 component = %#v", lockOn)
	}
}

func TestParseRobrowserEffectTableEntryIDsIgnoresCommentedEntries(t *testing.T) {
	ids, err := parseRobrowserEffectTableEntryIDs(`
export default {
	1: [],
	//2: [],
	/* 3: [] */
	4: [{ wav: 'effect/ok' }],
}
`)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := ids[1]; !ok {
		t.Fatal("active effect 1 was not parsed")
	}
	if _, ok := ids[4]; !ok {
		t.Fatal("active effect 4 was not parsed")
	}
	if _, ok := ids[2]; ok {
		t.Fatal("commented line effect 2 was parsed")
	}
	if _, ok := ids[3]; ok {
		t.Fatal("commented block effect 3 was parsed")
	}
}

func TestParseRobrowserEffectTableRealFileWhenAvailable(t *testing.T) {
	source, err := os.ReadFile("/home/kivutar/src/robr/src/DB/Effects/EffectTable.js")
	if os.IsNotExist(err) {
		t.Skip("roBrowser checkout not available")
	}
	if err != nil {
		t.Fatal(err)
	}
	ids, err := parseRobrowserEffectTableEntryIDs(string(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(ids) != robrowserActiveEffectTableEntries {
		t.Fatalf("active roBrowser effect entries = %d, want %d", len(ids), robrowserActiveEffectTableEntries)
	}
	specs, err := parseRobrowserEffectTableSubset(string(source))
	if err != nil {
		t.Fatal(err)
	}
	job, ok := specs[effectJobLevelUp]
	if !ok {
		t.Fatal("EF_JOBLVUP was not parsed")
	}
	if len(job.sfx) != 0 {
		t.Fatalf("EF_JOBLVUP sfx = %v, want none", job.sfx)
	}
	if len(job.components) != 1 || job.components[0].kind != effectComponentSTR || job.components[0].strFile != "joblvup" {
		t.Fatalf("EF_JOBLVUP component = %#v", job.components)
	}
	teleport, ok := specs[effectTeleportation]
	if !ok {
		t.Fatal("EF_TELEPORTATION2 was not parsed")
	}
	if len(teleport.components) != 4 {
		t.Fatalf("EF_TELEPORTATION2 component count = %d, want 4", len(teleport.components))
	}
	if !slices.Equal(teleport.sfx, []string{`effect\ef_teleportation.wav`}) {
		t.Fatalf("EF_TELEPORTATION2 sfx = %v", teleport.sfx)
	}
	portal, ok := specs[effectPortal]
	if !ok {
		t.Fatal("EF_PORTAL2 was not parsed")
	}
	if len(portal.components) != 4 {
		t.Fatalf("EF_PORTAL2 component count = %d, want 4", len(portal.components))
	}
	if !slices.Equal(portal.sfx, []string{`effect\ef_readyportal.wav`, `effect\ef_portal.wav`}) {
		t.Fatalf("EF_PORTAL2 sfx = %v", portal.sfx)
	}
}
