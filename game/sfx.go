package game

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/res"
	worldstate "github.com/kivutar/goro/world"
)

type scheduledSound struct {
	at     time.Time
	paths  []string
	volume float64
}

type actorSoundFrame struct {
	actionFamily int
	motion       int
	soundIndex   int
}

func (m *WorldMode) scheduleSound(at time.Time, paths ...string) {
	m.scheduleSoundVolume(at, 1, paths...)
}

func (m *WorldMode) scheduleSoundVolume(at time.Time, volume float64, paths ...string) {
	clean := paths[:0]
	for _, path := range paths {
		path = strings.TrimSpace(path)
		if path != "" {
			clean = append(clean, path)
		}
	}
	if len(clean) == 0 {
		return
	}
	m.scheduledSounds = append(m.scheduledSounds, scheduledSound{
		at:     at,
		paths:  append([]string(nil), clean...),
		volume: volume,
	})
}

func (m *WorldMode) playDueScheduledSounds(ctx client.Context, now time.Time) {
	if len(m.scheduledSounds) == 0 {
		return
	}
	active := m.scheduledSounds[:0]
	for _, sound := range m.scheduledSounds {
		if now.Before(sound.at) {
			active = append(active, sound)
			continue
		}
		m.playSFXFirstVolume(ctx, sound.volume, sound.paths...)
	}
	m.scheduledSounds = active
}

func (m *WorldMode) processMapSounds(ctx client.Context, now time.Time) {
	if ctx.World == nil || ctx.World.RSW == nil || ctx.World.GND == nil || len(ctx.World.RSW.Sounds) == 0 {
		return
	}
	playerX, playerY := ctx.World.Player.RenderPosition(now)
	width := float64(ctx.World.GND.Width)
	height := float64(ctx.World.GND.Height)
	if m.mapSoundNext == nil {
		m.mapSoundNext = make(map[int]time.Time)
	}
	for index, sound := range ctx.World.RSW.Sounds {
		if strings.TrimSpace(sound.File) == "" {
			continue
		}
		if sound.Volume <= 0 {
			continue
		}
		if next := m.mapSoundNext[index]; !next.IsZero() && now.Before(next) {
			continue
		}
		soundX := float64(sound.Position.X) + width
		soundY := float64(sound.Position.Z) + height
		maxDistance := float64(sound.Range)*0.2 + float64(sound.Height)
		if math.Hypot(soundX-playerX, soundY-playerY) > maxDistance {
			continue
		}
		m.scheduleSoundVolume(now, float64(sound.Volume), sound.File)
		delay := time.Duration(float64(time.Second) * float64(sound.Cycle))
		if delay < 100*time.Millisecond {
			delay = 100 * time.Millisecond
		}
		m.mapSoundNext[index] = now.Add(delay)
	}
}

func (m *WorldMode) processActorMotionSounds(ctx client.Context, now time.Time) {
	if ctx.World == nil || len(ctx.World.Actors) == 0 {
		return
	}
	for _, actor := range ctx.World.Actors {
		m.processNonPCMotionSound(ctx, actor, now)
	}
}

func (m *WorldMode) processNonPCMotionSound(ctx client.Context, actor worldstate.Actor, now time.Time) {
	if actor.ID == 0 || res.HasPlayerJobToken(int(actor.Job)) || !actorWithinSoundRange(ctx, actor, now) {
		return
	}
	view := m.nonPCSpriteView(ctx, actor)
	if view == nil || view.act == nil {
		return
	}
	state := m.nonPCSpriteState(actor, now)
	switch state.actionFamily {
	case spriteActionNonPCAttack, spriteActionNonPCHurt:
		return
	}
	_, action, ok := resolveSpriteAction(view.act, state.actionFamily, state.direction)
	if !ok || len(action.Animations) == 0 {
		return
	}
	motion := bodyMotionForState(action, state, view.started, now)
	if motion < 0 || motion >= len(action.Animations) {
		return
	}
	soundIndex := action.Animations[motion].Sound
	current := actorSoundFrame{actionFamily: state.actionFamily, motion: motion, soundIndex: soundIndex}
	if m.actorSoundFrames == nil {
		m.actorSoundFrames = make(map[uint32]actorSoundFrame)
	}
	if previous, ok := m.actorSoundFrames[actor.ID]; ok && previous == current {
		return
	}
	m.actorSoundFrames[actor.ID] = current
	if soundIndex < 0 {
		return
	}
	if sound := actionSoundName(view.act, action, motion); sound != "" {
		m.scheduleSound(now, sound)
	}
}

