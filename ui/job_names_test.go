package ui

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestJobName(t *testing.T) {
	tests := []struct {
		id   int
		want string
	}{
		{id: 0, want: "Novice"},
		{id: 1, want: "Swordman"},
		{id: 4001, want: "High Novice"},
		{id: 4008, want: "Lord Knight"},
		{id: 4046, want: "Taekwon"},
		{id: 4252, want: "Dragon Knight"},
		{id: 9999, want: "Job 9999"},
	}

	for _, tt := range tests {
		if got := JobName(tt.id); got != tt.want {
			t.Fatalf("JobName(%d) = %q, want %q", tt.id, got, tt.want)
		}
	}
}

func TestCharacterJobName(t *testing.T) {
	character := session.Character{Job: 7}
	if got := CharacterJobName(character); got != "Knight" {
		t.Fatalf("CharacterJobName() = %q, want Knight", got)
	}
}
