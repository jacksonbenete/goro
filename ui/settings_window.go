package ui

import (
	"log"
	"math"

	"github.com/gogpu/ui/core/checkbox"
	"github.com/gogpu/ui/core/slider"
	"github.com/gogpu/ui/geometry"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	settingsWindowWidth = 300
)

type SettingsWindow struct {
	window WindowState
	ctx    client.Context
}

func (w *SettingsWindow) OpenWindow(ctx client.Context) {
	w.ensureWindow(ctx)
	w.ctx = ctx
	w.window.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *SettingsWindow) Update(ctx client.Context) bool {
	w.ensureWindow(ctx)
	w.ctx = ctx
	if !w.window.IsOpen() {
		return false
	}
	consumed := w.window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *SettingsWindow) IsOpen() bool {
	w.ensureWindow(w.ctx)
	return w.window.IsOpen()
}

func (w *SettingsWindow) Close() {
	w.ensureWindow(w.ctx)
	w.window.Close()
	w.Publish(w.ctx)
}

func (w *SettingsWindow) Publish(ctx client.Context) {
	w.ensureWindow(ctx)
	if ctx.UIManager == nil {
		return
	}
	w.window.Publish(ctx)
}

func (w *SettingsWindow) Rebind(ctx client.Context) {
	if !w.IsOpen() {
		return
	}
	w.refresh(ctx)
}

func (w *SettingsWindow) ensureWindow(ctx client.Context) {
	height := settingsWindowHeight(ctx)
	if w.window.width == 0 {
		w.window = NewWindowState(settingsWindowWidth, height)
		return
	}
	w.window.SetSize(settingsWindowWidth, height)
}

func (w *SettingsWindow) widgetTree(ctx client.Context) widget.Widget {
	content := w.contentTree(ctx)
	return Window(
		Title("Settings"),
		CloseButton(true),
		OnClose(w.Close),
		Size(settingsWindowWidth, float32(settingsWindowHeightForContent(content))),

		Content(content),
	)
}

func (w *SettingsWindow) contentTree(ctx client.Context) widget.Widget {
	return primitives.Box(
		rotheme.SectionLabel("Display"),

		rotheme.Checkbox(
			checkbox.Checked(settingsRuntimeFullscreen(ctx)),
			checkbox.LabelOpt("Fullscreen"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetFullscreen(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsRuntimeVSync(ctx)),
			checkbox.LabelOpt("VSync (Restart)"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetVSync(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsRuntimeFPS(ctx)),
			checkbox.LabelOpt("FPS meter"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Runtime != nil {
					ctx.Runtime.SetFPS(enabled)
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.SectionLabel("Sound"),

		primitives.HBox(
			rotheme.Text("BGM Vol"),
			primitives.Expanded(
				rotheme.Slider(
					slider.Min(0),
					slider.Max(1),
					slider.Value(float32(settingsVolumeBGM(ctx))),
					slider.OnChange(func(v float32) {
						if ctx.Audio != nil {
							ctx.Audio.SetBGMVolume(float64(v))
						}
						w.saveSettings(ctx)
						w.refresh(ctx)
					}),
				),
			),
		).Gap(8),

		primitives.HBox(
			rotheme.Text("SFX Vol"),
			primitives.Expanded(
				rotheme.Slider(
					slider.Min(0),
					slider.Max(1),
					slider.Value(float32(settingsVolumeSFX(ctx))),
					slider.OnChange(func(v float32) {
						if ctx.Audio != nil {
							ctx.Audio.SetSFXVolume(float64(v))
						}
						w.saveSettings(ctx)
						w.refresh(ctx)
					}),
				),
			),
		).Gap(8),

		rotheme.SectionLabel("Gameplay"),

		rotheme.Checkbox(
			checkbox.Checked(settingsNoShift(ctx)),
			checkbox.LabelOpt("No Shift"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.NoShift = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsNoCtrl(ctx)),
			checkbox.LabelOpt("No Ctrl"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.NoCtrl = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsSnapTargets(ctx)),
			checkbox.LabelOpt("Snap to targets"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.SnapTargets = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),

		rotheme.Checkbox(
			checkbox.Checked(settingsSnapItems(ctx)),
			checkbox.LabelOpt("Snap to items"),
			checkbox.OnToggle(func(enabled bool) {
				if ctx.Session != nil {
					ctx.Session.SnapItems = enabled
				}
				w.saveSettings(ctx)
				w.refresh(ctx)
			}),
		),
	).
		Padding(14).
		Gap(8)
}

func settingsWindowHeight(ctx client.Context) int {
	return settingsWindowHeightForContent((&SettingsWindow{}).contentTree(ctx))
}

func settingsWindowHeightForContent(content widget.Widget) int {
	size := content.Layout(widget.NewContext(), geometry.TightWidth(settingsWindowWidth))
	return int(math.Ceil(float64(ROWindowTitleHeight + size.Height)))
}

func (w *SettingsWindow) refresh(ctx client.Context) {
	w.ensureWindow(ctx)
	w.ctx = ctx
	w.window.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *SettingsWindow) saveSettings(ctx client.Context) {
	settings := config.UserSettings{
		Fullscreen:  settingsRuntimeFullscreen(ctx),
		VSync:       settingsRuntimeVSync(ctx),
		FPS:         settingsRuntimeFPS(ctx),
		BGMVolume:   settingsVolumeBGM(ctx),
		SFXVolume:   settingsVolumeSFX(ctx),
		NoShift:     settingsNoShift(ctx),
		NoCtrl:      settingsNoCtrl(ctx),
		SnapTargets: settingsSnapTargets(ctx),
		SnapItems:   settingsSnapItems(ctx),
	}
	path, err := config.SaveUserSettings(settings)
	if err != nil {
		log.Printf("settings save failed: %v", err)
		return
	}
	log.Printf("settings saved path=%s", path)
}

func settingsVolumeBGM(ctx client.Context) float64 {
	if ctx.Audio != nil {
		return ctx.Audio.BGMVolume()
	}
	return ctx.Config.Audio.BGMVolume
}

func settingsVolumeSFX(ctx client.Context) float64 {
	if ctx.Audio != nil {
		return ctx.Audio.SFXVolume()
	}
	return ctx.Config.Audio.SFXVolume
}

func settingsRuntimeFullscreen(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.Fullscreen()
	}
	return ctx.Config.Window.Fullscreen
}

func settingsRuntimeVSync(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.VSync()
	}
	return ctx.Config.Render.VSync
}

func settingsRuntimeFPS(ctx client.Context) bool {
	if ctx.Runtime != nil {
		return ctx.Runtime.FPS()
	}
	return ctx.Config.Render.FPS
}

func settingsNoShift(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.NoShift
	}
	return ctx.Config.Gameplay.NoShift
}

func settingsNoCtrl(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.NoCtrl
	}
	return ctx.Config.Gameplay.NoCtrl
}

func settingsSnapTargets(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.SnapTargets
	}
	return ctx.Config.Gameplay.SnapTargets
}

func settingsSnapItems(ctx client.Context) bool {
	if ctx.Session != nil {
		return ctx.Session.SnapItems
	}
	return ctx.Config.Gameplay.SnapItems
}
