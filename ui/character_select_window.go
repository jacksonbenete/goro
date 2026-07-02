package ui

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

type CharacterPreviewFunc func(screen *render.Image, character session.Character, centerX, feetY int)

type CharacterSelectWindowOptions struct {
	X, Y, W, H      int
	TitleH          int
	FooterH         int
	FooterPadX      int
	FooterGap       int
	ButtonH         int
	PreviewFeetLift int
	SelectedSlot    int
	MaxSlots        int
	Characters      []session.Character
	MouseX, MouseY  int
	HasMouse        bool
	DrawPreview     CharacterPreviewFunc
}

func DrawCharacterSelectWindow(screen *render.Image, opts CharacterSelectWindowOptions) {
	if screen == nil {
		return
	}
	DrawTitledWindowFrame(screen, opts.X, opts.Y, opts.W, opts.H, opts.TitleH)
	DrawWindowTitle(screen, opts.X, opts.Y, opts.TitleH, 12, "Select Character", TitleTextColor)

	page := CharacterSelectPage(opts.SelectedSlot)
	pageStart := page * 3
	for localSlot := 0; localSlot < 3; localSlot++ {
		slot := pageStart + localSlot
		slotX, slotY, slotW, slotH := CharacterSelectSlotRect(opts, localSlot)
		selected := slot == opts.SelectedSlot
		bg := PanelBodyColor
		border := WindowBorderColor
		if selected {
			bg = SelectionColor
			border = SelectionBorder
		}
		DrawSurface(screen, slotX, slotY, slotW, slotH, bg, border)
		if character, ok := characterBySlot(opts.Characters, slot); ok {
			if opts.DrawPreview != nil {
				opts.DrawPreview(screen, character, slotX+slotW/2, slotY+slotH-15-opts.PreviewFeetLift)
			}
		} else {
			render.DebugPrintAtColor(screen, "Create", slotX+45, slotY+58, MutedTextColor)
		}
	}

	leftX, leftY, leftW, leftH := CharacterSelectLeftArrowRect(opts)
	rightX, rightY, rightW, rightH := CharacterSelectRightArrowRect(opts)
	drawCharSelectArrow(screen, leftX, leftY, leftW, leftH, "<")
	drawCharSelectArrow(screen, rightX, rightY, rightW, rightH, ">")
	drawSelectedCharacterInfo(screen, opts)
	drawCharacterSelectFooter(screen, opts)
}

func CharacterSelectSlotRect(opts CharacterSelectWindowOptions, localSlot int) (int, int, int, int) {
	lefts := [3]int{60, 224, 386}
	if localSlot < 0 || localSlot >= len(lefts) {
		localSlot = 0
	}
	return opts.X + lefts[localSlot] - 5, opts.Y + 40, 139, 144
}

func CharacterSelectLeftArrowRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return opts.X + 24, opts.Y + 105, 18, 18
}

func CharacterSelectRightArrowRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return opts.X + 534, opts.Y + 105, 18, 18
}

func CharacterSelectFooterRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return opts.X, opts.Y + opts.H - opts.FooterH, opts.W, opts.FooterH
}

func CharacterSelectInfoPanelRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return opts.X + 16, opts.Y + 202, 318, 88
}

func CharacterSelectPagerTextRect(opts CharacterSelectWindowOptions, label string) (int, int, int, int) {
	_, slotY, _, slotH := CharacterSelectSlotRect(opts, 0)
	_, panelY, _, _ := CharacterSelectInfoPanelRect(opts)
	textW, textH := render.DebugTextSize(label)
	gap := panelY - (slotY + slotH)
	if gap < textH {
		gap = textH
	}
	return opts.X + (opts.W-textW)/2, slotY + slotH + (gap-textH)/2, textW, textH
}

func CharacterSelectDeleteButtonRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	_, footerY, _, footerH := CharacterSelectFooterRect(opts)
	return opts.X + opts.FooterPadX, footerY + (footerH-opts.ButtonH)/2, ButtonLabelWidth("Delete"), opts.ButtonH
}

func CharacterSelectMakeButtonRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return characterSelectRightButtonRect(opts, 0)
}

func CharacterSelectOKButtonRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return characterSelectRightButtonRect(opts, 1)
}

func CharacterSelectCancelButtonRect(opts CharacterSelectWindowOptions) (int, int, int, int) {
	return characterSelectRightButtonRect(opts, 2)
}

