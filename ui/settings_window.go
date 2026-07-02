package ui

import (
	"fmt"
	"image/color"
	"log"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/config"
	"github.com/kivutar/goro/render"
)

const (
	settingsWindowWidth  = 300
	settingsWindowHeight = 272
	settingsWindowTitleH = 28
	settingsWindowPad    = 14
	settingsVolumeBarH   = 5
	settingsCheckboxGap  = 8
)

type SettingsWindow struct {
	open       bool
	x          int
	y          int
	positioned bool
	dragging   bool
	dragDX     int
	dragDY     int
	status     string
}

func (w *SettingsWindow) OpenWindow(ctx client.Context) {
	w.open = true
	w.EnsurePosition(ctx)
}

func (w *SettingsWindow) Update(ctx client.Context) bool {
	if !w.open || ctx.Input == nil {
		return false
	}
	w.EnsurePosition(ctx)
	width, height := ctx.ScreenSize()
	if w.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			w.x = clampSettingsWindowInt(ctx.Input.MouseX-w.dragDX, 8, maxInt(8, width-settingsWindowWidth-8))
			w.y = clampSettingsWindowInt(ctx.Input.MouseY-w.dragDY, 8, maxInt(8, height-settingsWindowHeight-8))
			return true
		}
		w.dragging = false
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) {
		w.open = false
		return true
	}
	inside := pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, w.x, w.y, settingsWindowWidth, settingsWindowHeight)
	if !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return inside
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if !inside {
		return false
	}
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		w.open = false
		return true
	}
	if pointInRect(mx, my, w.x, w.y, settingsWindowWidth, settingsWindowTitleH) {
		w.dragging = true
		w.dragDX = mx - w.x
		w.dragDY = my - w.y
		return true
	}
	if w.handleRuntimeToggleClick(ctx, mx, my) {
		return true
	}
	if w.handleVolumeClick(ctx, mx, my) {
		return true
	}
	return true
}

func (w *SettingsWindow) Draw(screen *render.Image, ctx client.Context) {
	if !w.open || screen == nil {
		return
	}
	w.EnsurePosition(ctx)
	x, y := w.x, w.y
	DrawTitledWindowFrame(screen, x, y, settingsWindowWidth, settingsWindowHeight, settingsWindowTitleH)
	DrawWindowTitle(screen, x, y, settingsWindowTitleH, settingsWindowPad, "Settings", TitleTextColor)
	cx, cy, cw, ch := w.closeBounds()
	DrawCloseButton(screen, cx, cy, cw, ch, ButtonColor, TextColor)

	labelX := x + settingsWindowPad
	rowY := y + settingsWindowTitleH + 18
	drawSettingsSectionLabel(screen, "Display", labelX, rowY)
	fullscreenX, fullscreenY, fullscreenW, fullscreenH := w.fullscreenToggleBounds()
	w.DrawRuntimeToggle(screen, ctx, "Fullscreen", fullscreenX, fullscreenY, fullscreenW, fullscreenH, settingsRuntimeFullscreen(ctx))
	vsyncX, vsyncY, vsyncW, vsyncH := w.vsyncToggleBounds()
	w.DrawRuntimeToggle(screen, ctx, "VSync (Restart)", vsyncX, vsyncY, vsyncW, vsyncH, settingsRuntimeVSync(ctx))
	fpsX, fpsY, fpsW, fpsH := w.fpsToggleBounds()
	w.DrawRuntimeToggle(screen, ctx, "FPS meter", fpsX, fpsY, fpsW, fpsH, settingsRuntimeFPS(ctx))

	soundY := rowY + 114
	drawSettingsSectionLabel(screen, "Sound", labelX, soundY)
	_, bgmBarY, _, bgmBarH := w.bgmVolumeBarBounds()
	render.DebugPrintAtColor(screen, "BGM Vol", labelX, settingsLabelY(bgmBarY, bgmBarH), TextColor)
	w.DrawVolumeControls(screen, ctx, settingsVolumeBGM(ctx), w.bgmVolumeMinusBounds, w.bgmVolumeBarBounds, w.bgmVolumePlusBounds)
	_, sfxBarY, _, sfxBarH := w.sfxVolumeBarBounds()
	render.DebugPrintAtColor(screen, "SFX Vol", labelX, settingsLabelY(sfxBarY, sfxBarH), TextColor)
	w.DrawVolumeControls(screen, ctx, settingsVolumeSFX(ctx), w.sfxVolumeMinusBounds, w.sfxVolumeBarBounds, w.sfxVolumePlusBounds)

}

func drawSettingsSectionLabel(screen *render.Image, label string, x, y int) {
	render.DebugPrintAtColor(screen, label, x, y, TitleTextColor)
	render.DebugPrintAtColor(screen, label, x+1, y, TitleTextColor)
}

