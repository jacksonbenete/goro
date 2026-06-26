package res

import "testing"

func TestDecodeTGAUncompressedTruecolor(t *testing.T) {
	data := []byte{
		0, 0, 2,
		0, 0, 0, 0, 0,
		0, 0, 0, 0,
		2, 0, 1, 0,
		24, 0x20,
		0, 0, 255,
		0, 255, 0,
	}
	img, err := decodeTGA(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := img.At(0, 0); got != nil {
		r, g, b, a := got.RGBA()
		if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 || a>>8 != 255 {
			t.Fatalf("pixel 0 = rgba(%d,%d,%d,%d), want red", r>>8, g>>8, b>>8, a>>8)
		}
	}
	r, g, b, a := img.At(1, 0).RGBA()
	if r>>8 != 0 || g>>8 != 255 || b>>8 != 0 || a>>8 != 255 {
		t.Fatalf("pixel 1 = rgba(%d,%d,%d,%d), want green", r>>8, g>>8, b>>8, a>>8)
	}
}
