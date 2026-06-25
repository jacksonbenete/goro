package gamemode

import (
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
)

func fsaaEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("GORO_FSAA")))
	return value != "0" && value != "false" && value != "off"
}

func triangleDrawOptions(filter ebiten.Filter, address ebiten.Address) *ebiten.DrawTrianglesOptions {
	return &ebiten.DrawTrianglesOptions{
		Filter:    filter,
		Address:   address,
		AntiAlias: fsaaEnabled(),
	}
}
