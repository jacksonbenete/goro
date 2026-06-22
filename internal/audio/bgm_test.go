package audio

import "testing"

func TestParseMP3NameTable(t *testing.T) {
	table := ParseMP3NameTable([]byte("prontera.rsw#bgm\\08.mp3#\r\nDATA\\geffen.rsw#09.mp3#\ninvalid\n"))
	if got := table["prontera.rsw"]; got != "bgm\\08.mp3" {
		t.Fatalf("prontera bgm = %q", got)
	}
	if got := table["geffen.rsw"]; got != "bgm\\09.mp3" {
		t.Fatalf("geffen bgm = %q", got)
	}
}

func TestCanonicalMapName(t *testing.T) {
	for _, tc := range []struct {
		in   string
		want string
	}{
		{"prontera", "prontera.rsw"},
		{"prontera.gat", "prontera.rsw"},
		{"data\\prontera.rsw", "prontera.rsw"},
		{"data/prontera.rsw", "prontera.rsw"},
	} {
		if got := canonicalMapName(tc.in); got != tc.want {
			t.Fatalf("canonicalMapName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestAudioPathCandidatesPreferBGMDirectory(t *testing.T) {
	got := audioPathCandidates("bgm\\01.mp3")
	want := "BGM\\01.mp3"
	for _, candidate := range got {
		if candidate == want {
			return
		}
	}
	t.Fatalf("candidates %v did not include %q", got, want)
}

func TestResamplePCM16StereoUpsamplePreservesConstantSignal(t *testing.T) {
	var src []byte
	for range 64 {
		src = appendPCMFrame(src, 10000, -12000)
	}

	got, err := resamplePCM16Stereo(src, 22050, 44100)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(src)*2 {
		t.Fatalf("resampled len = %d, want %d", len(got), len(src)*2)
	}
	for frame := 16; frame < len(got)/4-16; frame++ {
		left, right := readPCM16(got, frame, 0), readPCM16(got, frame, 1)
		if absInt16Diff(left, 10000) > 1 || absInt16Diff(right, -12000) > 1 {
			t.Fatalf("frame %d = %d,%d want near 10000,-12000", frame, left, right)
		}
	}
}

func TestMonoPCM16ToStereo(t *testing.T) {
	src := []byte{0x34, 0x12, 0x00, 0x80}
	got, err := monoPCM16ToStereo(src)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 8 {
		t.Fatalf("stereo len = %d, want 8", len(got))
	}
	if readPCM16(got, 0, 0) != 0x1234 || readPCM16(got, 0, 1) != 0x1234 {
		t.Fatalf("frame 0 = %d,%d", readPCM16(got, 0, 0), readPCM16(got, 0, 1))
	}
	if readPCM16(got, 1, 0) != -32768 || readPCM16(got, 1, 1) != -32768 {
		t.Fatalf("frame 1 = %d,%d", readPCM16(got, 1, 0), readPCM16(got, 1, 1))
	}
}

func appendPCMFrame(pcm []byte, left, right int16) []byte {
	frame := make([]byte, 4)
	writePCM16(frame, 0, 0, left)
	writePCM16(frame, 0, 1, right)
	return append(pcm, frame...)
}

func absInt16Diff(a, b int16) int {
	diff := int(a) - int(b)
	if diff < 0 {
		return -diff
	}
	return diff
}