func settingsLabelY(y, height int) int {
	return y + render.DebugTextTopForCenter(height)
}

func (w *SettingsWindow) IsOpen() bool {
	return w.open
}

func (w *SettingsWindow) DrawRuntimeToggle(screen *render.Image, ctx client.Context, label string, x, y, width, height int, on bool) {
	fill := ButtonColor
	if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, width, height) {
		fill = ButtonHoverColor
	}
	DrawCheckboxButton(screen, x, y, fill, TextColor, on)
	labelY := y + render.DebugTextTopForCenter(height)
	render.DebugPrintAtColor(screen, label, x+width+settingsCheckboxGap, labelY, TextColor)
}

func (w *SettingsWindow) DrawVolumeControls(screen *render.Image, ctx client.Context, volume float64, minusBounds, barBounds, plusBounds func() (int, int, int, int)) {
	minusX, minusY, minusW, minusH := minusBounds()
	plusX, plusY, plusW, plusH := plusBounds()
	barX, barY, barW, barH := barBounds()
	mx, my := -1, -1
	if ctx.Input != nil {
		mx, my = ctx.Input.MouseX, ctx.Input.MouseY
	}
	minusFill := ButtonColor
	if pointInRect(mx, my, minusX, minusY, minusW, minusH) {
		minusFill = ButtonHoverColor
	}
	plusFill := ButtonColor
	if pointInRect(mx, my, plusX, plusY, plusW, plusH) {
		plusFill = ButtonHoverColor
	}
	DrawMinusButton(screen, minusX, minusY, minusFill, TextColor)
	DrawSurface(screen, barX, barY, barW, barH, PanelBodyColor, ButtonBorderColor)
	fillW := int(volume * float64(barW-2))
	if fillW > 0 {
		render.DrawRect(screen, float64(barX+1), float64(barY+1), float64(fillW), float64(barH-2), color.RGBA{R: 104, G: 166, B: 224, A: 255})
	}
	DrawPlusButton(screen, plusX, plusY, plusFill, TextColor)
}

func (w *SettingsWindow) CursorAction(ctx client.Context) (int, bool) {
	if !w.open || ctx.Input == nil {
		return 0, false
	}
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	cx, cy, cw, ch := w.closeBounds()
	if pointInRect(mx, my, cx, cy, cw, ch) {
		return CursorActionClick, true
	}
	if pointInRect(mx, my, w.x, w.y, settingsWindowWidth, settingsWindowTitleH) {
		return CursorActionClick, true
	}
	for _, rect := range [][4]int{
		rectArray(w.fullscreenToggleBounds()),
		rectArray(w.vsyncToggleBounds()),
		rectArray(w.fpsToggleBounds()),
		rectArray(w.bgmVolumeMinusBounds()),
		rectArray(w.bgmVolumePlusBounds()),
		rectArray(w.sfxVolumeMinusBounds()),
		rectArray(w.sfxVolumePlusBounds()),
	} {
		if pointInRect(mx, my, rect[0], rect[1], rect[2], rect[3]) {
			return CursorActionClick, true
		}
	}
	if pointInRect(mx, my, w.x, w.y, settingsWindowWidth, settingsWindowHeight) {
		return CursorActionDefault, true
	}
	return 0, false
}

func (w *SettingsWindow) handleRuntimeToggleClick(ctx client.Context, mx, my int) bool {
	if ctx.Runtime == nil {
		return false
	}
	fullscreenX, fullscreenY, fullscreenW, fullscreenH := w.fullscreenToggleBounds()
	if pointInRect(mx, my, fullscreenX, fullscreenY, fullscreenW, fullscreenH) {
		next := !ctx.Runtime.Fullscreen()
		ctx.Runtime.SetFullscreen(next)
		w.saveSettings(ctx, fmt.Sprintf("fullscreen %s", settingsBoolText(next)))
		return true
	}
	vsyncX, vsyncY, vsyncW, vsyncH := w.vsyncToggleBounds()
	if pointInRect(mx, my, vsyncX, vsyncY, vsyncW, vsyncH) {
		next := !ctx.Runtime.VSync()
		ctx.Runtime.SetVSync(next)
		w.saveSettings(ctx, "vsync saved for restart")
		return true
	}
	fpsX, fpsY, fpsW, fpsH := w.fpsToggleBounds()
	if pointInRect(mx, my, fpsX, fpsY, fpsW, fpsH) {
		next := !ctx.Runtime.FPS()
		ctx.Runtime.SetFPS(next)
		w.saveSettings(ctx, fmt.Sprintf("fps meter %s", settingsBoolText(next)))
		return true
	}
	return false
}

