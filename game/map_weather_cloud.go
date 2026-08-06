package game

import (
	"image/color"
	"math"
	"time"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/render"
	worldstate "github.com/kivutar/goro/world"
)

const (
	weatherCloudFrameRate       = 60
	weatherCloudMaxStep         = 250 * time.Millisecond
	weatherCloudClassicUnit     = 0.1
	weatherCloudClassicAlphaMax = 170.0 / 255.0
)

type mapWeatherCloudParams struct {
	effectID     int
	textureFiles []string
	tint         color.RGBA
	count        int
	radius       float64
	zOffset      float64
	zRand        float64
	sizeBase     float64
	sizeRand     float64
	driftSpeed   float64
	ramp         time.Duration
	fadeOut      time.Duration
	rotStartMin  time.Duration
	rotStartRand time.Duration
	overlay      bool
	additive     bool
	blackKey     bool
	disableFog   bool
	screenHaze   color.RGBA
}

type mapWeatherCloudState struct {
	key        string
	effectID   int
	clouds     []mapWeatherCloud
	lastUpdate time.Time
}

type mapWeatherCloud struct {
	x            float64
	y            float64
	z            float64
	size         float64
	age          time.Duration
	rotStart     time.Duration
	phaseX       float64
	phaseY       float64
	phaseRateX   float64
	phaseRateY   float64
	breath       float64
	textureIndex int
	generation   int
}

func weatherCloudParamsForEffect(effectID int) (mapWeatherCloudParams, bool) {
	switch effectID {
	case effectCloud4:
		return mapWeatherCloudParams{
			effectID:     effectCloud4,
			textureFiles: []string{"effect/fog1.tga", "effect/fog2.tga", "effect/fog3.tga"},
			tint:         color.RGBA{R: 252, G: 171, B: 143, A: 255},
			count:        320,
			radius:       150 * weatherCloudClassicUnit,
			zOffset:      -20 * weatherCloudClassicUnit,
			zRand:        5 * weatherCloudClassicUnit,
			sizeBase:     35 * math.Sqrt2 * weatherCloudClassicUnit,
			sizeRand:     10 * math.Sqrt2 * weatherCloudClassicUnit,
			driftSpeed:   0.015 * weatherCloudFrameRate * weatherCloudClassicUnit,
			ramp:         170 * time.Second / weatherCloudFrameRate,
			fadeOut:      170 * time.Second / weatherCloudFrameRate,
			rotStartMin:  300 * time.Second / weatherCloudFrameRate,
			rotStartRand: 200 * time.Second / weatherCloudFrameRate,
			overlay:      true,
			additive:     false,
			blackKey:     false,
			disableFog:   true,
			screenHaze:   color.RGBA{R: 252, G: 171, B: 143, A: 70},
		}, true
	default:
		return mapWeatherCloudParams{}, false
	}
}

func (m *WorldMode) drawMapWeatherCloudEffect(screen *render.Frame, ctx client.Context, projection sceneProjection, effectID int, now time.Time) {
	params, ok := weatherCloudParamsForEffect(effectID)
	if !ok || ctx.World == nil {
		return
	}
	key := normalizeMapNameForWeather(ctx.World.MapName)
	centerX, centerY := mapWeatherCloudCenter(ctx.World, projection, now)
	m.mapWeatherCloud.ensure(key, params, ctx.World, centerX, centerY, now)
	m.mapWeatherCloud.update(params, ctx.World, centerX, centerY, now)
	options := texturedEffectBillboardDrawOptions(params.additive, params.overlay)
	options.DisableFog = params.disableFog
	for index := range m.mapWeatherCloud.clouds {
		cloud := &m.mapWeatherCloud.clouds[index]
		alpha := mapWeatherCloudAlpha(*cloud, params)
		if alpha <= 0 {
			continue
		}
		texture := m.effectFileTextureWithBlackKey(ctx.Resources, params.textureFiles[cloud.textureIndex%len(params.textureFiles)], params.blackKey)
		if texture == nil {
			continue
		}
		size := cloud.size * (1 + 0.05*math.Sin(cloud.breath))
		tint := params.tint
		tint.A = uint8(clampFloat(alpha, 0, 1) * 255)
		drawTexturedEffectBillboardRotatedXYWithOptions(screen, projection, texture, cloud.x, cloud.y, cloud.z, size, size, 0, tint, options)
	}
	drawMapWeatherScreenHaze(screen, params.screenHaze)
}

