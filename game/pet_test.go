package game

import (
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/network"
	worldstate "github.com/kivutar/goro/world"
)

func TestPetTalkNumberMatchesROBrowserLayout(t *testing.T) {
	if got := petTalkNumber(1063, petTalkFeeding, 1); got != 1063010 {
		t.Fatalf("petTalkNumber = %d, want 1063010", got)
	}
}

func TestPetHungryStateMatchesROBrowserBuckets(t *testing.T) {
	cases := []struct {
		fullness int
		want     int
	}{
		{0, 0},
		{10, 0},
		{11, 1},
		{25, 1},
		{26, 2},
		{75, 2},
		{76, 3},
		{90, 3},
		{91, 4},
		{100, 4},
	}
	for _, tc := range cases {
		if got := petHungryState(tc.fullness); got != tc.want {
			t.Fatalf("petHungryState(%d) = %d, want %d", tc.fullness, got, tc.want)
		}
	}
}

func TestPetTalkJobUsesActorJobBeforePropertyJob(t *testing.T) {
	mode := NewWorldMode()
	mode.petID = 123
	mode.petProperty = network.PetProperty{Job: 271}
	ctx := client.Context{World: &worldstate.World{Actors: map[uint32]worldstate.Actor{
		123: {ID: 123, Job: 1063},
	}}}
	if got := mode.petTalkJob(ctx); got != 1063 {
		t.Fatalf("petTalkJob = %d, want actor job 1063", got)
	}
}