func pointInAnyRect(mx, my int, bounds func() (int, int, int, int)) bool {
	x, y, width, height := bounds()
	return pointInRect(mx, my, x, y, width, height)
}

func (w *SettingsWindow) handleVolumeClick(ctx client.Context, mx, my int) bool {
	if ctx.Audio == nil {
		return false
	}
	if pointInAnyRect(mx, my, w.bgmVolumeMinusBounds) {
		ctx.Audio.SetBGMVolume(ctx.Audio.BGMVolume() - 0.1)
		w.saveSettings(ctx, fmt.Sprintf("bgm volume %d%%", int(ctx.Audio.BGMVolume()*100+0.5)))
		return true
	}
	if pointInAnyRect(mx, my, w.bgmVolumePlusBounds) {
		ctx.Audio.SetBGMVolume(ctx.Audio.BGMVolume() + 0.1)
		w.saveSettings(ctx, fmt.Sprintf("bgm volume %d%%", int(ctx.Audio.BGMVolume()*100+0.5)))
		return true
	}
	if pointInAnyRect(mx, my, w.sfxVolumeMinusBounds) {
		ctx.Audio.SetSFXVolume(ctx.Audio.SFXVolume() - 0.1)
		w.saveSettings(ctx, fmt.Sprintf("sfx volume %d%%", int(ctx.Audio.SFXVolume()*100+0.5)))
		return true
	}
	if pointInAnyRect(mx, my, w.sfxVolumePlusBounds) {
		ctx.Audio.SetSFXVolume(ctx.Audio.SFXVolume() + 0.1)
		w.saveSettings(ctx, fmt.Sprintf("sfx volume %d%%", int(ctx.Audio.SFXVolume()*100+0.5)))
		return true
	}
	return false
}

func (w *SettingsWindow) saveSettings(ctx client.Context, successStatus string) {
	settings := config.UserSettings{
		Fullscreen: settingsRuntimeFullscreen(ctx),
		VSync:      settingsRuntimeVSync(ctx),
		FPS:        settingsRuntimeFPS(ctx),
		BGMVolume:  settingsVolumeBGM(ctx),
		SFXVolume:  settingsVolumeSFX(ctx),
	}
	path, err := config.SaveUserSettings(settings)
	if err != nil {
		w.status = "settings save failed"
		log.Printf("settings save failed: %v", err)
		return
	}
	log.Printf("settings saved path=%s", path)
	w.status = successStatus
}

func (w *SettingsWindow) EnsurePosition(ctx client.Context) {
	if w.positioned {
		return
	}
	width, height := ctx.ScreenSize()
	w.x = maxInt(8, (width-settingsWindowWidth)/2)
	w.y = maxInt(8, (height-settingsWindowHeight)/2)
	w.positioned = true
}

func (w *SettingsWindow) closeBounds() (int, int, int, int) {
	return w.x + settingsWindowWidth - 25, w.y + 6, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) bgmVolumeMinusBounds() (int, int, int, int) {
	return w.x + 104, w.y + settingsWindowTitleH + 161, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) bgmVolumeBarBounds() (int, int, int, int) {
	x := w.x + 134
	plusX, _, _, _ := w.bgmVolumePlusBounds()
	return x, w.y + settingsWindowTitleH + 167, plusX - x - 6, settingsVolumeBarH
}

func (w *SettingsWindow) bgmVolumePlusBounds() (int, int, int, int) {
	return w.x + settingsWindowWidth - settingsWindowPad - IconButtonSize, w.y + settingsWindowTitleH + 161, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) sfxVolumeMinusBounds() (int, int, int, int) {
	return w.x + 104, w.y + settingsWindowTitleH + 199, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) sfxVolumeBarBounds() (int, int, int, int) {
	x := w.x + 134
	plusX, _, _, _ := w.sfxVolumePlusBounds()
	return x, w.y + settingsWindowTitleH + 205, plusX - x - 6, settingsVolumeBarH
}

func (w *SettingsWindow) sfxVolumePlusBounds() (int, int, int, int) {
	return w.x + settingsWindowWidth - settingsWindowPad - IconButtonSize, w.y + settingsWindowTitleH + 199, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) fullscreenToggleBounds() (int, int, int, int) {
	return w.x + settingsWindowPad, w.y + settingsWindowTitleH + 42, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) vsyncToggleBounds() (int, int, int, int) {
	return w.x + settingsWindowPad, w.y + settingsWindowTitleH + 72, IconButtonSize, IconButtonSize
}

func (w *SettingsWindow) fpsToggleBounds() (int, int, int, int) {
	return w.x + settingsWindowPad, w.y + settingsWindowTitleH + 102, IconButtonSize, IconButtonSize
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

func settingsBoolText(value bool) string {
	if value {
		return "on"
	}
	return "off"
}

func clampSettingsWindowInt(value, low, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}
