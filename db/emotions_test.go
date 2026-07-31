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

func TestEmotionListUsesRobrowserInterfaceOrder(t *testing.T) {
	emotes := EmotionList()
	if len(emotes) != 64 {
		t.Fatalf("EmotionList length = %d, want 64", len(emotes))
	}
	tests := []struct {
		index   int
		id      uint8
		frame   int
		command string
	}{
		{0, 0, 0, "!"},
		{1, 1, 1, "?"},
		{4, 14, 4, "lv2"},
		{5, 4, 5, "swt"},
		{63, 80, 79, "whisp"},
	}
	for _, tt := range tests {
		got := emotes[tt.index]
		if got.ID != tt.id || got.Frame != tt.frame || got.Command != tt.command {
			t.Fatalf("EmotionList[%d] = %+v, want id=%d frame=%d command=%q", tt.index, got, tt.id, tt.frame, tt.command)
		}
	}
	for _, got := range emotes {
		if got.ID == 58 {
			t.Fatalf("EmotionList includes dice id 58, but roBrowser does not place dice in the emote window")
		}
	}
}

func TestEmotionListReturnsCommandCopies(t *testing.T) {
	emotes := EmotionList()
	if len(emotes) == 0 || len(emotes[0].Commands) == 0 {
		t.Fatal("EmotionList returned no command data")
	}
	emotes[0].Commands[0] = "changed"
	next := EmotionList()
	if next[0].Commands[0] != "!" {
		t.Fatalf("EmotionList command copy leaked mutation: %q", next[0].Commands[0])
	}
}
