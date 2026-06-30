//go:build nofakecgo

package audio

import "testing"

func TestBGMAndSFXVolumesAreSeparate(t *testing.T) {
	bgm := NewBGM(nil, true, 0.25, 0.75)
	if bgm.BGMVolume() != 0.25 {
		t.Fatalf("bgm volume = %v, want 0.25", bgm.BGMVolume())
	}
	if bgm.SFXVolume() != 0.75 {
		t.Fatalf("sfx volume = %v, want 0.75", bgm.SFXVolume())
	}

	bgm.SetBGMVolume(0.4)
	if bgm.BGMVolume() != 0.4 || bgm.SFXVolume() != 0.75 {
		t.Fatalf("volumes after bgm change = %v/%v", bgm.BGMVolume(), bgm.SFXVolume())
	}
	bgm.SetSFXVolume(0.6)
	if bgm.BGMVolume() != 0.4 || bgm.SFXVolume() != 0.6 {
		t.Fatalf("volumes after sfx change = %v/%v", bgm.BGMVolume(), bgm.SFXVolume())
	}
}

func TestAudioVolumesClamp(t *testing.T) {
	bgm := NewBGM(nil, true, -1, 2)
	if bgm.BGMVolume() != 0 {
		t.Fatalf("bgm volume = %v, want 0", bgm.BGMVolume())
	}
	if bgm.SFXVolume() != 1 {
		t.Fatalf("sfx volume = %v, want 1", bgm.SFXVolume())
	}
}
