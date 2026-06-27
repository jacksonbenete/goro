package res

import (
	"errors"
	"path/filepath"
	"testing"
)

func TestManagerReadFileExactDoesNotUseGRFSuffixFallback(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.grf")
	if err := writeTestGRF(path, `data\wav\effect\provoke.wav`, []byte("sound")); err != nil {
		t.Fatal(err)
	}
	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	manager := &Manager{Root: dir, Archives: []*GRF{grf}}
	if _, err := manager.ReadFileExact(`effect\provoke.wav`); err == nil {
		t.Fatal("ReadFileExact matched a GRF suffix-only path")
	}
	if data, err := manager.ReadFileExact(`data\wav\effect\provoke.wav`); err != nil || string(data) != "sound" {
		t.Fatalf("ReadFileExact exact path data=%q err=%v", string(data), err)
	}
	if data, err := manager.ReadFile(`effect\provoke.wav`); err != nil || string(data) != "sound" {
		t.Fatalf("ReadFile legacy suffix data=%q err=%v", string(data), err)
	}
}
