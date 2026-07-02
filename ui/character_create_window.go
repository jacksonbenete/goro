package ui

import (
	"fmt"
	"image/color"

	"github.com/kivutar/goro/render"
)

const CharacterCreateStatCount = 6

const (
	CharacterCreateStatStr = iota
	CharacterCreateStatAgi
	CharacterCreateStatVit
	CharacterCreateStatInt
	CharacterCreateStatDex
	CharacterCreateStatLuk
)

type CharacterCreatePreviewFunc func(screen *render.Image, panelX, panelY, panelW, panelH int)

type CharacterCreateWindowOptions struct {
	X, Y, W, H  int
	TitleH      int
	FooterH     int
	FooterPadX  int
	FooterGap   int
	ButtonH     int
	PanelH      int
	Name        string
	FocusName   bool
	Stats       [CharacterCreateStatCount]uint8
	MouseX      int
	MouseY      int
	HasMouse    bool
	DrawPreview CharacterCreatePreviewFunc
}

func DrawCharacterCreateWindow(screen *render.Image, opts CharacterCreateWindowOptions) {
	if screen == nil {
		return
	}
	DrawTitledWindowFrame(screen, opts.X, opts.Y, opts.W, opts.H, opts.TitleH)
	DrawWindowTitle(screen, opts.X, opts.Y, opts.TitleH, 12, "Make Character", TitleTextColor)
	DrawWindowFooter(screen, opts.X, opts.Y, opts.W, opts.H, opts.FooterH)

	drawCharacterCreatePreview(screen, opts)
	drawCharacterCreateStats(screen, opts)

	nameX, nameY, nameW, nameH := CharacterCreateNameRect(opts)
	render.DebugPrintAtColor(screen, "Name", nameX, nameY-15, TextColor)
	DrawTextInput(screen, nameX, nameY, nameW, nameH, opts.Name, opts.FocusName)

	drawCharacterCreateButtonRect(screen, opts, rectArray(CharacterCreateMakeButtonRect(opts)), "Make")
	drawCharacterCreateButtonRect(screen, opts, rectArray(CharacterCreateCancelButtonRect(opts)), "Cancel")
}

func drawCharacterCreatePreview(screen *render.Image, opts CharacterCreateWindowOptions) {
	panelX, panelY, panelW, panelH := CharacterCreatePreviewPanelRect(opts)
	DrawPanelSurface(screen, panelX, panelY, panelW, panelH, PanelBodyColor)
	if opts.DrawPreview != nil {
		opts.DrawPreview(screen, panelX, panelY, panelW, panelH)
	}

	drawCharacterCreateButtonRect(screen, opts, rectArray(CharacterCreateHairPrevRect(opts)), "<")
	drawCharacterCreateButtonRect(screen, opts, rectArray(CharacterCreateHairNextRect(opts)), ">")
	drawCharacterCreateButtonRect(screen, opts, rectArray(CharacterCreateHairColorRect(opts)), "^")
}

func drawCharacterCreateStats(screen *render.Image, opts CharacterCreateWindowOptions) {
	graphX, graphY, graphW, graphH := CharacterCreateGraphPanelRect(opts)
	DrawCharacterCreateStatGraph(screen, graphX+graphW/2, graphY+graphH/2, opts.Stats)

	for i := 0; i < CharacterCreateStatCount; i++ {
		sx, sy, sw, sh := CharacterCreateStatButtonRect(opts, i)
		bg := ButtonColor
		if opts.HasMouse && pointInRect(opts.MouseX, opts.MouseY, sx, sy, sw, sh) {
			bg = ButtonHoverColor
		}
		label := CharacterCreateStatLabels()[i]
		DrawButtonSurface(screen, sx, sy, sw, sh, bg)
		DrawCenteredText(screen, sx, sy+2, sw, 15, label, TextColor)
		DrawCenteredText(screen, sx, sy+18, sw, 15, fmt.Sprintf("%d", opts.Stats[i]), TextColor)
	}

	listX, listY, listW, listH := CharacterCreateStatListPanelRect(opts)
	DrawPanelSurface(screen, listX, listY, listW, listH, PanelBodyColor)
	for i, label := range CharacterCreateStatLabels() {
		render.DebugPrintAtColor(screen, label, listX+18, listY+16+i*22, TextColor)
		render.DebugPrintAtColor(screen, fmt.Sprintf("%d", opts.Stats[i]), listX+92, listY+16+i*22, TextColor)
	}
}

