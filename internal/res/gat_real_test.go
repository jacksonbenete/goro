package res

import (
	"os"
	"testing"
)

func TestGATRealFileWhenConfigured(t *testing.T) {
	path := os.Getenv("GORO_TEST_GAT")
	if path == "" {
		t.Skip("set GORO_TEST_GAT to run against a real GAT file")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	gat, err := ParseGAT(data)
	if err != nil {
		t.Fatal(err)
	}
	if gat.Width <= 0 || gat.Height <= 0 || len(gat.Cells) != gat.Width*gat.Height {
		t.Fatalf("invalid parsed gat: %dx%d cells=%d", gat.Width, gat.Height, len(gat.Cells))
	}
}
