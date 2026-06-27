package core

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

type Config struct {
	DataDir string
	Window  WindowConfig
	Packet  PacketConfig
	Login   LoginConfig
	Audio   AudioConfig
	Render  RenderConfig
	Network NetworkConfig
	Fog     FogConfig
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

type RenderConfig struct {
	GraphicsAPI        string
	VSync              bool
	FPS                bool
	BenchSeconds       int
	BenchWarmupSeconds int
	CPUProfile         string
	Stats              bool
	WorldDebugStats    bool
}

type NetworkConfig struct {
	Trace bool
}

type FogConfig struct {
	Enabled        bool
	VeilStrength   float64
	VeilDepthScale float64
}

func LoadConfig(args []string) (Config, error) {
	cfg := defaultConfig()

	configPath, explicitConfig := configPathFromArgs(args)
	if configPath != "" {
		if err := applyINIFile(&cfg, configPath, explicitConfig); err != nil {
			return Config{}, err
		}
	}
	if err := applyCLI(&cfg, args); err != nil {
		return Config{}, err
	}
	cfg.DataDir = resolveDataDir(cfg.DataDir)
	return cfg, nil
}

func defaultConfig() Config {
	return Config{
		Window: WindowConfig{
			Title:      "goro",
			Width:      1280,
			Height:     720,
			Fullscreen: false,
		},
		Packet: PacketConfig{
			ClientDate: 20080910,
			Profile:    23,
		},
		Audio: AudioConfig{
			BGM:       true,
			BGMVolume: 0.55,
		},
		Render: RenderConfig{
			GraphicsAPI:        "vulkan",
			VSync:              true,
			BenchWarmupSeconds: 0,
		},
		Fog: FogConfig{
			Enabled:        true,
			VeilStrength:   0.10,
			VeilDepthScale: 1.2,
		},
	}
}

func configPathFromArgs(args []string) (string, bool) {
	for i := 0; i < len(args); i++ {
		arg := args[i]
		if arg == "--config" && i+1 < len(args) {
			return args[i+1], true
		}
		if strings.HasPrefix(arg, "--config=") {
			return strings.TrimPrefix(arg, "--config="), true
		}
	}
	if _, err := os.Stat("goro.ini"); err == nil {
		return "goro.ini", false
	}
	return "", false
}

func applyINIFile(cfg *Config, path string, explicit bool) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) && !explicit {
			return nil
		}
		return fmt.Errorf("open config %s: %w", path, err)
	}
	defer file.Close()

	if err := applyINI(cfg, file); err != nil {
		return fmt.Errorf("read config %s: %w", path, err)
	}
	return nil
}

