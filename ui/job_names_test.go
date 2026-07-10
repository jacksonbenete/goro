package ui

import (
	"testing"

	"github.com/kivutar/goro/session"
)

func TestCharacterJobName(t *testing.T) {
	character := session.Character{Job: 7}
	if got := CharacterJobName(character); got != "Knight" {
		t.Fatalf("CharacterJobName() = %q, want Knight", got)
	}
}
