package game

import (
	"testing"
	"time"
)

func TestMapWeatherForComodoMatchesReferenceWeatherTable(t *testing.T) {
	for _, name := range []string{"comodo", "comodo.gat", "data\\comodo.rsw", "DATA/COMODO.RSW"} {
		if got := mapWeatherForMap(name); got != mapWeatherFireworks {
			t.Fatalf("mapWeatherForMap(%q) = %d, want fireworks", name, got)
		}
	}
	if got := mapWeatherForMap("prontera"); got != mapWeatherNone {
		t.Fatalf("mapWeatherForMap(prontera) = %d, want none", got)
	}
}

func TestPokJukWeatherEffectSpecUsesAdditiveParticles(t *testing.T) {
	spec, ok := worldEffectSpecForID(effectPokJuk)
	if !ok {
		t.Fatal("pokjuk effect spec missing")
	}
	if spec.duration != 4*time.Second || len(spec.components) != 4 {
		t.Fatalf("pokjuk spec = %+v", spec)
	}
	launch := spec.components[0]
	if launch.kind != effectComponent3D || launch.textureFile != "effect/pok3.tga" || !launch.blendAdditive {
		t.Fatalf("pokjuk launch = %+v", launch)
	}
	if launch.posZ != 0 || launch.posZEnd != 8 {
		t.Fatalf("pokjuk launch z = %.1f..%.1f", launch.posZ, launch.posZEnd)
	}
	for i, component := range spec.components[1:] {
		if component.kind != effectComponent3D || component.textureFile != "effect/pok3.tga" || !component.blendAdditive {
			t.Fatalf("pokjuk burst component %d = %+v", i+1, component)
		}
		if component.duplicate == 0 || component.posXEndRand <= 0 || component.posYEndRand <= 0 || component.posZEndRand <= 0 {
			t.Fatalf("pokjuk burst spread %d = %+v", i+1, component)
		}
	}
}

func TestLoopingMapWeatherEffectStartStaggersFireworks(t *testing.T) {
	now := time.Unix(10, 500*int64(time.Millisecond))
	duration := 4 * time.Second
	first := loopingMapWeatherEffectStart(0, duration, now)
	second := loopingMapWeatherEffectStart(1, duration, now)
	if first.Equal(second) {
		t.Fatalf("weather effect starts were not staggered: %s", first)
	}
	if now.Sub(first) < 0 || now.Sub(first) >= duration {
		t.Fatalf("first start = %s for now %s", first, now)
	}
	if now.Sub(second) < 0 || now.Sub(second) >= duration {
		t.Fatalf("second start = %s for now %s", second, now)
	}
}

func TestScheduleMapWeatherSoundOnlyNearNewCycle(t *testing.T) {
	now := time.Unix(10, 0)
	mode := NewWorldMode()
	mode.scheduleMapWeatherSound(0, now.Add(-time.Second), now, pokJukLaunchSFX)
	if len(mode.scheduledSounds) != 0 {
		t.Fatalf("scheduled stale weather sounds = %+v", mode.scheduledSounds)
	}
	mode.scheduleMapWeatherSound(0, now.Add(-50*time.Millisecond), now, pokJukLaunchSFX)
	if len(mode.scheduledSounds) != 1 || mode.scheduledSounds[0].paths[0] != pokJukLaunchSFX {
		t.Fatalf("scheduled weather sounds = %+v", mode.scheduledSounds)
	}
	mode.scheduleMapWeatherSound(0, now.Add(-50*time.Millisecond), now, pokJukLaunchSFX)
	if len(mode.scheduledSounds) != 1 {
		t.Fatalf("weather sound scheduled twice = %+v", mode.scheduledSounds)
	}
}

func TestScheduleMapWeatherExplosionSoundUsesItemPokjukFallback(t *testing.T) {
	now := time.Unix(10, 0)
	mode := NewWorldMode()
	mode.scheduleMapWeatherSound(1, now.Add(-50*time.Millisecond), now, pokJukExplosionSFX, pokJukLaunchSFX)
	if len(mode.scheduledSounds) != 1 {
		t.Fatalf("scheduled weather sounds = %+v", mode.scheduledSounds)
	}
	got := mode.scheduledSounds[0].paths
	if len(got) != 2 || got[0] != pokJukExplosionSFX || got[1] != pokJukLaunchSFX {
		t.Fatalf("explosion sound paths = %#v", got)
	}
}
