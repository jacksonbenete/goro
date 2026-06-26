package render

import (
	"encoding/binary"
	"fmt"
	"image/color"
	"log"
	"math"
	"os"
	"time"

	"github.com/gogpu/gogpu"
	"github.com/gogpu/gputypes"
	"github.com/gogpu/wgpu"
)

const screenShaderWGSL = `
struct Uniforms {
	screen: vec2<f32>,
	padding: vec2<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;

struct VertexInput {
	@location(0) pos: vec2<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) color: vec4<f32>,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let x = (input.pos.x / uniforms.screen.x) * 2.0 - 1.0;
	let y = 1.0 - (input.pos.y / uniforms.screen.y) * 2.0;
	out.clip = vec4<f32>(x, y, 0.0, 1.0);
	out.uv = input.uv;
	out.color = input.color;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	return textureSample(tex, tex_sampler, input.uv) * input.color;
}
`

type gpuRenderer struct {
	dev             *wgpu.Device
	queue           *wgpu.Queue
	format          gputypes.TextureFormat
	bgl             *wgpu.BindGroupLayout
	layout          *wgpu.PipelineLayout
	pipelineAlpha   *wgpu.RenderPipeline
	pipelineAdd     *wgpu.RenderPipeline
	uniform         *wgpu.Buffer
	samplers        map[samplerKey]*wgpu.Sampler
	textures        map[*Image]*gpuTexture
	bindGroups      map[bindGroupKey]*wgpu.BindGroup
	frameBuffers    []*wgpu.Buffer
	frameBindGroups []*wgpu.BindGroup
	statsEnabled    bool
	statsLast       time.Time
}

type gpuTexture struct {
	tex     *gogpu.Texture
	version uint64
	width   int
	height  int
}

type samplerKey struct {
	filter  Filter
	address Address
}

type bindGroupKey struct {
	texture *gogpu.Texture
	sampler *wgpu.Sampler
}

type drawBatchKey struct {
	texture *Image
	options DrawTrianglesOptions
}

type drawBatch struct {
	key        drawBatchKey
	firstIndex uint32
	indexCount uint32
}

type drawFrame struct {
	floats  []float32
	indices []uint32
	batches []drawBatch
}

func newGPURenderer(ctx *gogpu.Context, app *gogpu.App) (*gpuRenderer, error) {
	provider := app.DeviceProvider()
	if provider == nil || provider.Device() == nil {
		return nil, fmt.Errorf("gogpu device provider is not ready")
	}
	r := &gpuRenderer{
		dev:          provider.Device(),
		queue:        provider.Queue(),
		format:       provider.SurfaceFormat(),
		samplers:     make(map[samplerKey]*wgpu.Sampler),
		textures:     make(map[*Image]*gpuTexture),
		bindGroups:   make(map[bindGroupKey]*wgpu.BindGroup),
		statsEnabled: os.Getenv("GORO_RENDER_STATS") == "1",
	}
	if r.queue == nil {
		r.queue = r.dev.Queue()
	}
	if err := r.init(ctx); err != nil {
		r.release()
		return nil, err
	}
	return r, nil
}

func (r *gpuRenderer) init(_ *gogpu.Context) error {
	var err error
	r.uniform, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "goro-screen-uniform",
		Size:  16,
		Usage: wgpu.BufferUsageUniform | wgpu.BufferUsageCopyDst,
	})
	if err != nil {
		return err
	}
	shader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label: "goro-screen-shader",
		WGSL:  screenShaderWGSL,
	})
	if err != nil {
		return err
	}
	defer shader.Release()

	r.bgl, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-screen-bind-layout",
		Entries: []wgpu.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageVertex, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: 16}},
			{Binding: 1, Visibility: wgpu.ShaderStageFragment, Sampler: &gputypes.SamplerBindingLayout{Type: gputypes.SamplerBindingTypeFiltering}},
			{Binding: 2, Visibility: wgpu.ShaderStageFragment, Texture: &gputypes.TextureBindingLayout{SampleType: gputypes.TextureSampleTypeFloat, ViewDimension: gputypes.TextureViewDimension2D}},
		},
	})
	if err != nil {
		return err
	}
	r.layout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-screen-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.bgl},
	})
	if err != nil {
		return err
	}
	r.pipelineAlpha, err = r.createPipeline(shader, gputypes.BlendStateAlpha(), "goro-screen-pipeline-alpha")
	if err != nil {
		return err
	}
	add := gputypes.BlendState{
		Color: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorSrcAlpha, DstFactor: gputypes.BlendFactorOne, Operation: gputypes.BlendOperationAdd},
		Alpha: gputypes.BlendComponent{SrcFactor: gputypes.BlendFactorOne, DstFactor: gputypes.BlendFactorOne, Operation: gputypes.BlendOperationAdd},
	}
	r.pipelineAdd, err = r.createPipeline(shader, add, "goro-screen-pipeline-add")
	return err
}

