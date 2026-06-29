package gamemode

import (
	"fmt"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
)

const (
	basicMenuX       = characterWindowX
	basicMenuY       = characterWindowY + characterWindowHeight + 6
	basicMenuCols    = 4
	basicMenuRows    = 2
	basicMenuButtonW = 72
	basicMenuButtonH = 24
	basicMenuGapX    = 6
	basicMenuGapY    = 5
	basicMenuPad     = 8
)

var (
	basicMenuTextColor   = uiTextColor
	basicMenuMutedColor  = uiMutedTextColor
	basicMenuButtonColor = uiButtonColor
	basicMenuHoverColor  = uiButtonHoverColor
	basicMenuDownColor   = uiButtonDownColor
	basicMenuPanelColor  = uiWindowBodyColor
)

type basicMenuState struct {
	lastAction string
	lastClick  time.Time
}

type basicMenuButton struct {
	key   string
	label string
}

var basicMenuButtons = []basicMenuButton{
	{key: "status", label: "Status"},
	{key: "option", label: "Option"},
	{key: "items", label: "Items"},
	{key: "equip", label: "Equip"},
	{key: "skill", label: "Skill"},
	{key: "map", label: "Map"},
	{key: "comm", label: "Comm"},
	{key: "friend", label: "Friend"},
}

func (m *basicMenuState) update(ctx Context) bool {
	if ctx.Input == nil || !ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		return false
	}
	index, ok := basicMenuButtonAt(ctx.Input.MouseX, ctx.Input.MouseY)
	if !ok {
		return false
	}
	m.lastAction = basicMenuButtons[index].key
	m.lastClick = time.Now()
	return true
}

func (m *basicMenuState) draw(screen *render.Image, ctx Context) {
	if screen == nil {
		return
	}
	x, y, w, h := basicMenuBounds()
	drawUIPanelSurface(screen, x, y, w, h, basicMenuPanelColor)

	mouseX, mouseY := -1, -1
	mouseDown := false
	if ctx.Input != nil {
		mouseX, mouseY = ctx.Input.MouseX, ctx.Input.MouseY
		mouseDown = ctx.Input.MousePressed(render.MouseButtonLeft)
	}
	hoverIndex, hoverOK := basicMenuButtonAt(mouseX, mouseY)
	for i, button := range basicMenuButtons {
		bx, by, bw, bh := basicMenuButtonBounds(i)
		fill := basicMenuButtonColor
		if hoverOK && hoverIndex == i {
			if mouseDown {
				fill = basicMenuDownColor
			} else {
				fill = basicMenuHoverColor
			}
		}
		drawUIButtonLabel(screen, bx, by, bw, bh, button.label, fill, basicMenuTextColor)
	}
	if m.lastAction != "" && m.lastAction != "status" && m.lastAction != "skill" && m.lastAction != "items" && m.lastAction != "equip" && time.Since(m.lastClick) < 1500*time.Millisecond {
		label := strings.ToUpper(m.lastAction[:1]) + m.lastAction[1:]
		render.DebugPrintAtColor(screen, fmt.Sprintf("%s: not implemented", label), x+basicMenuPad, y+h+6, basicMenuMutedColor)
	}
}

func (m *basicMenuState) cursorAction(ctx Context) (int, bool) {
	if ctx.Input == nil {
		return 0, false
	}
	if _, ok := basicMenuButtonAt(ctx.Input.MouseX, ctx.Input.MouseY); ok {
		return cursorActionClick, true
	}
	return 0, false
}

func basicMenuBounds() (int, int, int, int) {
	w := basicMenuPad*2 + basicMenuCols*basicMenuButtonW + (basicMenuCols-1)*basicMenuGapX
	h := basicMenuPad*2 + basicMenuRows*basicMenuButtonH + (basicMenuRows-1)*basicMenuGapY
	return basicMenuX, basicMenuY, w, h
}

func basicMenuButtonBounds(index int) (int, int, int, int) {
	col := index % basicMenuCols
	row := index / basicMenuCols
	x := basicMenuX + basicMenuPad + col*(basicMenuButtonW+basicMenuGapX)
	y := basicMenuY + basicMenuPad + row*(basicMenuButtonH+basicMenuGapY)
	return x, y, basicMenuButtonW, basicMenuButtonH
}

func basicMenuButtonAt(mouseX, mouseY int) (int, bool) {
	x, y, w, h := basicMenuBounds()
	if !pointInRect(mouseX, mouseY, x, y, w, h) {
		return 0, false
	}
	for i := range basicMenuButtons {
		bx, by, bw, bh := basicMenuButtonBounds(i)
		if pointInRect(mouseX, mouseY, bx, by, bw, bh) {
			return i, true
		}
	}
	return 0, false
}
