package gamemode

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
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
	options := triangleDrawOptions(ebiten.FilterLinear, ebiten.AddressRepeat)
	if options.Filter != ebiten.FilterLinear || options.Address != ebiten.AddressRepeat {
		t.Fatalf("options filter/address = %v/%v", options.Filter, options.Address)
	}
	if !options.AntiAlias {
		t.Fatal("triangle options should enable anti-aliasing")
	}
}
