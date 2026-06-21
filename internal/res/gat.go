package res

import (
	"encoding/binary"
	"fmt"
	"math"
)

const (
	GATTypeNone     uint32 = 1 << 0
	GATTypeWalkable uint32 = 1 << 1
	GATTypeWater    uint32 = 1 << 2
	GATTypeSnipable uint32 = 1 << 3
)

type GAT struct {
	VersionMajor byte
	VersionMinor byte
	Width        int
	Height       int
	Cells        []GATCell
}

type GATCell struct {
	Heights [4]float32
	RawType uint32
	Type    uint32
}

func ParseGAT(data []byte) (*GAT, error) {
	if len(data) < 14 {
		return nil, fmt.Errorf("gat too short: %d", len(data))
	}
	if string(data[:4]) != "GRAT" {
		return nil, fmt.Errorf("invalid gat signature %q", string(data[:4]))
	}

	width := int(binary.LittleEndian.Uint32(data[6:10]))
	height := int(binary.LittleEndian.Uint32(data[10:14]))
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("invalid gat size %dx%d", width, height)
	}

	cellCount := width * height
	expected := 14 + cellCount*20
	if len(data) < expected {
		return nil, fmt.Errorf("gat truncated: have %d want %d", len(data), expected)
	}

	gat := &GAT{
		VersionMajor: data[4],
		VersionMinor: data[5],
		Width:        width,
		Height:       height,
		Cells:        make([]GATCell, cellCount),
	}

	offset := 14
	for i := 0; i < cellCount; i++ {
		cell := GATCell{}
		for h := 0; h < 4; h++ {
			cell.Heights[h] = math.Float32frombits(binary.LittleEndian.Uint32(data[offset:offset+4])) * 0.2
			offset += 4
		}
		cell.RawType = binary.LittleEndian.Uint32(data[offset : offset+4])
		cell.Type = GATCellType(cell.RawType)
		offset += 4
		gat.Cells[i] = cell
	}

	return gat, nil
}

func (g *GAT) InBounds(x, y int) bool {
	return g != nil && x >= 0 && y >= 0 && x < g.Width && y < g.Height
}

func (g *GAT) Cell(x, y int) (GATCell, bool) {
	if !g.InBounds(x, y) {
		return GATCell{}, false
	}
	return g.Cells[y*g.Width+x], true
}

func (g *GAT) Walkable(x, y int) bool {
	cell, ok := g.Cell(x, y)
	return ok && cell.Type&GATTypeWalkable != 0
}

func GATCellType(raw uint32) uint32 {
	switch raw {
	case 0, 2, 4, 6, 0x80000000, 0x80000002, 0x80000005, 0x80000006:
		return GATTypeWalkable | GATTypeSnipable
	case 3, 0x80000003:
		return GATTypeWalkable | GATTypeSnipable | GATTypeWater
	case 5:
		return GATTypeSnipable
	default:
		return GATTypeNone
	}
}
