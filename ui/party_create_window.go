package ui

import (
	"strings"

	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/core/textfield"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	partyCreateW        = 306
	partyCreateContentH = 158
	partyCreateFooterH  = 42
)

type PartyCreateWindowAction struct {
	Name         string
	ItemPickup   uint8
	ItemDivision uint8
}

type PartyCreateWindow struct {
	Window
	name         string
	itemPickup   uint8
	itemDivision uint8
	nameField    *textfield.Widget
	action       PartyCreateWindowAction
}

func (w *PartyCreateWindow) Open(ctx Context) {
	w.EnsureWindow(partyCreateW, ROWindowTitleHeight+partyCreateContentH+partyCreateFooterH)
	w.ctx = ctx
	w.name = ""
	w.itemPickup = 0
	w.itemDivision = 0
	w.nameField = nil
	w.Window.Open(ctx, w.widgetTree(ctx))
	w.focusName()
	w.Publish(ctx)
}

func (w *PartyCreateWindow) Update(ctx Context) bool {
	w.EnsureWindow(partyCreateW, ROWindowTitleHeight+partyCreateContentH+partyCreateFooterH)
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

func (w *PartyCreateWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.nameField = nil
	w.SetContent(w.widgetTree(ctx))
	w.focusName()
	w.Publish(ctx)
}

func (w *PartyCreateWindow) PopAction() PartyCreateWindowAction {
	action := w.action
	w.action = PartyCreateWindowAction{}
	return action
}

func (w *PartyCreateWindow) widgetTree(ctx Context) widget.Widget {
	return Win(
		Title("Create Party"),
		CloseButton(true),
		OnClose(w.Close),
		Size(partyCreateW, ROWindowTitleHeight+partyCreateContentH+partyCreateFooterH),
		Content(
			primitives.Box(
				rotheme.SectionLabel("Party Name"),
				primitives.Box(w.nameInput(ctx)).
					Height(24).
					CrossAlign(primitives.CrossAxisStretch),
				rotheme.SectionLabel("Item Pickup"),
				rotheme.Radio(
					radio.Items(
						radio.ItemDef{Value: "0", Label: "Each Take"},
						radio.ItemDef{Value: "1", Label: "Party Share"},
					),
					radio.Selected(partyRuleValue(w.itemPickup)),
					radio.OnChange(func(value string) {
						w.itemPickup = parsePartySettingUint8(value)
					}),
				),
				rotheme.SectionLabel("Item Sharing"),
				rotheme.Radio(
					radio.Items(
						radio.ItemDef{Value: "0", Label: "Individual"},
						radio.ItemDef{Value: "1", Label: "Shared"},
					),
					radio.Selected(partyRuleValue(w.itemDivision)),
					radio.OnChange(func(value string) {
						w.itemDivision = parsePartySettingUint8(value)
					}),
				),
			).
				Padding(14).
				Gap(6),
		),
		FooterHeight(partyCreateFooterH),
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

func (w *PartyCreateWindow) nameInput(ctx Context) *textfield.Widget {
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
		textfield.Placeholder("Party Name"),
	)
	return w.nameField
}

func (w *PartyCreateWindow) submit(ctx Context) {
	name := strings.TrimSpace(w.name)
	if name == "" {
		return
	}
	w.action = PartyCreateWindowAction{Name: name, ItemPickup: w.itemPickup, ItemDivision: w.itemDivision}
	w.Close()
	w.Publish(ctx)
}

func (w *PartyCreateWindow) submitFromFocusedEnter(ctx Context) bool {
	if ctx.Input == nil || w.nameField == nil || !w.nameField.IsFocused() {
		return false
	}
	if !ctx.Input.JustPressed(render.KeyEnter) {
		return false
	}
	w.submit(ctx)
	return true
}

func (w *PartyCreateWindow) focusName() {
	if w.nameField != nil {
		w.nameField.SetFocused(true)
	}
}

func partyRuleValue(value uint8) string {
	if value != 0 {
		return "1"
	}
	return "0"
}
