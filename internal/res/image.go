package res

import (
	"bytes"
	"fmt"
	"image"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/bmp"
)

func LoadImage(manager *Manager, candidates []string) (image.Image, string, error) {
	for _, candidate := range candidates {
		data, err := manager.ReadFile(candidate)
		if err != nil {
			continue
		}
		img, _, err := image.Decode(bytes.NewReader(data))
		if err != nil {
			return nil, candidate, fmt.Errorf("decode image: %w", err)
		}
		return applyROTransparency(img), candidate, nil
	}
	return nil, "", fmt.Errorf("image not found: %s", strings.Join(candidates, ", "))
}

func applyROTransparency(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	out := image.NewNRGBA(bounds)
	draw.Draw(out, bounds, img, bounds.Min, draw.Src)
	for i := 0; i < len(out.Pix); i += 4 {
		if isROMagenta(out.Pix[i], out.Pix[i+1], out.Pix[i+2]) {
			out.Pix[i+3] = 0
		}
	}
	return out
}

func isROMagenta(r, g, b uint8) bool {
	return r >= 248 && g <= 8 && b >= 248
}

func GroundTextureCandidates(name string) []string {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	return []string{
		"data\\texture\\" + name,
		"data/texture/" + strings.ReplaceAll(name, "\\", "/"),
		"texture\\" + name,
		"texture/" + strings.ReplaceAll(name, "\\", "/"),
		name,
		strings.ReplaceAll(name, "\\", "/"),
	}
}
