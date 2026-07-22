package res

var gr2Rev4 = [16]byte{
	0x0, 0x8, 0x4, 0xc, 0x2, 0xa, 0x6, 0xe,
	0x1, 0x9, 0x5, 0xd, 0x3, 0xb, 0x7, 0xf,
}

func gr2Reverse4(n uint32) uint32 {
	return uint32(gr2Rev4[n&0xf])
}

func gr2Reverse8(b uint32) uint32 {
	return gr2Reverse4(b)<<4 | gr2Reverse4(b>>4)
}

func gr2LEU32OrZero(data []byte, off int) uint32 {
	var b [4]byte
	for i := range b {
		if off+i < len(data) {
			b[i] = data[off+i]
		}
	}
	return uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
}

type gr2Reservoir struct {
	stream   []byte
	pos      int
	bitbuf   uint64
	bitcount uint32
}

func newGR2Reservoir(data []byte, start int) gr2Reservoir {
	return gr2Reservoir{stream: data, pos: start}
}

func (r *gr2Reservoir) pull(n uint32) uint32 {
	if n == 0 {
		return 0
	}
	if r.bitcount < n {
		dword := gr2LEU32OrZero(r.stream, r.pos)
		r.pos += 4
		r.bitbuf |= uint64(dword) << r.bitcount
		r.bitcount += 32
	}
	v := uint32(r.bitbuf & ((uint64(1) << n) - 1))
	r.bitbuf >>= n
	r.bitcount -= n
	return v
}

type gr2RangeDecoder struct {
	res       gr2Reservoir
	low, high uint32
	code      uint32
}

func newGR2RangeDecoder(data []byte, start int) gr2RangeDecoder {
	first := gr2LEU32OrZero(data, start)
	res := newGR2Reservoir(data, start+4)
	res.bitbuf = uint64(first >> 31)
	res.bitcount = 1
	return gr2RangeDecoder{
		res:  res,
		low:  0,
		high: 0x7fff_ffff,
		code: bitsReverse32(first&0x7fff_ffff) >> 1,
	}
}

func bitsReverse32(v uint32) uint32 {
	v = (v&0x55555555)<<1 | (v>>1)&0x55555555
	v = (v&0x33333333)<<2 | (v>>2)&0x33333333
	v = (v&0x0f0f0f0f)<<4 | (v>>4)&0x0f0f0f0f
	v = (v&0x00ff00ff)<<8 | (v>>8)&0x00ff00ff
	return v<<16 | v>>16
}

func (d *gr2RangeDecoder) pull(n uint32) uint32 {
	return d.res.pull(n)
}

func (d *gr2RangeDecoder) decode(total uint32) uint16 {
	rng := uint64(d.high - d.low + 1)
	v := (uint64(d.code-d.low+1)*uint64(total) - 1) / rng
	return uint16(v)
}

func (d *gr2RangeDecoder) commit(total uint32, val, width uint16) uint16 {
	rng := uint64(d.high - d.low + 1)
	total64 := uint64(total)
	d.high = d.low + uint32(rng*(uint64(val)+uint64(width))/total64) - 1
	d.low += uint32(rng * uint64(val) / total64)
	d.renormalize()
	return val
}

func (d *gr2RangeDecoder) decodeCommit(total uint32) uint16 {
	v := d.decode(total)
	return d.commit(total, v, 1)
}

func (d *gr2RangeDecoder) renormalize() {
	if (d.low^d.high)&0x4000_0000 == 0 {
		for (d.low^d.high)&0x7f80_0000 == 0 {
			b := d.pull(8)
			d.high = d.high<<8 | 0xff
			d.low <<= 8
			d.code = d.code<<8 | gr2Reverse8(b)
		}
		if (d.low^d.high)&0x7800_0000 == 0 {
			nib := d.pull(4)
			d.high = d.high<<4 | 0xf
			d.low <<= 4
			d.code = d.code<<4 | gr2Reverse4(nib)
		}
		for (d.low^d.high)&0x4000_0000 == 0 {
			bit := d.pull(1)
			d.high = d.high<<1 | 1
			d.low <<= 1
			d.code = d.code<<1 | bit
		}
	}
	for d.low&0x2000_0000 != 0 && d.high&0x2000_0000 == 0 {
		bit := d.pull(1)
		d.low = (d.low & 0x1fff_ffff) << 1
		d.high = d.high<<1 | 0x4000_0001
		d.code = ((d.code ^ 0x2000_0000) << 1) | bit
	}
	d.low &= 0x7fff_ffff
	d.high &= 0x7fff_ffff
	d.code &= 0x7fff_ffff
}

