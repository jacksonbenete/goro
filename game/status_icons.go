package game

import (
	"image/color"
	"log"
	"math"
	"sort"
	"time"

	"github.com/kivutar/goro/network"
	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
	"github.com/kivutar/goro/session"
	gameui "github.com/kivutar/goro/ui"
)

const (
	statusIconSize    = 32
	statusIconSpacing = 36
	statusIconGap     = 8
)

type statusIconInfo struct {
	icon  string
	label string
}

var statusIconInfos = map[uint16]statusIconInfo{
	0:  {icon: "\xc7\xc1\xb7\xce\xba\xb8\xc5\xa9.tga", label: "Provoke"},
	1:  {icon: "\xc0\xce\xb5\xe0\xbe\xee.tga", label: "Endure"},
	2:  {icon: "\xc5\xf5\xc7\xda\xb5\xe5\xc4\xfb\xc5\xab.tga", label: "Two Hand Quicken"},
	3:  {icon: "\xc1\xfd\xc1\xdf\xb7\xc2\xc7\xe2\xbb\xf3.tga", label: "Attention Concentration"},
	9:  {icon: "\xbe\xc8\xc1\xa9\xb7\xe7\xbd\xba.tga", label: "Angelus"},
	10: {icon: "\xba\xed\xb7\xb9\xbd\xcc.tga", label: "Blessing"},
	11: {icon: "\xbd\xc3\xb1\xd7\xb3\xd1\xc5\xa9\xb7\xe7\xbd\xc3\xbd\xba.tga", label: "Signum Crucis"},
	12: {icon: "\xb9\xce\xc3\xb8\xbc\xba\xc1\xf5\xb0\xa1.tga", label: "Increase Agility"},
	13: {icon: "\xb9\xce\xc3\xb8\xbc\xba\xb0\xa8\xbc\xd2.tga", label: "Decrease Agility"},
	15: {icon: "\xc0\xd3\xc6\xf7\xbd\xc3\xc6\xbc\xbf\xc0\xb8\xb6\xb4\xa9\xbd\xba.tga", label: "Impositio Manus"},
	16: {icon: "\xbc\xf6\xc1\xdd\xc0\xba\xc7\xcf\xb7\xe7\xc0\xc7\xbf\xec\xbf\xef.tga", label: "Suffragium"},
	19: {icon: "\xb1\xe2\xb8\xae\xbf\xa1\xbf\xa4\xb7\xb9\xc0\xcc\xbc\xd5.tga", label: "Kyrie Eleison"},
	20: {icon: "\xb8\xb6\xb4\xcf\xc7\xc7\xc4\xb1.tga", label: "Magnificat"},
	21: {icon: "\xb1\xdb\xb7\xce\xb8\xae\xbe\xc6.tga", label: "Gloria"},
	23: {icon: "\xbe\xc6\xb5\xe5\xb7\xb9\xb3\xaf\xb8\xb0\xb7\xaf\xbd\xac.tga", label: "Adrenaline Rush"},
	25: {icon: "\xbf\xc0\xb9\xf6\xc6\xae\xb7\xaf\xbd\xba\xc6\xae.tga", label: "Over Thrust"},
	26: {icon: "\xb8\xc6\xbd\xc3\xb8\xb6\xc0\xcc\xc1\xee\xc6\xc4\xbf\xf6.tga", label: "Maximize Power"},
	37: {icon: "\xb0\xf8\xbc\xd3\xb9\xb0\xbe\xe0.tga", label: "Attack Speed Potion"},
	38: {icon: "\xb0\xf8\xbc\xd3\xb9\xb0\xbe\xe0.tga", label: "Attack Speed Potion"},
	39: {icon: "\xb0\xf8\xbc\xd3\xb9\xb0\xbe\xe0.tga", label: "Attack Speed Potion"},
	41: {icon: "\xb9\xce\xc3\xb8\xbc\xba\xc1\xf5\xb0\xa1.tga", label: "Movement Speed Potion"},
}

func (m *WorldMode) applyStatusEffectChange(ctx Context, change network.StatusEffectChange) {
	if ctx.Session == nil || change.StatusID == 0xFFFF {
		return
	}
	localID := localSkillTarget(ctx)
	if change.ActorID != 0 && localID != 0 && change.ActorID != localID && change.ActorID != ctx.Session.CharID {
		return
	}
	if ctx.Session.Statuses.Active == nil {
		ctx.Session.Statuses.Active = make(map[uint16]session.StatusEffect)
	}
	if !change.Active {
		delete(ctx.Session.Statuses.Active, change.StatusID)
		log.Printf("status effect inactive id=%d actor=%d", change.StatusID, change.ActorID)
		return
	}
	now := time.Now()
	effect := session.StatusEffect{
		ID:          change.StatusID,
		Source:      change.ActorID,
		StartedAt:   now,
		HasDuration: change.HasDuration,
	}
	if change.HasDuration {
		effect.ExpiresAt = now.Add(change.Duration)
	}
	ctx.Session.Statuses.Active[change.StatusID] = effect
	log.Printf("status effect active id=%d actor=%d duration_ms=%d", change.StatusID, change.ActorID, change.Duration.Milliseconds())
}