func characterSelectRightButtonRect(opts CharacterSelectWindowOptions, index int) (int, int, int, int) {
	labels := [...]string{"Make", "OK", "Cancel"}
	if index < 0 || index >= len(labels) {
		index = 0
	}
	_, footerY, _, footerH := CharacterSelectFooterRect(opts)
	totalW := 0
	for _, label := range labels {
		totalW += ButtonLabelWidth(label)
	}
	totalW += opts.FooterGap * (len(labels) - 1)
	bx := opts.X + opts.W - opts.FooterPadX - totalW
	for i := 0; i < index; i++ {
		bx += ButtonLabelWidth(labels[i]) + opts.FooterGap
	}
	return bx, footerY + (footerH-opts.ButtonH)/2, ButtonLabelWidth(labels[index]), opts.ButtonH
}

func CharacterSelectPage(slot int) int {
	if slot < 0 {
		return 0
	}
	return slot / 3
}

func drawSelectedCharacterInfo(screen *render.Image, opts CharacterSelectWindowOptions) {
	character, ok := characterBySlot(opts.Characters, opts.SelectedSlot)
	panelX, panelY, panelW, panelH := CharacterSelectInfoPanelRect(opts)
	DrawPanelSurface(screen, panelX, panelY, panelW, panelH, PanelBodyColor)
	if !ok {
		render.DebugPrintAtColor(screen, "Empty Slot", panelX+18, panelY+14, TextColor)
		render.DebugPrintAtColor(screen, "Use Make to create a character later.", panelX+18, panelY+34, TextColor)
		return
	}
	render.DebugPrintAtColor(screen, trimRunes(character.Name, 24), panelX+14, panelY+8, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Job: %s", trimRunes(CharacterJobName(character), 18)), panelX+14, panelY+24, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Lv: %d / Job %d", character.Level, character.JobLevel), panelX+14, panelY+40, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("HP: %d / %d", character.HP, character.MaxHP), panelX+14, panelY+56, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("SP: %d / %d", character.SP, character.MaxSP), panelX+14, panelY+72, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("STR %d", character.Str), panelX+180, panelY+8, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("AGI %d", character.Agi), panelX+180, panelY+24, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("VIT %d", character.Vit), panelX+180, panelY+40, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("INT %d", character.Int), panelX+246, panelY+8, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("DEX %d", character.Dex), panelX+246, panelY+24, TextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("LUK %d", character.Luk), panelX+246, panelY+40, TextColor)
}

func drawCharacterSelectFooter(screen *render.Image, opts CharacterSelectWindowOptions) {
	page := CharacterSelectPage(opts.SelectedSlot)
	pageCount := maxInt(1, (opts.MaxSlots+2)/3)
	DrawWindowFooter(screen, opts.X, opts.Y, opts.W, opts.H, opts.FooterH)
	pageLabel := fmt.Sprintf("%d / %d", page+1, pageCount)
	pageX, pageY, _, _ := CharacterSelectPagerTextRect(opts, pageLabel)
	render.DebugPrintAtColor(screen, pageLabel, pageX, pageY, MutedTextColor)

	drawCharSelectButtonRect(screen, opts, rectArray(CharacterSelectDeleteButtonRect(opts)), "Delete", TextColor)
	drawCharSelectButtonRect(screen, opts, rectArray(CharacterSelectMakeButtonRect(opts)), "Make", TextColor)
	drawCharSelectButtonRect(screen, opts, rectArray(CharacterSelectOKButtonRect(opts)), "OK", TextColor)
	drawCharSelectButtonRect(screen, opts, rectArray(CharacterSelectCancelButtonRect(opts)), "Cancel", TextColor)
}

func drawCharSelectButtonRect(screen *render.Image, opts CharacterSelectWindowOptions, rect [4]int, label string, textColor color.RGBA) {
	x, y, w, h := rect[0], rect[1], rect[2], rect[3]
	bg := ButtonColor
	if opts.HasMouse && pointInRect(opts.MouseX, opts.MouseY, x, y, w, h) {
		bg = ButtonHoverColor
	}
	DrawButtonLabel(screen, x, y, w, h, label, bg, textColor)
}

func drawCharSelectArrow(screen *render.Image, x, y, w, h int, label string) {
	DrawButtonLabel(screen, x, y, w, h, label, ButtonColor, TextColor)
}

func characterBySlot(characters []session.Character, slot int) (session.Character, bool) {
	for _, character := range characters {
		if int(character.Slot) == slot {
			return character, true
		}
	}
	return session.Character{}, false
}
