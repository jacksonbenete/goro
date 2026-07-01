package game

import (
	"image/color"

	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
	goroui "github.com/kivutar/goro/ui"
)

var (
	uiWindowRadius      = goroui.WindowRadius
	uiButtonRadius      = goroui.ButtonRadius
	uiWindowBodyColor   = goroui.WindowBodyColor
	uiWindowTitleTop    = goroui.WindowTitleTop
	uiWindowTitleColor  = goroui.WindowTitleColor
	uiWindowBorderColor = goroui.WindowBorderColor
	uiPanelBodyColor    = goroui.PanelBodyColor
	uiPanelAltColor     = goroui.PanelAltColor
	uiPanelHoverColor   = goroui.PanelHoverColor
	uiDisabledColor     = goroui.DisabledColor
	uiTextColor         = goroui.TextColor
	uiMutedTextColor    = goroui.MutedTextColor
	uiTitleTextColor    = goroui.TitleTextColor
	uiGoodTextColor     = goroui.GoodTextColor
	uiErrorTextColor    = goroui.ErrorTextColor
	uiButtonColor       = goroui.ButtonColor
	uiButtonHoverColor  = goroui.ButtonHoverColor
	uiButtonDownColor   = goroui.ButtonDownColor
	uiButtonBorderColor = goroui.ButtonBorderColor
	uiSeparatorColor    = goroui.SeparatorColor
	uiSelectionColor    = goroui.SelectionColor
	uiSelectionBorder   = goroui.SelectionBorder
)

func drawUIWindowFrame(screen *render.Image, x, y, w, h int) {
	goroui.DrawWindowFrame(screen, x, y, w, h)
}

func drawUITitledWindowFrame(screen *render.Image, x, y, w, h, titleH int) {
	goroui.DrawTitledWindowFrame(screen, x, y, w, h, titleH)
}

func drawUIWindowTitle(screen *render.Image, x, y, titleH, pad int, title string, text color.RGBA) {
	goroui.DrawWindowTitle(screen, x, y, titleH, pad, title, text)
}

func drawUITitleTextAt(screen *render.Image, x, y, titleH int, title string, text color.RGBA) {
	goroui.DrawTitleTextAt(screen, x, y, titleH, title, text)
}

func drawUIPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	goroui.DrawPanelSurface(screen, x, y, w, h, bg)
}

func drawUIButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	goroui.DrawButtonSurface(screen, x, y, w, h, bg)
}

func drawUIButtonLabel(screen *render.Image, x, y, w, h int, label string, bg, text color.RGBA) {
	goroui.DrawButtonLabel(screen, x, y, w, h, label, bg, text)
}

func drawUICloseButton(screen *render.Image, x, y, w, h int, bg, line color.RGBA) {
	goroui.DrawCloseButton(screen, x, y, w, h, bg, line)
}

func drawUICenteredText(screen *render.Image, x, y, w, h int, label string, text color.RGBA) {
	goroui.DrawCenteredText(screen, x, y, w, h, label, text)
}

func drawUIRowSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	goroui.DrawRowSurface(screen, x, y, w, h, bg)
}

func drawUISurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA) {
	goroui.DrawSurface(screen, x, y, w, h, bg, border)
}

func uiColor(c color.RGBA) uiwidget.Color {
	return goroui.Color(c)
}
