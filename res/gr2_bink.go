package res

import "fmt"

var gr2Run6Escape = [4]uint32{0x80, 0x100, 0x200, 0x400}
var gr2Run8Escape = [4]uint32{0x200, 0x400, 0x800, 0xc00}

func gr2BinkRound16(v int32) int16 {
	bias := (v >> 31) ^ 0x7fff
	a := bias + v
	if a < 0 {
		a += 0xffff
	}
	return int16(a >> 16)
}

type gr2BinkTap struct {
	detail bool
	rel    int
	coeff  int32
}

var gr2PassAEven = []gr2BinkTap{
	{rel: -1, coeff: -2667}, {rel: 0, coeff: 51674}, {rel: 1, coeff: -2667},
	{detail: true, rel: -2, coeff: -1563}, {detail: true, rel: -1, coeff: 24733}, {detail: true, rel: 0, coeff: 24733}, {detail: true, rel: 1, coeff: -1563},
}

var gr2PassAOdd = []gr2BinkTap{
	{rel: -1, coeff: -4230}, {rel: 0, coeff: 27400}, {rel: 1, coeff: 27400}, {rel: 2, coeff: -4230},
	{detail: true, rel: -2, coeff: -2479}, {detail: true, rel: -1, coeff: 7250}, {detail: true, rel: 0, coeff: -55882}, {detail: true, rel: 1, coeff: 7250}, {detail: true, rel: 2, coeff: -2479},
}

var gr2PassALeft = [][]gr2BinkTap{
	{{rel: 0, coeff: 51674}, {rel: 1, coeff: -5334}, {detail: true, rel: 0, coeff: 49466}, {detail: true, rel: 1, coeff: -3126}},
	{{rel: 0, coeff: 27400}, {rel: 1, coeff: 23170}, {rel: 2, coeff: -4230}, {detail: true, rel: 0, coeff: -48632}, {detail: true, rel: 1, coeff: 4771}, {detail: true, rel: 2, coeff: -2479}},
	{{rel: 0, coeff: -2667}, {rel: 1, coeff: 51674}, {rel: 2, coeff: -2667}, {detail: true, rel: 0, coeff: 23170}, {detail: true, rel: 1, coeff: 24733}, {detail: true, rel: 2, coeff: -1563}},
	{{rel: 0, coeff: -4230}, {rel: 1, coeff: 27400}, {rel: 2, coeff: 27400}, {rel: 3, coeff: -4230}, {detail: true, rel: 0, coeff: 4771}, {detail: true, rel: 1, coeff: -55882}, {detail: true, rel: 2, coeff: 7250}, {detail: true, rel: 3, coeff: -2479}},
}

var gr2PassARight = [][]gr2BinkTap{
	{{rel: -3, coeff: -2667}, {rel: -2, coeff: 51674}, {rel: -1, coeff: -2667}, {detail: true, rel: -4, coeff: -1563}, {detail: true, rel: -3, coeff: 24733}, {detail: true, rel: -2, coeff: 24733}, {detail: true, rel: -1, coeff: -1563}},
	{{rel: -3, coeff: -4230}, {rel: -2, coeff: 27400}, {rel: -1, coeff: 23170}, {detail: true, rel: -4, coeff: -2479}, {detail: true, rel: -3, coeff: 7250}, {detail: true, rel: -2, coeff: -58361}, {detail: true, rel: -1, coeff: 7250}},
	{{rel: -2, coeff: -2667}, {rel: -1, coeff: 49007}, {detail: true, rel: -3, coeff: -1563}, {detail: true, rel: -2, coeff: 23170}, {detail: true, rel: -1, coeff: 24733}},
	{{rel: -2, coeff: -8460}, {rel: -1, coeff: 54800}, {detail: true, rel: -3, coeff: -4958}, {detail: true, rel: -2, coeff: 14500}, {detail: true, rel: -1, coeff: -55882}},
}

func gr2BinkPlaneCount(hasAlpha bool) int {
	if hasAlpha {
		return 4
	}
	return 3
}

func gr2BinkMagClass(x uint32) int {
	if x == 0 {
		return 0
	}
	c := 32 - leadingZeros32(x)
	if c > 15 {
		return 15
	}
	return c
}

