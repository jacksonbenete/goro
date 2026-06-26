package render

import (
	"log"
	"os"
	"runtime/pprof"
	"strconv"
	"time"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gpucontext"
	"github.com/kivutar/goro/internal/core"
	"github.com/kivutar/goro/internal/input"
)

const BackendName = "gogpu-wgpu"

type Game interface {
	Update() error
	Draw(*Image)
	Resize(width, height int)
	InputState() *input.State
}

type runner struct {
	app            *gogpu.App
	game           Game
	screen         *Image
	gpu            *gpuRenderer
	width          int
	height         int
	duration       time.Duration
	warmup         time.Duration
	started        time.Time
	measureStarted time.Time
	lastLog        time.Time
	lastFrame      int64
	frames         int64
	measuredFrames int64
	quit           func()
	cpuProfile     *os.File
}

func Run(game Game, cfg core.WindowConfig) error {
	appConfig := gogpu.DefaultConfig().
		WithTitle(cfg.Title).
		WithSize(cfg.Width, cfg.Height).
		WithResizable(true).
		WithContinuousRender(true).
		WithVSync(os.Getenv("GORO_VSYNC") != "0" && os.Getenv("GORO_BENCH_SECONDS") == "")
	if cfg.Fullscreen {
		appConfig = appConfig.WithFullscreen()
	}
	gg := gogpu.NewApp(appConfig)
	setCursorApp(gg)
	defer setCursorApp(nil)

	r := &runner{
		app:      gg,
		game:     game,
		width:    cfg.Width,
		height:   cfg.Height,
		duration: time.Duration(intEnv("GORO_BENCH_SECONDS", 0)) * time.Second,
		warmup:   time.Duration(intEnv("GORO_BENCH_WARMUP_SECONDS", 0)) * time.Second,
		quit:     gg.Quit,
	}
	game.Resize(cfg.Width, cfg.Height)
	wireInput(gg, game.InputState())

	gg.OnResize(func(width, height int) {
		if width <= 0 || height <= 0 {
			return
		}
		r.width, r.height = width, height
		r.screen = nil
		r.game.Resize(width, height)
	})
	gg.OnUpdate(func(float64) {
		if err := r.update(); err != nil {
			log.Printf("update error: %v", err)
			gg.Quit()
		}
	})
	gg.OnDraw(func(ctx *gogpu.Context) {
		if err := r.draw(ctx); err != nil {
			log.Printf("draw error: %v", err)
			gg.Quit()
		}
	})
	gg.OnClose(func() {
		if r.cpuProfile != nil {
			pprof.StopCPUProfile()
			_ = r.cpuProfile.Close()
			r.cpuProfile = nil
		}
		if r.gpu != nil {
			r.gpu.release()
			r.gpu = nil
		}
	})
	return gg.Run()
}

func wireInput(app *gogpu.App, state *input.State) {
	if state == nil {
		return
	}
	events := app.EventSource()
	events.OnKeyPress(func(key gpucontext.Key, _ gpucontext.Modifiers) {
		if mapped, ok := mapKey(key); ok {
			state.SetKey(mapped, true)
		}
	})
	events.OnKeyRelease(func(key gpucontext.Key, _ gpucontext.Modifiers) {
		if mapped, ok := mapKey(key); ok {
			state.SetKey(mapped, false)
		}
	})
	events.OnMouseMove(func(x, y float64) {
		state.SetMousePosition(int(x+0.5), int(y+0.5))
	})
	events.OnMousePress(func(button gpucontext.MouseButton, x, y float64) {
		state.SetMousePosition(int(x+0.5), int(y+0.5))
		if mapped, ok := mapMouseButton(button); ok {
			state.SetMouseButton(mapped, true)
		}
	})
	events.OnMouseRelease(func(button gpucontext.MouseButton, x, y float64) {
		state.SetMousePosition(int(x+0.5), int(y+0.5))
		if mapped, ok := mapMouseButton(button); ok {
			state.SetMouseButton(mapped, false)
		}
	})
	events.OnScroll(func(x, y float64) {
		state.AddWheel(x, y)
	})
	events.OnTextInput(func(text string) {
		state.AddTextInput(text)
	})
}

