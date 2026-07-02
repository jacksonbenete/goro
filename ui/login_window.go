package ui

import (
	"strings"

	"github.com/kivutar/goro/render"
)

type LoginWindowDrawOptions struct {
	X, Y, W, H       int
	TitleH           int
	FooterH          int
	FormTopPad       int
	FieldGap         int
	FieldLeft        int
	FieldRightPad    int
	FieldH           int
	Username         string
	Password         string
	FocusUser        bool
	FocusPassword    bool
	LoginButtonHover bool
}

func DrawLoginWindow(screen *render.Image, opts LoginWindowDrawOptions) {
	if screen == nil {
		return
	}
	DrawTitledWindowFrame(screen, opts.X, opts.Y, opts.W, opts.H, opts.TitleH)
	DrawWindowTitle(screen, opts.X, opts.Y, opts.TitleH, 10, "Login", TitleTextColor)
	DrawWindowFooter(screen, opts.X, opts.Y, opts.W, opts.H, opts.FooterH)

	userX, userY, userW, userH := LoginWindowFieldRect(opts, 0)
	passX, passY, passW, passH := LoginWindowFieldRect(opts, 1)
	render.DebugPrintAtColor(screen, "Account", LoginWindowLabelX(userX, "Account"), LoginWindowLabelY(userY, userH), TextColor)
	render.DebugPrintAtColor(screen, "Password", LoginWindowLabelX(passX, "Password"), LoginWindowLabelY(passY, passH), TextColor)
	DrawTextInput(screen, userX, userY, userW, userH, opts.Username, opts.FocusUser)
	DrawTextInput(screen, passX, passY, passW, passH, strings.Repeat("*", len([]rune(opts.Password))), opts.FocusPassword)

	buttonX, buttonY, buttonW, buttonH := LoginWindowButtonRect(opts)
	buttonBG := ButtonColor
	if opts.LoginButtonHover {
		buttonBG = ButtonHoverColor
	}
	DrawButtonLabel(screen, buttonX, buttonY, buttonW, buttonH, "Login", buttonBG, TextColor)
}

func LoginWindowButtonRect(opts LoginWindowDrawOptions) (int, int, int, int) {
	fieldX, _, fieldW, _ := LoginWindowFieldRect(opts, 0)
	buttonW := ButtonLabelWidth("Login")
	_, footerY, _, footerH := LoginWindowFooterRect(opts)
	buttonH := 24
	return fieldX + fieldW - buttonW, footerY + (footerH-buttonH)/2, buttonW, buttonH
}

func LoginWindowFooterRect(opts LoginWindowDrawOptions) (int, int, int, int) {
	return opts.X, opts.Y + opts.H - opts.FooterH, opts.W, opts.FooterH
}

func LoginWindowFieldRect(opts LoginWindowDrawOptions, row int) (int, int, int, int) {
	fieldX := opts.X + opts.FieldLeft
	fieldY := opts.Y + opts.TitleH + opts.FormTopPad + row*(opts.FieldH+opts.FieldGap)
	fieldW := opts.W - opts.FieldLeft - opts.FieldRightPad
	return fieldX, fieldY, fieldW, opts.FieldH
}

func LoginWindowLabelX(fieldX int, label string) int {
	return fieldX - 12 - len([]rune(label))*7
}

func LoginWindowLabelY(fieldY, fieldH int) int {
	return fieldY + maxInt(0, (fieldH-14)/2)
}