func leadingZeros32(x uint32) int {
	if x == 0 {
		return 32
	}
	n := 0
	for bit := uint32(1 << 31); x&bit == 0; bit >>= 1 {
		n++
	}
	return n
}

type gr2BinkBand struct {
	plane  []int16
	base   int
	stride int
	w      int
	h      int
}

func (b gr2BinkBand) fill(v int16) {
	for r := 0; r < b.h; r++ {
		start := b.base + r*b.stride
		for i := 0; i < b.w; i++ {
			b.plane[start+i] = v
		}
	}
}

type gr2BinkBandModels struct {
	levels []gr2Window
	run6   gr2Window
	run8   gr2Window
}

func setupGR2BinkModels(token uint32, classes int) gr2BinkBandModels {
	levels := make([]gr2Window, classes)
	for i := range levels {
		levels[i] = newGR2Window(token, uint16(token+1))
	}
	return gr2BinkBandModels{
		levels: levels,
		run6:   newGR2Window(0x3f, 0x40),
		run8:   newGR2Window(0xff, 0x100),
	}
}

func gr2BinkDCBand(dec *gr2RangeDecoder, rd *gr2Reservoir, band gr2BinkBand) {
	if rd.pull(1) != 0 {
		band.fill(int16(rd.pull(16)))
		return
	}
	maxValue := rd.pull(16)
	total := maxValue + 1
	model := newGR2Window(maxValue, uint16(total))
	delta := func() int32 {
		v := int32(model.decodeSymbol(dec, func(decoder *gr2RangeDecoder) uint16 {
			return decoder.decodeCommit(total)
		}))
		if v != 0 && rd.pull(1) != 0 {
			v = -v
		}
		return v
	}

	dst := band.base
	left := int32(rd.pull(16))
	band.plane[dst] = int16(left)
	dst++
	for i := 1; i < band.w; i++ {
		left += delta()
		band.plane[dst] = int16(left)
		dst++
	}
	rowGap := band.stride - band.w
	for row := 1; row < band.h; row++ {
		dst += rowGap
		above := dst - band.stride
		left = int32(band.plane[above]) + delta()
		band.plane[dst] = int16(left)
		dst++
		above++
		for col := 1; col < band.w; col++ {
			s := int32(band.plane[above]) + left
			pred := s
			if pred < 0 {
				pred++
			}
			pred >>= 1
			left = pred + delta()
			band.plane[dst] = int16(left)
			dst++
			above++
		}
	}
}