func mapKey(key gpucontext.Key) (input.Key, bool) {
	switch key {
	case gpucontext.KeyEnter:
		return input.KeyEnter, true
	case gpucontext.KeyEscape:
		return input.KeyEscape, true
	case gpucontext.KeyTab:
		return input.KeyTab, true
	case gpucontext.KeyUp:
		return input.KeyArrowUp, true
	case gpucontext.KeyDown:
		return input.KeyArrowDown, true
	case gpucontext.KeyLeft:
		return input.KeyArrowLeft, true
	case gpucontext.KeyRight:
		return input.KeyArrowRight, true
	case gpucontext.KeyBackspace:
		return input.KeyBackspace, true
	default:
		return 0, false
	}
}

func mapMouseButton(button gpucontext.MouseButton) (input.MouseButton, bool) {
	switch button {
	case gpucontext.MouseButtonLeft:
		return input.MouseButtonLeft, true
	case gpucontext.MouseButtonRight:
		return input.MouseButtonRight, true
	default:
		return 0, false
	}
}

func (r *runner) update() error {
	if r.duration > 0 && r.started.IsZero() {
		r.started = time.Now()
		r.lastLog = r.started
		if path := os.Getenv("GORO_CPU_PROFILE"); path != "" {
			file, err := os.Create(path)
			if err != nil {
				log.Printf("cpu profile start failed: %v", err)
			} else if err := pprof.StartCPUProfile(file); err != nil {
				log.Printf("cpu profile start failed: %v", err)
				_ = file.Close()
			} else {
				r.cpuProfile = file
				log.Printf("cpu profile writing %s", path)
			}
		}
		log.Printf("benchmark start duration=%s warmup=%s vsync=%v", r.duration, r.warmup, os.Getenv("GORO_VSYNC") != "0")
	}
	if err := r.game.Update(); err != nil {
		return err
	}
	if r.duration <= 0 {
		return nil
	}
	now := time.Now()
	if r.measureStarted.IsZero() && now.Sub(r.started) >= r.warmup {
		r.measureStarted = now
		r.measuredFrames = 0
		log.Printf("benchmark measure start elapsed=%.3fs", now.Sub(r.started).Seconds())
	}
	if now.Sub(r.lastLog) >= time.Second {
		elapsed := now.Sub(r.started).Seconds()
		interval := now.Sub(r.lastLog).Seconds()
		frames := r.frames - r.lastFrame
		log.Printf("benchmark fps interval=%.1f average=%.1f frames=%d elapsed=%.1fs", float64(frames)/interval, float64(r.frames)/elapsed, r.frames, elapsed)
		r.lastLog = now
		r.lastFrame = r.frames
	}
	if now.Sub(r.started) >= r.duration {
		elapsed := now.Sub(r.started).Seconds()
		measuredElapsed := elapsed
		measuredFPS := float64(r.frames) / elapsed
		if !r.measureStarted.IsZero() {
			measuredElapsed = now.Sub(r.measureStarted).Seconds()
			if measuredElapsed > 0 {
				measuredFPS = float64(r.measuredFrames) / measuredElapsed
			}
		}
		log.Printf("benchmark result fps=%.1f measured_fps=%.1f frames=%d measured_frames=%d elapsed=%.3fs measured_elapsed=%.3fs", float64(r.frames)/elapsed, measuredFPS, r.frames, r.measuredFrames, elapsed, measuredElapsed)
		if r.cpuProfile != nil {
			pprof.StopCPUProfile()
			_ = r.cpuProfile.Close()
			r.cpuProfile = nil
		}
		r.quit()
	}
	return nil
}

func (r *runner) draw(ctx *gogpu.Context) error {
	width, height := ctx.Size()
	if width <= 0 || height <= 0 {
		width, height = r.width, r.height
	}
	if width <= 0 || height <= 0 {
		return nil
	}
	if r.screen == nil || r.screen.Bounds().Dx() != width || r.screen.Bounds().Dy() != height {
		r.screen = NewScreenImage(width, height)
		r.width, r.height = width, height
		r.game.Resize(width, height)
	}
	if r.gpu == nil {
		gpu, err := newGPURenderer(ctx, r.app)
		if err != nil {
			return err
		}
		r.gpu = gpu
		log.Printf("render backend=%s surface_format=%s", ctx.Backend(), r.gpu.format)
	}
	r.screen.BeginFrame()
	r.game.Draw(r.screen)
	if err := r.gpu.Draw(ctx, r.screen); err != nil {
		return err
	}
	r.frames++
	if !r.measureStarted.IsZero() {
		r.measuredFrames++
	}
	return nil
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
