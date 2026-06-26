package gamemode

import (
	"os"
	"strings"

	"github.com/kivutar/goro/internal/render"
)

func fsaaEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("GORO_FSAA")))
	return value != "0" && value != "false" && value != "off"
}

func triangleDrawOptions(filter render.Filter, address render.Address) *render.DrawTrianglesOptions {
	return &render.DrawTrianglesOptions{
		Filter:    filter,
		Address:   address,
		AntiAlias: fsaaEnabled(),
		DepthTest: true,
	}
}

func worldOpaqueTriangleDrawOptions(filter render.Filter, address render.Address) *render.DrawTrianglesOptions {
	options := triangleDrawOptions(filter, address)
	options.DepthWrite = true
	return options
}
