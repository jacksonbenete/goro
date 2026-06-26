package core

import (
	"os"
	"path/filepath"
	"strconv"
)

type Config struct {
	DataDir string
	Window  WindowConfig
	Packet  PacketConfig
	Login   LoginConfig
	Audio   AudioConfig
}

type WindowConfig struct {
	Title      string
	Width      int
	Height     int
	Fullscreen bool
}

type PacketConfig struct {
	ClientDate int
	Profile    int
}

type LoginConfig struct {
	Username  string
	Password  string
	AutoLogin bool
}

type AudioConfig struct {
	BGM       bool
	BGMVolume float64
}

func LoadConfig() Config {
	return Config{
		DataDir: resolveDataDir(),
		Window: WindowConfig{
			Title:      "goro",
			Width:      1280,
			Height:     720,
			Fullscreen: os.Getenv("GORO_FULLSCREEN") == "1",
		},
		Packet: PacketConfig{
			ClientDate: intEnv("GORO_PACKET_CLIENT_DATE", 20080910),
			Profile:    intEnv("GORO_PACKET_PROFILE", 23),
		},
		Login: LoginConfig{
			Username:  os.Getenv("GORO_USERNAME"),
			Password:  os.Getenv("GORO_PASSWORD"),
			AutoLogin: os.Getenv("GORO_AUTOLOGIN") == "1",
		},
		Audio: AudioConfig{
			BGM:       os.Getenv("GORO_BGM") != "0",
			BGMVolume: floatEnv("GORO_BGM_VOLUME", 0.55),
		},
	}
}

func intEnv(name string, fallback int) int {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return fallback
	}
	return value
}

func floatEnv(name string, fallback float64) float64 {
	raw := os.Getenv(name)
	if raw == "" {
		return fallback
	}
	value, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return fallback
	}
	return value
}

func resolveDataDir() string {
	for _, key := range []string{"GORO_DATA_DIR", "OPEN_MIDGARD_DATA_DIR"} {
		if value := os.Getenv(key); value != "" {
			if abs, err := filepath.Abs(value); err == nil {
				return abs
			}
			return value
		}
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
