package game

import "github.com/kivutar/goro/render"

const groundDecalDepthBias = 1.0 / 32768.0

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

func groundTextureDrawOptions() *render.DrawTrianglesOptions {
	return worldOpaqueTriangleDrawOptions(render.FilterLinear, render.AddressClampToZero)
}
