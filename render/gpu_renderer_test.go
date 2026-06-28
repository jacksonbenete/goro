package render

import (
	"encoding/binary"
	"math"
	"testing"

	"github.com/gogpu/naga"
	"github.com/gogpu/naga/spirv"
)

func TestRendererShadersParse(t *testing.T) {
	for name, source := range map[string]string{
		"screen":          screenShaderWGSL,
		"world":           worldShaderWGSL,
		"world-billboard": worldBillboardShaderWGSL,
	} {
		ast, err := naga.Parse(source)
		if err != nil {
			t.Fatalf("%s shader parse: %v", name, err)
		}
		if _, err := naga.Lower(ast); err != nil {
			t.Fatalf("%s shader lower: %v", name, err)
		}
	}
}

func TestRendererShadersGenerateSPIRV(t *testing.T) {
	for name, source := range map[string]string{
		"screen":          screenShaderWGSL,
		"world":           worldShaderWGSL,
		"world-billboard": worldBillboardShaderWGSL,
	} {
		ast, err := naga.Parse(source)
		if err != nil {
			t.Fatalf("%s shader parse: %v", name, err)
		}
		module, err := naga.Lower(ast)
		if err != nil {
			t.Fatalf("%s shader lower: %v", name, err)
		}
		data, err := naga.GenerateSPIRV(module, spirv.Options{Version: spirv.Version1_3})
		if err != nil {
			t.Fatalf("%s shader SPIR-V: %v", name, err)
		}
		if len(data) == 0 {
			t.Fatalf("%s shader SPIR-V is empty", name)
		}
	}
}

func TestWorldMeshSubmissionDoesNotCreateDynamicWorldCommand(t *testing.T) {
	screen := NewScreenImage(320, 240)
	texture := WhiteImage()
	mesh := NewWorldMesh([]Vertex3D{
		{X: 0, Y: 0, Z: 0, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{X: 1, Y: 0, Z: 0, SrcX: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
		{X: 1, Y: 1, Z: 0, SrcX: 1, SrcY: 1, ColorR: 1, ColorG: 1, ColorB: 1, ColorA: 1},
	}, []uint16{0, 1, 2}, texture, &DrawTrianglesOptions{DepthWrite: true})

	screen.BeginFrame()
	screen.DrawWorldMesh(mesh)

	if got := len(screen.worldMeshes); got != 1 {
		t.Fatalf("world mesh commands = %d, want 1", got)
	}
	if got := len(screen.worldCommands); got != 0 {
		t.Fatalf("dynamic world commands = %d, want 0", got)
	}
}

func TestWorldBillboardSubmissionDoesNotCreateDynamicWorldCommand(t *testing.T) {
	screen := NewScreenImage(320, 240)
	screen.BeginFrame()
	screen.DrawWorldBillboard(WorldBillboardCommand{
		Texture: WhiteImage(),
		Width:   16,
		Height:  16,
		ColorR:  1,
		ColorG:  1,
		ColorB:  1,
		ColorA:  1,
	})

	if got := len(screen.worldBillboards); got != 1 {
		t.Fatalf("world billboard commands = %d, want 1", got)
	}
	if got := len(screen.worldCommands); got != 0 {
		t.Fatalf("dynamic world commands = %d, want 0", got)
	}
}

func TestWorldBillboardCommandsKeepSeparateInstanceData(t *testing.T) {
	screen := NewScreenImage(320, 240)
	texture := WhiteImage()
	screen.BeginFrame()
	screen.DrawWorldBillboard(WorldBillboardCommand{
		Texture: texture,
		Center:  [3]float32{1, 2, 3},
		Width:   16,
		Height:  16,
		ColorA:  1,
	})
	screen.DrawWorldBillboard(WorldBillboardCommand{
		Texture: texture,
		Center:  [3]float32{7, 8, 9},
		Width:   16,
		Height:  16,
		ColorA:  0.5,
	})

	if got := len(screen.worldBillboards); got != 2 {
		t.Fatalf("world billboard commands = %d, want 2", got)
	}
	if screen.worldBillboards[0].Center == screen.worldBillboards[1].Center {
		t.Fatalf("billboard centers unexpectedly alias: %v", screen.worldBillboards[0].Center)
	}
	if screen.worldBillboards[0].ColorA == screen.worldBillboards[1].ColorA {
		t.Fatalf("billboard alpha unexpectedly identical: %v", screen.worldBillboards[0].ColorA)
	}
}

func TestWorldUniformBytesPacksMatrixAndFog(t *testing.T) {
	camera := Camera3D{
		Enabled: true,
		ViewProjection: [16]float32{
			0: 1, 5: 2, 10: 3, 15: 4,
		},
		Fog: Fog3D{
			Enabled: true,
			Near:    10,
			Far:     20,
			ColorR:  0.25,
			ColorG:  0.5,
			ColorB:  0.75,
			Factor:  0,
		},
	}
	data := worldUniformBytes(camera)
	if len(data) != 96 {
		t.Fatalf("world uniform len = %d, want 96", len(data))
	}
	if got := f32At(data, 0); got != 1 {
		t.Fatalf("matrix[0] = %v, want 1", got)
	}
	if got := f32At(data, 64); got != 10 {
		t.Fatalf("fog near = %v, want 10", got)
	}
	if got := f32At(data, 68); got != 20 {
		t.Fatalf("fog far = %v, want 20", got)
	}
	if got := f32At(data, 72); got != 1 {
		t.Fatalf("fog enabled = %v, want 1", got)
	}
	if got := f32At(data, 80); got != 0.25 {
		t.Fatalf("fog red = %v, want 0.25", got)
	}
}

func f32At(data []byte, offset int) float32 {
	return math.Float32frombits(binary.LittleEndian.Uint32(data[offset : offset+4]))
}