func DrawCharacterCreateStatGraph(screen *render.Image, cx, cy int, stats [CharacterCreateStatCount]uint8) {
	outer := 64.0
	inner := 32.0
	points := CharacterCreateGraphPoints(cx, cy, outer)
	mid := CharacterCreateGraphPoints(cx, cy, inner)
	order := CharacterCreateGraphDrawOrder()
	for i := 0; i < CharacterCreateStatCount; i++ {
		current := order[i]
		next := order[(i+1)%CharacterCreateStatCount]
		render.DrawLine(screen, points[current][0], points[current][1], points[next][0], points[next][1], SeparatorColor)
		render.DrawLine(screen, mid[current][0], mid[current][1], mid[next][0], mid[next][1], SeparatorColor)
		render.DrawLine(screen, float64(cx), float64(cy), points[current][0], points[current][1], color.RGBA{R: 185, G: 204, B: 224, A: 150})
	}
	statPoints := [CharacterCreateStatCount][2]float64{}
	for i := 0; i < CharacterCreateStatCount; i++ {
		scale := 0.22 + float64(stats[i])/9.0*0.78
		statPoints[i][0] = float64(cx) + (points[i][0]-float64(cx))*scale
		statPoints[i][1] = float64(cy) + (points[i][1]-float64(cy))*scale
	}
	drawFilledCharacterCreateStatPolygon(screen, statPoints, order, color.RGBA{R: 36, G: 92, B: 154, A: 220})
}

func drawFilledCharacterCreateStatPolygon(screen *render.Image, points [CharacterCreateStatCount][2]float64, order [CharacterCreateStatCount]int, fill color.RGBA) {
	if screen == nil {
		return
	}
	r := float32(fill.R) / 255
	g := float32(fill.G) / 255
	b := float32(fill.B) / 255
	a := float32(fill.A) / 255
	vertices := make([]render.Vertex, 0, CharacterCreateStatCount)
	for _, stat := range order {
		vertices = append(vertices, render.Vertex{
			DstX:   float32(points[stat][0]),
			DstY:   float32(points[stat][1]),
			ColorR: r,
			ColorG: g,
			ColorB: b,
			ColorA: a,
		})
	}
	indices := make([]uint16, 0, (CharacterCreateStatCount-2)*3)
	for i := 1; i < CharacterCreateStatCount-1; i++ {
		indices = append(indices, 0, uint16(i), uint16(i+1))
	}
	screen.DrawTrianglesOwned(vertices, indices, render.WhiteImage(), &render.DrawTrianglesOptions{Filter: render.FilterNearest, Address: render.AddressUnsafe})
}

func CharacterCreateGraphDrawOrder() [CharacterCreateStatCount]int {
	return [CharacterCreateStatCount]int{
		CharacterCreateStatStr,
		CharacterCreateStatVit,
		CharacterCreateStatLuk,
		CharacterCreateStatInt,
		CharacterCreateStatDex,
		CharacterCreateStatAgi,
	}
}

func CharacterCreateGraphPoints(cx, cy int, radius float64) [CharacterCreateStatCount][2]float64 {
	dirs := [CharacterCreateStatCount][2]float64{
		{0, -1},
		{-0.866, -0.5},
		{0.866, -0.5},
		{0, 1},
		{-0.866, 0.5},
		{0.866, 0.5},
	}
	points := [CharacterCreateStatCount][2]float64{}
	for i := range dirs {
		points[i][0] = float64(cx) + dirs[i][0]*radius
		points[i][1] = float64(cy) + dirs[i][1]*radius
	}
	return points
}

