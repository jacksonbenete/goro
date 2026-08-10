// Package appicon provides Goro's embedded application icon.
package appicon

import (
	"bytes"
	_ "embed"
	"image"
	"image/png"
)

//go:generate go run generate.go

//go:embed icon.png
var iconPNG []byte

var windowIcon = mustDecode(iconPNG)

// Image returns the application icon used by the native window backend.
func Image() image.Image {
	return windowIcon
}

func mustDecode(data []byte) image.Image {
	img, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		panic("decode embedded application icon: " + err.Error())
	}
	return img
}
