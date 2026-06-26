package gamemode

import (
	"testing"

	"github.com/kivutar/goro/internal/render"
)

func TestFSAAEnabledByDefault(t *testing.T) {
	t.Setenv("GORO_FSAA", "")
	if !fsaaEnabled() {
		t.Fatal("FSAA should be enabled by default")
	}
}

func TestFSAACanBeDisabled(t *testing.T) {
	for _, value := range []string{"0", "false", "off"} {
		t.Run(value, func(t *testing.T) {
			t.Setenv("GORO_FSAA", value)
			if fsaaEnabled() {
				t.Fatalf("FSAA enabled for %q", value)
			}
		})
	}
}

func TestTriangleDrawOptionsCarryFSAA(t *testing.T) {
	t.Setenv("GORO_FSAA", "1")
	options := triangleDrawOptions(render.FilterLinear, render.AddressRepeat)
	if options.Filter != render.FilterLinear || options.Address != render.AddressRepeat {
		t.Fatalf("options filter/address = %v/%v", options.Filter, options.Address)
	}
	if !options.AntiAlias {
		t.Fatal("triangle options should enable anti-aliasing")
	}
}
