package gamemode

import (
	"fmt"
	"strings"
	"time"

	"github.com/gogpu/ui/offscreen"
	"github.com/gogpu/ui/primitives"
	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/internal/render"
)

const (
	consoleMargin   = 16
	consoleWidth    = 480
	consoleHeight   = 88
	consoleMaxLines = 3
	consoleMaxInput = 120
	consoleFieldH   = 24
)

type chatConsole struct {
	active        bool
	input         string
	messages      []string
	lastMessage   string
	lastMessageAt time.Time

	cacheKey string
	image    *render.Image
}

func (c *chatConsole) update(ctx Context) bool {
	if ctx.Input == nil {
		return false
	}
	w, h := ctx.ScreenSize()
	x, y, cw, ch := consoleBounds(w, h)
	inside := ctx.Input.MouseX >= x && ctx.Input.MouseX < x+cw && ctx.Input.MouseY >= y && ctx.Input.MouseY < y+ch
	if ctx.Input.MouseJustPressed(render.MouseButtonLeft) && inside {
		c.active = true
		c.invalidate()
		return true
	}
	if ctx.Input.JustPressed(render.KeyEscape) && c.active {
		c.active = false
		c.invalidate()
		return true
	}
	if ctx.Input.JustPressed(render.KeyEnter) {
		if c.active {
			c.submit(ctx)
		} else {
			c.active = true
			c.invalidate()
		}
		return true
	}
	if !c.active {
		return false
	}
	if text := ctx.Input.TextInput(); text != "" {
		c.appendInput(text)
	}
	if ctx.Input.JustPressed(render.KeyBackspace) {
		c.backspace()
	}
	return true
}

func (c *chatConsole) addMessage(format string, args ...any) {
	text := strings.TrimSpace(fmt.Sprintf(format, args...))
	if text == "" {
		return
	}
	now := time.Now()
	if text == c.lastMessage && now.Sub(c.lastMessageAt) < time.Second {
		return
	}
	c.lastMessage = text
	c.lastMessageAt = now
	c.messages = append(c.messages, text)
	if len(c.messages) > 80 {
		copy(c.messages, c.messages[len(c.messages)-80:])
		c.messages = c.messages[:80]
	}
	c.invalidate()
}

func (c *chatConsole) submit(ctx Context) {
	text := strings.TrimSpace(c.input)
	if text == "" {
		c.active = false
		c.invalidate()
		return
	}
	name := "Player"
	if ctx.Session != nil {
		name = selectedCharacter(ctx.Session).Name
	}
	if ctx.Network == nil {
		c.addMessage("send failed: not connected")
		return
	}
	if err := ctx.Network.SendGlobalChat(name, text); err != nil {
		c.addMessage("send failed: %s", err)
		return
	}
	c.input = ""
	c.active = false
	c.invalidate()
}

func (c *chatConsole) appendInput(text string) {
	if text == "" {
		return
	}
	runes := []rune(c.input + text)
	if len(runes) > consoleMaxInput {
		runes = runes[:consoleMaxInput]
	}
	c.input = string(runes)
	c.invalidate()
}

func (c *chatConsole) backspace() {
	runes := []rune(c.input)
	if len(runes) == 0 {
		return
	}
	c.input = string(runes[:len(runes)-1])
	c.invalidate()
}

func (c *chatConsole) draw(screen *render.Image, width, height int) {
	if screen == nil {
		return
	}
	x, y, cw, ch := consoleBounds(width, height)
	key := c.renderKey(cw, ch)
	if c.image == nil || c.cacheKey != key {
		c.cacheKey = key
		c.image = c.renderImage(cw, ch)
	}
	if c.image == nil {
		return
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(float64(x), float64(y))
	opts.Filter = render.FilterNearest
	screen.DrawImage(c.image, &opts)
}

func (c *chatConsole) renderImage(width, height int) *render.Image {
	root := c.widgetTree(width, height)
	r := offscreen.NewRenderer(width, height, offscreen.WithBackground(uiwidget.ColorTransparent))
	r.Render(root)
	img := r.Image()
	if img == nil {
		return nil
	}
	out := render.NewImageFromImage(img)
	c.drawCrispText(out, width, height)
	return out
}

func (c *chatConsole) widgetTree(width, height int) uiwidget.Widget {
	contentWidth := maxInt(1, width-16)
	field := primitives.Box().
		Width(float32(contentWidth)).
		Height(consoleFieldH).
		PaddingXY(6, 3).
		Background(uiwidget.RGBA8(5, 8, 13, 205)).
		BorderStyle(1, uiwidget.RGBA8(190, 208, 230, 120))
	messageHeight := maxInt(20, height-16-consoleFieldH-4)
	messages := primitives.Box().
		Width(float32(contentWidth)).
		Height(float32(messageHeight)).
		Gap(1)
	return primitives.Box(messages, field).
		Width(float32(width)).
		Height(float32(height)).
		PaddingXY(8, 6).
		Gap(4).
		Background(uiwidget.RGBA8(14, 18, 24, 188)).
		BorderStyle(1, uiwidget.RGBA8(180, 198, 218, 95))
}

func (c *chatConsole) drawCrispText(img *render.Image, width, height int) {
	if img == nil {
		return
	}
	contentWidth := maxInt(1, width-16)
	maxRunes := maxInt(8, (contentWidth-12)/7)
	for i, line := range c.visibleLines(width) {
		render.DebugPrintAt(img, trimRunes(line, maxRunes), 14, 10+i*14)
	}
	prompt := c.input
	if c.active && time.Now().UnixMilli()/500%2 == 0 {
		prompt += "|"
	}
	if prompt == "" && !c.active {
		prompt = "Press Enter to chat"
	}
	fieldY := height - 6 - consoleFieldH
	render.DebugPrintAt(img, trimRunes(prompt, maxRunes), 15, fieldY+5)
}

func (c *chatConsole) visibleLines(width int) []string {
	if len(c.messages) == 0 {
		return []string{"Server messages will appear here."}
	}
	start := len(c.messages) - consoleMaxLines
	if start < 0 {
		start = 0
	}
	out := make([]string, 0, len(c.messages)-start)
	maxRunes := maxInt(20, (width-24)/7)
	for _, msg := range c.messages[start:] {
		out = append(out, trimRunes(msg, maxRunes))
	}
	return out
}

func (c *chatConsole) renderKey(width, height int) string {
	blink := int64(0)
	if c.active {
		blink = time.Now().UnixMilli() / 500
	}
	return fmt.Sprintf("%dx%d:%t:%s:%s:%d", width, height, c.active, c.input, strings.Join(c.messages, "\n"), blink)
}

func (c *chatConsole) invalidate() {
	c.cacheKey = ""
}

func consoleBounds(screenW, screenH int) (x, y, w, h int) {
	w = minInt(consoleWidth, maxInt(260, screenW-2*consoleMargin))
	h = minInt(consoleHeight, maxInt(128, screenH-2*consoleMargin))
	x = consoleMargin
	y = maxInt(consoleMargin, screenH-h-consoleMargin)
	return x, y, w, h
}

func trimRunes(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}
