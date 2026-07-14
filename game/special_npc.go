package game

import (
	"github.com/kivutar/goro/client"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/render"
	worldstate "github.com/kivutar/goro/world"
)

type specialNPCVisual uint8

const (
	specialNPCVisualNone specialNPCVisual = iota
	specialNPCVisualTorch
)

func specialNPCVisualForActor(ctx client.Context, actor worldstate.Actor) specialNPCVisual {
	resourceName := ""
	if ctx.Resources != nil {
		resourceName, _ = ctx.Resources.NonPCResourceName(int(actor.Job))
	}
	return specialNPCVisualForActorResource(actor, resourceName)
}

func specialNPCVisualForActorResource(actor worldstate.Actor, resourceName string) specialNPCVisual {
	if int(actor.Job) == actorJobClearNPC {
		switch normalizeSpecialNPCActorName(actor.Name) {
		case "BOBBING TORCH", "WET FIREWOOD":
			return specialNPCVisualTorch
		}
	}
	return specialNPCVisualNone
}

func normalizeSpecialNPCResourceName(name string) string {
	name = strings.TrimSpace(strings.ReplaceAll(name, "/", "\\"))
	name = strings.TrimPrefix(name, "data\\sprite\\")
	if i := strings.LastIndex(name, "\\"); i >= 0 {
		name = name[i+1:]
	}
	return strings.ToUpper(name)
}

func normalizeSpecialNPCActorName(name string) string {
	name = sanitizeActorName(name)
	if i := strings.IndexByte(name, '#'); i >= 0 {
		name = name[:i]
	}
	return strings.ToUpper(strings.TrimSpace(name))
}

func actorJobHasSpecialNoShadow(job int) bool {
	if actorJobHasNoSprite(job) {
		return true
	}
	switch job {
	case actorJobClearNPC:
		return true
	default:
		return false
	}
}

func (m *WorldMode) drawSpecialNPCVisual(screen *render.Frame, ctx client.Context, projection sceneProjection, entry sceneActorDrawEntry, visual specialNPCVisual, now time.Time) bool {
	switch visual {
	case specialNPCVisualTorch:
		m.drawPersistentWorldEffectAt(screen, ctx, projection, effectTorch, entry, now)
		return true
	default:
		return false
	}
}

func (m *WorldMode) drawPersistentWorldEffectAt(screen *render.Frame, ctx client.Context, projection sceneProjection, effectID int, entry sceneActorDrawEntry, now time.Time) {
	spec, ok := worldEffectSpecForID(effectID)
	if !ok {
		return
	}
	effect := worldEffect{
		effectID: effectID,
		actorID:  entry.actor.ID,
		x:        int(math.Round(entry.worldX - 0.5)),
		y:        int(math.Round(entry.worldY - 0.5)),
		starts:   now,
		expires:  now.Add(24 * time.Hour),
		duration: 24 * time.Hour,
	}
	for index, component := range spec.components {
		component = mapEffectComponent(component)
		duration := worldEffectComponentDuration(spec, component)
		if component.duration > 0 {
			duration = component.duration
		}
		if duration <= 0 {
			duration = 600 * time.Millisecond
		}
		elapsed := now.Sub(time.Unix(0, 0))
		componentStart := now.Add(-(elapsed % duration))
		effect.starts = componentStart
		progress := worldEffectComponentProgress(componentStart, duration, now)
		m.drawWorldEffectComponent(screen, ctx, projection, effect, component, index, entry.worldX, entry.worldY, entry.worldZ+0.07, progress, now)
	}
}
