package ui

import (
	"math"

	"github.com/kivutar/goro/render"
	"github.com/kivutar/goro/res"
)

// ROBrowser's 2D SpriteRenderer subtracts half of a 35px cell from canvas Y.
const uiSpriteRendererMidCellY = 17.5

func drawUISpriteLayer(target, img *render.Image, layer res.ACTLayer, anchorX, anchorY float64) {
	bounds := img.Bounds()
	width := float64(bounds.Dx())
	height := float64(bounds.Dy())
	scaleX := float64(layer.ScaleX)
	scaleY := float64(layer.ScaleY)
	if scaleX == 0 {
		scaleX = 1
	}
	if scaleY == 0 {
		scaleY = 1
	}
	if layer.Mirror {
		scaleX = -scaleX
	}
	var opts render.DrawImageOptions
	opts.GeoM.Translate(-width/2, -height/2)
	opts.GeoM.Scale(scaleX, scaleY)
	if layer.Angle != 0 {
		opts.GeoM.Rotate(float64(-layer.Angle) * math.Pi / 180)
	}
	opts.GeoM.Translate(anchorX+float64(layer.X), anchorY+float64(layer.Y))
	opts.Filter = render.FilterNearest
	opts.ColorScale.Scale(layer.Color[0], layer.Color[1], layer.Color[2], layer.Color[3])
	target.DrawImage(img, &opts)
}