func drawMapWeatherScreenHaze(screen *render.Frame, tint color.RGBA) {
	if screen == nil || tint.A == 0 {
		return
	}
	bounds := screen.Bounds()
	width, height := bounds.Dx(), bounds.Dy()
	if width <= 0 || height <= 0 {
		return
	}
	r := float32(tint.R) / 255
	g := float32(tint.G) / 255
	b := float32(tint.B) / 255
	a := float32(tint.A) / 255
	vertices := []render.Vertex{
		{DstX: 0, DstY: 0, SrcX: 0, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: float32(width), DstY: 0, SrcX: 1, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: 0, DstY: float32(height), SrcX: 0, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		{DstX: float32(width), DstY: float32(height), SrcX: 1, SrcY: 1, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
	}
	screen.DrawTrianglesOwned(vertices, []uint16{0, 1, 2, 2, 1, 3}, render.WhiteImage(), &render.DrawTrianglesOptions{Filter: render.FilterNearest, Address: render.AddressUnsafe})
}

func mapWeatherCloudCenter(world *worldstate.World, projection sceneProjection, now time.Time) (float64, float64) {
	if world == nil {
		return projection.playerX, projection.playerY
	}
	playerX, playerY := actorRenderPosition(world.Player, now)
	return cellCenter(playerX), cellCenter(playerY)
}

func (s *mapWeatherCloudState) ensure(key string, params mapWeatherCloudParams, world *worldstate.World, centerX, centerY float64, now time.Time) {
	if s.key == key && s.effectID == params.effectID && len(s.clouds) == params.count {
		return
	}
	s.key = key
	s.effectID = params.effectID
	s.lastUpdate = now
	s.clouds = make([]mapWeatherCloud, params.count)
	for i := range s.clouds {
		s.spawn(i, params, world, centerX, centerY)
	}
}

func (s *mapWeatherCloudState) reset() {
	*s = mapWeatherCloudState{}
}

func (s *mapWeatherCloudState) update(params mapWeatherCloudParams, world *worldstate.World, centerX, centerY float64, now time.Time) {
	if s.lastUpdate.IsZero() {
		s.lastUpdate = now
		return
	}
	step := now.Sub(s.lastUpdate)
	s.lastUpdate = now
	if step <= 0 {
		return
	}
	if step > weatherCloudMaxStep {
		step = weatherCloudMaxStep
	}
	seconds := step.Seconds()
	for i := range s.clouds {
		cloud := &s.clouds[i]
		cloud.age += step
		if cloud.age >= cloud.rotStart+params.fadeOut {
			cloud.generation++
			s.spawn(i, params, world, centerX, centerY)
			continue
		}
		cloud.x += params.driftSpeed * math.Sin(cloud.phaseX) * seconds
		cloud.y += params.driftSpeed * math.Sin(cloud.phaseY) * seconds
		cloud.phaseX += cloud.phaseRateX * seconds
		cloud.phaseY += cloud.phaseRateY * seconds
		cloud.breath += seconds
	}
}

func (s *mapWeatherCloudState) spawn(index int, params mapWeatherCloudParams, world *worldstate.World, centerX, centerY float64) {
	cloud := &s.clouds[index]
	generation := cloud.generation
	offsetX := weatherCloudHashSigned(index, generation, 1) * params.radius
	offsetY := weatherCloudHashSigned(index, generation, 2) * params.radius
	cloud.x = centerX + offsetX
	cloud.y = centerY + offsetY
	ground := terrainHeightAtRenderPoint(world, cloud.x, cloud.y)
	cloud.z = ground + params.zOffset - weatherCloudHash01(index, generation, 3)*params.zRand
	cloud.size = params.sizeBase + weatherCloudHash01(index, generation, 4)*params.sizeRand
	cloud.age = 0
	cloud.rotStart = params.rotStartMin + time.Duration(weatherCloudHash01(index, generation, 12)*float64(params.rotStartRand))
	cloud.phaseX = weatherCloudHash01(index, generation, 5) * 2 * math.Pi
	cloud.phaseY = weatherCloudHash01(index, generation, 6) * 2 * math.Pi
	cloud.phaseRateX = weatherCloudPhaseRate(index, generation, 10)
	cloud.phaseRateY = weatherCloudPhaseRate(index, generation, 11)
	cloud.breath = weatherCloudHash01(index, generation, 7) * 2 * math.Pi
	cloud.textureIndex = int(weatherCloudHash01(index/4, 0, 9) * float64(len(params.textureFiles)))
	if cloud.textureIndex >= len(params.textureFiles) {
		cloud.textureIndex = len(params.textureFiles) - 1
	}
}

func weatherCloudPhaseRate(index, generation, salt int) float64 {
	rate := degreesToRadians(weatherCloudFrameRate)
	if weatherCloudHash01(index, generation, salt) < 0.5 {
		return -rate
	}
	return rate
}

func mapWeatherCloudAlpha(cloud mapWeatherCloud, params mapWeatherCloudParams) float64 {
	switch {
	case cloud.age < params.ramp:
		return weatherCloudClassicAlphaMax * float64(cloud.age) / float64(params.ramp)
	case cloud.age <= cloud.rotStart:
		return weatherCloudClassicAlphaMax
	case cloud.age < cloud.rotStart+params.fadeOut:
		return weatherCloudClassicAlphaMax * (1 - float64(cloud.age-cloud.rotStart)/float64(params.fadeOut))
	default:
		return 0
	}
}

func weatherCloudHashSigned(index, generation, salt int) float64 {
	return weatherCloudHash01(index, generation, salt)*2 - 1
}

func weatherCloudHash01(index, generation, salt int) float64 {
	x := uint32(index)*2_654_435_761 + uint32(generation*11+salt)*40_503 + 0x9E37_79B9
	x ^= x >> 15
	return float64(x%100_000) / 100_000
}
