package res

import "testing"

func TestParsePAL(t *testing.T) {
	data := make([]byte, 1024)
	data[4] = 10
	data[5] = 20
	data[6] = 30
	data[7] = 40

	palette, err := ParsePAL(data)
	if err != nil {
		t.Fatal(err)
	}
	if got := palette[1]; got != [4]byte{10, 20, 30, 40} {
		t.Fatalf("palette[1] = %v", got)
	}
}

func TestParsePALRejectsTruncatedData(t *testing.T) {
	if _, err := ParsePAL(make([]byte, 1023)); err == nil {
		t.Fatal("expected truncated palette error")
	}
}
