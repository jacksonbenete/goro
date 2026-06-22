package network

import "testing"

func TestBuildAccountLoginPacket(t *testing.T) {
	packet := BuildAccountLoginPacket(AccountLogin{
		Version:    55,
		Username:   "alice",
		Password:   "secret",
		ClientType: 0,
	})

	if len(packet) != 55 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x64 || packet[1] != 0x00 {
		t.Fatalf("opcode = %02x %02x", packet[0], packet[1])
	}
	if packet[6] != 'a' || packet[30] != 's' {
		t.Fatalf("username/password offsets are wrong")
	}
}

func TestBuildCharServerEnterPacket(t *testing.T) {
	packet := BuildCharServerEnterPacket(CharServerEnter{
		AccountID: 10,
		AuthCode:  20,
		UserLevel: 30,
		Sex:       1,
	})
	if len(packet) != 17 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x65 || packet[1] != 0x00 || packet[16] != 1 {
		t.Fatalf("unexpected packet bytes: % x", packet)
	}
}

func TestBuildSelectCharacterPacket(t *testing.T) {
	packet := BuildSelectCharacterPacket(2)
	if len(packet) != 3 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x66 || packet[1] != 0x00 || packet[2] != 2 {
		t.Fatalf("unexpected packet bytes: % x", packet)
	}
}

func TestBuildLoadEndAckPacket(t *testing.T) {
	packet := BuildLoadEndAckPacket()
	if len(packet) != 2 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x7D || packet[1] != 0x00 {
		t.Fatalf("unexpected packet bytes: % x", packet)
	}
}

func TestBuildMapServerEnterPacket(t *testing.T) {
	packet := BuildMapServerEnterPacket(MapServerEnter{
		AccountID:  0x11223344,
		CharID:     0x55667788,
		AuthCode:   0x99aabbcc,
		ClientTick: 0xddeeff00,
		Sex:        1,
	})
	if len(packet) != 19 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x36 || packet[1] != 0x04 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	if packet[18] != 1 {
		t.Fatalf("unexpected sex offset: % x", packet)
	}
}

func TestBuildMapServerEnterPacketFor2021ClientDate(t *testing.T) {
	packet := BuildMapServerEnterPacketForClientDate(MapServerEnter{
		AccountID:  0x11223344,
		CharID:     0x55667788,
		AuthCode:   0x99aabbcc,
		ClientTick: 0xddeeff00,
		Sex:        1,
	}, 20211103)
	if len(packet) != 23 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x36 || packet[1] != 0x04 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	if packet[22] != 1 {
		t.Fatalf("unexpected sex offset: % x", packet)
	}
	for i := 18; i < 22; i++ {
		if packet[i] != 0 {
			t.Fatalf("expected zero padding at %d: % x", i, packet)
		}
	}
}

func TestBuildTickSendPacket(t *testing.T) {
	packet := BuildTickSendPacket(0x11223344)
	if len(packet) != 8 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x89 || packet[1] != 0x00 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	if packet[4] != 0x44 || packet[5] != 0x33 || packet[6] != 0x22 || packet[7] != 0x11 {
		t.Fatalf("unexpected tick bytes: % x", packet[4:8])
	}
}

func TestBuildTickSendPacketFor2021ClientDate(t *testing.T) {
	packet := BuildTickSendPacketForClientDate(0x11223344, 20211103)
	if len(packet) != 6 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x60 || packet[1] != 0x03 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	if packet[2] != 0x44 || packet[3] != 0x33 || packet[4] != 0x22 || packet[5] != 0x11 {
		t.Fatalf("unexpected tick bytes: % x", packet[2:6])
	}
}

func TestBuildNameRequestPacket(t *testing.T) {
	packet := BuildNameRequestPacket(0x11223344)
	want := []byte{0x94, 0x00, 0x44, 0x33, 0x22, 0x11}
	if len(packet) != len(want) {
		t.Fatalf("len = %d", len(packet))
	}
	for i := range want {
		if packet[i] != want[i] {
			t.Fatalf("packet = % x, want % x", packet, want)
		}
	}
}

func TestBuildWalkToXYPacket(t *testing.T) {
	packet, ok := BuildWalkToXYPacket(150, 200)
	if !ok {
		t.Fatal("expected packet")
	}
	if len(packet) != 8 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0xA7 || packet[1] != 0x00 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	want0, want1, want2 := packPosition(150, 200, 0)
	if packet[5] != want0 || packet[6] != want1 || packet[7] != want2 {
		t.Fatalf("unexpected dest: got % x want %02x %02x %02x", packet[5:8], want0, want1, want2)
	}
}

func TestBuildWalkToXYPacketFor2008ClientDate(t *testing.T) {
	packet, ok := BuildWalkToXYPacketForClientDate(150, 200, 20080910)
	if !ok {
		t.Fatal("expected packet")
	}
	if len(packet) != 8 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0xA7 || packet[1] != 0x00 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	want0, want1, want2 := packPosition(150, 200, 0)
	if packet[5] != want0 || packet[6] != want1 || packet[7] != want2 {
		t.Fatalf("unexpected dest: got % x want %02x %02x %02x", packet[5:8], want0, want1, want2)
	}
}

func TestBuildWalkToXYPacketFor2021ClientDate(t *testing.T) {
	packet, ok := BuildWalkToXYPacketForClientDate(150, 200, 20211103)
	if !ok {
		t.Fatal("expected packet")
	}
	if len(packet) != 5 {
		t.Fatalf("len = %d", len(packet))
	}
	if packet[0] != 0x5F || packet[1] != 0x03 {
		t.Fatalf("unexpected opcode: % x", packet[:2])
	}
	want0, want1, want2 := packPosition(150, 200, 0)
	if packet[2] != want0 || packet[3] != want1 || packet[4] != want2 {
		t.Fatalf("unexpected dest: got % x want %02x %02x %02x", packet[2:5], want0, want1, want2)
	}
}
