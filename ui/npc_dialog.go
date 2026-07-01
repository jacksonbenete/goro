package ui

import (
	"fmt"
	"image/color"
	"math"
	"strings"
	"unicode/utf8"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
)

const (
	npcDialogWidth       = 560
	npcDialogHeight      = 178
	npcDialogPad         = 12
	npcDialogTitleH      = 30
	npcDialogButtonW     = 78
	npcDialogButtonH     = 24
	npcDialogLineH       = 14
	npcDialogMaxMessages = 32
	npcMenuWidth         = 260
	npcMenuMinRows       = 4
	npcMenuMaxRows       = 8
	npcMenuRowH          = 24
	npcMenuTitleH        = 26
	npcMenuTopPad        = 8
	npcMenuBottomH       = 30
	npcMenuMinHeight     = npcMenuTitleH + npcMenuTopPad + npcMenuMinRows*npcMenuRowH + npcMenuBottomH
)

var (
	npcDialogTextColor   = TextColor
	npcDialogMutedColor  = MutedTextColor
	npcDialogTitleColor  = TitleTextColor
	npcDialogOptionColor = TextColor
)

type npcDialogAction int

const (
	npcDialogActionNone npcDialogAction = iota
	npcDialogActionNext
	npcDialogActionClose
	npcDialogActionMenu
)

type NPCDialog struct {
	open        bool
	npcID       uint32
	lines       []string
	options     []string
	action      npcDialogAction
	clearOnText bool
	status      string

	positioned bool
	x          int
	y          int
	dragging   bool
	dragDX     int
	dragDY     int

	menuPositioned bool
	menuX          int
	menuY          int
	menuDragging   bool
	menuDragDX     int
	menuDragDY     int
	menuScroll     int
}

type npcDialogTextRun struct {
	text  string
	color color.RGBA
}

func (d *NPCDialog) Apply(packet network.NPCDialog) {
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
		d.menuScroll = 0
	case network.NPCDialogClear:
		if !d.open || d.npcID == 0 || packet.NPCID == 0 || d.npcID == packet.NPCID {
			d.Reset()
		}
	}
}

func (d *NPCDialog) Reset() {
	*d = NPCDialog{}
}

