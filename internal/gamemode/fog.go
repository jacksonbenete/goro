package gamemode

import (
	"image/color"
	"math"
	"os"

	"github.com/kivutar/goro/internal/res"
)

type sceneFog struct {
	enabled bool
	near    float64
	far     float64
	color   color.RGBA
	factor  float64
}

func sceneFogFromMap(manager *res.Manager, mapName string) sceneFog {
	if manager == nil || os.Getenv("GORO_FOG") == "0" {
		return sceneFog{}
	}
	parameter, ok := manager.FogParameter(mapName)
	if !ok || parameter.Far <= parameter.Near {
		return sceneFog{}
	}
	return sceneFog{
		enabled: true,
		near:    parameter.Near * 240,
		far:     parameter.Far * 240,
		color:   parameter.Color,
		factor:  parameter.Factor,
	}
}

func (f sceneFog) mixColor(c color.RGBA, depth float64) color.RGBA {
	if !f.enabled || !isFinite(depth) {
		return c
	}
	amount := smoothstep(f.near, f.far, depth)
	if amount <= 0 {
		return c
	}
	if amount >= 1 {
		return color.RGBA{R: f.color.R, G: f.color.G, B: f.color.B, A: c.A}
	}
	inverse := 1 - amount
	return color.RGBA{
		R: clampColor(float64(c.R)*inverse + float64(f.color.R)*amount),
		G: clampColor(float64(c.G)*inverse + float64(f.color.G)*amount),
		B: clampColor(float64(c.B)*inverse + float64(f.color.B)*amount),
		A: c.A,
	}
}

func (f sceneFog) mixVertexTints(projection sceneProjection, verts [4]modelPoint3, tints [4]color.RGBA) [4]color.RGBA {
	if !f.enabled {
		return tints
	}
	for i, vert := range verts {
		tints[i] = f.mixColor(tints[i], projection.FogDepth(vert.x, vert.z, vert.y))
	}
	return tints
}

func smoothstep(edge0, edge1, x float64) float64 {
	if edge0 == edge1 {
		if x < edge0 {
			return 0
		}
		return 1
	}
	t := (x - edge0) / (edge1 - edge0)
	t = math.Max(0, math.Min(1, t))
	return t * t * (3 - 2*t)
}
