package game

import (
	"testing"

	"github.com/kivutar/goro/render"
)

func TestTriangleDrawOptionsEnableDepthTest(t *testing.T) {
	options := triangleDrawOptions(render.FilterLinear, render.AddressRepeat)
	if options.Filter != render.FilterLinear || options.Address != render.AddressRepeat {
		t.Fatalf("options filter/address = %v/%v", options.Filter, options.Address)
	}
	if !options.DepthTest {
		t.Fatal("triangle options should enable depth testing")
	}
}

func TestGroundTextureDrawOptionsClampTileEdges(t *testing.T) {
	options := groundTextureDrawOptions()
	if options.Filter != render.FilterLinear {
		t.Fatalf("ground filter = %v, want linear", options.Filter)
	}
	if options.Address != render.AddressClampToZero {
		t.Fatalf("ground address = %v, want clamp", options.Address)
	}
	if !options.DepthTest || !options.DepthWrite {
		t.Fatalf("ground depth flags test=%t write=%t, want both true", options.DepthTest, options.DepthWrite)
	}
}