func (r *gpuRenderer) createPipeline(shader *wgpu.ShaderModule, blend gputypes.BlendState, label string) (*wgpu.RenderPipeline, error) {
	return r.dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  label,
		Layout: r.layout,
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{{
				ArrayStride: 32,
				StepMode:    gputypes.VertexStepModeVertex,
				Attributes: []gputypes.VertexAttribute{
					{Format: gputypes.VertexFormatFloat32x2, Offset: 0, ShaderLocation: 0},
					{Format: gputypes.VertexFormatFloat32x2, Offset: 8, ShaderLocation: 1},
					{Format: gputypes.VertexFormatFloat32x4, Offset: 16, ShaderLocation: 2},
				},
			}},
		},
		Primitive: gputypes.PrimitiveState{
			Topology:  gputypes.PrimitiveTopologyTriangleList,
			FrontFace: gputypes.FrontFaceCCW,
			CullMode:  gputypes.CullModeNone,
		},
		Fragment: &wgpu.FragmentState{
			Module:     shader,
			EntryPoint: "fs_main",
			Targets: []gputypes.ColorTargetState{{
				Format:    r.format,
				Blend:     &blend,
				WriteMask: gputypes.ColorWriteMaskAll,
			}},
		},
	})
}

func (r *gpuRenderer) Draw(ctx *gogpu.Context, screen *Image) error {
	if screen == nil {
		return nil
	}
	surface := ctx.SurfaceView()
	if surface == nil {
		return nil
	}
	width, height := ctx.FramebufferSize()
	if width <= 0 || height <= 0 {
		b := screen.Bounds()
		width, height = b.Dx(), b.Dy()
	}
	if width <= 0 || height <= 0 {
		return nil
	}
	if err := r.queue.WriteBuffer(r.uniform, 0, uniformBytes(float32(width), float32(height))); err != nil {
		return fmt.Errorf("upload screen uniform: %w", err)
	}
	r.releaseFrameResources()
	frame := r.buildFrame(screen)
	if r.statsEnabled && time.Since(r.statsLast) >= time.Second {
		log.Printf("render stats commands=%d batches=%d vertices=%d indices=%d textures=%d bindgroups=%d", len(screen.commands), len(frame.batches), len(frame.floats)/8, len(frame.indices), len(r.textures), len(r.bindGroups))
		r.statsLast = time.Now()
	}
	var vertexBuf, indexBuf *wgpu.Buffer
	var err error
	if len(frame.floats) > 0 && len(frame.indices) > 0 {
		vertexBuf, err = r.buffer("goro-screen-vertices", len(frame.floats)*4, wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, floatBytes(frame.floats))
		if err != nil {
			return err
		}
		indexBuf, err = r.buffer("goro-screen-indices", len(frame.indices)*4, wgpu.BufferUsageIndex|wgpu.BufferUsageCopyDst, u32Bytes(frame.indices))
		if err != nil {
			return err
		}
	}

	enc, err := r.dev.CreateCommandEncoder(&wgpu.CommandEncoderDescriptor{Label: "goro-screen-encoder"})
	if err != nil {
		return err
	}
	clear := clearValue(screen.clear)
	pass, err := enc.BeginRenderPass(&wgpu.RenderPassDescriptor{
		ColorAttachments: []wgpu.RenderPassColorAttachment{{
			View:       surface,
			LoadOp:     gputypes.LoadOpClear,
			StoreOp:    gputypes.StoreOpStore,
			ClearValue: clear,
		}},
	})
	if err != nil {
		return err
	}
	if vertexBuf != nil && indexBuf != nil {
		pass.SetVertexBuffer(0, vertexBuf, 0)
		pass.SetIndexBuffer(indexBuf, gputypes.IndexFormatUint32, 0)
	}
	for _, batch := range frame.batches {
		if batch.indexCount == 0 {
			continue
		}
		tex, err := r.ensureTexture(ctx, batch.key.texture, batch.key.options)
		if err != nil {
			_ = pass.End()
			return err
		}
		sampler, err := r.sampler(batch.key.options)
		if err != nil {
			_ = pass.End()
			return err
		}
		bg, err := r.bindGroup(tex.tex, sampler)
		if err != nil {
			_ = pass.End()
			return err
		}
		pass.SetPipeline(r.pipeline(batch.key.options.Blend))
		pass.SetBindGroup(0, bg, nil)
		pass.DrawIndexed(batch.indexCount, 1, batch.firstIndex, 0, 0)
	}
	if err := pass.End(); err != nil {
		return err
	}
	cmd, err := enc.Finish()
	if err != nil {
		return err
	}
	_, err = r.queue.Submit(cmd)
	return err
}