func gr2BinkACBand(dec *gr2RangeDecoder, rd *gr2Reservoir, band gr2BinkBand) error {
	scale := int32(rd.pull(16))
	if rd.pull(1) != 0 {
		v := int16(int32(rd.pull(16)) * scale)
		band.fill(v)
		return nil
	}
	token := rd.pull(16)
	escTotal := token + 1
	classes := gr2BinkMagClass(token*uint32(scale)) + 1
	models := setupGR2BinkModels(token, classes)

	v := int32(dec.decodeCommit(escTotal))
	if v != 0 {
		if rd.pull(1) != 0 {
			v = -v
		}
		v *= scale
	}
	band.plane[band.base] = int16(v)
	dst := band.base + 1
	above := band.base
	aboveVal := v
	aa := v
	left := v
	rows := band.h
	rowGap := band.stride - band.w
	cols := 0
	if band.w != 1 {
		cols = band.w - 1
	}
	r1 := uint32(0)
	r2 := uint32(0)
	budget := band.w*band.h*4 + 4096

	for {
		if r1 == 0 {
			if r2 == 0 {
				budget--
				if budget < 0 {
					return fmt.Errorf("gr2 bink: band overrun")
				}
				t1 := uint32(models.run6.decodeSymbol(dec, func(_ *gr2RangeDecoder) uint16 {
					return uint16(rd.pull(6))
				}))
				if t1 >= 0x3c {
					r1 = gr2Run6Escape[t1-0x3c]
				} else {
					r1 = t1
				}
				t2 := uint32(models.run8.decodeSymbol(dec, func(_ *gr2RangeDecoder) uint16 {
					return uint16(rd.pull(8))
				}))
				if t2 >= 0xfc {
					r2 = gr2Run8Escape[t2-0xfc] + 2
				} else if t2 != 0 {
					r2 = t2 + 2
				} else {
					r2 = 0
				}
				continue
			}
			if int(r2) < cols {
				cols -= int(r2)
				for i := uint32(0); i < r2; i++ {
					band.plane[dst] = 0
					dst++
				}
				above += int(r2)
				aa = int32(band.plane[above-2])
				aboveVal = int32(band.plane[above-1])
				left = 0
				r2 = 0
			} else {
				r2 -= uint32(cols)
				for i := 0; i < cols; i++ {
					band.plane[dst] = 0
					dst++
				}
				rows--
				if rows == 0 {
					return nil
				}
				dst += rowGap
				above = dst - band.stride
				aboveVal = int32(band.plane[above])
				above++
				aa = aboveVal
				left = aboveVal
				cols = band.w
			}
			continue
		}
		if cols == 0 {
			rows--
			if rows == 0 {
				return nil
			}
			dst += rowGap
			above = dst - band.stride
			aboveVal = int32(band.plane[above])
			above++
			aa = aboveVal
			left = aboveVal
			cols = band.w
			continue
		}
		lastCol := cols == 1
		var pred uint32
		if lastCol {
			pred = (uint32(abs32(left*2)) + uint32(abs32(aa)) + uint32(abs32(aboveVal))) >> 2
		} else {
			pred = (uint32(abs32(int32(band.plane[above]))) + uint32(abs32(aa)) + uint32(abs32(aboveVal)) + uint32(abs32(left))) >> 2
		}
		cls := gr2BinkMagClass(pred)
		lvl := int32(models.levels[cls].decodeSymbol(dec, func(decoder *gr2RangeDecoder) uint16 {
			return decoder.decodeCommit(escTotal)
		}))
		if lvl != 0 {
			if rd.pull(1) != 0 {
				lvl = -lvl
			}
			lvl *= scale
		}
		band.plane[dst] = int16(lvl)
		dst++
		r1--
		if lastCol {
			rows--
			if rows == 0 {
				return nil
			}
			dst += rowGap
			above = dst - band.stride
			aboveVal = int32(band.plane[above])
			above++
			aa = aboveVal
			left = aboveVal
			cols = band.w
		} else {
			aa = aboveVal
			aboveVal = int32(band.plane[above])
			above++
			left = lvl
			cols--
		}
	}
}

func abs32(v int32) int32 {
	if v < 0 {
		return -v
	}
	return v
}

func decodeGR2BinkPlane(src []byte, plane []int16, w, h int, rowFlags []byte) (int, error) {
	if len(src) < 8 {
		return 0, fmt.Errorf("gr2 bink: short plane")
	}
	lenA := int(gr2LEU32OrZero(src, 0))
	lenB := int(gr2LEU32OrZero(src, 4))
	end := 8 + lenA + lenB
	if len(src) < end {
		return 0, fmt.Errorf("gr2 bink: plane truncated")
	}
	dec := newGR2RangeDecoder(src, 8)
	rd := newGR2Reservoir(src, 8+lenA)

	gr2BinkDCBand(&dec, &rd, gr2BinkBand{plane: plane, base: 0, stride: w << 4, w: w >> 4, h: h >> 4})
	for sh := 4; sh >= 1; sh-- {
		stride := w << sh
		bw, bh := w>>sh, h>>sh
		if err := gr2BinkACBand(&dec, &rd, gr2BinkBand{plane: plane, base: w >> sh, stride: stride, w: bw, h: bh}); err != nil {
			return 0, err
		}
		if err := gr2BinkACBand(&dec, &rd, gr2BinkBand{plane: plane, base: w << (sh - 1), stride: stride, w: bw, h: bh}); err != nil {
			return 0, err
		}
		if err := gr2BinkACBand(&dec, &rd, gr2BinkBand{plane: plane, base: (w >> sh) + (w << (sh - 1)), stride: stride, w: bw, h: bh}); err != nil {
			return 0, err
		}
	}

	count := uint32(h)
	threshold := dec.decodeCommit(count + 1)
	for i := 0; i < h; i++ {
		t := dec.decode(count)
		if t < threshold {
			rowFlags[i] = 0
			dec.commit(count, 0, threshold)
		} else {
			rowFlags[i] = 1
			dec.commit(count, threshold, uint16(count)-threshold)
		}
	}
	return end, nil
}

