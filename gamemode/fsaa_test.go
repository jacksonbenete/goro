package gamemode

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
