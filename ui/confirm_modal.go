package ui

import (
	"strings"

	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	smallPromptWidth    = 286
	smallPromptHeight   = 128
	smallPromptFooterH  = 42
	smallPromptSidePad  = 12
	smallPromptMessageH = 28
	smallPromptLineH    = 14
)

type ConfirmModal struct {
	WindowHandle
	open     bool
	title    string
	message  string
	onOK     func()
	onCancel func()
	ctx      client.Context
}

func (m *ConfirmModal) Open(ctx client.Context, title, message string, onOK, onCancel func()) {
	m.open = true
	m.title = title
	m.message = message
	m.onOK = onOK
	m.onCancel = onCancel
	m.ctx = ctx
	m.EnsureWindow(smallPromptWidth, smallPromptHeight)
	m.window.Open(ctx, m.widgetTree(ctx))
	m.Publish(ctx)
}

func (m *ConfirmModal) Update(ctx client.Context) bool {
	m.ctx = ctx
	if !m.open {
		return false
	}
	if ctx.Input != nil {
		if ctx.Input.JustPressed(render.KeyEscape) {
			m.Cancel(ctx)
			return true
		}
		if ctx.Input.JustPressed(render.KeyEnter) {
			m.Confirm(ctx)
			return true
		}
	}
	m.openWindow(ctx)
	if m.window.Update(ctx) {
		m.Publish(ctx)
		return true
	}
	m.Publish(ctx)
	return true
}

func (m *ConfirmModal) IsOpen() bool {
	return m.open
}

func (m *ConfirmModal) Confirm(ctx client.Context) {
	m.Close(ctx)
	if m.onOK != nil {
		m.onOK()
	}
}

func (m *ConfirmModal) Cancel(ctx client.Context) {
	m.Close(ctx)
	if m.onCancel != nil {
		m.onCancel()
	}
}

func (m *ConfirmModal) Close(ctx client.Context) {
	m.open = false
	m.window.Close()
	m.Publish(ctx)
}

func (m *ConfirmModal) openWindow(ctx client.Context) {
	m.EnsureWindow(smallPromptWidth, smallPromptHeight)
	if !m.window.IsOpen() {
		m.window.Open(ctx, m.widgetTree(ctx))
	}
}

func (m *ConfirmModal) widgetTree(ctx client.Context) widget.Widget {
	okW := float32(ButtonLabelWidth("OK"))
	cancelW := float32(ButtonLabelWidth("Cancel"))
	return Window(
		Title(m.title),
		CloseButton(false),
		Size(smallPromptWidth, smallPromptHeight),
		FooterHeight(smallPromptFooterH),
		FooterPadding(18),
		Content(smallPromptContent(m.message)),
		Footer(
			primitives.HBox(
				primitives.Expanded(primitives.Box()),
				rotheme.Button("OK", func() {
					m.Confirm(ctx)
				}).
					Width(okW),
				rotheme.Button("Cancel", func() {
					m.Cancel(ctx)
				}).
					Width(cancelW),
			).Gap(8),
		),
	)
}

func smallPromptContent(message string) widget.Widget {
	return primitives.Box(
		primitives.Expanded(primitives.Box()),
		smallPromptMessage(message),
		primitives.Expanded(primitives.Box()),
	).
		PaddingLeft(smallPromptSidePad).
		PaddingRight(smallPromptSidePad)
}

func smallPromptMessage(message string) widget.Widget {
	lines := strings.Split(message, "\n")
	first := ""
	second := ""
	if len(lines) > 0 {
		first = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		second = strings.TrimSpace(strings.Join(lines[1:], " "))
	}
	return primitives.Box(
		smallPromptLine(first),
		smallPromptLine(second),
	).
		Height(smallPromptMessageH)
}

func smallPromptLine(line string) widget.Widget {
	if line == "" {
		return primitives.Box().Height(smallPromptLineH)
	}
	return primitives.Box(
		rotheme.Text(line).
			MaxLines(1),
	).
		Height(smallPromptLineH)
}