type gr2BinkDecodedPlanes struct {
	planes   [][]int16
	rowFlags [][]byte
	w, h     int
}

func decodeGR2BinkPlanes(pixels []byte, width, height int, hasAlpha bool) (gr2BinkDecodedPlanes, error) {
	if width*height <= 0x100 {
		return gr2BinkDecodedPlanes{}, fmt.Errorf("gr2 bink tiny-texture path is raw")
	}
	w := (width + 15) &^ 15
	h := (height + 15) &^ 15
	n := gr2BinkPlaneCount(hasAlpha)
	decoded := gr2BinkDecodedPlanes{
		planes:   make([][]int16, 0, n),
		rowFlags: make([][]byte, 0, n),
		w:        w,
		h:        h,
	}
	off := 4
	for i := 0; i < n; i++ {
		plane := make([]int16, w*h)
		flags := make([]byte, h)
		consumed, err := decodeGR2BinkPlane(pixels[off:], plane, w, h, flags)
		if err != nil {
			return gr2BinkDecodedPlanes{}, err
		}
		off += consumed
		decoded.planes = append(decoded.planes, plane)
		decoded.rowFlags = append(decoded.rowFlags, flags)
	}
	return decoded, nil
}

func gr2BinkColorTransformRGBA(planes [][]int16, width, height int) []byte {
	hasAlpha := len(planes) >= 4
	out := make([]byte, width*height*4)
	for i := 0; i < width*height; i++ {
		p0 := int32(planes[0][i])
		p1 := int32(planes[1][i])
		p2 := int32(planes[2][i])
		s := p1 + p2
		delta := s
		if delta < 0 {
			delta += 3
		}
		delta >>= 2
		y := p0 - delta
		dst := i * 4
		out[dst+0] = clampByteInt32(p1 + y)
		out[dst+1] = clampByteInt32(y)
		out[dst+2] = clampByteInt32(p2 + y)
		if hasAlpha {
			out[dst+3] = clampByteInt32(int32(planes[3][i]))
		} else {
			out[dst+3] = 255
		}
	}
	return out
}

func clampByteInt32(v int32) byte {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return byte(v)
}

func gr2BinkPassARow(coarse, detail []int16, out []int16) {
	w := len(out)
	half := w / 2
	band := func(t gr2BinkTap, k int) int32 {
		arr := coarse
		if t.detail {
			arr = detail
		}
		i := k + t.rel
		if i >= 0 && i < half {
			return int32(arr[i])
		}
		return 0
	}
	for o := 0; o < w; o++ {
		acc := int32(0)
		switch {
		case o < 4:
			for _, t := range gr2PassALeft[o] {
				arr := coarse
				if t.detail {
					arr = detail
				}
				acc += t.coeff * int32(arr[t.rel])
			}
		case o >= w-4:
			for _, t := range gr2PassARight[o-(w-4)] {
				arr := coarse
				if t.detail {
					arr = detail
				}
				idx := half + t.rel
				acc += t.coeff * int32(arr[idx])
			}
		default:
			k := o / 2
			stencil := gr2PassAEven
			if o%2 != 0 {
				stencil = gr2PassAOdd
			}
			for _, t := range stencil {
				acc += t.coeff * band(t, k)
			}
		}
		out[o] = gr2BinkRound16(acc)
	}
}

func gr2MirrorCoarse(i int, n int) int {
	if i < 0 {
		return -i
	}
	if i >= n {
		return 2*n - 1 - i
	}
	return i
}

func gr2MirrorDetail(i int, n int) int {
	if i < 0 {
		return -i - 1
	}
	if i >= n {
		return 2*n - 2 - i
	}
	return i
}