func actorWithinSoundRange(ctx client.Context, actor worldstate.Actor, now time.Time) bool {
	if ctx.World == nil {
		return false
	}
	actorX, actorY := actor.RenderPosition(now)
	playerX, playerY := ctx.World.Player.RenderPosition(now)
	const soundRangeCells = 25
	return math.Hypot(actorX-playerX, actorY-playerY) <= soundRangeCells
}

func (m *WorldMode) playSFXFirstVolume(ctx client.Context, volume float64, paths ...string) {
	if ctx.Audio == nil {
		return
	}
	var lastErr error
	for _, path := range paths {
		if strings.TrimSpace(path) == "" {
			continue
		}
		source, err := ctx.Audio.PlaySFXVolume(path, volume)
		if err == nil {
			if source != "" {
				log.Printf("sfx playing path=%s source=%s", path, source)
			}
			return
		}
		lastErr = err
	}
	if lastErr != nil {
		log.Printf("sfx failed paths=%v: %v", paths, lastErr)
	}
}

func actionSoundName(act *res.ACT, action res.ACTAction, motion int) string {
	if act == nil || motion < 0 || motion >= len(action.Animations) {
		return ""
	}
	soundIndex := action.Animations[motion].Sound
	if soundIndex < 0 || soundIndex >= len(act.Sounds) {
		return ""
	}
	sound := strings.TrimSpace(act.Sounds[soundIndex])
	if strings.EqualFold(sound, "atk") {
		return ""
	}
	return sound
}

func combatHitSFXCandidates(source worldstate.Actor, sourceOK bool, target worldstate.Actor, targetOK bool) []string {
	if targetOK && res.HasPlayerJobToken(int(target.Job)) {
		return playerJobHitSFX(int(target.Job))
	}
	if sourceOK && res.HasPlayerJobToken(int(source.Job)) {
		return weaponHitSFXCandidates(res.PlayerWeaponType(int(source.Weapon)))
	}
	return nil
}

func weaponHitSFXCandidates(weaponType int) []string {
	switch weaponType {
	case 0:
		return []string{"_hit_fist1.wav", "_hit_fist2.wav", "_hit_fist3.wav", "_hit_fist4.wav"}
	case 1, 2, 3:
		return []string{"_hit_sword.wav"}
	case 4, 5:
		return []string{"_hit_spear.wav"}
	case 6, 7:
		return []string{"_hit_axe.wav"}
	case 8, 9, 13, 14, 15, 16, 22:
		return []string{"_hit_mace.wav"}
	case 10, 23:
		return []string{"_hit_rod.wav"}
	case 11:
		return []string{"_hit_arrow.wav"}
	case 12:
		return []string{"_HIT_FIST2.wav"}
	case 17:
		return []string{"_hit_\xB1\xC7\xC3\xD1.wav"}
	case 18:
		return []string{"_hit_\xB6\xF3\xC0\xCC\xC7\xC3.wav"}
	case 19:
		return []string{"_hit_\xB0\xB3\xC6\xB2\xB8\xB5\xC7\xD1\xB9\xDF.wav"}
	case 20:
		return []string{"_hit_\xBC\xA6\xB0\xC7.wav"}
	case 21:
		return []string{"_hit_\xB1\xD7\xB7\xB9\xB3\xD7\xC0\xCC\xB5\xE5\xB7\xB1\xC3\xC4.wav"}
	case 102:
		return []string{"_hit_fist4.wav"}
	default:
		return []string{"_hit_mace.wav"}
	}
}

func playerJobHitSFX(job int) []string {
	switch job {
	case 1, 7, 13, 14, 21, 23, 4008, 4015, 4022, 4028, 4036, 4044, 4054, 4066, 4080:
		return []string{"player_metal.wav"}
	case 3, 6, 11, 12, 17, 19, 20, 24, 25, 4047, 4048, 4056, 4057, 4069, 4070, 4083, 4084:
		return []string{"player_wooden_male.wav"}
	default:
		return []string{"player_clothes.wav"}
	}
}
