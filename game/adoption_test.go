package game

import (
	"encoding/binary"
	"testing"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/db"
	"github.com/kivutar/goro/network"
	worldstate "github.com/kivutar/goro/world"
)

func TestAdoptionRequestPacketOpensConfirmation(t *testing.T) {
	data := make([]byte, 34)
	binary.LittleEndian.PutUint16(data[0:2], network.PacketZCReqBaby)
	binary.LittleEndian.PutUint32(data[2:6], 2000001)
	binary.LittleEndian.PutUint32(data[6:10], 2000000)
	copy(data[10:34], "Zambla")

	mode := &WorldMode{}
	ctx := client.Context{ScreenW: 800, ScreenH: 600}
	mode.handleNetworkPacket(ctx, network.Packet{ID: network.PacketZCReqBaby, Data: data}, time.Now())

	if !mode.ui.adoptionRequest.IsOpen() {
		t.Fatal("adoption request did not open a confirmation")
	}
}

func TestBabyPlayerBodyScale(t *testing.T) {
	if got := playerBodyScaleForJob(db.JobNovice); got != 1 {
		t.Fatalf("adult scale = %v, want 1", got)
	}
	if got := playerBodyScaleForJob(db.JobNoviceB); got != babyPlayerBodyScale {
		t.Fatalf("baby scale = %v, want %v", got, babyPlayerBodyScale)
	}
}

func TestBodySizeEffectUsesAbsoluteScaleForBaby(t *testing.T) {
	starts := time.Unix(10, 0)
	mode := WorldMode{worldEffects: []worldEffect{{
		effectID: effectBabyBody,
		actorID:  42,
		starts:   starts,
		expires:  starts.Add(300 * time.Millisecond),
		duration: 300 * time.Millisecond,
	}}}
	ctx := client.Context{}
	actor := worldstate.Actor{ID: 42, Job: db.JobNoviceB}
	baseScale := 10.0 * babyPlayerBodyScale

	for _, tc := range []struct {
		name string
		at   time.Duration
		want float64
	}{
		{"start", 0, 8},
		{"middle", 150 * time.Millisecond, 6.5},
		{"end", 300 * time.Millisecond, 5},
	} {
		if got := mode.playerRenderScale(ctx, actor, baseScale, starts.Add(tc.at)); got != tc.want {
			t.Fatalf("baby effect scale at %s = %v, want %v", tc.name, got, tc.want)
		}
	}
}