func gr2BinkPassBColumn(scratch []int16, ow, j, n int, plane []int16, pitch int) {
	c := func(i int) int32 {
		return int32(scratch[2*gr2MirrorCoarse(i, n)*ow+j])
	}
	d := func(i int) int32 {
		return int32(scratch[(2*gr2MirrorDetail(i, n)+1)*ow+j])
	}
	for k := 0; k < n; k++ {
		even := c(k)*51674 -
			(c(k-1)+c(k+1))*2667 +
			(d(k-1)+d(k))*24733 -
			(d(k-2)+d(k+1))*1563
		odd := (c(k)+c(k+1))*27400 -
			(c(k-1)+c(k+2))*4230 +
			(d(k-1)+d(k+1))*7250 -
			(d(k-2)+d(k+2))*2479 -
			d(k)*55882
		o := 2*k*pitch + j
		plane[o] = gr2BinkRound16(even)
		plane[o+pitch] = gr2BinkRound16(odd)
	}
}

func gr2BinkHaarRound(v int32) int16 {
	t := v + ((v >> 31) ^ 1)
	t -= t >> 31
	return int16(t >> 1)
}

func gr2BinkPassARowHaar(coarse, detail []int16, out []int16, flag0 bool) {
	for k, c16 := range coarse {
		c := int32(c16)
		if flag0 {
			out[2*k] = int16(c)
			out[2*k+1] = int16(c)
			continue
		}
		d := int32(detail[k])
		out[2*k] = gr2BinkHaarRound(2*c + d)
		out[2*k+1] = gr2BinkHaarRound(2*c - d)
	}
}

func gr2BinkPassBColumnHaar(scratch []int16, ow, j, n int, plane []int16, pitch int) {
	for k := 0; k < n; k++ {
		c := int32(scratch[2*k*ow+j])
		d := int32(scratch[(2*k+1)*ow+j])
		o := 2*k*pitch + j
		plane[o] = gr2BinkHaarRound(2*c + d)
		plane[o+pitch] = gr2BinkHaarRound(2*c - d)
	}
}

func synthesizeGR2BinkStage(plane []int16, w, h, ow, oh int, flags []byte) {
	pitch := (h / oh) * w
	half := ow / 2
	aFull := ow >= 12
	bFull := oh >= 10
	zeros := make([]int16, half)
	scratch := make([]int16, ow*oh)
	for r := 0; r < oh; r++ {
		src := plane[r*pitch : r*pitch+ow]
		out := scratch[r*ow : r*ow+ow]
		flag0 := flags != nil && flags[r] == 0
		if aFull {
			detail := src[half:]
			if flag0 {
				detail = zeros
			}
			gr2BinkPassARow(src[:half], detail, out)
		} else {
			gr2BinkPassARowHaar(src[:half], src[half:], out, flag0)
		}
	}
	for j := 0; j < ow; j++ {
		if bFull {
			gr2BinkPassBColumn(scratch, ow, j, oh/2, plane, pitch)
		} else {
			gr2BinkPassBColumnHaar(scratch, ow, j, oh/2, plane, pitch)
		}
	}
}

func synthesizeGR2BinkPlane(plane []int16, w, h int, flags []byte) {
	ow, oh := w/8, h/8
	for ow <= w {
		var stageFlags []byte
		if ow == w {
			stageFlags = flags
		}
		synthesizeGR2BinkStage(plane, w, h, ow, oh, stageFlags)
		ow *= 2
		oh *= 2
	}
}

func decodeGR2Bink(pixels []byte, width, height int, hasAlpha bool) ([]byte, error) {
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("gr2 bink: invalid texture size %dx%d", width, height)
	}
	if width*height <= 0x100 {
		expected := width * height * 4
		if len(pixels) < expected {
			return nil, fmt.Errorf("gr2 bink: raw tiny texture truncated")
		}
		return append([]byte(nil), pixels[:expected]...), nil
	}
	decoded, err := decodeGR2BinkPlanes(pixels, width, height, hasAlpha)
	if err != nil {
		return nil, err
	}
	for i := range decoded.planes {
		synthesizeGR2BinkPlane(decoded.planes[i], decoded.w, decoded.h, decoded.rowFlags[i])
	}
	rgba := gr2BinkColorTransformRGBA(decoded.planes, decoded.w, decoded.h)
	if decoded.w == width && decoded.h == height {
		return rgba, nil
	}
	out := make([]byte, 0, width*height*4)
	for row := 0; row < height; row++ {
		start := row * decoded.w * 4
		out = append(out, rgba[start:start+width*4]...)
	}
	return out, nil
}
