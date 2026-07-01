package ui

import (
	"image/color"

	uiwidget "github.com/gogpu/ui/widget"
	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
)

const (
	CursorActionDefault = 0
	CursorActionTalk    = 1
	CursorActionClick   = 2
	CursorActionRotate  = 4
	CursorActionAttack  = 5
	CursorActionWarp    = 7
	CursorActionPick    = 9
	CursorActionTarget  = 10
	CursorActionTarget2 = 11
	CursorActionNoWalk  = 13
)

const (
	actorObjectTypeMob       = 5
	actorObjectTypeNPC       = 6
	actorObjectTypeNPCABR    = 13
	actorObjectTypeNPCBionic = 14
)

type Context = client.Context

type WorldRenderer interface {
	DrawInventoryItemIcon(screen *render.Image, manager *res.Manager, item session.InventoryItem, x, y int)
	DrawSkillIcon(screen *render.Image, manager *res.Manager, skill session.Skill, x, y, size int)
	DrawItemInfoIllustration(screen *render.Image, manager *res.Manager, item session.InventoryItem, x, y, width, height int)
	DrawEquipmentPreview(screen *render.Image, ctx client.Context, x, y, width, height int)
	UseShortcutSkill(ctx client.Context, skill session.Skill) error
	AddTeleportEffect(ctx client.Context)
}

var (
	uiWindowRadius      = WindowRadius
	uiButtonRadius      = ButtonRadius
	uiWindowBodyColor   = WindowBodyColor
	uiWindowTitleTop    = WindowTitleTop
	uiWindowTitleColor  = WindowTitleColor
	uiWindowBorderColor = WindowBorderColor
	uiPanelBodyColor    = PanelBodyColor
	uiPanelAltColor     = PanelAltColor
	uiPanelHoverColor   = PanelHoverColor
	uiDisabledColor     = DisabledColor
	uiTextColor         = TextColor
	uiMutedTextColor    = MutedTextColor
	uiTitleTextColor    = TitleTextColor
	uiGoodTextColor     = GoodTextColor
	uiErrorTextColor    = ErrorTextColor
	uiButtonColor       = ButtonColor
	uiButtonHoverColor  = ButtonHoverColor
	uiButtonDownColor   = ButtonDownColor
	uiButtonBorderColor = ButtonBorderColor
)

func drawUIWindowFrame(screen *render.Image, x, y, w, h int) {
	DrawWindowFrame(screen, x, y, w, h)
}

func drawUITitledWindowFrame(screen *render.Image, x, y, w, h, titleH int) {
	DrawTitledWindowFrame(screen, x, y, w, h, titleH)
}

func drawUIWindowTitle(screen *render.Image, x, y, titleH, pad int, title string, text color.RGBA) {
	DrawWindowTitle(screen, x, y, titleH, pad, title, text)
}

func drawUITitleTextAt(screen *render.Image, x, y, titleH int, title string, text color.RGBA) {
	DrawTitleTextAt(screen, x, y, titleH, title, text)
}

func drawUIPanelSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawPanelSurface(screen, x, y, w, h, bg)
}

func drawUIButtonSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawButtonSurface(screen, x, y, w, h, bg)
}

func drawUIButtonLabel(screen *render.Image, x, y, w, h int, label string, bg, text color.RGBA) {
	DrawButtonLabel(screen, x, y, w, h, label, bg, text)
}

func drawUICloseButton(screen *render.Image, x, y, w, h int, bg, line color.RGBA) {
	DrawCloseButton(screen, x, y, w, h, bg, line)
}

func drawUICenteredText(screen *render.Image, x, y, w, h int, label string, text color.RGBA) {
	DrawCenteredText(screen, x, y, w, h, label, text)
}

func drawUIRowSurface(screen *render.Image, x, y, w, h int, bg color.RGBA) {
	DrawRowSurface(screen, x, y, w, h, bg)
}

func drawUISurface(screen *render.Image, x, y, w, h int, bg, border color.RGBA) {
	DrawSurface(screen, x, y, w, h, bg, border)
}

func uiColor(c color.RGBA) uiwidget.Color {
	return Color(c)
}

func pointInRect(px, py, x, y, w, h int) bool {
	return px >= x && py >= y && px < x+w && py < y+h
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func trimRunes(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes <= 0 {
		return ""
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func selectedCharacter(s *session.Session) session.Character {
	if s == nil {
		return session.Character{Name: "Player"}
	}
	if s.Selected.ID != 0 {
		return s.Selected
	}
	for _, character := range s.Characters {
		if character.ID == s.CharID {
			return character
		}
	}
	if len(s.Characters) > 0 {
		return s.Characters[0]
	}
	return session.Character{ID: s.CharID, Name: "Player", Job: 0}
}

func skillByID(s *session.Session, skillID uint16) (session.Skill, bool) {
	if s == nil || skillID == 0 {
		return session.Skill{}, false
	}
	for _, skill := range s.Skills.List {
		if skill.ID == skillID {
			return skill, true
		}
	}
	return session.Skill{}, false
}

func rectArray(x, y, w, h int) [4]int {
	return [4]int{x, y, w, h}
}
