package render

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/kivutar/goro/internal/core"
)

const BackendName = "ebitengine-opengl"

type Game interface {
	Update() error
	Draw(*ebiten.Image)
	Layout(outsideWidth, outsideHeight int) (screenWidth, screenHeight int)
}

func Run(game Game, cfg core.WindowConfig) error {
	ebiten.SetWindowTitle(cfg.Title)
	ebiten.SetWindowSize(cfg.Width, cfg.Height)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	if os.Getenv("GORO_VSYNC") == "0" || os.Getenv("GORO_BENCH_SECONDS") != "" {
		ebiten.SetVsyncEnabled(false)
	}
	if seconds := intEnv("GORO_BENCH_SECONDS", 0); seconds > 0 {
		warmupSeconds := intEnv("GORO_BENCH_WARMUP_SECONDS", 0)
		game = newBenchmarkGame(game, time.Duration(seconds)*time.Second, time.Duration(warmupSeconds)*time.Second)
	}

	return ebiten.RunGameWithOptions(game, &ebiten.RunGameOptions{
		GraphicsLibrary: ebiten.GraphicsLibraryOpenGL,
	})
}

type benchmarkGame struct {
	game           Game
	duration       time.Duration
	warmup         time.Duration
	started        time.Time
	measureStarted time.Time
	lastLog        time.Time
	lastFrame      int64
	frames         int64
	measuredFrames int64
}

func newBenchmarkGame(game Game, duration, warmup time.Duration) *benchmarkGame {
	return &benchmarkGame{
		game:     game,
		duration: duration,
		warmup:   warmup,
	}
}

func (g *benchmarkGame) Update() error {
	if g.started.IsZero() {
		g.started = time.Now()
		g.lastLog = g.started
		log.Printf("benchmark start duration=%s warmup=%s vsync=%v", g.duration, g.warmup, ebiten.IsVsyncEnabled())
	}
	if err := g.game.Update(); err != nil {
		return err
	}
	now := time.Now()
	if g.measureStarted.IsZero() && now.Sub(g.started) >= g.warmup {
		g.measureStarted = now
		g.measuredFrames = 0
		log.Printf("benchmark measure start elapsed=%.3fs", now.Sub(g.started).Seconds())
	}
	if now.Sub(g.lastLog) >= time.Second {
		elapsed := now.Sub(g.started).Seconds()
		interval := now.Sub(g.lastLog).Seconds()
		frames := g.frames - g.lastFrame
		log.Printf("benchmark fps interval=%.1f average=%.1f actual=%.1f frames=%d elapsed=%.1fs", float64(frames)/interval, float64(g.frames)/elapsed, ebiten.ActualFPS(), g.frames, elapsed)
		g.lastLog = now
		g.lastFrame = g.frames
	}
	if now.Sub(g.started) >= g.duration {
		elapsed := now.Sub(g.started).Seconds()
		measuredElapsed := elapsed
		measuredFPS := float64(g.frames) / elapsed
		if !g.measureStarted.IsZero() {
			measuredElapsed = now.Sub(g.measureStarted).Seconds()
			if measuredElapsed > 0 {
				measuredFPS = float64(g.measuredFrames) / measuredElapsed
			}
		}
		log.Printf("benchmark result fps=%.1f measured_fps=%.1f frames=%d measured_frames=%d elapsed=%.3fs measured_elapsed=%.3fs actual=%.1f", float64(g.frames)/elapsed, measuredFPS, g.frames, g.measuredFrames, elapsed, measuredElapsed, ebiten.ActualFPS())
		return ebiten.Termination
	}
	return nil
}

func (g *benchmarkGame) Draw(screen *ebiten.Image) {
	g.game.Draw(screen)
	g.frames++
	if !g.measureStarted.IsZero() {
		g.measuredFrames++
	}
}

func (g *benchmarkGame) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.game.Layout(outsideWidth, outsideHeight)
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
