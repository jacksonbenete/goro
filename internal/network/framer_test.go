package network

import "testing"

func TestFramerFixedPacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0073: 3})
	packets, err := framer.Push([]byte{0x73, 0x00, 0x01})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x0073 || len(packets[0].Data) != 3 {
		t.Fatalf("unexpected packet: %+v", packets[0])
	}
}

func TestFramerVariablePacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0069: -1})
	packets, err := framer.Push([]byte{0x69, 0x00, 0x06, 0x00, 0xaa, 0xbb})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
	if packets[0].ID != 0x0069 || len(packets[0].Data) != 6 {
		t.Fatalf("unexpected packet: %+v", packets[0])
	}
}

func TestFramerWaitsForFullPacket(t *testing.T) {
	framer := NewFramer(LengthTable{0x0073: 3})
	packets, err := framer.Push([]byte{0x73, 0x00})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 0 {
		t.Fatalf("packets = %d", len(packets))
	}
	packets, err = framer.Push([]byte{0x01})
	if err != nil {
		t.Fatal(err)
	}
	if len(packets) != 1 {
		t.Fatalf("packets = %d", len(packets))
	}
}
