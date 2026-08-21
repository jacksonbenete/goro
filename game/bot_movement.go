package game

import (
	"time"

	"github.com/kivutar/goro/client"
)

func (m *WorldMode) scriptWalk(ctx client.Context, x, y int) bool {
	now := time.Now()
	if ctx.World == nil || ctx.Network == nil || !m.walkReady(now) || !walkTargetInBounds(ctx, x, y) {
		return false
	}
	if ctx.World.GAT != nil && !ctx.World.GAT.Walkable(x, y) {
		return false
	}
	if playerAtWalkTarget(ctx.World.Player, x, y, now) {
		return false
	}
	m.cancelAttackIntent()
	return m.requestWalk(ctx, x, y, "script")
}

func (m *WorldMode) scriptStop(ctx client.Context) bool {
	m.cancelAttackIntent()
	return m.requestWalkStop(ctx, "script")
}
