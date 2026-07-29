package ui

import (
	"strconv"
	"strings"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/input"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	shopAmountPromptW = 218
	shopAmountPromptH = 112
)

type shopAmountPrompt struct {
	Window
	value    string
	max      uint16
	onSubmit func(uint16)
	field    *textfield.Widget
}

func (p *shopAmountPrompt) Open(ctx Context, initial, maxAmount uint16, onSubmit func(uint16)) {
	p.max = maxAmount
	p.value = strconv.FormatUint(uint64(clampShopAmount(initial, maxAmount)), 10)
	p.onSubmit = onSubmit
	p.field = nil
	p.EnsureWindow(shopAmountPromptW, shopAmountPromptH)
	p.Window.Open(ctx, p.widgetTree(ctx))
	p.focusInput()
	p.Publish(ctx)
}

func (p *shopAmountPrompt) Update(ctx Context) bool {
	if !p.IsOpen() {
		return false
	}
	if ctx.Input != nil {
		if ctx.Input.JustPressed(input.KeyEscape) {
			p.Close(ctx)
			return true
		}
		if p.field != nil && p.field.IsFocused() && ctx.Input.JustPressed(input.KeyEnter) {
			p.submit(ctx)
			return true
		}
	}
	if p.Window.Update(ctx) {
		p.Publish(ctx)
		return true
	}
	p.Publish(ctx)
	return true
}

func (p *shopAmountPrompt) Close(ctx Context) {
	p.Window.Close()
	p.Publish(ctx)
}

func (p *shopAmountPrompt) BringToFront(ctx Context) {
	if !p.IsOpen() || p.published == nil || ctx.UIManager == nil {
		return
	}
	ctx.UIManager.RemoveOverlay(p.published)
	ctx.UIManager.AddOverlay(p.published)
}

func (p *shopAmountPrompt) widgetTree(ctx Context) widget.Widget {
	return Win(
		TitleBar(false),
		Size(shopAmountPromptW, shopAmountPromptH),
		Content(
			primitives.Box(
				rotheme.Text("Input number"),
				primitives.Box(p.input(ctx)).
					Height(24).
					CrossAlign(primitives.CrossAxisStretch),
			).
				Padding(10).
				Gap(7),
		),
		Footer(
			primitives.Expanded(primitives.Box()),
			rotheme.Button("OK", func() {
				p.submit(ctx)
			}),
			rotheme.Button("Cancel", func() {
				p.Close(ctx)
			}),
		),
	)
}

func (p *shopAmountPrompt) input(ctx Context) *textfield.Widget {
	if p.field != nil {
		return p.field
	}
	p.field = rotheme.TextField(
		p.value,
		textfield.TypeNumber,
		func(value string) {
			p.value = value
		},
		func(string) {
			p.submit(ctx)
		},
		textfield.MaxLength(5),
	)
	return p.field
}

func (p *shopAmountPrompt) submit(ctx Context) {
	amount, ok := parseShopAmount(p.value, p.max)
	if !ok {
		return
	}
	onSubmit := p.onSubmit
	p.Close(ctx)
	if amount > 0 && onSubmit != nil {
		onSubmit(amount)
	}
}

func (p *shopAmountPrompt) focusInput() {
	if p.field != nil {
		p.field.SetFocused(true)
	}
}

func parseShopAmount(value string, maxAmount uint16) (uint16, bool) {
	text := strings.TrimSpace(value)
	if text == "" {
		return 0, false
	}
	parsed, err := strconv.ParseUint(text, 10, 32)
	if err != nil {
		return 0, false
	}
	if parsed == 0 {
		return 0, true
	}
	if parsed > uint64(^uint16(0)) {
		parsed = uint64(^uint16(0))
	}
	return clampShopAmount(uint16(parsed), maxAmount), true
}
