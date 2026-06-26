package gamemode

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/kivutar/goro/internal/network"
	"github.com/kivutar/goro/internal/render"
	worldstate "github.com/kivutar/goro/internal/world"
)

const (
	npcDialogWidth       = 560
	npcDialogHeight      = 178
	npcDialogMarginY     = 56
	npcDialogPad         = 12
	npcDialogButtonW     = 78
	npcDialogButtonH     = 24
	npcDialogLineH       = 14
	npcDialogMaxMessages = 32
)

var (
	npcDialogTextColor     = color.RGBA{R: 246, G: 242, B: 232, A: 255}
	npcDialogMutedColor    = color.RGBA{R: 174, G: 184, B: 194, A: 255}
	npcDialogTitleColor    = color.RGBA{R: 255, G: 230, B: 150, A: 255}
	npcDialogOptionColor   = color.RGBA{R: 214, G: 232, B: 255, A: 255}
	npcDialogSelectedColor = color.RGBA{R: 255, G: 245, B: 178, A: 255}
)

type npcDialogAction int

const (
	npcDialogActionNone npcDialogAction = iota
	npcDialogActionNext
	npcDialogActionClose
	npcDialogActionMenu
)

type npcDialogState struct {
	open        bool
	npcID       uint32
	lines       []string
	options     []string
	action      npcDialogAction
	clearOnText bool
	status      string
}

func (d *npcDialogState) apply(packet network.NPCDialog) {
	switch packet.Kind {
	case network.NPCDialogSay:
		if d.clearOnText {
			d.lines = d.lines[:0]
			d.clearOnText = false
		}
		d.open = true
		d.npcID = packet.NPCID
		d.action = npcDialogActionNone
		d.options = nil
		if packet.Message != "" {
			d.lines = append(d.lines, packet.Message)
			if len(d.lines) > npcDialogMaxMessages {
				d.lines = append([]string(nil), d.lines[len(d.lines)-npcDialogMaxMessages:]...)
			}
		}
	case network.NPCDialogNext:
		d.open = true
		d.npcID = packet.NPCID
		d.action = npcDialogActionNext
		d.options = nil
	case network.NPCDialogClose:
		d.open = true
		d.npcID = packet.NPCID
		d.action = npcDialogActionClose
		d.options = nil
	case network.NPCDialogMenu:
		d.open = true
		d.npcID = packet.NPCID
		d.action = npcDialogActionMenu
		d.options = append([]string(nil), packet.Options...)
	case network.NPCDialogClear:
		if !d.open || d.npcID == 0 || packet.NPCID == 0 || d.npcID == packet.NPCID {
			d.reset()
		}
	}
}

func (d *npcDialogState) reset() {
	*d = npcDialogState{}
}

func (d *npcDialogState) update(ctx Context) bool {
	if !d.open || ctx.Input == nil {
		return false
	}
	width, height := ctx.ScreenSize()
	x, y, w, h := npcDialogBounds(width, height)
	if ctx.Input.JustPressed(render.KeyEscape) {
		if d.action == npcDialogActionMenu {
			d.choose(ctx, 255)
		} else {
			d.close(ctx)
		}
		return true
	}
	if ctx.Input.JustPressed(render.KeyEnter) {
		switch d.action {
		case npcDialogActionNext:
			d.next(ctx)
		case npcDialogActionClose:
			d.close(ctx)
		}
		return true
	}
	if ctx.Input.MouseJustPressed(render.MouseButtonLeft) {
		mx, my := ctx.Input.MouseX, ctx.Input.MouseY
		if !pointInRect(mx, my, x, y, w, h) {
			return true
		}
		if d.action == npcDialogActionMenu {
			for i := range d.options {
				ox, oy, ow, oh := npcDialogOptionBounds(x, y, w, i)
				if pointInRect(mx, my, ox, oy, ow, oh) {
					d.choose(ctx, i+1)
					return true
				}
			}
		}
		if d.action == npcDialogActionNext || d.action == npcDialogActionClose {
			bx, by, bw, bh := npcDialogButtonBounds(x, y, w, h)
			if pointInRect(mx, my, bx, by, bw, bh) {
				if d.action == npcDialogActionNext {
					d.next(ctx)
				} else {
					d.close(ctx)
				}
				return true
			}
		}
		return true
	}
	return true
}

