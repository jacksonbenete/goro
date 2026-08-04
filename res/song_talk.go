package res

import (
	"math/rand/v2"
	"strings"
)

type SongTalkKind int

const (
	SongTalkScream SongTalkKind = iota
	SongTalkFrostJoke
)

var songTalkCandidates = map[SongTalkKind][]string{
	SongTalkScream: {
		"data\\dc_scream.txt",
		"data/dc_scream.txt",
		"data\\DC_scream.txt",
		"data/DC_scream.txt",
		"dc_scream.txt",
		"DC_scream.txt",
		"data\\english\\dc_scream.txt",
		"data/english/dc_scream.txt",
		"data\\english\\DC_scream.txt",
		"data/english/DC_scream.txt",
		"english\\dc_scream.txt",
		"english/dc_scream.txt",
		"english\\DC_scream.txt",
		"english/DC_scream.txt",
	},
	SongTalkFrostJoke: {
		"data\\ba_frostjoke.txt",
		"data/ba_frostjoke.txt",
		"data\\BA_frostjoke.txt",
		"data/BA_frostjoke.txt",
		"ba_frostjoke.txt",
		"BA_frostjoke.txt",
		"data\\english\\ba_frostjoke.txt",
		"data/english/ba_frostjoke.txt",
		"data\\english\\BA_frostjoke.txt",
		"data/english/BA_frostjoke.txt",
		"english\\ba_frostjoke.txt",
		"english/ba_frostjoke.txt",
		"english\\BA_frostjoke.txt",
		"english/BA_frostjoke.txt",
	},
}

func (m *Manager) SongTalkLine(kind SongTalkKind) (string, bool) {
	m.loadSongTalk(kind)
	lines := m.songTalks[kind]
	if len(lines) == 0 {
		return "", false
	}
	return lines[rand.IntN(len(lines))], true
}

func (m *Manager) loadSongTalk(kind SongTalkKind) {
	if m.songTalksLoaded == nil {
		m.songTalksLoaded = make(map[SongTalkKind]bool)
	}
	if m.songTalksLoaded[kind] {
		return
	}
	m.songTalksLoaded[kind] = true
	if m.songTalks == nil {
		m.songTalks = make(map[SongTalkKind][]string)
	}
	candidates := songTalkCandidates[kind]
	if len(candidates) == 0 {
		return
	}
	_, data, ok := m.ReadFirst(candidates)
	if !ok {
		return
	}
	m.songTalks[kind] = parseSongTalkLines(data)
}

func parseSongTalkLines(data []byte) []string {
	lines := make([]string, 0)
	headerSkipped := false
	for _, rawLine := range strings.Split(decodeROText(data), "\n") {
		line := strings.TrimSpace(strings.TrimRight(rawLine, "\r"))
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}
		if !headerSkipped {
			headerSkipped = true
			continue
		}
		lines = append(lines, line)
	}
	return lines
}