func (d *NPCDialog) Update(ctx Context) bool {
	if !d.open || ctx.Input == nil {
		return false
	}
	width, height := ctx.ScreenSize()
	x, y, w, h := d.resolvedDialogBounds(width, height)
	menuX, menuY, menuW, menuH := d.menuBounds(width, height, x, y, w, h)
	if d.updateMenuScrollAt(ctx, menuX, menuY, menuW, menuH) {
		return true
	}
	if d.dragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			d.x = clampNPCDialogInt(ctx.Input.MouseX-d.dragDX, 8, maxInt(8, width-w-8))
			d.y = clampNPCDialogInt(ctx.Input.MouseY-d.dragDY, 8, maxInt(8, height-h-8))
			d.positioned = true
			return true
		}
		d.dragging = false
	}
	if d.menuDragging {
		if ctx.Input.MousePressed(render.MouseButtonLeft) {
			d.menuX = clampNPCDialogInt(ctx.Input.MouseX-d.menuDragDX, 8, maxInt(8, width-menuW-8))
			d.menuY = clampNPCDialogInt(ctx.Input.MouseY-d.menuDragDY, 8, maxInt(8, height-menuH-8))
			d.menuPositioned = true
			return true
		}
		d.menuDragging = false
	}
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
		if d.action == npcDialogActionMenu && pointInRect(mx, my, menuX, menuY, menuW, menuH) {
			if pointInRect(mx, my, menuX, menuY, menuW, npcMenuTitleH) {
				d.menuDragging = true
				d.menuDragDX = mx - menuX
				d.menuDragDY = my - menuY
				return true
			}
			start, end := d.visibleMenuRange(menuH)
			for i, optionIndex := 0, start; optionIndex < end; i, optionIndex = i+1, optionIndex+1 {
				ox, oy, ow, oh := npcDialogOptionBounds(menuX, menuY, menuW, i)
				if pointInRect(mx, my, ox, oy, ow, oh) {
					d.choose(ctx, optionIndex+1)
					return true
				}
			}
			cancelX, cancelY, cancelW, cancelH := npcDialogMenuCancelBounds(menuX, menuY, menuW, menuH)
			if pointInRect(mx, my, cancelX, cancelY, cancelW, cancelH) {
				d.choose(ctx, 255)
				return true
			}
			return true
		}
		if !pointInRect(mx, my, x, y, w, h) {
			return true
		}
		if pointInRect(mx, my, x, y, w, npcDialogTitleH) {
			d.dragging = true
			d.dragDX = mx - x
			d.dragDY = my - y
			return true
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

func (d *NPCDialog) next(ctx Context) {
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

func (d *NPCDialog) close(ctx Context) {
	if ctx.Network != nil && d.npcID != 0 {
		if err := ctx.Network.SendNPCClose(d.npcID); err != nil {
			d.status = err.Error()
			return
		}
	}
	d.Reset()
}

func (d *NPCDialog) choose(ctx Context, choice int) {
	if ctx.Network == nil {
		d.status = "not connected"
		return
	}
	cancel := choice < 1 || choice > 254
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
	if cancel || choice == 255 {
		d.Reset()
		return
	}
	d.action = npcDialogActionNone
	d.options = nil
	d.menuDragging = false
	d.status = ""
}

func (d *NPCDialog) Draw(screen *render.Image, ctx Context, width, height int) {
	if !d.open || screen == nil {
		return
	}
	x, y, w, h := d.resolvedDialogBounds(width, height)
	DrawTitledWindowFrame(screen, x, y, w, h, npcDialogTitleH)

	title := d.title(ctx)
	DrawWindowTitle(screen, x, y, npcDialogTitleH, npcDialogPad, title, npcDialogTitleColor)
	if d.status != "" {
		render.DebugPrintAtColor(screen, trimRunes(d.status, 48), x+w-260, y+10, ErrorTextColor)
	}

	lineY := y + 38
	lineMax := maxInt(12, (w-2*npcDialogPad)/7)
	for _, line := range wrapNPCDialogLines(d.lines, lineMax) {
		if lineY > y+h-38 {
			break
		}
		drawNPCDialogTextRuns(screen, line, x+npcDialogPad, lineY)
		lineY += npcDialogLineH
	}

	if d.action == npcDialogActionMenu {
		menuX, menuY, menuW, menuH := d.menuBounds(width, height, x, y, w, h)
		d.drawMenu(screen, menuX, menuY, menuW, menuH)
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

func (d *NPCDialog) drawMenu(screen *render.Image, x, y, w, h int) {
	DrawTitledWindowFrame(screen, x, y, w, h, npcMenuTitleH)
	DrawWindowTitle(screen, x, y, npcMenuTitleH, npcDialogPad, "Choose", npcDialogTitleColor)
	if len(d.options) == 0 {
		render.DebugPrintAtColor(screen, "No options.", x+npcDialogPad, y+npcMenuTitleH+12, npcDialogMutedColor)
		return
	}
	start, end := d.visibleMenuRange(h)
	for i, optionIndex := 0, start; optionIndex < end; i, optionIndex = i+1, optionIndex+1 {
		ox, oy, ow, oh := npcDialogOptionBounds(x, y, w, i)
		DrawButtonSurface(screen, ox, oy, ow, oh, ButtonColor)
		runs := npcDialogTextRuns(d.options[optionIndex], npcDialogOptionColor)
		runs = append([]npcDialogTextRun{{text: fmt.Sprintf("%d. ", optionIndex+1), color: npcDialogOptionColor}}, runs...)
		runs = trimNPCDialogTextRuns(runs, maxInt(8, (ow-12)/7))
		_, textH := render.DebugTextSize("Ag")
		drawNPCDialogTextRuns(screen, runs, ox+6, oy+(oh-textH)/2)
	}
	if len(d.options) > maxNPCMenuVisibleRows(h) {
		d.drawMenuScrollBar(screen, x, y, w, h)
	}
	cancelX, cancelY, cancelW, cancelH := npcDialogMenuCancelBounds(x, y, w, h)
	DrawButtonLabel(screen, cancelX, cancelY, cancelW, cancelH, "Cancel", ButtonColor, TextColor)
}

func npcDialogBounds(width, height int) (int, int, int, int) {
	w := minInt(npcDialogWidth, maxInt(260, width-40))
	h := minInt(npcDialogHeight, maxInt(130, height-40))
	x := (width - w) / 2
	y := (height - h) / 2
	if y < 16 {
		y = 16
	}
	return x, y, w, h
}

func (d *NPCDialog) resolvedDialogBounds(width, height int) (int, int, int, int) {
	x, y, w, h := npcDialogBounds(width, height)
	if d.positioned {
		x = clampNPCDialogInt(d.x, 8, maxInt(8, width-w-8))
		y = clampNPCDialogInt(d.y, 8, maxInt(8, height-h-8))
	}
	return x, y, w, h
}

func (d *NPCDialog) menuBounds(width, height, dialogX, dialogY, dialogW, dialogH int) (int, int, int, int) {
	w := minInt(npcMenuWidth, maxInt(220, width-40))
	rows := maxInt(npcMenuMinRows, minInt(len(d.options), npcMenuMaxRows))
	h := maxInt(npcMenuMinHeight, npcMenuTitleH+npcMenuTopPad+rows*npcMenuRowH+npcMenuBottomH)
	x := dialogX + (dialogW-w)/2
	y := dialogY + dialogH + 8
	if d.menuPositioned {
		x = d.menuX
		y = d.menuY
	}
	x = clampNPCDialogInt(x, 8, maxInt(8, width-w-8))
	y = clampNPCDialogInt(y, 8, maxInt(8, height-h-8))
	return x, y, w, h
}

func npcDialogButtonBounds(x, y, w, h int) (int, int, int, int) {
	return x + w - npcDialogButtonW - npcDialogPad, y + h - npcDialogButtonH - 10, npcDialogButtonW, npcDialogButtonH
}

func npcDialogOptionBounds(x, y, w, index int) (int, int, int, int) {
	return x + npcDialogPad, y + npcMenuTitleH + npcMenuTopPad + index*npcMenuRowH, w - 2*npcDialogPad, 22
}

func npcDialogMenuCancelBounds(x, y, w, h int) (int, int, int, int) {
	return x + w - 68, y + h - 24, 56, 18
}

func maxNPCMenuVisibleRows(height int) int {
	return maxInt(1, (height-npcMenuTitleH-npcMenuTopPad-npcMenuBottomH)/npcMenuRowH)
}

func (d *NPCDialog) UpdateMenuScroll(ctx Context) bool {
	if !d.open || d.action != npcDialogActionMenu || ctx.Input == nil || ctx.Input.WheelY == 0 {
		return false
	}
	width, height := ctx.ScreenSize()
	dialogX, dialogY, dialogW, dialogH := d.resolvedDialogBounds(width, height)
	x, y, w, h := d.menuBounds(width, height, dialogX, dialogY, dialogW, dialogH)
	return d.updateMenuScrollAt(ctx, x, y, w, h)
}

func (d *NPCDialog) updateMenuScrollAt(ctx Context, x, y, w, h int) bool {
	if d.action != npcDialogActionMenu || ctx.Input == nil || ctx.Input.WheelY == 0 {
		return false
	}
	if !pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, w, h) {
		return false
	}
	d.scrollMenuBy(ctx.Input.WheelY, h)
	ctx.Input.WheelY = 0
	return true
}

func (d *NPCDialog) scrollMenuBy(wheelY float64, height int) {
	step := int(math.Ceil(math.Abs(wheelY)))
	if step < 1 {
		step = 1
	}
	if wheelY > 0 {
		d.menuScroll -= step
	} else {
		d.menuScroll += step
	}
	d.clampMenuScroll(height)
}

func (d *NPCDialog) visibleMenuRange(height int) (int, int) {
	d.clampMenuScroll(height)
	visible := minInt(len(d.options), maxNPCMenuVisibleRows(height))
	start := d.menuScroll
	end := minInt(len(d.options), start+visible)
	return start, end
}

func (d *NPCDialog) clampMenuScroll(height int) {
	maxScroll := maxInt(0, len(d.options)-maxNPCMenuVisibleRows(height))
	if d.menuScroll < 0 {
		d.menuScroll = 0
	}
	if d.menuScroll > maxScroll {
		d.menuScroll = maxScroll
	}
}

func (d *NPCDialog) drawMenuScrollBar(screen *render.Image, x, y, w, h int) {
	visible := maxNPCMenuVisibleRows(h)
	if len(d.options) <= visible {
		return
	}
	trackX := x + w - 9
	trackY := y + npcMenuTitleH + npcMenuTopPad
	trackH := maxInt(1, visible*npcMenuRowH-2)
	total := len(d.options)
	maxScroll := maxInt(1, total-visible)
	thumbH := maxInt(18, trackH*visible/total)
	thumbTravel := maxInt(1, trackH-thumbH)
	thumbY := trackY + thumbTravel*d.menuScroll/maxScroll
	render.DrawRect(screen, float64(trackX), float64(trackY), 3, float64(trackH), PanelAltColor)
	render.DrawRect(screen, float64(trackX), float64(thumbY), 3, float64(thumbH), WindowBorderColor)
}

func drawNPCWindowFrame(screen *render.Image, x, y, w, h int) {
	DrawTitledWindowFrame(screen, x, y, w, h, 28)
}

func drawNPCDialogButton(screen *render.Image, x, y, w, h int, label string) {
	bx, by, bw, bh := npcDialogButtonBounds(x, y, w, h)
	DrawButtonLabel(screen, bx, by, bw, bh, label, ButtonColor, TextColor)
}

func (d *NPCDialog) title(ctx Context) string {
	name := ""
	if ctx.World != nil && d.npcID != 0 {
		if actor, ok := ctx.World.Actors[d.npcID]; ok {
			name = strings.TrimSpace(actor.Name)
		}
	}
	if name == "" {
		name = "NPC"
	}
	return name
}

func wrapNPCDialogLines(lines []string, maxRunes int) [][]npcDialogTextRun {
	if maxRunes < 8 {
		maxRunes = 8
	}
	var out [][]npcDialogTextRun
	for _, line := range lines {
		for _, split := range strings.Split(line, "\n") {
			out = append(out, wrapNPCDialogTextRuns(npcDialogTextRuns(split, npcDialogTextColor), maxRunes)...)
		}
	}
	return out
}

type npcDialogColoredRune struct {
	r     rune
	color color.RGBA
}

func npcDialogTextRuns(text string, base color.RGBA) []npcDialogTextRun {
	current := base
	var runs []npcDialogTextRun
	var b strings.Builder
	flush := func() {
		if b.Len() == 0 {
			return
		}
		runs = append(runs, npcDialogTextRun{text: b.String(), color: current})
		b.Reset()
	}
	for i := 0; i < len(text); {
		if c, ok := parseNPCDialogColorCode(text, i, base); ok {
			flush()
			current = c
			i += 7
			continue
		}
		r, size := utf8.DecodeRuneInString(text[i:])
		b.WriteRune(r)
		i += size
	}
	flush()
	return runs
}

func parseNPCDialogColorCode(text string, at int, base color.RGBA) (color.RGBA, bool) {
	if at+7 > len(text) || text[at] != '^' {
		return color.RGBA{}, false
	}
	var value [6]byte
	for i := 0; i < 6; i++ {
		c := text[at+1+i]
		if !isNPCDialogHex(c) {
			return color.RGBA{}, false
		}
		value[i] = c
	}
	if strings.EqualFold(string(value[:]), "000000") {
		return base, true
	}
	return color.RGBA{
		R: npcDialogHexByte(value[0], value[1]),
		G: npcDialogHexByte(value[2], value[3]),
		B: npcDialogHexByte(value[4], value[5]),
		A: 255,
	}, true
}

func isNPCDialogHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'A' && c <= 'F') || (c >= 'a' && c <= 'f')
}

func npcDialogHexByte(hi, lo byte) uint8 {
	return npcDialogHexNibble(hi)<<4 | npcDialogHexNibble(lo)
}

func npcDialogHexNibble(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	default:
		return 0
	}
}

