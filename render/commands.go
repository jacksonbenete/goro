package render

import (
	"image/color"
	"math"
)

var quadIndices012213 = []uint16{0, 1, 2, 2, 1, 3}

type DrawCommand struct {
	Vertices []Vertex
	Indices  []uint16
	Texture  *Image
	Options  DrawTrianglesOptions
}

type WorldCommand struct {
	Vertices     []Vertex3D
	Indices      []uint16
	Texture      *Image
	LightTexture *Image
	Options      DrawTrianglesOptions
}

type WorldBillboardCommand struct {
	Texture     *Image
	Options     DrawTrianglesOptions
	Center      [3]float32
	RightAxis   [3]float32
	UpAxis      [3]float32
	DepthUpAxis [3]float32
	Width       float32
	Height      float32
	AnchorX     float32
	AnchorY     float32
	ColorR      float32
	ColorG      float32
	ColorB      float32
	ColorA      float32
	DepthBias   float32
}

type UITextLabelCommand struct {
	Text       string
	X          float64
	Y          float64
	Foreground color.RGBA
	Outline    color.RGBA
	Centered   bool
	Bold       bool
	Size       float32
}

type UIRectCommand struct {
	X, Y, W, H float64
	Color      color.RGBA
}

type UITextBoxAnchor int

const (
	UITextBoxAnchorTopLeft UITextBoxAnchor = iota
	UITextBoxAnchorBottomCenter
	UITextBoxAnchorTooltipCenter
)

type UITextBoxCommand struct {
	Text     string
	X        float64
	Y        float64
	AltY     float64
	Anchor   UITextBoxAnchor
	MaxWidth float64
	MaxLines int
}

func NewScreenImage(width, height int) *Image {
	img := NewImage(width, height)
	img.screen = true
	img.screenScaleX = 1
	img.screenScaleY = 1
	return img
}

func (i *Image) BeginFrame() {
	if i == nil {
		return
	}
	i.commands = i.commands[:0]
	i.worldCommands = i.worldCommands[:0]
	i.worldMeshes = i.worldMeshes[:0]
	i.worldBillboards = i.worldBillboards[:0]
	i.uiRects = i.uiRects[:0]
	i.uiTextBoxes = i.uiTextBoxes[:0]
	i.uiTextLabels = i.uiTextLabels[:0]
	i.camera = Camera3D{}
}

func (i *Image) clearUIOverlayCommands() {
	if i == nil {
		return
	}
	i.uiRects = i.uiRects[:0]
	i.uiTextBoxes = i.uiTextBoxes[:0]
	i.uiTextLabels = i.uiTextLabels[:0]
}

func (i *Image) SetScreenScale(x, y float32) {
	if i == nil || !i.screen {
		return
	}
	if x <= 0 || math.IsNaN(float64(x)) || math.IsInf(float64(x), 0) {
		x = 1
	}
	if y <= 0 || math.IsNaN(float64(y)) || math.IsInf(float64(y), 0) {
		y = 1
	}
	i.screenScaleX = x
	i.screenScaleY = y
}

func (i *Image) SetCamera3D(camera Camera3D) {
	if i == nil {
		return
	}
	i.camera = camera
}

