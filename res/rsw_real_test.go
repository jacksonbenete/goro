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
