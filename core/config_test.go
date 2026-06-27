package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigReadsINIAndCLIOverrides(t *testing.T) {
	root := t.TempDir()
	dataDir := filepath.Join(root, "OldRO")
	if err := os.Mkdir(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "goro.ini")
	if err := os.WriteFile(configPath, []byte(`
data_dir = ./ignored

[window]
width = 1024
height = 768
fullscreen = true

[packet]
client_date = 20211103

[audio]
bgm = false
bgm_volume = 0.25

[render]
graphics_api = gles
vsync = false

[network]
trace = true

[fog]
enabled = false
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig([]string{
		"--config", configPath,
		"--data-dir", dataDir,
		"--width", "1280",
		"--fullscreen=false",
		"--bgm=true",
		"--bgm-volume", "0.75",
		"--graphics-api", "vulkan",
	})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.DataDir != dataDir {
		t.Fatalf("data dir = %q, want %q", cfg.DataDir, dataDir)
	}
	if cfg.Window.Width != 1280 || cfg.Window.Height != 768 || cfg.Window.Fullscreen {
		t.Fatalf("unexpected window config: %#v", cfg.Window)
	}
	if cfg.Packet.ClientDate != 20211103 {
		t.Fatalf("packet client date = %d", cfg.Packet.ClientDate)
	}
	if !cfg.Audio.BGM || cfg.Audio.BGMVolume != 0.75 {
		t.Fatalf("unexpected audio config: %#v", cfg.Audio)
	}
	if cfg.Render.GraphicsAPI != "vulkan" || cfg.Render.VSync {
		t.Fatalf("unexpected render config: %#v", cfg.Render)
	}
	if !cfg.Network.Trace {
		t.Fatalf("network trace = false, want true")
	}
	if cfg.Fog.Enabled {
		t.Fatalf("fog enabled = true, want false")
	}
}

func TestLoadConfigWindowedOverridesFullscreenINI(t *testing.T) {
	root := t.TempDir()
	configPath := filepath.Join(root, "goro.ini")
	if err := os.WriteFile(configPath, []byte("[window]\nfullscreen = true\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := LoadConfig([]string{"--config", configPath, "--windowed"})
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Window.Fullscreen {
		t.Fatal("fullscreen = true, want false")
	}
}
