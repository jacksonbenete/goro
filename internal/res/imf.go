package res

import (
	"encoding/binary"
	"fmt"
)

type IMF struct {
	Version  int32
	Checksum int32
	Layers   []IMFLayer
}

type IMFLayer struct {
	Actions []IMFAction
}

type IMFAction struct {
	Motions []IMFMotion
}

type IMFMotion struct {
	Priority int32
	X        int32
	Y        int32
}

func ParseIMF(data []byte) (*IMF, error) {
	reader := imfReader{data: data}
	imf := &IMF{
		Version:  reader.i32(),
		Checksum: reader.i32(),
	}
	layerCount := int(reader.i32())
	if layerCount < 0 || layerCount > 15 {
		return nil, fmt.Errorf("imf invalid layer count %d", layerCount)
	}
	imf.Layers = make([]IMFLayer, layerCount)
	for layerIndex := range imf.Layers {
		actionCount := int(reader.i32())
		if actionCount < 0 {
			return nil, fmt.Errorf("imf negative action count")
		}
		actions := make([]IMFAction, actionCount)
		for actionIndex := range actions {
			motionCount := int(reader.i32())
			if motionCount < 0 {
				return nil, fmt.Errorf("imf negative motion count")
			}
			motions := make([]IMFMotion, motionCount)
			for motionIndex := range motions {
				motions[motionIndex] = IMFMotion{
					Priority: reader.i32(),
					X:        reader.i32(),
					Y:        reader.i32(),
				}
			}
			actions[actionIndex] = IMFAction{Motions: motions}
		}
		imf.Layers[layerIndex] = IMFLayer{Actions: actions}
	}
	if reader.err != nil {
		return nil, reader.err
	}
	return imf, nil
}

func (i *IMF) LayerForPriority(priority, action, motion int) int {
	if i == nil || priority < 0 || priority >= 15 {
		return -1
	}
	if action < 0 || motion < 0 {
		return -1
	}
	for layerIndex := range i.Layers {
		if i.Priority(layerIndex, action, motion) == int32(priority) {
			return layerIndex
		}
	}
	return -1
}

func (i *IMF) Priority(layer, action, motion int) int32 {
	if i == nil || layer < 0 || layer >= len(i.Layers) {
		return 0
	}
	layerData := i.Layers[layer]
	if action < 0 || action >= len(layerData.Actions) {
		return 0
	}
	actionData := layerData.Actions[action]
	if motion < 0 || motion >= len(actionData.Motions) {
		return 0
	}
	return actionData.Motions[motion].Priority
}

func (i *IMF) Point(layer, action, motion int) (int32, int32) {
	if i == nil || layer < 0 || layer >= len(i.Layers) {
		return 0, 0
	}
	layerData := i.Layers[layer]
	if action < 0 || action >= len(layerData.Actions) {
		return 0, 0
	}
	actionData := layerData.Actions[action]
	if motion < 0 || motion >= len(actionData.Motions) {
		return 0, 0
	}
	point := actionData.Motions[motion]
	return point.X, point.Y
}

type imfReader struct {
	data   []byte
	offset int
	err    error
}

func (r *imfReader) i32() int32 {
	if r.err != nil {
		return 0
	}
	if r.offset+4 > len(r.data) {
		r.err = fmt.Errorf("imf truncated at offset %d", r.offset)
		return 0
	}
	value := int32(binary.LittleEndian.Uint32(r.data[r.offset:]))
	r.offset += 4
	return value
}
