package render

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
	fog: vec4<f32>,
	fog_color: vec4<f32>,
}

@group(0) @binding(0) var<uniform> uniforms: Uniforms;
@group(0) @binding(1) var tex_sampler: sampler;
@group(0) @binding(2) var tex: texture_2d<f32>;

struct VertexInput {
	@location(0) pos: vec3<f32>,
	@location(1) uv: vec2<f32>,
	@location(2) color: vec4<f32>,
	@location(3) depth_pos: vec3<f32>,
	@location(4) depth_bias: f32,
}

struct VertexOutput {
	@builtin(position) clip: vec4<f32>,
	@location(0) uv: vec2<f32>,
	@location(1) color: vec4<f32>,
	@location(2) fog_amount: f32,
}

@vertex
fn vs_main(input: VertexInput) -> VertexOutput {
	var out: VertexOutput;
	let clip = uniforms.c0 * input.pos[0] + uniforms.c1 * input.pos[1] + uniforms.c2 * input.pos[2] + uniforms.c3;
	let depth_clip = uniforms.c0 * input.depth_pos[0] + uniforms.c1 * input.depth_pos[1] + uniforms.c2 * input.depth_pos[2] + uniforms.c3;
	let clip_z = clip[2] * 0.5 + clip[3] * 0.5;
	let depth_z = depth_clip[2] * 0.5 + depth_clip[3] * 0.5;
	let fog_depth = clip_z;
	let occlusion_z = min(clip_z, depth_z * clip[3] / max(depth_clip[3], 0.000001));
	let final_z = max(0.0, occlusion_z - input.depth_bias * clip[3]);
	out.clip = vec4<f32>(clip[0], clip[1], final_z, clip[3]);
	out.uv = input.uv;
	out.color = input.color;
	out.fog_amount = smoothstep(uniforms.fog[0], uniforms.fog[1], fog_depth) * uniforms.fog[2];
	return out;
}

@fragment
fn fs_main(input: VertexOutput) -> @location(0) vec4<f32> {
	var color = textureSample(tex, tex_sampler, input.uv) * input.color;
	if (color[3] < 0.01) {
		discard;
	}
	let fog = clamp(input.fog_amount, 0.0, 1.0);
	return vec4<f32>(
		color[0] * (1.0 - fog) + uniforms.fog_color[0] * fog,
		color[1] * (1.0 - fog) + uniforms.fog_color[1] * fog,
		color[2] * (1.0 - fog) + uniforms.fog_color[2] * fog,
		color[3],
	);
}
`
