package ui

import (
	"image/color"

	"github.com/kivutar/goro/render"
)

type ConfirmModalOptions struct {
	ScreenW, ScreenH int
	X, Y, W, H       int
	Title            string
	Message          string
	MouseX, MouseY   int
	HasMouse         bool
}

func DrawConfirmModal(screen *render.Image, opts ConfirmModalOptions) {
	if screen == nil {
		return
	}
	DrawSurface(screen, 0, 0, opts.ScreenW, opts.ScreenH, color.RGBA{A: 80}, color.RGBA{})
	DrawTitledWindowFrame(screen, opts.X, opts.Y, opts.W, opts.H, 24)
	DrawWindowTitle(screen, opts.X, opts.Y, 24, 12, opts.Title, TitleTextColor)
	render.DebugPrintAtColor(screen, opts.Message, opts.X+28, opts.Y+52, TextColor)

	okX, okY, okW, okH := ConfirmModalOKRect(opts)
	cancelX, cancelY, cancelW, cancelH := ConfirmModalCancelRect(opts)
	drawConfirmButton(screen, opts, okX, okY, okW, okH, "OK")
	drawConfirmButton(screen, opts, cancelX, cancelY, cancelW, cancelH, "Cancel")
}

func ConfirmModalOKRect(opts ConfirmModalOptions) (int, int, int, int) {
	return opts.X + opts.W - 150, opts.Y + opts.H - 36, 56, 23
}

func ConfirmModalCancelRect(opts ConfirmModalOptions) (int, int, int, int) {
	return opts.X + opts.W - 84, opts.Y + opts.H - 36, 66, 23
}

func drawConfirmButton(screen *render.Image, opts ConfirmModalOptions, x, y, w, h int, label string) {
	fill := ButtonColor
	if opts.HasMouse && pointInRect(opts.MouseX, opts.MouseY, x, y, w, h) {
		fill = ButtonHoverColor
	}
	DrawButtonLabel(screen, x, y, w, h, label, fill, TextColor)
}
