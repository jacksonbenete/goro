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
			attachedEntity: true
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
	if component.kind != effectPrimitiveSTR || component.strFile != "joblvup" {
		t.Fatalf("effect 158 component = %#v, want STR joblvup", component)
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
	if component.kind != effectPrimitiveCylinder || component.textureName != "ring_blue" {
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
	if len(job.components) != 1 || job.components[0].kind != effectPrimitiveSTR || job.components[0].strFile != "joblvup" {
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