func wrapNPCDialogTextRuns(runs []npcDialogTextRun, maxRunes int) [][]npcDialogTextRun {
	chars := npcDialogRunsToRunes(runs)
	if len(chars) == 0 {
		return nil
	}
	var out [][]npcDialogTextRun
	for len(chars) > maxRunes {
		breakAt := maxRunes
		for i := maxRunes - 1; i > 0; i-- {
			if chars[i].r == ' ' || chars[i].r == '\t' {
				breakAt = i
				break
			}
		}
		out = append(out, npcDialogRunesToRuns(chars[:breakAt]))
		chars = chars[breakAt:]
		for len(chars) > 0 && (chars[0].r == ' ' || chars[0].r == '\t') {
			chars = chars[1:]
		}
	}
	if len(chars) > 0 {
		out = append(out, npcDialogRunesToRuns(chars))
	}
	return out
}

func trimNPCDialogTextRuns(runs []npcDialogTextRun, maxRunes int) []npcDialogTextRun {
	chars := npcDialogRunsToRunes(runs)
	if len(chars) <= maxRunes {
		return runs
	}
	if maxRunes <= 3 {
		return npcDialogRunesToRuns(chars[:maxRunes])
	}
	trimmed := npcDialogRunesToRuns(chars[:maxRunes-3])
	trimmed = append(trimmed, npcDialogTextRun{text: "...", color: runs[len(runs)-1].color})
	return trimmed
}

