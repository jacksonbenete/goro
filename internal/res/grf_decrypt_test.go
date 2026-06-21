package res

import (
	"bytes"
	"testing"
)

func TestDecryptGRFHeaderOnlyTouchesFirstTwentyBlocks(t *testing.T) {
	data := makeSequentialBytes(256)
	originalTail := append([]byte(nil), data[160:]...)

	decryptGRFHeader(data, uint32(len(data)))

	if bytes.Equal(data[:160], makeSequentialBytes(160)) {
		t.Fatal("expected first twenty blocks to change")
	}
	if !bytes.Equal(data[160:], originalTail) {
		t.Fatal("expected bytes after first twenty blocks to stay unchanged")
	}
}

func TestDecryptGRFFullDiffersAfterHeader(t *testing.T) {
	headerOnly := makeSequentialBytes(320)
	full := makeSequentialBytes(320)

	decryptGRFHeader(headerOnly, uint32(len(headerOnly)))
	decryptGRFFull(full, uint32(len(full)), 100)

	if !bytes.Equal(headerOnly[:160], full[:160]) {
		t.Fatal("first twenty blocks should match")
	}
	if bytes.Equal(headerOnly[160:], full[160:]) {
		t.Fatal("full decrypt should continue processing after header blocks")
	}
}

func makeSequentialBytes(size int) []byte {
	out := make([]byte, size)
	for i := range out {
		out[i] = byte(i)
	}
	return out
}
