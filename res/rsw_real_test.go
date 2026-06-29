package res

import (
	"testing"
)

func TestRSWRealFileWhenConfigured(t *testing.T) {
	data := readRealDataFile(t, "data\\geffen_in.rsw")
	rsw, err := ParseRSW(data)
	if err != nil {
		t.Fatal(err)
	}
	if rsw.Files.GND == "" || rsw.Files.GAT == "" {
		t.Fatalf("real rsw has missing subfiles: %+v", rsw.Files)
	}
}

func TestRSWRealArchiveWhenConfigured(t *testing.T) {
	grf, name := realDataArchiveFile(t, "geffen_in.rsw")
	data, err := grf.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	rsw, err := ParseRSW(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if rsw.Files.GND == "" || rsw.Files.GAT == "" {
		t.Fatalf("invalid parsed rsw %s: files=%+v", name, rsw.Files)
	}
	t.Logf("parsed %s version=%d.%d models=%d lights=%d sounds=%d effects=%d water=%+v", name, rsw.VersionMajor, rsw.VersionMinor, len(rsw.Models), len(rsw.Lights), len(rsw.Sounds), len(rsw.Effects), rsw.Water)
}

func TestPayonDungeonRSWMoodRealWhenConfigured(t *testing.T) {
	for _, name := range []string{"pay_dun00.rsw", "pay_dun01.rsw", "pay_dun02.rsw", "pay_dun03.rsw"} {
		grf, path := realDataArchiveFile(t, name)
		data, err := grf.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		rsw, err := ParseRSW(data)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		t.Logf("%s light longitude=%d latitude=%d diffuse=%v ambient=%v opacity=%.3f effects=%d", path, rsw.Light.Longitude, rsw.Light.Latitude, rsw.Light.Diffuse, rsw.Light.Ambient, rsw.Light.Opacity, len(rsw.Effects))
		if len(rsw.Effects) == 0 {
			t.Fatalf("%s has no RSW effects", path)
		}
		for _, effect := range rsw.Effects {
			if effect.ID != 47 {
				t.Fatalf("%s effect id = %d, want warp-zone id 47", path, effect.ID)
			}
		}
	}
}

func TestByalanDungeonRSWEffectsRealWhenConfigured(t *testing.T) {
	for _, name := range []string{"iz_dun00.rsw", "iz_dun01.rsw", "iz_dun02.rsw", "iz_dun03.rsw", "iz_dun04.rsw"} {
		grf, path := realDataArchiveFile(t, name)
		data, err := grf.ReadFile(path)
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		rsw, err := ParseRSW(data)
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		ids := make(map[int32]int)
		for _, effect := range rsw.Effects {
			ids[effect.ID]++
		}
		t.Logf("%s effects=%d ids=%v", path, len(rsw.Effects), ids)
		if len(rsw.Effects) == 0 {
			t.Fatalf("%s has no RSW effects", path)
		}
		if ids[45]+ids[47]+ids[109] == 0 {
			t.Fatalf("%s has no known Byalan RSW effects", path)
		}
	}
}
