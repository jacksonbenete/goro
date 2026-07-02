package ui

import (
	"github.com/kivutar/goro/render"
)

func DrawTextInput(screen *render.Image, x, y, w, h int, text string, focused bool) {
	bg := PanelBodyColor
	border := ButtonBorderColor
	if focused {
		border = SelectionBorder
	}
	DrawTextBoxSurface(screen, x, y, w, h, bg, border)
	visibleText := trimRunes(text, maxInt(1, (w-14)/7))
	render.DebugPrintAtColor(screen, visibleText, x+6, y+4, TextColor)
	if focused {
		textW, _ := render.DebugTextSize(visibleText)
		caretX := minInt(x+w-6, x+6+textW)
		DrawBlinkingCaret(screen, caretX, y, h, TextColor)
	}
}
