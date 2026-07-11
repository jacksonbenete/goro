package db

import "testing"

func TestEmotionCommandID(t *testing.T) {
	tests := map[string]uint8{
		"!":     0,
		"?":     1,
		"swt":   4,
		"heart": 3,
		"dice":  58,
		"ok":    33,
	}
	for command, want := range tests {
		got, ok := EmotionCommandID(command)
		if !ok {
			t.Fatalf("EmotionCommandID(%q) not found", command)
		}
		if got != want {
			t.Fatalf("EmotionCommandID(%q) = %d, want %d", command, got, want)
		}
	}
}

func TestEmotionSpriteFrame(t *testing.T) {
	tests := map[uint8]int{
		0:  0,
		3:  3,
		4:  5,
		14: 4,
		58: 57,
	}
	for id, want := range tests {
		got, ok := EmotionSpriteFrame(id)
		if !ok {
			t.Fatalf("EmotionSpriteFrame(%d) not found", id)
		}
		if got != want {
			t.Fatalf("EmotionSpriteFrame(%d) = %d, want %d", id, got, want)
		}
	}
}
