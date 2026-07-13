package ui

import (
	"strings"

	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	partyInviteW        = 286
	partyInviteContentH = 70
	partyInviteFooterH  = 42
)

type PartyInviteWindow struct {
	Window
	name      string
	nameField *textfield.Widget
	action    string
}

func (w *PartyInviteWindow) Open(ctx Context) {
	w.EnsureWindow(partyInviteW, ROWindowTitleHeight+partyInviteContentH+partyInviteFooterH)
	w.ctx = ctx
	w.name = ""
	w.nameField = nil
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.focusName()
	w.Publish(ctx)
}

func (w *PartyInviteWindow) Update(ctx Context) bool {
	w.EnsureWindow(partyInviteW, ROWindowTitleHeight+partyInviteContentH+partyInviteFooterH)
	w.ctx = ctx
	if !w.IsOpen() {
		return false
	}
	if w.submitFromFocusedEnter(ctx) {
		w.Publish(ctx)
		return true
	}
	consumed := w.Window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *PartyInviteWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.nameField = nil
	w.SetContent(w.widgetTree(ctx))
	w.focusName()
	w.Publish(ctx)
}

func (w *PartyInviteWindow) PopAction() string {
	action := w.action
	w.action = ""
	return action
}

func (w *PartyInviteWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Party Invitation"),
		CloseButton(true),
		OnClose(w.Close),
		Size(partyInviteW, ROWindowTitleHeight+partyInviteContentH+partyInviteFooterH),
		Content(
			primitives.Box(
				rotheme.SectionLabel("Player Name"),
				primitives.Box(w.nameInput(ctx)).
					Height(24).
					CrossAlign(primitives.CrossAxisStretch),
			).
				Padding(14).
				Gap(8),
		),
		FooterHeight(partyInviteFooterH),
		Footer(
			primitives.HBox(
				primitives.Expanded(primitives.Box()),
				rotheme.Button("OK", func() {
					w.submit(ctx)
				}).Width(float32(ButtonLabelWidth("OK"))),
				rotheme.Button("Cancel", w.Close).Width(float32(ButtonLabelWidth("Cancel"))),
			).Gap(8),
		),
	)
}

func (w *PartyInviteWindow) nameInput(ctx Context) *textfield.Widget {
	if w.nameField != nil {
		return w.nameField
	}
	w.nameField = rotheme.TextField(
		w.name,
		textfield.TypeText,
		func(value string) {
			w.name = value
		},
		func(string) {
			w.submit(ctx)
		},
		textfield.MaxLength(23),
		textfield.Placeholder("Player Name"),
	)
	return w.nameField
}

func (w *PartyInviteWindow) submit(ctx Context) {
	name := strings.TrimSpace(w.name)
	if name == "" {
		return
	}
	w.action = name
	w.Close()
	w.Publish(ctx)
}

func (w *PartyInviteWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil || w.nameField == nil || !w.nameField.IsFocused() {
		return false
	}
	if !ctx.Input.JustPressed(render.KeyEnter) {
		return false
	}
	w.submit(ctx)
	return true
}

func (w *PartyInviteWindow) focusName() {
	if w.nameField != nil {
		w.nameField.SetFocused(true)
	}
}
