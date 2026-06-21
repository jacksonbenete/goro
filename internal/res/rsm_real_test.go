package res

import (
	"errors"
	"os"
	"testing"
)

func TestRSMRealArchiveWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	name := os.Getenv("GORO_TEST_RSM_FILE")
	if name == "" {
		name = "data\\model\\¿öÇÁ¿¡·ºº£ÀÌÅÍ.rsm"
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
	rsm, err := ParseRSM(data)
	if err != nil {
		t.Fatalf("parse %s: %v", name, err)
	}
	if len(rsm.Nodes) == 0 {
		t.Fatalf("parsed %s has no nodes", name)
	}
	t.Logf("parsed %s version=%d.%d textures=%d nodes=%d", name, rsm.VersionMajor, rsm.VersionMinor, len(rsm.Textures), len(rsm.Nodes))
}

func TestRSMRealArchiveFromRSWWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GRF")
	if path == "" {
		t.Skip("set GORO_TEST_GRF to run against a real archive")
	}
	rswName := os.Getenv("GORO_TEST_RSW_FILE")
	if rswName == "" {
		rswName = "geffen_in.rsw"
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if !grf.Has(rswName) {
		matches := grf.NamesWithSuffix(rswName)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", rswName, path)
		}
		rswName = matches[0]
	}
	rswData, err := grf.ReadFile(rswName)
	if err != nil {
		t.Fatalf("read %s: %v", rswName, err)
	}
	rsw, err := ParseRSW(rswData)
	if err != nil {
		t.Fatalf("parse %s: %v", rswName, err)
	}

	seen := make(map[string]struct{})
	parsed := 0
	for _, model := range rsw.Models {
		if model.Filename == "" {
			continue
		}
		if _, ok := seen[model.Filename]; ok {
			continue
		}
		seen[model.Filename] = struct{}{}

		var data []byte
		var source string
		for _, candidate := range RSMModelCandidates(model.Filename) {
			data, err = grf.ReadFile(candidate)
			if err == nil {
				source = candidate
				break
			}
		}
		if data == nil {
			continue
		}
		rsm, err := ParseRSM(data)
		if err != nil {
			t.Fatalf("parse model %s from %s: %v", model.Filename, source, err)
		}
		if len(rsm.Nodes) == 0 {
			t.Fatalf("model %s from %s has no nodes", model.Filename, source)
		}
		parsed++
		if parsed >= 8 {
			break
		}
	}
	if parsed == 0 {
		t.Fatalf("parsed no models from %s (%d placements)", rswName, len(rsw.Models))
	}
	t.Logf("parsed %d unique RSM models from %s", parsed, rswName)
}
