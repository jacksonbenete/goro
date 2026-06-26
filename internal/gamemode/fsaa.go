package gamemode

import "github.com/kivutar/goro/internal/render"

func triangleDrawOptions(filter render.Filter, address render.Address) *render.DrawTrianglesOptions {
	return &render.DrawTrianglesOptions{
		Filter:    filter,
		Address:   address,
		DepthTest: true,
	}
}

func worldOpaqueTriangleDrawOptions(filter render.Filter, address render.Address) *render.DrawTrianglesOptions {
	options := triangleDrawOptions(filter, address)
	options.DepthWrite = true
	return options
}