func (i *Image) DrawImage(src *Image, opts *DrawImageOptions) {
	if i == nil || i.pix == nil || src == nil || src.pix == nil {
		return
	}
	var o DrawImageOptions
	if opts != nil {
		o = *opts
	}
	w, h := src.Bounds().Dx(), src.Bounds().Dy()
	if i.screen {
		r, g, b, a := o.ColorScale.rgba()
		p0x, p0y := o.GeoM.apply(0, 0)
		p1x, p1y := o.GeoM.apply(float64(w), 0)
		p2x, p2y := o.GeoM.apply(0, float64(h))
		p3x, p3y := o.GeoM.apply(float64(w), float64(h))
		vertices := []Vertex{
			{DstX: float32(p0x), DstY: float32(p0y), SrcX: 0, SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
			{DstX: float32(p1x), DstY: float32(p1y), SrcX: float32(w), SrcY: 0, ColorR: r, ColorG: g, ColorB: b, ColorA: a},
			{DstX: float32(p2x), DstY: float32(p2y), SrcX: 0, SrcY: float32(h), ColorR: r, ColorG: g, ColorB: b, ColorA: a},
			{DstX: float32(p3x), DstY: float32(p3y), SrcX: float32(w), SrcY: float32(h), ColorR: r, ColorG: g, ColorB: b, ColorA: a},
		}
		i.DrawTrianglesOwned(vertices, quadIndices012213, src, &DrawTrianglesOptions{Filter: o.Filter, Address: AddressClampToZero, Blend: o.Blend})
		return
	}
	corners := [4][2]float64{{0, 0}, {float64(w), 0}, {0, float64(h)}, {float64(w), float64(h)}}
	minX, minY := math.Inf(1), math.Inf(1)
	maxX, maxY := math.Inf(-1), math.Inf(-1)
	for _, corner := range corners {
		x, y := o.GeoM.apply(corner[0], corner[1])
		minX, minY = math.Min(minX, x), math.Min(minY, y)
		maxX, maxY = math.Max(maxX, x), math.Max(maxY, y)
	}
	inv, ok := o.GeoM.invert()
	if !ok {
		return
	}
	r, g, b, a := o.ColorScale.rgba()
	dstBounds := i.pix.Bounds()
	x0 := clampInt(int(math.Floor(minX)), dstBounds.Min.X, dstBounds.Max.X)
	y0 := clampInt(int(math.Floor(minY)), dstBounds.Min.Y, dstBounds.Max.Y)
	x1 := clampInt(int(math.Ceil(maxX)), dstBounds.Min.X, dstBounds.Max.X)
	y1 := clampInt(int(math.Ceil(maxY)), dstBounds.Min.Y, dstBounds.Max.Y)
	for y := y0; y < y1; y++ {
		for x := x0; x < x1; x++ {
			sx, sy := inv.apply(float64(x)+0.5, float64(y)+0.5)
			if sx < 0 || sy < 0 || sx >= float64(w) || sy >= float64(h) {
				continue
			}
			sc := src.sample(float32(sx), float32(sy), o.Filter, AddressClampToZero)
			sc.R, sc.G, sc.B, sc.A = scaleByte(sc.R, r), scaleByte(sc.G, g), scaleByte(sc.B, b), scaleByte(sc.A, a)
			i.blendPixel(x, y, sc, o.Blend)
		}
	}
}

func (i *Image) DrawTriangles(vertices []Vertex, indices []uint16, texture *Image, opts *DrawTrianglesOptions) {
	if i == nil || i.pix == nil || texture == nil || texture.pix == nil {
		return
	}
	var o DrawTrianglesOptions
	if opts != nil {
		o = *opts
	}
	if i.screen {
		i.DrawTrianglesOwned(append([]Vertex(nil), vertices...), append([]uint16(nil), indices...), texture, &o)
		return
	}
	for n := 0; n+2 < len(indices); n += 3 {
		i.drawTriangle(vertices[indices[n]], vertices[indices[n+1]], vertices[indices[n+2]], texture, o)
	}
	i.version++
}

func (i *Image) DrawTrianglesOwned(vertices []Vertex, indices []uint16, texture *Image, opts *DrawTrianglesOptions) {
	if i == nil || i.pix == nil || texture == nil || texture.pix == nil {
		return
	}
	var o DrawTrianglesOptions
	if opts != nil {
		o = *opts
	}
	if i.screen {
		i.commands = append(i.commands, DrawCommand{
			Vertices: vertices,
			Indices:  indices,
			Texture:  texture,
			Options:  o,
		})
		return
	}
	i.DrawTriangles(vertices, indices, texture, &o)
}

func (i *Image) DrawTriangles3D(vertices []Vertex3D, indices []uint16, texture *Image, opts *DrawTrianglesOptions) {
	if i == nil || i.pix == nil || texture == nil || texture.pix == nil {
		return
	}
	var o DrawTrianglesOptions
	if opts != nil {
		o = *opts
	}
	if i.screen {
		i.DrawTriangles3DOwned(append([]Vertex3D(nil), vertices...), append([]uint16(nil), indices...), texture, &o)
		return
	}
	vertices2D := make([]Vertex, len(vertices))
	for n, vertex := range vertices {
		vertices2D[n] = Vertex{
			DstX:   vertex.X,
			DstY:   vertex.Y,
			SrcX:   vertex.SrcX,
			SrcY:   vertex.SrcY,
			ColorR: vertex.ColorR,
			ColorG: vertex.ColorG,
			ColorB: vertex.ColorB,
			ColorA: vertex.ColorA,
		}
	}
	i.DrawTriangles(vertices2D, indices, texture, opts)
}

func (i *Image) DrawTriangles3DOwned(vertices []Vertex3D, indices []uint16, texture *Image, opts *DrawTrianglesOptions) {
	if i == nil || i.pix == nil || texture == nil || texture.pix == nil {
		return
	}
	var o DrawTrianglesOptions
	if opts != nil {
		o = *opts
	}
	if i.screen {
		i.worldCommands = append(i.worldCommands, WorldCommand{
			Vertices: vertices,
			Indices:  indices,
			Texture:  texture,
			Options:  o,
		})
		return
	}
	i.DrawTriangles3D(vertices, indices, texture, &o)
}

func (i *Image) DrawWorldBillboard(cmd WorldBillboardCommand) {
	if i == nil || i.pix == nil || cmd.Texture == nil || cmd.Texture.pix == nil || cmd.Width <= 0 || cmd.Height <= 0 {
		return
	}
	if i.screen {
		i.worldBillboards = append(i.worldBillboards, cmd)
	}
}