func (d *npcDialogState) next(ctx Context) {
	if ctx.Network == nil {
		d.status = "not connected"
		return
	}
	if err := ctx.Network.SendNPCNext(d.npcID); err != nil {
		d.status = err.Error()
		return
	}
	d.action = npcDialogActionNone
	d.clearOnText = true
	d.status = ""
}

func (d *npcDialogState) close(ctx Context) {
	if ctx.Network != nil && d.npcID != 0 {
		if err := ctx.Network.SendNPCClose(d.npcID); err != nil {
			d.status = err.Error()
			return
		}
	}
	d.reset()
}

func (d *npcDialogState) choose(ctx Context, choice int) {
	if ctx.Network == nil {
		d.status = "not connected"
		return
	}
	if choice < 1 {
		choice = 255
	}
	if choice > 255 {
		choice = 255
	}
	if err := ctx.Network.SendNPCMenuChoice(d.npcID, uint8(choice)); err != nil {
		d.status = err.Error()
		return
	}
	d.action = npcDialogActionNone
	d.options = nil
	d.status = ""
}

func (d *npcDialogState) draw(screen *render.Image, ctx Context, width, height int) {
	if !d.open || screen == nil {
		return
	}
	x, y, w, h := npcDialogBounds(width, height)
	render.DrawRect(screen, float64(x+3), float64(y+4), float64(w), float64(h), color.RGBA{A: 110})
	render.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), color.RGBA{R: 24, G: 26, B: 31, A: 232})
	render.DrawRect(screen, float64(x), float64(y), float64(w), 1, color.RGBA{R: 232, G: 218, B: 172, A: 180})
	render.DrawRect(screen, float64(x), float64(y+h-1), float64(w), 1, color.RGBA{R: 64, G: 58, B: 48, A: 220})
	render.DrawRect(screen, float64(x), float64(y), 1, float64(h), color.RGBA{R: 232, G: 218, B: 172, A: 150})
	render.DrawRect(screen, float64(x+w-1), float64(y), 1, float64(h), color.RGBA{R: 64, G: 58, B: 48, A: 220})
	render.DrawRect(screen, float64(x+8), float64(y+28), float64(w-16), 1, color.RGBA{R: 210, G: 200, B: 170, A: 80})

	title := d.title(ctx)
	render.DebugPrintAtColor(screen, title, x+npcDialogPad, y+10, npcDialogTitleColor)
	if d.status != "" {
		render.DebugPrintAtColor(screen, trimRunes(d.status, 48), x+w-260, y+10, color.RGBA{R: 255, G: 132, B: 132, A: 255})
	}

	lineY := y + 38
	lineMax := maxInt(12, (w-2*npcDialogPad)/7)
	for _, line := range wrapNPCDialogLines(d.lines, lineMax) {
		if lineY > y+h-38 {
			break
		}
		render.DebugPrintAtColor(screen, line, x+npcDialogPad, lineY, npcDialogTextColor)
		lineY += npcDialogLineH
	}

	if d.action == npcDialogActionMenu {
		d.drawMenu(screen, x, y, w, h)
		return
	}
	if d.action == npcDialogActionNext || d.action == npcDialogActionClose {
		label := "Next"
		if d.action == npcDialogActionClose {
			label = "Close"
		}
		drawNPCDialogButton(screen, x, y, w, h, label)
		return
	}
	render.DebugPrintAtColor(screen, "Waiting...", x+w-92, y+h-25, npcDialogMutedColor)
}

