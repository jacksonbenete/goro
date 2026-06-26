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
	let color = textureSample(tex, tex_sampler, input.uv) * input.color;
	if (color.a < 0.01) {
		discard;
	}
	return color;
}
`

const worldShaderWGSL = `
struct Uniforms {
	c0: vec4<f32>,
	c1: vec4<f32>,
	c2: vec4<f32>,
	c3: vec4<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;

struct VertexInput {
	@location(0) pos: vec3<f32>,
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
	let p = vec4<f32>(input.pos, 1.0);
	var clip = uniforms.c0 * p.x + uniforms.c1 * p.y + uniforms.c2 * p.z + uniforms.c3 * p.w;
	clip.z = clip.z * 0.5 + clip.w * 0.5;
	out.clip = clip;
	out.uv = input.uv;
	out.color = input.color;
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	let color = textureSample(tex, tex_sampler, input.uv) * input.color;
	if (color.a < 0.01) {
		discard;
	}
	return color;
}
`

type gpuRenderer struct {
	dev             *wgpu.Device
	queue           *wgpu.Queue
	format          gputypes.TextureFormat
	bgl             *wgpu.BindGroupLayout
	worldBGL        *wgpu.BindGroupLayout
	layout          *wgpu.PipelineLayout
	worldLayout     *wgpu.PipelineLayout
	pipelineAlpha   *wgpu.RenderPipeline
	pipelineAdd     *wgpu.RenderPipeline
	worldAlphaWrite *wgpu.RenderPipeline
	worldAddWrite   *wgpu.RenderPipeline
	worldAlphaRead  *wgpu.RenderPipeline
	worldAddRead    *wgpu.RenderPipeline
	uniform         *wgpu.Buffer
	worldUniform    *wgpu.Buffer
	samplers        map[samplerKey]*wgpu.Sampler
	textures        map[*Image]*gpuTexture
	bindGroups      map[bindGroupKey]*wgpu.BindGroup
	depthTex        *wgpu.Texture
	depthView       *wgpu.TextureView
	depthWidth      int
	depthHeight     int
	worldVertexBuf  dynamicGPUBuffer
	worldIndexBuf   dynamicGPUBuffer
	screenVertexBuf dynamicGPUBuffer
	screenIndexBuf  dynamicGPUBuffer
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

type dynamicGPUBuffer struct {
	buf      *wgpu.Buffer
	capacity int
}

type samplerKey struct {
	filter  Filter
	address Address
}

type bindGroupKey struct {
	texture *gogpu.Texture
	sampler *wgpu.Sampler
	layout  *wgpu.BindGroupLayout
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

type worldFrame struct {
	floats  []float32
	indices []uint32
	batches []drawBatch
}

type pendingWorldBatch struct {
	key     drawBatchKey
	floats  []float32
	indices []uint32
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
	r.worldUniform, err = r.dev.CreateBuffer(&wgpu.BufferDescriptor{
		Label: "goro-world-uniform",
		Size:  64,
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
	worldShader, err := r.dev.CreateShaderModule(&wgpu.ShaderModuleDescriptor{
		Label: "goro-world-shader",
		WGSL:  worldShaderWGSL,
	})
	if err != nil {
		return err
	}
	defer worldShader.Release()

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
	r.worldBGL, err = r.dev.CreateBindGroupLayout(&wgpu.BindGroupLayoutDescriptor{
		Label: "goro-world-bind-layout",
		Entries: []wgpu.BindGroupLayoutEntry{
			{Binding: 0, Visibility: wgpu.ShaderStageVertex, Buffer: &gputypes.BufferBindingLayout{Type: gputypes.BufferBindingTypeUniform, MinBindingSize: 64}},
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
	r.worldLayout, err = r.dev.CreatePipelineLayout(&wgpu.PipelineLayoutDescriptor{
		Label:            "goro-world-pipeline-layout",
		BindGroupLayouts: []*wgpu.BindGroupLayout{r.worldBGL},
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
	if err != nil {
		return err
	}
	r.worldAlphaWrite, err = r.createWorldPipeline(worldShader, gputypes.BlendStateAlpha(), true, "goro-world-pipeline-alpha-write")
	if err != nil {
		return err
	}
	r.worldAddWrite, err = r.createWorldPipeline(worldShader, add, true, "goro-world-pipeline-add-write")
	if err != nil {
		return err
	}
	r.worldAlphaRead, err = r.createWorldPipeline(worldShader, gputypes.BlendStateAlpha(), false, "goro-world-pipeline-alpha-read")
	if err != nil {
		return err
	}
	r.worldAddRead, err = r.createWorldPipeline(worldShader, add, false, "goro-world-pipeline-add-read")
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

func (r *gpuRenderer) createWorldPipeline(shader *wgpu.ShaderModule, blend gputypes.BlendState, depthWrite bool, label string) (*wgpu.RenderPipeline, error) {
	return r.dev.CreateRenderPipeline(&wgpu.RenderPipelineDescriptor{
		Label:  label,
		Layout: r.worldLayout,
		Vertex: wgpu.VertexState{
			Module:     shader,
			EntryPoint: "vs_main",
			Buffers: []wgpu.VertexBufferLayout{{
				ArrayStride: 36,
				StepMode:    gputypes.VertexStepModeVertex,
				Attributes: []gputypes.VertexAttribute{
					{Format: gputypes.VertexFormatFloat32x3, Offset: 0, ShaderLocation: 0},
					{Format: gputypes.VertexFormatFloat32x2, Offset: 12, ShaderLocation: 1},
					{Format: gputypes.VertexFormatFloat32x4, Offset: 20, ShaderLocation: 2},
				},
			}},
		},
		Primitive: gputypes.PrimitiveState{
			Topology:  gputypes.PrimitiveTopologyTriangleList,
			FrontFace: gputypes.FrontFaceCCW,
			CullMode:  gputypes.CullModeNone,
		},
		DepthStencil: &wgpu.DepthStencilState{
			Format:            gputypes.TextureFormatDepth24Plus,
			DepthWriteEnabled: depthWrite,
			DepthCompare:      gputypes.CompareFunctionLessEqual,
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
	if screen.camera.Enabled {
		if err := r.queue.WriteBuffer(r.worldUniform, 0, matrixBytes(screen.camera.ViewProjection)); err != nil {
			return fmt.Errorf("upload world uniform: %w", err)
		}
	}
	if err := r.ensureDepth(width, height); err != nil {
		return err
	}
	r.releaseFrameResources()
	world := r.buildWorldFrame(screen)
	frame := r.buildFrame(screen)
	if r.statsEnabled && time.Since(r.statsLast) >= time.Second {
		log.Printf("render stats world_commands=%d world_batches=%d world_vertices=%d world_indices=%d commands=%d batches=%d vertices=%d indices=%d textures=%d bindgroups=%d", len(screen.worldCommands), len(world.batches), len(world.floats)/9, len(world.indices), len(screen.commands), len(frame.batches), len(frame.floats)/8, len(frame.indices), len(r.textures), len(r.bindGroups))
		r.statsLast = time.Now()
	}
	var worldVertexBuf, worldIndexBuf, vertexBuf, indexBuf *wgpu.Buffer
	var err error
	if len(world.floats) > 0 && len(world.indices) > 0 {
		worldVertexBuf, err = r.dynamicBuffer(&r.worldVertexBuf, "goro-world-vertices", len(world.floats)*4, wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, floatBytes(world.floats))
		if err != nil {
			return err
		}
		worldIndexBuf, err = r.dynamicBuffer(&r.worldIndexBuf, "goro-world-indices", len(world.indices)*4, wgpu.BufferUsageIndex|wgpu.BufferUsageCopyDst, u32Bytes(world.indices))
		if err != nil {
			return err
		}
	}
	if len(frame.floats) > 0 && len(frame.indices) > 0 {
		vertexBuf, err = r.dynamicBuffer(&r.screenVertexBuf, "goro-screen-vertices", len(frame.floats)*4, wgpu.BufferUsageVertex|wgpu.BufferUsageCopyDst, floatBytes(frame.floats))
		if err != nil {
			return err
		}
		indexBuf, err = r.dynamicBuffer(&r.screenIndexBuf, "goro-screen-indices", len(frame.indices)*4, wgpu.BufferUsageIndex|wgpu.BufferUsageCopyDst, u32Bytes(frame.indices))
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
		DepthStencilAttachment: &wgpu.RenderPassDepthStencilAttachment{
			View:            r.depthView,
			DepthLoadOp:     gputypes.LoadOpClear,
			DepthStoreOp:    gputypes.StoreOpStore,
			DepthClearValue: 1,
		},
	})
	if err != nil {
		return err
	}
	if worldVertexBuf != nil && worldIndexBuf != nil && screen.camera.Enabled {
		pass.SetVertexBuffer(0, worldVertexBuf, 0)
		pass.SetIndexBuffer(worldIndexBuf, gputypes.IndexFormatUint32, 0)
		for _, batch := range world.batches {
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
			bg, err := r.bindGroup(r.worldBGL, r.worldUniform, 64, tex.tex, sampler)
			if err != nil {
				_ = pass.End()
				return err
			}
			pass.SetPipeline(r.worldPipelineFor(batch.key.options.Blend, batch.key.options.DepthWrite))
			pass.SetBindGroup(0, bg, nil)
			pass.DrawIndexed(batch.indexCount, 1, batch.firstIndex, 0, 0)
		}
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
		bg, err := r.bindGroup(r.bgl, r.uniform, 16, tex.tex, sampler)
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

func (r *gpuRenderer) buildWorldFrame(screen *Image) worldFrame {
	var pending []pendingWorldBatch
	batchByKey := make(map[drawBatchKey]int)
	var alpha []WorldCommand
	for _, cmd := range screen.worldCommands {
		if cmd.Texture == nil || cmd.Texture.pix == nil {
			continue
		}
		if !cmd.Options.DepthWrite {
			alpha = append(alpha, cmd)
			continue
		}
		key := drawBatchKey{texture: cmd.Texture, options: cmd.Options}
		batchIndex, ok := batchByKey[key]
		if !ok {
			pending = append(pending, pendingWorldBatch{key: key})
			batchIndex = len(pending) - 1
			batchByKey[key] = batchIndex
		}
		w, h := cmd.Texture.Bounds().Dx(), cmd.Texture.Bounds().Dy()
		if w <= 0 || h <= 0 {
			continue
		}
		batch := &pending[batchIndex]
		base := uint32(len(batch.floats) / 9)
		for _, v := range cmd.Vertices {
			batch.floats = append(batch.floats,
				v.X, v.Y, v.Z,
				v.SrcX/float32(w), v.SrcY/float32(h),
				saneColor(v.ColorR), saneColor(v.ColorG), saneColor(v.ColorB), saneColor(v.ColorA),
			)
		}
		for _, idx := range cmd.Indices {
			batch.indices = append(batch.indices, base+idx)
		}
	}
	var frame worldFrame
	for _, batch := range pending {
		if len(batch.indices) == 0 || len(batch.floats) == 0 {
			continue
		}
		vertexBase := uint32(len(frame.floats) / 9)
		firstIndex := uint32(len(frame.indices))
		frame.floats = append(frame.floats, batch.floats...)
		for _, index := range batch.indices {
			frame.indices = append(frame.indices, vertexBase+index)
		}
		frame.batches = append(frame.batches, drawBatch{
			key:        batch.key,
			firstIndex: firstIndex,
			indexCount: uint32(len(batch.indices)),
		})
	}
	var current *drawBatch
	for _, cmd := range alpha {
		key := drawBatchKey{texture: cmd.Texture, options: cmd.Options}
		if current == nil || current.key != key {
			frame.batches = append(frame.batches, drawBatch{key: key, firstIndex: uint32(len(frame.indices))})
			current = &frame.batches[len(frame.batches)-1]
		}
		w, h := cmd.Texture.Bounds().Dx(), cmd.Texture.Bounds().Dy()
		if w <= 0 || h <= 0 {
			continue
		}
		base := uint32(len(frame.floats) / 9)
		for _, v := range cmd.Vertices {
			frame.floats = append(frame.floats,
				v.X, v.Y, v.Z,
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

func (r *gpuRenderer) bindGroup(layout *wgpu.BindGroupLayout, uniform *wgpu.Buffer, uniformSize uint64, tex *gogpu.Texture, sampler *wgpu.Sampler) (*wgpu.BindGroup, error) {
	key := bindGroupKey{texture: tex, sampler: sampler, layout: layout}
	if bg := r.bindGroups[key]; bg != nil {
		return bg, nil
	}
	bg, err := r.dev.CreateBindGroup(&wgpu.BindGroupDescriptor{
		Label:  "goro-screen-bind-group",
		Layout: layout,
		Entries: []wgpu.BindGroupEntry{
			{Binding: 0, Buffer: uniform, Size: uniformSize},
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

func (r *gpuRenderer) dynamicBuffer(slot *dynamicGPUBuffer, label string, size int, usage wgpu.BufferUsage, data []byte) (*wgpu.Buffer, error) {
	if size <= 0 {
		size = 4
	}
	if slot.buf == nil || slot.capacity < size {
		if slot.buf != nil {
			slot.buf.Release()
			slot.buf = nil
		}
		capacity := nextBufferCapacity(size)
		buf, err := r.dev.CreateBuffer(&wgpu.BufferDescriptor{
			Label: label,
			Size:  uint64(capacity),
			Usage: usage,
		})
		if err != nil {
			return nil, err
		}
		slot.buf = buf
		slot.capacity = capacity
	}
	if len(data) > 0 {
		if err := r.queue.WriteBuffer(slot.buf, 0, data); err != nil {
			return nil, err
		}
	}
	return slot.buf, nil
}

func nextBufferCapacity(size int) int {
	capacity := 4096
	for capacity < size {
		capacity *= 2
	}
	return capacity
}

func (r *gpuRenderer) pipeline(blend Blend) *wgpu.RenderPipeline {
	if blend == BlendLighter {
		return r.pipelineAdd
	}
	return r.pipelineAlpha
}

func (r *gpuRenderer) worldPipelineFor(blend Blend, depthWrite bool) *wgpu.RenderPipeline {
	if blend == BlendLighter {
		if depthWrite {
			return r.worldAddWrite
		}
		return r.worldAddRead
	}
	if depthWrite {
		return r.worldAlphaWrite
	}
	return r.worldAlphaRead
}

func (r *gpuRenderer) ensureDepth(width, height int) error {
	if r.depthView != nil && r.depthWidth == width && r.depthHeight == height {
		return nil
	}
	if r.depthView != nil {
		r.depthView.Release()
		r.depthView = nil
	}
	if r.depthTex != nil {
		r.depthTex.Release()
		r.depthTex = nil
	}
	tex, err := r.dev.CreateTexture(&wgpu.TextureDescriptor{
		Label: "goro-depth",
		Size: wgpu.Extent3D{
			Width:              uint32(width),
			Height:             uint32(height),
			DepthOrArrayLayers: 1,
		},
		MipLevelCount: 1,
		SampleCount:   1,
		Dimension:     gputypes.TextureDimension2D,
		Format:        gputypes.TextureFormatDepth24Plus,
		Usage:         wgpu.TextureUsageRenderAttachment,
	})
	if err != nil {
		return err
	}
	view, err := r.dev.CreateTextureView(tex, nil)
	if err != nil {
		tex.Release()
		return err
	}
	r.depthTex = tex
	r.depthView = view
	r.depthWidth = width
	r.depthHeight = height
	return nil
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
	for _, buf := range []*wgpu.Buffer{
		r.worldVertexBuf.buf,
		r.worldIndexBuf.buf,
		r.screenVertexBuf.buf,
		r.screenIndexBuf.buf,
	} {
		if buf != nil {
			buf.Release()
		}
	}
	for _, res := range []interface{ Release() }{
		r.pipelineAlpha, r.pipelineAdd, r.worldAlphaWrite, r.worldAddWrite, r.worldAlphaRead, r.worldAddRead,
		r.layout, r.worldLayout, r.bgl, r.worldBGL,
		r.uniform, r.worldUniform, r.depthView, r.depthTex,
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

func matrixBytes(values [16]float32) []byte {
	data := make([]byte, 64)
	for i, v := range values {
		binary.LittleEndian.PutUint32(data[i*4:i*4+4], math.Float32bits(v))
	}
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
