package res

import "fmt"

type gr2OodleParameter struct {
	decodedValueMax uint32
	backrefValueMax uint32
	decodedCount    uint32
	highbitCount    uint32
	sizesCount      [4]byte
}

func readGR2OodleParameter(buf []byte) gr2OodleParameter {
	w0 := gr2LEU32OrZero(buf, 0)
	w1 := gr2LEU32OrZero(buf, 4)
	return gr2OodleParameter{
		decodedValueMax: w0 & 0x1ff,
		backrefValueMax: w0 >> 9,
		decodedCount:    w1 & 0x1ff,
		highbitCount:    w1 >> 9,
		sizesCount:      [4]byte{buf[8], buf[9], buf[10], buf[11]},
	}
}

type gr2OodleDictionary struct {
	decodedSize     uint32
	backrefSize     uint32
	decodedValueMax uint32
	backrefValueMax uint32
	lowbitValueMax  uint32
	lowbitWindow    gr2Window
	highbitWindow   gr2Window
	decodedWindow   gr2Window
	sizeWindows     []gr2Window
}

func newGR2OodleDictionary(p gr2OodleParameter) gr2OodleDictionary {
	backrefValueMax := p.backrefValueMax
	lowbitValueMax := min(backrefValueMax, uint32(4))

	sizeWindows := make([]gr2Window, 0, 65)
	for i := 0; i < 4; i++ {
		for j := 0; j < 16; j++ {
			sizeWindows = append(sizeWindows, newGR2Window(64, uint16(p.sizesCount[3-i])))
		}
	}
	sizeWindows = append(sizeWindows, newGR2Window(64, uint16(p.sizesCount[0])))

	return gr2OodleDictionary{
		decodedValueMax: p.decodedValueMax,
		backrefValueMax: backrefValueMax,
		lowbitValueMax:  lowbitValueMax,
		lowbitWindow:    newGR2Window(lowbitValueMax, uint16(lowbitValueMax)),
		highbitWindow:   newGR2Window(p.highbitCount, uint16(p.highbitCount)),
		decodedWindow:   newGR2Window(p.decodedValueMax, uint16(p.decodedCount)),
		sizeWindows:     sizeWindows,
	}
}

func (d *gr2OodleDictionary) decompressBlock(decoder *gr2RangeDecoder, out []byte, pos int) (int, error) {
	sizeIndex := int(d.backrefSize)
	if sizeIndex < 0 || sizeIndex >= len(d.sizeWindows) {
		return 0, fmt.Errorf("gr2 oodle size window out of range: %d", sizeIndex)
	}
	d.backrefSize = uint32(d.sizeWindows[sizeIndex].decodeSymbol(decoder, func(decoder *gr2RangeDecoder) uint16 {
		return decoder.decodeCommit(65)
	}))

	if d.backrefSize > 0 {
		sizes := [4]uint32{128, 192, 256, 512}
		backrefSize := d.backrefSize + 1
		if d.backrefSize >= 61 {
			backrefSize = sizes[d.backrefSize-61]
		}
		backrefRange := min(d.backrefValueMax, d.decodedSize)

		lowbitMax := d.lowbitValueMax
		lowValue := d.lowbitWindow.decodeSymbol(decoder, func(decoder *gr2RangeDecoder) uint16 {
			return decoder.decodeCommit(lowbitMax)
		})

		highTotal := backrefRange/4 + 1
		highValue := d.highbitWindow.decodeSymbol(decoder, func(decoder *gr2RangeDecoder) uint16 {
			return decoder.decodeCommit(highTotal)
		})

		backrefOffset := (uint32(highValue) << 2) + uint32(lowValue) + 1
		d.decodedSize += backrefSize

		if pos < int(backrefOffset) || pos+int(backrefSize) > len(out) {
			return 0, fmt.Errorf("gr2 oodle backref out of range: pos=%d off=%d size=%d", pos, backrefOffset, backrefSize)
		}
		src := pos - int(backrefOffset)
		for k := 0; k < int(backrefSize); k++ {
			out[pos+k] = out[src+k]
		}
		return int(backrefSize), nil
	}

	decodedMax := d.decodedValueMax
	value := d.decodedWindow.decodeSymbol(decoder, func(decoder *gr2RangeDecoder) uint16 {
		return decoder.decodeCommit(decodedMax)
	})
	if pos >= len(out) {
		return 0, fmt.Errorf("gr2 oodle output overflow")
	}
	out[pos] = byte(value)
	d.decodedSize++
	return 1, nil
}

func decompressGR2Oodle(compressed []byte, out []byte, stop0, stop1 uint32) error {
	if len(out) == 0 {
		return nil
	}
	if len(compressed) < 36 {
		return fmt.Errorf("gr2 oodle: short header")
	}
	params := [3]gr2OodleParameter{
		readGR2OodleParameter(compressed[0:12]),
		readGR2OodleParameter(compressed[12:24]),
		readGR2OodleParameter(compressed[24:36]),
	}
	decoder := newGR2RangeDecoder(compressed, 36)
	steps := [3]int{int(stop0), int(stop1), len(out)}
	pos := 0
	for i, step := range steps {
		dict := newGR2OodleDictionary(params[i])
		for pos < step {
			n, err := dict.decompressBlock(&decoder, out, pos)
			if err != nil {
				return err
			}
			pos += n
		}
	}
	return nil
}
