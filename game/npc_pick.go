package game

import (
	"time"

	worldstate "github.com/kivutar/goro/world"
)

func clickedTalkTarget(ctx Context, projection sceneProjection, mouseX, mouseY int, now time.Time, deadActors map[uint32]time.Time) (worldstate.Actor, bool) {
	actor, ok := hoveredCursorActor(ctx, projection, mouseX, mouseY, now, deadActors)
	if !ok || !cursorActorCanTalk(actor) {
		return worldstate.Actor{}, false
	}
	return actor, true
}
