package game

import "github.com/kivutar/goro/render"

const groundDecalDepthBias = 1.0 / 32768.0
const rsmModelDepthBias = groundDecalDepthBias
const rsmHorizontalDepthBiasLayers = 64
const rsmHorizontalDepthBiasStep = groundDecalDepthBias / 256

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

func rsmHorizontalSurfaceDrawOptions(filter render.Filter, address render.Address, placementIndex int) *render.DrawTrianglesOptions {
	options := rsmModelDrawOptions(filter, address)
	if placementIndex < 0 {
		return options
	}
	layer := placementIndex % rsmHorizontalDepthBiasLayers
	options.DepthBias += float32(layer) * rsmHorizontalDepthBiasStep
	return options
}
