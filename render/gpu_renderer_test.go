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
		"screen": screenShaderWGSL,
		"world":  worldShaderWGSL,
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
		"screen": screenShaderWGSL,
		"world":  worldShaderWGSL,
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
