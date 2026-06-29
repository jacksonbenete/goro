package gamemode

import (
	"testing"

	worldstate "github.com/kivutar/goro/world"
)

func TestSpecialNPCVisualForActorResource(t *testing.T) {
	tests := []struct {
		name     string
		actor    worldstate.Actor
		resource string
		want     specialNPCVisual
	}{
		{
			name:     "guild flag gr2 remains sprite fallback",
			actor:    worldstate.Actor{Job: 722},
			resource: "Guildflag90_1.gr2",
			want:     specialNPCVisualNone,
		},
		{
			name:     "city flag a remains sprite",
			actor:    worldstate.Actor{Job: 1912},
			resource: "OBJ_FLAG_A",
			want:     specialNPCVisualNone,
		},
		{
			name:     "city flag b remains sprite",
			actor:    worldstate.Actor{Job: 1913},
			resource: "OBJ_FLAG_B",
			want:     specialNPCVisualNone,
		},
		{
			name:     "sprite flag remains sprite",
			actor:    worldstate.Actor{Job: 973},
			resource: "1_FLAG_LION",
			want:     specialNPCVisualNone,
		},
		{
			name:     "clear npc torch",
			actor:    worldstate.Actor{Job: actorJobClearNPC, Name: "Bobbing Torch#7"},
			resource: "CLEAR_NPC",
			want:     specialNPCVisualTorch,
		},
		{
			name:     "clear npc firewood",
			actor:    worldstate.Actor{Job: actorJobClearNPC, Name: "Wet Firewood#moc2"},
			resource: "CLEAR_NPC",
			want:     specialNPCVisualTorch,
		},
		{
			name:     "flame monster remains sprite",
			actor:    worldstate.Actor{Job: 1869, HasObjectType: true, ObjectType: actorObjectTypeMob},
			resource: "FLAME_SKULL",
			want:     specialNPCVisualNone,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := specialNPCVisualForActorResource(tt.actor, tt.resource); got != tt.want {
				t.Fatalf("specialNPCVisualForActorResource(%+v, %q) = %d, want %d", tt.actor, tt.resource, got, tt.want)
			}
		})
	}
}

func TestSpecialNPCResourceNormalization(t *testing.T) {
	got := normalizeSpecialNPCResourceName(`data/sprite/npc/Guildflag90_1.gr2`)
	if got != "GUILDFLAG90_1.GR2" {
		t.Fatalf("normalized resource = %q", got)
	}
}