func (r *gpuRenderer) buildFrame(screen *Image) drawFrame {
	var frame drawFrame
	var current *drawBatch
	for _, cmd := range screen.commands {
		if cmd.Texture == nil || cmd.Texture.pix == nil {
			continue
		}
		key := drawBatchKey{texture: cmd.Texture, options: cmd.Options}
		if current == nil || current.key != key {
			frame.batches = append(frame.batches, drawBatch{key: key, firstIndex: uint32(len(frame.indices))})
			current = &frame.batches[len(frame.batches)-1]
		}
		w, h := cmd.Texture.Bounds().Dx(), cmd.Texture.Bounds().Dy()
		if w <= 0 || h <= 0 {
			continue
		}
		base := uint32(len(frame.floats) / 8)
		for _, v := range cmd.Vertices {
			frame.floats = append(frame.floats,
				v.DstX, v.DstY,
				v.SrcX/float32(w), v.SrcY/float32(h),
				saneColor(v.ColorR), saneColor(v.ColorG), saneColor(v.ColorB), saneColor(v.ColorA),
			)
		}
		for _, idx := range cmd.Indices {
			frame.indices = append(frame.indices, base+idx)
			current.indexCount++
		}
	}
	return frame
}

func (r *gpuRenderer) ensureTexture(ctx *gogpu.Context, img *Image, opts DrawTrianglesOptions) (*gpuTexture, error) {
	if img == nil || img.pix == nil {
		return nil, fmt.Errorf("nil render texture")
	}
	w, h := img.Bounds().Dx(), img.Bounds().Dy()
	existing := r.textures[img]
	if existing != nil && existing.version == img.version && existing.width == w && existing.height == h {
		return existing, nil
	}
	if existing != nil {
		if existing.width == w && existing.height == h {
			if err := existing.tex.UpdateData(img.RGBA().Pix); err != nil {
				return nil, fmt.Errorf("update render texture: %w", err)
			}
			existing.version = img.version
			return existing, nil
		}
		r.releaseTexture(existing.tex)
		delete(r.textures, img)
	}
	tex, err := ctx.Renderer().NewTextureFromRGBAWithOptions(w, h, img.RGBA().Pix, gogpu.TextureOptions{
		Label:        "goro-image-texture",
		MagFilter:    gpuFilter(opts.Filter),
		MinFilter:    gpuFilter(opts.Filter),
		AddressModeU: gpuAddress(opts.Address),
		AddressModeV: gpuAddress(opts.Address),
	})
	if err != nil {
		return nil, fmt.Errorf("create render texture: %w", err)
	}
	out := &gpuTexture{tex: tex, version: img.version, width: w, height: h}
	r.textures[img] = out
	return out, nil
}

func (r *gpuRenderer) releaseTexture(tex *gogpu.Texture) {
	if tex == nil {
		return
	}
	for key, bg := range r.bindGroups {
		if key.texture == tex {
			if bg != nil {
				bg.Release()
			}
			delete(r.bindGroups, key)
		}
	}
	tex.Destroy()
}

