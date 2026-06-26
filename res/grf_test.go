package res

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestGRFReadPlainCompressedEntry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "data.grf")

	if err := writeTestGRF(path, "data\\hello.txt", []byte("hello world")); err != nil {
		t.Fatal(err)
	}

	grf, err := OpenGRF(path)
	if err != nil {
		if errors.Is(err, ErrGRFUnsupportedVersion) {
			t.Skip(err)
		}
		t.Fatal(err)
	}
	defer grf.Close()

	if grf.Count() != 1 {
		t.Fatalf("count = %d", grf.Count())
	}

	data, err := grf.ReadFile("data/hello.txt")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "hello world" {
		t.Fatalf("data = %q", string(data))
	}
}

func TestNormalizeGRFNamePreservesNonUTF8Bytes(t *testing.T) {
	input := string([]byte{'D', 'A', 'T', 'A', '/', 0xc7, 0xca, '/', 'T', 'E', 'X', 0xf8, '.', 'B', 'M', 'P'})
	got := normalizeGRFName(input)
	want := string([]byte{'d', 'a', 't', 'a', '\\', 0xc7, 0xca, '\\', 't', 'e', 'x', 0xf8, '.', 'b', 'm', 'p'})
	if got != want {
		t.Fatalf("normalizeGRFName bytes = % x, want % x", []byte(got), []byte(want))
	}
	if bytes.Contains([]byte(got), []byte{0xef, 0xbf, 0xbd}) {
		t.Fatalf("normalizeGRFName introduced UTF-8 replacement bytes: % x", []byte(got))
	}
}

func TestGRFRealArchiveWhenConfigured(t *testing.T) {
	grf := realDataArchive(t)
	name := "prontera.gat"

	if grf.Count() == 0 {
		t.Fatal("archive has no entries")
	}
	if !grf.Has(name) {
		matches := grf.NamesWithSuffix(name)
		if len(matches) == 0 {
			t.Skipf("%s not present in %s", name, grf.Path())
		}
		name = matches[0]
	}
	data, err := grf.ReadFile(name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	if len(data) == 0 {
		t.Fatalf("read %s returned empty data", name)
	}
}

func writeTestGRF(path, name string, content []byte) error {
	compressedFile := zlibBytes(content)

	var table bytes.Buffer
	table.WriteString(name)
	table.WriteByte(0)
	writeU32(&table, uint32(len(compressedFile)))
	writeU32(&table, uint32(len(compressedFile)))
	writeU32(&table, uint32(len(content)))
	table.WriteByte(0x01)
	writeU32(&table, 0)

	compressedTable := zlibBytes(table.Bytes())

	var out bytes.Buffer
	header := make([]byte, grfHeaderSize)
	copy(header[:15], []byte("Master of Magic"))
	binary.LittleEndian.PutUint32(header[30:34], uint32(len(compressedFile)))
	binary.LittleEndian.PutUint32(header[34:38], 0)
	binary.LittleEndian.PutUint32(header[38:42], 8)
	binary.LittleEndian.PutUint32(header[42:46], grfVersion200)
	out.Write(header)
	out.Write(compressedFile)
	writeU32(&out, uint32(len(compressedTable)))
	writeU32(&out, uint32(len(table.Bytes())))
	out.Write(compressedTable)

	return os.WriteFile(path, out.Bytes(), 0o644)
}

func zlibBytes(data []byte) []byte {
	var out bytes.Buffer
	writer := zlib.NewWriter(&out)
	_, _ = writer.Write(data)
	_ = writer.Close()
	return out.Bytes()
}

func writeU32(buf *bytes.Buffer, value uint32) {
	var tmp [4]byte
	binary.LittleEndian.PutUint32(tmp[:], value)
	buf.Write(tmp[:])
}
