package res

import "fmt"

type Palette [256][4]byte

func ParsePAL(data []byte) (Palette, error) {
	var palette Palette
	if len(data) < 1024 {
		return palette, fmt.Errorf("pal truncated: %d bytes", len(data))
	}
	for i := 0; i < 256; i++ {
		copy(palette[i][:], data[i*4:i*4+4])
	}
	return palette, nil
}