func CharacterCreateStatLabels() [CharacterCreateStatCount]string {
	return [CharacterCreateStatCount]string{"STR", "AGI", "VIT", "INT", "DEX", "LUK"}
}

func CharacterCreateNameRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	previewX, _, previewW, _ := CharacterCreatePreviewPanelRect(opts)
	return previewX, opts.Y + 252, previewW, 22
}

func CharacterCreateHairPrevRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 44, opts.Y + 126, 24, 22
}

func CharacterCreateHairNextRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 138, opts.Y + 126, 24, 22
}

func CharacterCreateHairColorRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 91, opts.Y + 66, 24, 22
}

func CharacterCreatePreviewPanelRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 32, opts.Y + 58, 142, opts.PanelH
}

func CharacterCreatePreviewSpriteOrigin(panelX, panelY, panelW, panelH, imageW, imageH int, scale float64) (float64, float64) {
	return float64(panelX) + float64(panelW)/2 - float64(imageW)*scale/2,
		float64(panelY) + float64(panelH)/2 - float64(imageH)*scale/2
}

func CharacterCreateGraphPanelRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 204, opts.Y + 58, 166, opts.PanelH
}

func CharacterCreateStatListPanelRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X + 402, opts.Y + 58, 136, opts.PanelH
}

func CharacterCreateStatButtonRect(opts CharacterCreateWindowOptions, stat int) (int, int, int, int) {
	rects := [CharacterCreateStatCount][4]int{
		{opts.X + 269, opts.Y + 36, 38, 36},
		{opts.X + 181, opts.Y + 100, 38, 36},
		{opts.X + 356, opts.Y + 100, 38, 36},
		{opts.X + 269, opts.Y + 210, 38, 36},
		{opts.X + 181, opts.Y + 156, 38, 36},
		{opts.X + 356, opts.Y + 156, 38, 36},
	}
	if stat < 0 || stat >= len(rects) {
		stat = 0
	}
	rect := rects[stat]
	return rect[0], rect[1], rect[2], rect[3]
}

func CharacterCreateFooterRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return opts.X, opts.Y + opts.H - opts.FooterH, opts.W, opts.FooterH
}

func CharacterCreateMakeButtonRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return characterCreateFooterButtonRect(opts, 0)
}

func CharacterCreateCancelButtonRect(opts CharacterCreateWindowOptions) (int, int, int, int) {
	return characterCreateFooterButtonRect(opts, 1)
}

func characterCreateFooterButtonRect(opts CharacterCreateWindowOptions, index int) (int, int, int, int) {
	labels := [...]string{"Make", "Cancel"}
	if index < 0 || index >= len(labels) {
		index = 0
	}
	_, footerY, _, footerH := CharacterCreateFooterRect(opts)
	totalW := 0
	for _, label := range labels {
		totalW += ButtonLabelWidth(label)
	}
	totalW += opts.FooterGap * (len(labels) - 1)
	bx := opts.X + opts.W - opts.FooterPadX - totalW
	for i := 0; i < index; i++ {
		bx += ButtonLabelWidth(labels[i]) + opts.FooterGap
	}
	return bx, footerY + (footerH-opts.ButtonH)/2, ButtonLabelWidth(labels[index]), opts.ButtonH
}

func drawCharacterCreateButtonRect(screen *render.Image, opts CharacterCreateWindowOptions, rect [4]int, label string) {
	x, y, w, h := rect[0], rect[1], rect[2], rect[3]
	bg := ButtonColor
	if opts.HasMouse && pointInRect(opts.MouseX, opts.MouseY, x, y, w, h) {
		bg = ButtonHoverColor
	}
	DrawButtonLabel(screen, x, y, w, h, label, bg, TextColor)
}