func (r *gpuRenderer) sampler(opts DrawTrianglesOptions) (*wgpu.Sampler, error) {
	key := samplerKey{filter: opts.Filter, address: opts.Address}
	if sampler := r.samplers[key]; sampler != nil {
		return sampler, nil
	}
	sampler, err := r.dev.CreateSampler(&wgpu.SamplerDescriptor{
		Label:        "goro-screen-sampler",
		AddressModeU: gpuAddress(opts.Address),
		AddressModeV: gpuAddress(opts.Address),
		AddressModeW: gputypes.AddressModeClampToEdge,
		MagFilter:    gpuFilter(opts.Filter),
		MinFilter:    gpuFilter(opts.Filter),
		MipmapFilter: gputypes.FilterModeNearest,
		LodMaxClamp:  32,
	})
	if err != nil {
		return nil, err
	}
	r.samplers[key] = sampler
	return sampler, nil
}

func (r *gpuRenderer) bindGroup(tex *gogpu.Texture, sampler *wgpu.Sampler) (*wgpu.BindGroup, error) {
	key := bindGroupKey{texture: tex, sampler: sampler}
	if bg := r.bindGroups[key]; bg != nil {
		return bg, nil
	}
	bg, err := r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "goro-screen-bind-group",
		Layout: r.bgl,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, Buffer: r.uniform, Size: 16},
			{Binding: 1, Sampler: sampler},
			{Binding: 2, TextureView: tex.View()},
		},
	})
	if err != nil {
		return nil, err
	}
	r.bindGroups[key] = bg
	return bg, nil
}

func (r *gpuRenderer) buffer(label string, size int, usage wgpu.BufferUsage, data []byte) (*wgpu.Buffer, error) {
	if size <= 0 {
		size = 4
	}
	buf, err := r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: label,
		Size:  uint64(size),
		Usage: usage,
	})
	if err != nil {
		return nil, err
	}
	r.frameBuffers = append(r.frameBuffers, buf)
	if len(data) > 0 {
		if err := r.queue.WriteBuffer(buf, 0, data); err != nil {
			return nil, err
		}
	}
	return buf, nil
}

func (r *gpuRenderer) pipeline(blend Blend) *wgpu.RenderPipeline {
	if blend == BlendLighter {
		return r.pipelineAdd
	}
	return r.pipelineAlpha
}

func (r *gpuRenderer) releaseFrameResources() {
	for _, buf := range r.frameBuffers {
		if buf != nil {
			buf.Release()
		}
	}
	r.frameBuffers = r.frameBuffers[:0]
	for _, bg := range r.frameBindGroups {
		if bg != nil {
			bg.Release()
		}
	}
	r.frameBindGroups = r.frameBindGroups[:0]
}

func (r *gpuRenderer) release() {
	r.releaseFrameResources()
	for _, bg := range r.bindGroups {
		if bg != nil {
			bg.Release()
		}
	}
	for _, tex := range r.textures {
		if tex != nil && tex.tex != nil {
			tex.tex.Destroy()
		}
	}
	for _, sampler := range r.samplers {
		if sampler != nil {
			sampler.Release()
		}
	}
	for _, res := range []interface{ Release() }{
		r.pipelineAlpha, r.pipelineAdd, r.layout, r.bgl, r.uniform,
	} {
		if res != nil {
			res.Release()
		}
	}
}

func gpuFilter(filter Filter) gputypes.FilterMode {
	if filter == FilterNearest {
		return gputypes.FilterModeNearest
	}
	return gputypes.FilterModeLinear
}

func gpuAddress(address Address) gputypes.AddressMode {
	if address == AddressRepeat {
		return gputypes.AddressModeRepeat
	}
	return gputypes.AddressModeClampToEdge
}

func clearValue(c color.RGBA) gputypes.Color {
	return gputypes.Color{
		R: float64(c.R) / 255,
		G: float64(c.G) / 255,
		B: float64(c.B) / 255,
		A: float64(c.A) / 255,
	}
}

func saneColor(v float32) float32 {
	if math.IsNaN(float64(v)) || math.IsInf(float64(v), 0) {
		return 1
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

func uniformBytes(width, height float32) []byte {
	data := make([]byte, 16)
	binary.LittleEndian.PutUint32(data[0:4], math.Float32bits(width))
	binary.LittleEndian.PutUint32(data[4:8], math.Float32bits(height))
	return data
}

func floatBytes(values []float32) []byte {
	data := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(data[i*4:i*4+4], math.Float32bits(v))
	}
	return data
}

func u32Bytes(values []uint32) []byte {
	data := make([]byte, len(values)*4)
	for i, v := range values {
		binary.LittleEndian.PutUint32(data[i*4:i*4+4], v)
	}
	return data
}