func (d *npcDialogState) drawMenu(screen *render.Image, x, y, w, h int) {
	if len(d.options) == 0 {
		render.DebugPrintAtColor(screen, "No options.", x+npcDialogPad, y+h-58, npcDialogMutedColor)
		return
	}
	visible := minInt(len(d.options), 5)
	for i := 0; i < visible; i++ {
		ox, oy, ow, oh := npcDialogOptionBounds(x, y, w, i)
		render.DrawRect(screen, float64(ox), float64(oy), float64(ow), float64(oh), color.RGBA{R: 42, G: 48, B: 58, A: 190})
		render.DrawRect(screen, float64(ox), float64(oy+oh-1), float64(ow), 1, color.RGBA{R: 86, G: 98, B: 114, A: 160})
		text := fmt.Sprintf("%d. %s", i+1, d.options[i])
		render.DebugPrintAtColor(screen, trimRunes(text, maxInt(8, (ow-12)/7)), ox+6, oy+5, npcDialogOptionColor)
	}
	if len(d.options) > visible {
		render.DebugPrintAtColor(screen, fmt.Sprintf("+%d more", len(d.options)-visible), x+w-80, y+h-34, npcDialogMutedColor)
	}
	render.DebugPrintAtColor(screen, "Esc: Cancel", x+w-96, y+h-18, npcDialogMutedColor)
}

func npcDialogBounds(width, height int) (int, int, int, int) {
	w := minInt(npcDialogWidth, maxInt(260, width-40))
	h := minInt(npcDialogHeight, maxInt(130, height-40))
	x := (width - w) / 2
	y := height - h - npcDialogMarginY
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func npcDialogButtonBounds(x, y, w, h int) (int, int, int, int) {
	return x + w - npcDialogButtonW - npcDialogPad, y + h - npcDialogButtonH - 10, npcDialogButtonW, npcDialogButtonH
}

func npcDialogOptionBounds(x, y, w, index int) (int, int, int, int) {
	return x + npcDialogPad, y + 92 + index*24, w - 2*npcDialogPad, 22
}

func drawNPCDialogButton(screen *render.Image, x, y, w, h int, label string) {
	bx, by, bw, bh := npcDialogButtonBounds(x, y, w, h)
	render.DrawRect(screen, float64(bx), float64(by), float64(bw), float64(bh), color.RGBA{R: 70, G: 76, B: 86, A: 235})
	render.DrawRect(screen, float64(bx), float64(by), float64(bw), 1, color.RGBA{R: 228, G: 218, B: 184, A: 150})
	render.DrawRect(screen, float64(bx), float64(by+bh-1), float64(bw), 1, color.RGBA{A: 220})
	tx := bx + (bw-len(label)*7)/2
	render.DebugPrintAtColor(screen, label, tx, by+6, npcDialogSelectedColor)
}

func (d *npcDialogState) title(ctx Context) string {
	name := ""
	if ctx.World != nil && d.npcID != 0 {
		if actor, ok := ctx.World.Actors[d.npcID]; ok {
			name = actorDisplayName(ctx, actor, false)
		}
	}
	if name == "" {
		name = "NPC"
	}
	return name
}

func wrapNPCDialogLines(lines []string, maxRunes int) []string {
	if maxRunes < 8 {
		maxRunes = 8
	}
	var out []string
	for _, line := range lines {
		for _, split := range strings.Split(line, "\n") {
			out = append(out, wrapNPCDialogLine(split, maxRunes)...)
		}
	}
	return out
}

func wrapNPCDialogLine(line string, maxRunes int) []string {
	words := strings.Fields(line)
	if len(words) == 0 {
		return nil
	}
	var out []string
	current := ""
	for _, word := range words {
		if current == "" {
			current = word
			continue
		}
		if len([]rune(current))+1+len([]rune(word)) <= maxRunes {
			current += " " + word
			continue
		}
		out = append(out, trimRunes(current, maxRunes))
		current = word
	}
	if current != "" {
		out = append(out, trimRunes(current, maxRunes))
	}
	return out
}

func pointInRect(px, py, x, y, w, h int) bool {
	return px >= x && py >= y && px < x+w && py < y+h
}

func clickedTalkTarget(ctx Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	actor, ok := hoveredCursorActor(ctx, projection, mouseX, mouseY, now, deadActors)
	if !ok || !cursorActorCanTalk(actor) {
		return worldstate.Actor{}, false
	}
	return actor, true
}
