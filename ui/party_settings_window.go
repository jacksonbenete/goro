package ui

import (
	"log"
	"strconv"

	"github.com/gogpu/ui/core/radio"
	"github.com/gogpu/ui/primitives"
	"github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/ui/rotheme"
)

const (
	partySettingsW       = 286
	partySettingsContent = 100
	partySettingsFooterH = 42
)

type PartySettingsWindow struct {
	window   WindowState
	ctx      Context
	expShare uint32
}

func (w *PartySettingsWindow) Open(ctx Context) {
	w.ensureWindow()
	w.ctx = ctx
	party := sessionParty(ctx.Session)
	w.expShare = party.ExpShare
	w.window.Open(ctx, w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *PartySettingsWindow) Update(ctx Context) bool {
	w.ensureWindow()
	w.ctx = ctx
	if !w.window.IsOpen() {
		return false
	}
	consumed := w.window.Update(ctx)
	w.Publish(ctx)
	return consumed
}

func (w *PartySettingsWindow) IsOpen() bool {
	w.ensureWindow()
	return w.window.IsOpen()
}

func (w *PartySettingsWindow) Close() {
	w.ensureWindow()
	w.window.Close()
	w.Publish(w.ctx)
}

func (w *PartySettingsWindow) Publish(ctx Context) {
	w.ensureWindow()
	w.window.Publish(ctx)
}

func (w *PartySettingsWindow) Rebind(ctx Context) {
	if !w.IsOpen() {
		return
	}
	w.ctx = ctx
	w.window.SetContent(w.widgetTree(ctx))
	w.Publish(ctx)
}

func (w *PartySettingsWindow) ensureWindow() {
	if w.window.width != 0 {
		return
	}
	w.window = NewWindowState(partySettingsW, ROWindowTitleHeight+partySettingsContent+partySettingsFooterH)
}

func (w *PartySettingsWindow) widgetTree(ctx Context) widget.Widget {
	return Window(
		Title("Party Settings"),
		CloseButton(true),
		OnClose(w.Close),
		Size(partySettingsW, ROWindowTitleHeight+partySettingsContent+partySettingsFooterH),
		Content(
			primitives.Box(
				rotheme.SectionLabel("EXP"),
				rotheme.Radio(
					radio.Items(
						radio.ItemDef{Value: "0", Label: "Each Take"},
						radio.ItemDef{Value: "1", Label: "Even Share"},
					),
					radio.Selected(strconv.Itoa(int(w.expShare))),
					radio.OnChange(func(value string) {
						w.expShare = parsePartySettingUint32(value)
					}),
				),
			).
				Padding(14).
				Gap(8),
		),
		FooterHeight(partySettingsFooterH),
		Footer(
			primitives.HBox(
				primitives.Expanded(primitives.Box()),
				rotheme.Button("OK", func() {
					w.apply(ctx)
				}).Width(float32(ButtonLabelWidth("OK"))),
				rotheme.Button("Cancel", w.Close).Width(float32(ButtonLabelWidth("Cancel"))),
			).
				Gap(8),
		),
	)
}

func (w *PartySettingsWindow) apply(ctx Context) {
	if ctx.Session != nil {
		ctx.Session.Party.ExpShare = w.expShare
	}
	if ctx.Network != nil {
		if err := ctx.Network.SendPartyOption(w.expShare); err != nil {
			log.Printf("party settings failed: %v", err)
		}
	}
	w.Close()
}

func parsePartySettingUint32(value string) uint32 {
	n, err := strconv.Atoi(value)
	if err != nil || n < 0 {
		return 0
	}
	return uint32(n)
}