func (m *WorldMode) drawStatusIcons(screen *render.Image, ctx Context, now time.Time) {
	if screen == nil || ctx.Session == nil || len(ctx.Session.Statuses.Active) == 0 {
		return
	}
	m.removeExpiredStatusIcons(ctx.Session, now)
	ids := visibleStatusIconIDs(ctx.Session.Statuses.Active)
	if len(ids) == 0 {
		return
	}
	width, height := ctx.ScreenSize()
	minimapX, minimapY, minimapW, minimapH := gameui.MinimapBounds(width, height)
	startX := minimapX + minimapW - statusIconSize
	startY := minimapY + minimapH + statusIconGap
	maxRows := maxInt(1, (height-startY-16)/statusIconSpacing)
	hovered := -1
	for i, id := range ids {
		col := i / maxRows
		row := i % maxRows
		x := startX - col*(statusIconSize+statusIconGap)
		y := startY + row*statusIconSpacing
		effect := ctx.Session.Statuses.Active[id]
		m.drawStatusIcon(screen, ctx.Resources, id, effect, x, y, now)
		if ctx.Input != nil && pointInRect(ctx.Input.MouseX, ctx.Input.MouseY, x, y, statusIconSize, statusIconSize) {
			hovered = int(id)
		}
	}
	if hovered >= 0 && ctx.Input != nil {
		m.drawStatusIconTooltip(screen, hovered, ctx.Input.MouseX, ctx.Input.MouseY, width, height)
	}
}

func (m *WorldMode) removeExpiredStatusIcons(s *session.Session, now time.Time) {
	if s == nil {
		return
	}
	for id, effect := range s.Statuses.Active {
		if effect.HasDuration && !effect.ExpiresAt.IsZero() && now.After(effect.ExpiresAt) {
			delete(s.Statuses.Active, id)
		}
	}
}

func visibleStatusIconIDs(active map[uint16]session.StatusEffect) []uint16 {
	ids := make([]uint16, 0, len(active))
	for id := range active {
		if _, ok := statusIconInfos[id]; ok {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}

func (m *WorldMode) drawStatusIcon(screen *render.Image, manager *res.Manager, id uint16, effect session.StatusEffect, x, y int, now time.Time) {
	render.DrawRect(screen, float64(x-1), float64(y-1), statusIconSize+2, statusIconSize+2, color.RGBA{R: 60, G: 74, B: 96, A: 170})
	render.DrawRect(screen, float64(x), float64(y), statusIconSize, statusIconSize, color.RGBA{R: 236, G: 242, B: 250, A: 215})
	if icon := m.statusIconTexture(manager, id); icon != nil {
		bounds := icon.Bounds()
		if bounds.Dx() > 0 && bounds.Dy() > 0 {
			scale := math.Min(float64(statusIconSize)/float64(bounds.Dx()), float64(statusIconSize)/float64(bounds.Dy()))
			var opts render.DrawImageOptions
			opts.GeoM.Scale(scale, scale)
			opts.GeoM.Translate(float64(x)+(statusIconSize-float64(bounds.Dx())*scale)/2, float64(y)+(statusIconSize-float64(bounds.Dy())*scale)/2)
			opts.Filter = spriteDrawFilter()
			screen.DrawImage(icon, &opts)
		}
	} else {
		render.DebugPrintAtColor(screen, "?", x+12, y+9, gameui.MutedTextColor)
	}
	if effect.HasDuration && !effect.ExpiresAt.IsZero() && effect.ExpiresAt.After(effect.StartedAt) {
		total := effect.ExpiresAt.Sub(effect.StartedAt)
		remaining := effect.ExpiresAt.Sub(now)
		if remaining < 0 {
			remaining = 0
		}
		frac := float64(remaining) / float64(total)
		fillW := int(float64(statusIconSize) * clampUnit(frac))
		render.DrawRect(screen, float64(x), float64(y+statusIconSize-4), statusIconSize, 4, color.RGBA{R: 18, G: 24, B: 34, A: 180})
		if fillW > 0 {
			render.DrawRect(screen, float64(x), float64(y+statusIconSize-4), float64(fillW), 4, color.RGBA{R: 244, G: 228, B: 130, A: 230})
		}
	}
}

func (m *WorldMode) statusIconTexture(manager *res.Manager, id uint16) *render.Image {
	info, ok := statusIconInfos[id]
	if !ok || manager == nil {
		return nil
	}
	key := "__status_icon_" + info.icon
	if m.textures == nil {
		m.textures = make(map[string]*render.Image)
	}
	if m.textureMiss == nil {
		m.textureMiss = make(map[string]struct{})
	}
	if texture, ok := m.textures[key]; ok {
		return texture
	}
	if _, ok := m.textureMiss[key]; ok {
		return nil
	}
	img, _, err := res.LoadImage(manager, res.EffectTextureCandidates(info.icon))
	if err != nil {
		m.textureMiss[key] = struct{}{}
		return nil
	}
	texture := render.NewImageFromImage(img)
	m.textures[key] = texture
	return texture
}

func (m *WorldMode) drawStatusIconTooltip(screen *render.Image, statusID int, mouseX, mouseY, width, height int) {
	info, ok := statusIconInfos[uint16(statusID)]
	if !ok || info.label == "" {
		return
	}
	text := info.label
	w := len(text)*7 + 12
	h := 20
	x := clampInt(mouseX+12, 4, maxInt(4, width-w-4))
	y := clampInt(mouseY+12, 4, maxInt(4, height-h-4))
	render.DrawRect(screen, float64(x), float64(y), float64(w), float64(h), color.RGBA{R: 32, G: 36, B: 44, A: 230})
	render.DrawRect(screen, float64(x), float64(y), float64(w), 1, gameui.WindowBorderColor)
	render.DrawRect(screen, float64(x), float64(y+h-1), float64(w), 1, gameui.WindowBorderColor)
	render.DebugPrintAtColor(screen, text, x+6, y+4, color.RGBA{R: 246, G: 246, B: 246, A: 255})
}