func applyCLI(cfg *Config, args []string) error {
	fs := flag.NewFlagSet("goro", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	configPath := ""
	windowed := false
	fs.StringVar(&configPath, "config", "", "path to goro ini configuration")
	fs.StringVar(&cfg.DataDir, "data-dir", cfg.DataDir, "Ragnarok data directory")
	fs.StringVar(&cfg.Window.Title, "title", cfg.Window.Title, "window title")
	fs.IntVar(&cfg.Window.Width, "width", cfg.Window.Width, "window width")
	fs.IntVar(&cfg.Window.Height, "height", cfg.Window.Height, "window height")
	fs.BoolVar(&cfg.Window.Fullscreen, "fullscreen", cfg.Window.Fullscreen, "start in fullscreen mode")
	fs.BoolVar(&windowed, "windowed", false, "force windowed mode")
	fs.IntVar(&cfg.Packet.ClientDate, "packet-client-date", cfg.Packet.ClientDate, "packet client date")
	fs.IntVar(&cfg.Packet.Profile, "packet-profile", cfg.Packet.Profile, "packet profile")
	fs.StringVar(&cfg.Login.Username, "username", cfg.Login.Username, "login username")
	fs.StringVar(&cfg.Login.Password, "password", cfg.Login.Password, "login password")
	fs.BoolVar(&cfg.Login.AutoLogin, "autologin", cfg.Login.AutoLogin, "attempt login automatically")
	fs.BoolVar(&cfg.Audio.BGM, "bgm", cfg.Audio.BGM, "enable BGM")
	fs.Float64Var(&cfg.Audio.BGMVolume, "bgm-volume", cfg.Audio.BGMVolume, "BGM and SFX volume from 0 to 1")
	fs.StringVar(&cfg.Render.GraphicsAPI, "graphics-api", cfg.Render.GraphicsAPI, "graphics API: auto, vulkan, dx12, metal, gles, software")
	fs.BoolVar(&cfg.Render.VSync, "vsync", cfg.Render.VSync, "enable vsync")
	fs.BoolVar(&cfg.Render.FPS, "fps", cfg.Render.FPS, "show measured FPS counter")
	fs.IntVar(&cfg.Render.BenchSeconds, "bench-seconds", cfg.Render.BenchSeconds, "quit after benchmarking for this many seconds")
	fs.IntVar(&cfg.Render.BenchWarmupSeconds, "bench-warmup-seconds", cfg.Render.BenchWarmupSeconds, "benchmark warmup seconds")
	fs.StringVar(&cfg.Render.CPUProfile, "cpu-profile", cfg.Render.CPUProfile, "write CPU profile to this path during benchmark")
	fs.BoolVar(&cfg.Render.Stats, "render-stats", cfg.Render.Stats, "show render stats")
	fs.BoolVar(&cfg.Render.WorldDebugStats, "world-debug-stats", cfg.Render.WorldDebugStats, "show world renderer debug stats")
	fs.BoolVar(&cfg.Network.Trace, "net-trace", cfg.Network.Trace, "log network reads and writes")
	fs.BoolVar(&cfg.Fog.Enabled, "fog", cfg.Fog.Enabled, "enable map fog")
	fs.Float64Var(&cfg.Fog.VeilStrength, "fog-veil-strength", cfg.Fog.VeilStrength, "fog veil strength")
	fs.Float64Var(&cfg.Fog.VeilDepthScale, "fog-veil-depth-scale", cfg.Fog.VeilDepthScale, "fog veil depth scale")

	if err := fs.Parse(args); err != nil {
		return err
	}
	if windowed {
		cfg.Window.Fullscreen = false
	}
	return validateConfig(cfg)
}

func applyINI(cfg *Config, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	section := ""
	lineNo := 0
	for scanner.Scan() {
		lineNo++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			section = normalizeKey(strings.TrimSpace(line[1 : len(line)-1]))
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			return fmt.Errorf("line %d: expected key=value", lineNo)
		}
		if err := applyConfigValue(cfg, section, normalizeKey(key), cleanINIValue(value)); err != nil {
			return fmt.Errorf("line %d: %w", lineNo, err)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return validateConfig(cfg)
}

func applyConfigValue(cfg *Config, section, key, value string) error {
	switch section + "." + key {
	case ".datadir", "data.dir", "data.datadir", "core.datadir":
		cfg.DataDir = value
	case "window.title":
		cfg.Window.Title = value
	case "window.width":
		return setInt(value, &cfg.Window.Width)
	case "window.height":
		return setInt(value, &cfg.Window.Height)
	case "window.fullscreen":
		return setBool(value, &cfg.Window.Fullscreen)
	case "packet.clientdate":
		return setInt(value, &cfg.Packet.ClientDate)
	case "packet.profile":
		return setInt(value, &cfg.Packet.Profile)
	case "login.username":
		cfg.Login.Username = value
	case "login.password":
		cfg.Login.Password = value
	case "login.autologin":
		return setBool(value, &cfg.Login.AutoLogin)
	case "audio.bgm":
		return setBool(value, &cfg.Audio.BGM)
	case "audio.bgmvolume":
		return setFloat(value, &cfg.Audio.BGMVolume)
	case "render.graphicsapi":
		cfg.Render.GraphicsAPI = value
	case "render.vsync":
		return setBool(value, &cfg.Render.VSync)
	case "render.fps":
		return setBool(value, &cfg.Render.FPS)
	case "render.benchseconds":
		return setInt(value, &cfg.Render.BenchSeconds)
	case "render.benchwarmupseconds":
		return setInt(value, &cfg.Render.BenchWarmupSeconds)
	case "render.cpuprofile":
		cfg.Render.CPUProfile = value
	case "render.stats":
		return setBool(value, &cfg.Render.Stats)
	case "render.worlddebugstats":
		return setBool(value, &cfg.Render.WorldDebugStats)
	case "network.trace":
		return setBool(value, &cfg.Network.Trace)
	case "fog.enabled":
		return setBool(value, &cfg.Fog.Enabled)
	case "fog.veilstrength":
		return setFloat(value, &cfg.Fog.VeilStrength)
	case "fog.veildepthscale":
		return setFloat(value, &cfg.Fog.VeilDepthScale)
	default:
		return fmt.Errorf("unknown key %q in section %q", key, section)
	}
	return nil
}

func validateConfig(cfg *Config) error {
	if cfg.Window.Width <= 0 {
		return fmt.Errorf("window width must be positive")
	}
	if cfg.Window.Height <= 0 {
		return fmt.Errorf("window height must be positive")
	}
	if cfg.Packet.ClientDate <= 0 {
		return fmt.Errorf("packet client date must be positive")
	}
	if cfg.Audio.BGMVolume < 0 || cfg.Audio.BGMVolume > 1 {
		return fmt.Errorf("bgm volume must be between 0 and 1")
	}
	if cfg.Render.BenchSeconds < 0 || cfg.Render.BenchWarmupSeconds < 0 {
		return fmt.Errorf("benchmark durations must be non-negative")
	}
	if cfg.Fog.VeilStrength < 0 || cfg.Fog.VeilDepthScale < 0 {
		return fmt.Errorf("fog values must be non-negative")
	}
	return nil
}

func cleanINIValue(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		if (value[0] == '"' && value[len(value)-1] == '"') || (value[0] == '\'' && value[len(value)-1] == '\'') {
			return value[1 : len(value)-1]
		}
	}
	return value
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "-", "")
	value = strings.ReplaceAll(value, "_", "")
	return value
}

func setBool(raw string, dst *bool) error {
	value, err := strconv.ParseBool(raw)
	if err == nil {
		*dst = value
		return nil
	}
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "yes", "on", "enabled":
		*dst = true
		return nil
	case "no", "off", "disabled":
		*dst = false
		return nil
	default:
		return fmt.Errorf("invalid bool %q", raw)
	}
}

func setInt(raw string, dst *int) error {
	value, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return fmt.Errorf("invalid int %q", raw)
	}
	*dst = value
	return nil
}

func setFloat(raw string, dst *float64) error {
	value, err := strconv.ParseFloat(strings.TrimSpace(raw), 64)
	if err != nil {
		return fmt.Errorf("invalid float %q", raw)
	}
	*dst = value
	return nil
}

func resolveDataDir(value string) string {
	if value != "" {
		if abs, err := filepath.Abs(value); err == nil {
			return abs
		}
		return value
	}

	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}
