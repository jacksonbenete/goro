package res

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestParseSongTalkLinesSkipsHeaderAndTrimsTabs(t *testing.T) {
	got := parseSongTalkLines([]byte("SCREAM\r\n\tKyaaaa----!\r\n\tNyang-*\r\n"))
	want := []string{"Kyaaaa----!", "Nyang-*"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("song talk lines = %v, want %v", got, want)
	}
}

func TestSongTalkLineLoadsClientData(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "dc_scream.txt"), []byte("SCREAM\r\n\tLine one\r\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	manager, err := NewManager(dir)
	if err != nil {
		t.Fatal(err)
	}

	line, ok := manager.SongTalkLine(SongTalkScream)
	if !ok || line != "Line one" {
		t.Fatalf("scream line = %q ok=%t, want Line one", line, ok)
	}
}
