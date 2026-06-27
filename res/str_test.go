package res

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func TestParseSTR(t *testing.T) {
	var data bytes.Buffer
	data.WriteString("STRM")
	writeSTRU32(&data, 0x94)
	writeSTRU32(&data, 60)
	writeSTRU32(&data, 12)
	writeSTRU32(&data, 1)
	data.Write(make([]byte, 16))
	writeSTRI32(&data, 1)
	texture := make([]byte, 128)
	copy(texture, []byte("spark.tga"))
	data.Write(texture)
	writeSTRI32(&data, 1)
	writeSTRI32(&data, 3)
	writeSTRU32(&data, 0)
	writeSTRF32(&data, 320)
	writeSTRF32(&data, 310)
	for i := 0; i < 8; i++ {
		writeSTRF32(&data, float32(i))
	}
	for i := 0; i < 8; i++ {
		writeSTRF32(&data, float32(i+10))
	}
	writeSTRF32(&data, 0)
	writeSTRU32(&data, 1)
	writeSTRF32(&data, 1)
	writeSTRF32(&data, 512)
	for _, value := range []float32{255, 128, 64, 255} {
		writeSTRF32(&data, value)
	}
	writeSTRU32(&data, 5)
	writeSTRU32(&data, 2)
	writeSTRU32(&data, 0)

	str, err := ParseSTR(data.Bytes(), "sub\\")
	if err != nil {
		t.Fatal(err)
	}
	if str.FPS != 60 || str.MaxKey != 12 || len(str.Layers) != 1 {
		t.Fatalf("str = %+v", str)
	}
	layer := str.Layers[0]
	if len(layer.Textures) != 1 || layer.Textures[0] != `data\texture\effect\sub\spark.tga` {
		t.Fatalf("textures = %#v", layer.Textures)
	}
	if len(layer.Animations) != 1 {
		t.Fatalf("animations = %d", len(layer.Animations))
	}
	anim := layer.Animations[0]
	if anim.Frame != 3 || anim.Pos != ([2]float32{320, 310}) || anim.XY[0] != 10 || anim.Angle != 180 || anim.Color[1] != float32(128.0/255.0) || anim.SrcAlpha != 5 || anim.DestAlpha != 2 {
		t.Fatalf("anim = %+v", anim)
	}
}

func TestRealPotionAndProvokeSTRExactResources(t *testing.T) {
	manager := realDataManager(t)
	for _, path := range []string{
		`data\texture\effect\provoke.str`,
		"data\\texture\\effect\\\xbb\xa1\xb0\xa3\xc6\xf7\xbc\xc7.str",
	} {
		data, err := manager.ReadFileExact(path)
		if err != nil {
			t.Fatalf("read exact %s: %v", path, err)
		}
		str, err := ParseSTR(data, "")
		if err != nil {
			t.Fatalf("parse %s: %v", path, err)
		}
		foundTexture := false
		for _, layer := range str.Layers {
			for _, texture := range layer.Textures {
				if texture == "" {
					continue
				}
				if _, err := manager.ReadFileExact(texture); err != nil {
					t.Fatalf("read exact texture %s from %s: %v", texture, path, err)
				}
				foundTexture = true
				break
			}
			if foundTexture {
				break
			}
		}
		if !foundTexture {
			t.Fatalf("%s references no textures", path)
		}
	}
}

func TestRealBashEffectTextures(t *testing.T) {
	manager := realDataManager(t)
	for _, name := range []string{"alpha_down", "alpha_center", "ring_yellow", "\xb4\xeb\xc6\xf8\xb9\xdf"} {
		if _, source, err := LoadImageExact(manager, EffectTextureCandidates(name)); err != nil {
			t.Fatalf("load exact effect texture %s: %v", name, err)
		} else if source == "" {
			t.Fatalf("load exact effect texture %s returned empty source", name)
		}
	}
	if _, _, err := LoadImageExact(manager, []string{`data\texture\effect\endure.tga`, `data/texture/effect/endure.tga`}); err != nil {
		t.Fatalf("load exact endure texture: %v", err)
	}
}

func writeSTRI32(out *bytes.Buffer, value int32) {
	writeSTRU32(out, uint32(value))
}

func writeSTRF32(out *bytes.Buffer, value float32) {
	writeSTRU32(out, math.Float32bits(value))
}

func writeSTRU32(out *bytes.Buffer, value uint32) {
	var buf [4]byte
	binary.LittleEndian.PutUint32(buf[:], value)
	out.Write(buf[:])
}