type gr2Window struct {
	total       uint16
	numValues   uint16
	stepTimes15 uint16
	shift       uint8
	step        uint16
	countCap    uint16
	values      []uint16
	weights     []uint16
}

func newGR2Window(_ uint32, countCap uint16) gr2Window {
	capacity := (int(countCap) + 5) &^ 3
	w := gr2Window{
		countCap: countCap,
		values:   make([]uint16, capacity),
		weights:  make([]uint16, capacity),
	}
	w.granularity(uint32(countCap) + 1)
	w.update(0, 3)
	return w
}

func (w *gr2Window) granularity(n uint32) {
	if n < 6 {
		w.step = 0
		w.shift = 15
		w.stepTimes15 = 0
		return
	}
	bestShift := uint32(0)
	bestSpan := ^uint32(0)
	for shift := uint32(0); shift < 0x10; shift++ {
		step := uint32(1) << shift
		buckets := (step + n - 1) / step
		if buckets > 0x10 {
			buckets = 0x10
		}
		span := n - (buckets-1)*step
		if span < step {
			span = step
		}
		if span < bestSpan {
			bestSpan = span
			bestShift = shift
		}
		if step > n {
			break
		}
	}
	w.step = uint16(1) << bestShift
	w.shift = uint8(bestShift)
	w.stepTimes15 = uint16(uint32(w.step) * 15)
}

func (w *gr2Window) update(index int, delta uint16) {
	w.weights[index] += delta
	w.total += delta
}

func (w *gr2Window) rebuild() {
	w.granularity(uint32(w.numValues) + 1)
	w.weights[0] >>= 1

	maxWeight := uint32(0)
	maxSym := 0
	if w.numValues >= 1 {
		d := 1
		for {
			for w.weights[d] <= 1 {
				if uint16(d) >= w.numValues {
					w.weights[d] = 0
					w.numValues--
					goto compacted
				}
				last := int(w.numValues)
				w.weights[d] = w.weights[last]
				w.weights[last] = 0
				w.values[d] = w.values[last]
				w.numValues--
			}
			w.weights[d] >>= 1
			if uint32(w.weights[d]) > maxWeight {
				maxWeight = uint32(w.weights[d])
				maxSym = d
			}
			d++
			if uint16(d) > w.numValues {
				break
			}
		}
	}

compacted:
	if maxWeight != 0 {
		step15 := int(w.stepTimes15)
		target := step15
		if int(w.numValues) < step15 {
			target = (int(w.numValues) >> w.shift) << w.shift
		}
		if target == 0 {
			target = 1
		}
		if maxSym != target {
			w.weights[target], w.weights[maxSym] = w.weights[maxSym], w.weights[target]
			w.values[target], w.values[maxSym] = w.values[maxSym], w.values[target]
		}
	}

	if w.numValues != w.countCap && w.weights[0] == 0 {
		w.weights[0] = 2
	}

	sum := uint16(0)
	for i := 0; i <= int(w.numValues); i++ {
		sum += w.weights[i]
	}
	w.total = sum
}

func (w *gr2Window) decodeSymbol(decoder *gr2RangeDecoder, readValue func(*gr2RangeDecoder) uint16) uint16 {
	existing, value, slot := w.tryDecode(decoder)
	if existing {
		return value
	}
	v := readValue(decoder)
	w.values[slot] = v
	return v
}

func (w *gr2Window) tryDecode(decoder *gr2RangeDecoder) (existing bool, value uint16, slot int) {
	if w.total >= 0x4000 {
		w.rebuild()
	}

	total := uint32(w.total)
	freq := decoder.decode(total)
	cum := uint16(0)
	index := 0
	for {
		next := cum + w.weights[index]
		if freq < next {
			break
		}
		cum = next
		index++
	}

	weight := w.weights[index]
	decoder.commit(total, cum, weight)
	w.weights[index]++
	w.total++

	if index != 0 {
		return true, w.values[index], 0
	}

	w.numValues++
	newSlot := int(w.numValues)
	w.values[newSlot] = 0
	w.update(newSlot, 2)
	if w.numValues == w.countCap {
		w0 := w.weights[0]
		w.update(0, -w0)
	}
	return false, 0, newSlot
}