func npcDialogRunsToRunes(runs []npcDialogTextRun) []npcDialogColoredRune {
	var chars []npcDialogColoredRune
	for _, run := range runs {
		for _, r := range run.text {
			chars = append(chars, npcDialogColoredRune{r: r, color: run.color})
		}
	}
	return chars
}

func npcDialogRunesToRuns(chars []npcDialogColoredRune) []npcDialogTextRun {
	if len(chars) == 0 {
		return nil
	}
	runs := []npcDialogTextRun{{color: chars[0].color}}
	var b strings.Builder
	current := chars[0].color
	for _, char := range chars {
		if char.color != current {
			runs[len(runs)-1].text = b.String()
			b.Reset()
			current = char.color
			runs = append(runs, npcDialogTextRun{color: current})
		}
		b.WriteRune(char.r)
	}
	runs[len(runs)-1].text = b.String()
	return runs
}

func drawNPCDialogTextRuns(screen *render.Image, runs []npcDialogTextRun, x, y int) {
	offset := 0
	for _, run := range runs {
		if run.text == "" {
			continue
		}
		render.DebugPrintAtColor(screen, run.text, x+offset, y, run.color)
		offset += len([]rune(run.text)) * 7
	}
}

func clampNPCDialogInt(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func (d *NPCDialog) CursorAction(ctx Context) (int, bool) {
	if !d.open || ctx.Input == nil {
		return 0, false
	}
	width, height := ctx.ScreenSize()
	x, y, w, h := d.resolvedDialogBounds(width, height)
	mx, my := ctx.Input.MouseX, ctx.Input.MouseY
	if d.action == npcDialogActionMenu {
		menuX, menuY, menuW, menuH := d.menuBounds(width, height, x, y, w, h)
		if pointInRect(mx, my, menuX, menuY, menuW, menuH) {
			if pointInRect(mx, my, menuX, menuY, menuW, npcMenuTitleH) {
				return CursorActionClick, true
			}
			start, end := d.visibleMenuRange(menuH)
			for i := 0; start+i < end; i++ {
				ox, oy, ow, oh := npcDialogOptionBounds(menuX, menuY, menuW, i)
				if pointInRect(mx, my, ox, oy, ow, oh) {
					return CursorActionClick, true
				}
			}
			cancelX, cancelY, cancelW, cancelH := npcDialogMenuCancelBounds(menuX, menuY, menuW, menuH)
			if pointInRect(mx, my, cancelX, cancelY, cancelW, cancelH) {
				return CursorActionClick, true
			}
			return CursorActionDefault, true
		}
	}
	if pointInRect(mx, my, x, y, w, h) {
		if pointInRect(mx, my, x, y, w, npcDialogTitleH) {
			return CursorActionClick, true
		}
		if d.action == npcDialogActionNext || d.action == npcDialogActionClose {
			bx, by, bw, bh := npcDialogButtonBounds(x, y, w, h)
			if pointInRect(mx, my, bx, by, bw, bh) {
				return CursorActionClick, true
			}
		}
		return CursorActionDefault, true
	}
	return 0, false
}
