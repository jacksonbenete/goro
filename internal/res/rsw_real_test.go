package res

import (
	"errors"
	"os"
	"testing"
)

func TestRSWRealFileWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_RSW")
	if path == "" {
		t.Skip("set GORO_TEST_RSW to run against a real RSW file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	rsw, err := ParseRSW(data)
	if err != nil {
		t.Fatal(err)
	}
	if rsw.Files.GND == "" || rsw.Files.GAT == "" {
		t.Fatalf("real rsw has missing subfiles: %+v", rsw.Files)
	}
}

func TestRSWRealArchiveWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	name := os.Getenv("GORO_TEST_RSW_FILE")
	if name == "" {
		name = "geffen_in.rsw"
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if !grf.Has(name) {
		matches := grf.NamesWithSuffix(name)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", name, path)
		}
		name = matches[0]
	}
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
