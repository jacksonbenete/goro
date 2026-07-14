package game

import (
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
)

type mapWeatherKind int

const (
	mapWeatherNone mapWeatherKind = iota
	mapWeatherLoopingEffect
	mapWeatherFireworks
)

const (
	pokJukLaunchSFX        = "effect\\\xc6\xf8\xc1\xd7.wav"
	pokJukExplosionSFX     = "effect\\itempokjuk.wav"
	pokJukExplosionDelay   = 900 * time.Millisecond
	pokJukWeatherFireworks = 2
)

func mapWeatherForMap(name string) mapWeatherKind {
	if mapWeatherEffectIDForMap(name) != 0 {
		return mapWeatherLoopingEffect
	}
	if normalizeMapNameForWeather(name) == "comodo.rsw" {
		return mapWeatherFireworks
	}
	return mapWeatherNone
}

func mapWeatherEffectIDForMap(name string) int {
	switch normalizeMapNameForWeather(name) {
	case "xmas.rsw":
		return effectSnow
	case "einbroch.rsw":
		return effectCloud3
	default:
		return 0
	}
}

func normalizeMapNameForWeather(name string) string {
	name = strings.ToLower(strings.TrimSpace(name))
	if index := strings.LastIndexAny(name, `\/`); index >= 0 {
		name = name[index+1:]
	}
	if strings.HasSuffix(name, ".gat") {
		name = strings.TrimSuffix(name, ".gat") + ".rsw"
	}
	if name != "" && !strings.Contains(name, ".") {
		name += ".rsw"
	}
	return name
}

func (m *WorldMode) drawMapWeatherEffects(screen *render.Frame, ctx client.Context, projection sceneProjection, now time.Time) {
	if screen == nil || ctx.World == nil || ctx.World.GND == nil {
		return
	}
	switch mapWeatherForMap(ctx.World.MapName) {
	case mapWeatherFireworks:
		m.drawFireworksWeather(screen, ctx, projection, now)
	case mapWeatherLoopingEffect:
		m.drawLoopingMapWeatherEffect(screen, ctx, projection, mapWeatherEffectIDForMap(ctx.World.MapName), now)
	}
}

func (m *WorldMode) drawLoopingMapWeatherEffect(screen *render.Frame, ctx client.Context, projection sceneProjection, effectID int, now time.Time) {
	spec, ok := worldEffectSpecForID(effectID)
	if !ok {
		return
	}
	starts := loopingMapWeatherEffectStart(effectID, spec.duration, now)
	effect := worldEffect{
		effectID: effectID,
		actorID:  uint32(effectID),
		starts:   starts,
		expires:  now.Add(24 * time.Hour),
		duration: spec.duration,
		x:        int(math.Round(projection.playerX)),
		y:        int(math.Round(projection.playerY)),
	}
	worldX := projection.playerX
	worldY := projection.playerY
	worldZ := terrainHeightAt(ctx.World, worldX-0.5, worldY-0.5) + 1
	for componentIndex, component := range spec.components {
		duration := m.worldEffectResolvedComponentDuration(ctx, spec, component)
		if duration <= 0 {
			duration = spec.duration
		}
		componentStarts := loopingMapWeatherEffectStart(effectID+componentIndex*17, duration, now)
		effect.starts = componentStarts
		progress := worldEffectComponentProgress(componentStarts, duration, now)
		m.drawWorldEffectComponent(screen, ctx, projection, effect, component, componentIndex, worldX, worldY, worldZ, progress, now)
	}
}

func (m *WorldMode) drawFireworksWeather(screen *render.Frame, ctx client.Context, projection sceneProjection, now time.Time) {
	spec, ok := worldEffectSpecForID(effectPokJuk)
	if !ok {
		return
	}
	const count = pokJukWeatherFireworks
	for i := 0; i < count; i++ {
		starts := loopingMapWeatherEffectStart(i, spec.duration, now)
		m.scheduleMapWeatherSound(i*2, starts, now, pokJukLaunchSFX)
		m.scheduleMapWeatherSound(i*2+1, starts.Add(pokJukExplosionDelay), now, pokJukExplosionSFX, pokJukLaunchSFX)
		effect := worldEffect{
			effectID: effectPokJuk,
			actorID:  uint32(i + 1),
			starts:   starts,
			expires:  now.Add(24 * time.Hour),
			duration: spec.duration,
			x:        int(math.Round(projection.playerX)),
			y:        int(math.Round(projection.playerY)),
		}
		worldX := projection.playerX + deterministicSigned(effect, 11)*10
		worldY := projection.playerY + deterministicSigned(effect, 12)*10
		worldZ := terrainHeightAt(ctx.World, worldX-0.5, worldY-0.5) + 1
		for componentIndex, component := range spec.components {
			duration := m.worldEffectResolvedComponentDuration(ctx, spec, component)
			if duration <= 0 {
				duration = spec.duration
			}
			progress := worldEffectComponentProgress(starts, duration, now)
			m.drawWorldEffectComponent(screen, ctx, projection, effect, component, componentIndex, worldX, worldY, worldZ, progress, now)
		}
	}
}

func (m *WorldMode) scheduleMapWeatherSound(index int, starts time.Time, now time.Time, paths ...string) {
	elapsed := now.Sub(starts)
	if elapsed < 0 {
		return
	}
	if m.mapWeatherSounds == nil {
		m.mapWeatherSounds = make(map[int]time.Time)
	}
	if m.mapWeatherSounds[index].Equal(starts) {
		return
	}
	m.mapWeatherSounds[index] = starts
	if elapsed > 120*time.Millisecond {
		return
	}
	m.scheduleSound(now, paths...)
}

func loopingMapWeatherEffectStart(index int, duration time.Duration, now time.Time) time.Time {
	if duration <= 0 {
		return now
	}
	phase := time.Duration(index) * duration / 3
	origin := time.Unix(0, 0).Add(phase)
	elapsed := now.Sub(origin)
	if elapsed < 0 {
		return origin
	}
	return now.Add(-(elapsed % duration))
}
