package network

import (
	"encoding/binary"
	"fmt"
)

type Framer struct {
	lengths LengthTable
	buffer  []byte
}

func NewFramer(lengths LengthTable) *Framer {
	return &Framer{lengths: lengths}
}

func (f *Framer) Push(data []byte) ([]Packet, error) {
	f.buffer = append(f.buffer, data...)

	var packets []Packet
	for {
		if len(f.buffer) < 2 {
			return packets, nil
		}

		id := ID(f.buffer)
		length, ok := f.lengths[id]
		if !ok {
			if f.resyncToKnownPacket() {
				continue
			}
			f.buffer = f.buffer[1:]
			return packets, fmt.Errorf("unknown packet id 0x%04X", id)
		}

		if length == -1 {
			if len(f.buffer) < 4 {
				return packets, nil
			}
			length = int(binary.LittleEndian.Uint16(f.buffer[2:4]))
			if length < 4 {
				f.buffer = f.buffer[2:]
				return packets, fmt.Errorf("invalid variable packet 0x%04X length %d", id, length)
			}
		}

		if len(f.buffer) < length {
			return packets, nil
		}

		payload := make([]byte, length)
		copy(payload, f.buffer[:length])
		f.buffer = f.buffer[length:]
		packets = append(packets, Packet{ID: id, Data: payload})
	}
}

func (f *Framer) resyncToKnownPacket() bool {
	for offset := 1; offset+1 < len(f.buffer); offset++ {
		id := ID(f.buffer[offset:])
		length, ok := f.lengths[id]
		if !ok {
			continue
		}
		if length == -1 {
			if len(f.buffer)-offset < 4 {
				f.buffer = f.buffer[offset:]
				return true
			}
			packetLen := int(binary.LittleEndian.Uint16(f.buffer[offset+2 : offset+4]))
			if packetLen < 4 {
				continue
			}
			f.buffer = f.buffer[offset:]
			return true
		}
		if length > 0 {
			f.buffer = f.buffer[offset:]
			return true
		}
	}
	return false
}
