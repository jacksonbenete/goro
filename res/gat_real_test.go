package res

import "testing"

func TestGATRealFileWhenConfigured(t *testing.T) {
	data := readRealDataFile(t, "data\\geffen_in.gat")
	gat, err := ParseGAT(data)
	if err != nil {
		t.Fatal(err)
	}
	if gat.Width <= 0 || gat.Height <= 0 || len(gat.Cells) != gat.Width*gat.Height {
		t.Fatalf("invalid parsed gat: %dx%d cells=%d", gat.Width, gat.Height, len(gat.Cells))
	}
}
