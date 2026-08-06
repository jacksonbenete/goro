package game

import "github.com/kivutar/goro/render"

const groundDecalDepthBias = 1.0 / 32768.0

// Some map RSMs are decorative floor slabs that intersect the GND by a few
// hundredths of a world unit. Keep the nudge small, but stronger than decals.
const rsmModelDepthBias = groundDecalDepthBias * 4

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

func rsmModelDrawOptions(filter render.Filter, address render.Address) *render.DrawTrianglesOptions {
	options := worldOpaqueTriangleDrawOptions(filter, address)
	options.DepthBias = rsmModelDepthBias
	return options
}
