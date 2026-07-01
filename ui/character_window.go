package ui

import (
	"fmt"
	"image/color"
	"math"
	"strconv"
	"strings"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/session"
)

const (
	characterWindowX      = 16
	characterWindowY      = 16
	characterWindowWidth  = 324
	characterWindowHeight = 158
)

var (
	characterWindowTextColor   = TextColor
	characterWindowMutedColor  = MutedTextColor
	characterWindowTitleColor  = TitleTextColor
	characterWindowBarBack     = color.RGBA{R: 224, G: 232, B: 242, A: 255}
	characterWindowHPColor     = color.RGBA{R: 210, G: 72, B: 72, A: 255}
	characterWindowSPColor     = color.RGBA{R: 70, G: 112, B: 214, A: 255}
	characterWindowEXPColor    = color.RGBA{R: 74, G: 174, B: 98, A: 255}
	characterWindowJobEXPColor = color.RGBA{R: 190, G: 148, B: 58, A: 255}
	characterWindowWeightWarn  = ErrorTextColor
)

func DrawCharacterWindow(screen *render.Image, ctx Context) {
	if ctx.Session == nil {
		return
	}
	x, y, w, h := characterWindowX, characterWindowY, characterWindowWidth, characterWindowHeight
	DrawTitledWindowFrame(screen, x, y, w, h, 29)

	character := selectedCharacter(ctx.Session)
	name := strings.TrimSpace(character.Name)
	if name == "" {
		name = "Player"
	}
	DrawTitleTextAt(screen, x+12, y, 29, trimRunes(name, 20), characterWindowTitleColor)
	DrawTitleTextAt(screen, x+166, y, 29, trimRunes(CharacterJobName(character), 20), characterWindowMutedColor)

	vitals := ctx.Session.Vitals
	if vitals.HP == 0 && vitals.MaxHP == 0 && vitals.SP == 0 && vitals.MaxSP == 0 {
		vitals = sessionVitalsFromCharacter(character)
	}
	progress := ctx.Session.Progress
	if progress.BaseLevel == 0 {
		progress = sessionProgressFromCharacter(character)
		progress.BaseExp = ctx.Session.Progress.BaseExp
		progress.NextBaseExp = ctx.Session.Progress.NextBaseExp
		progress.JobExp = ctx.Session.Progress.JobExp
		progress.NextJobExp = ctx.Session.Progress.NextJobExp
	}
	if progress.JobLevel == 0 && character.JobLevel > 0 {
		progress.JobLevel = int(character.JobLevel)
	}

	render.DebugPrintAtColor(screen, fmt.Sprintf("Base Lv. %d", progress.BaseLevel), x+12, y+38, characterWindowTextColor)
	render.DebugPrintAtColor(screen, fmt.Sprintf("Job Lv. %d", progress.JobLevel), x+166, y+38, characterWindowTextColor)
	drawCharacterWindowBar(screen, x+12, y+58, 146, "HP", vitals.HP, vitals.MaxHP, characterWindowHPColor)
	drawCharacterWindowBar(screen, x+166, y+58, 146, "SP", vitals.SP, vitals.MaxSP, characterWindowSPColor)
	drawCharacterProgressBar(screen, x+12, y+88, w-24, "Base EXP", progress.BaseExp, progress.NextBaseExp, characterWindowEXPColor)
	drawCharacterProgressBar(screen, x+12, y+110, w-24, "Job EXP", progress.JobExp, progress.NextJobExp, characterWindowJobEXPColor)

	inventory := ctx.Session.Inventory
	render.DebugPrintAtColor(screen, fmt.Sprintf("Zeny : %s", formatHUDNumber(inventory.Zeny)), x+12, y+136, characterWindowTextColor)
	weightColor := characterWindowTextColor
	if inventory.MaxWeight > 0 && inventory.Weight*100 >= inventory.MaxWeight*50 {
		weightColor = characterWindowWeightWarn
	}
	render.DebugPrintAtColor(screen, fmt.Sprintf("Weight : %d / %d", displayWeight(inventory.Weight), displayWeight(inventory.MaxWeight)), x+166, y+136, weightColor)
}

func displayWeight(raw int) int {
	return raw / 10
}

func DisplayWeight(raw int) int {
	return displayWeight(raw)
}

func drawCharacterWindowBar(screen *render.Image, x, y, w int, label string, current, maxValue int, fill color.RGBA) {
	render.DebugPrintAtColor(screen, fmt.Sprintf("%s %d / %d", label, current, maxValue), x, y, characterWindowMutedColor)
	drawRatioBar(screen, x, y+14, w, 7, ratioInt(current, maxValue), fill)
}

func drawCharacterProgressBar(screen *render.Image, x, y, w int, label string, current, next int64, fill color.RGBA) {
	render.DebugPrintAtColor(screen, fmt.Sprintf("%s %s", label, formatEXPPercent(current, next)), x, y, characterWindowMutedColor)
	drawRatioBar(screen, x, y+13, w, 6, ratioInt64(current, next), fill)
}

func drawRatioBar(screen *render.Image, x, y, w, h int, ratio float64, fill color.RGBA) {
	render.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), characterWindowBarBack)
	if ratio > 0 {
		fillW := int(math.Round(float64(w) * ratio))
		if fillW < 1 {
			fillW = 1
		}
		if fillW > w {
			fillW = w
		}
		render.DrawRect(screen, float64(x), float64(y), float64(fillW), float64(h), fill)
	}
	render.DrawRect(screen, float64(x), float64(y), float64(w), 1, WindowBorderColor)
	render.DrawRect(screen, float64(x), float64(y+h-1), float64(w), 1, WindowBorderColor)
}

func ratioInt(current, maxValue int) float64 {
	if maxValue <= 0 {
		return 0
	}
	return clampUnit(float64(current) / float64(maxValue))
}

func ratioInt64(current, maxValue int64) float64 {
	if maxValue <= 0 {
		return 0
	}
	return clampUnit(float64(current) / float64(maxValue))
}

func sessionVitalsFromCharacter(character session.Character) session.Vitals {
	return session.Vitals{
		HP:    int(character.HP),
		MaxHP: int(character.MaxHP),
		SP:    int(character.SP),
		MaxSP: int(character.MaxSP),
	}
}

func sessionProgressFromCharacter(character session.Character) session.Progress {
	return session.Progress{
		BaseLevel: int(character.Level),
		JobLevel:  int(character.JobLevel),
	}
}

func formatEXPPercent(current, next int64) string {
	if next <= 0 {
		return "--"
	}
	percent := 100 * float64(current) / float64(next)
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	return fmt.Sprintf("%.1f%%", math.Floor(percent*10)/10)
}

func FormatEXPPercent(current, next int64) string {
	return formatEXPPercent(current, next)
}

func formatHUDNumber(value int64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}
	text := strconv.FormatInt(value, 10)
	if len(text) <= 3 {
		return sign + text
	}
	var b strings.Builder
	prefix := len(text) % 3
	if prefix == 0 {
		prefix = 3
	}
	b.WriteString(text[:prefix])
	for i := prefix; i < len(text); i += 3 {
		b.WriteByte(',')
		b.WriteString(text[i : i+3])
	}
	return sign + b.String()
}

func FormatHUDNumber(value int64) string {
	return formatHUDNumber(value)
}
