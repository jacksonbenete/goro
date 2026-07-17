package res

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"math/rand/v2"
	"strconv"
	"strings"

	"github.com/kivutar/goro/db"
)

var petTalkTableCandidates = []string{
	"data\\pettalktable.xml",
	"data/pettalktable.xml",
	"pettalktable.xml",
}

func (m *Manager) PetTalk(data uint32) (string, bool) {
	if !m.petTalksLoaded {
		m.loadPetTalks()
	}
	mobName, hunger, action, ok := petTalkKeys(data)
	if !ok {
		return "", false
	}
	byHunger := m.petTalks[mobName]
	if byHunger == nil {
		byHunger = m.petTalks["poring"]
	}
	byAction := byHunger[hunger]
	if byAction == nil {
		return "", false
	}
	texts := byAction[action]
	if len(texts) == 0 {
		return "", false
	}
	text := texts[0]
	if len(texts) > 1 {
		text = texts[rand.IntN(len(texts))]
	}
	return strings.TrimSpace(text), strings.TrimSpace(text) != ""
}

func (m *Manager) loadPetTalks() {
	m.petTalksLoaded = true
	m.petTalks = make(map[string]map[string]map[string][]string)
	_, data, ok := m.ReadFirst(petTalkTableCandidates)
	if !ok {
		return
	}
	table, err := parsePetTalkTable(data)
	if err == nil {
		m.petTalks = table
	}
}

func petTalkKeys(data uint32) (string, string, string, bool) {
	s := strconv.FormatUint(uint64(data), 10)
	mobID := parsePetTalkInt(s, 0, 4, 1001)
	hunger := parsePetTalkInt(s, 4, 6, 0)
	action := parsePetTalkInt(s, 6, 7, 0)
	if hunger >= 5 || action >= 11 {
		return "", "", "", false
	}
	name := strings.ToLower(db.MonsterResourceName[mobID])
	if name == "" {
		name = strings.ToLower(db.MonsterResourceName[1001])
	}
	if name == "" {
		name = "poring"
	}
	return name, petHungryText(hunger), petActionText(action), true
}

func parsePetTalkInt(s string, start, end int, fallback int) int {
	if start >= len(s) {
		return fallback
	}
	if end > len(s) {
		end = len(s)
	}
	i, err := strconv.Atoi(s[start:end])
	if err != nil {
		return fallback
	}
	return i
}

func petHungryText(state int) string {
	switch state {
	case 1:
		return "bit_hungry"
	case 2:
		return "noting"
	case 3:
		return "full"
	case 4:
		return "so_full"
	default:
		return "hungry"
	}
}

func petActionText(action int) string {
	switch action {
	case 0:
		return "feeding"
	case 1:
		return "hunting"
	case 2:
		return "danger"
	case 3:
		return "dead"
	case 4:
		return "stand"
	case 5:
		return "perfor_s"
	case 6:
		return "levelup"
	case 7:
		return "perfor_1"
	case 8:
		return "perfor_2"
	case 9:
		return "perfor_3"
	case 10:
		return "connect"
	default:
		return "stand"
	}
}

func parsePetTalkTable(data []byte) (map[string]map[string]map[string][]string, error) {
	table := make(map[string]map[string]map[string][]string)
	decoder := xml.NewDecoder(bytes.NewReader(bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})))
	decoder.CharsetReader = clientInfoCharsetReader

	var stack []string
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return table, nil
		}
		if err != nil {
			return nil, fmt.Errorf("parse pet talk table: %w", err)
		}
		switch token := token.(type) {
		case xml.StartElement:
			stack = append(stack, strings.ToLower(strings.TrimSpace(token.Name.Local)))
		case xml.EndElement:
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		case xml.CharData:
			if len(stack) != 4 || stack[0] != "monster_talk_table" {
				continue
			}
			text := strings.TrimSpace(string(token))
			if text == "" {
				continue
			}
			monster, hunger, action := stack[1], stack[2], stack[3]
			if table[monster] == nil {
				table[monster] = make(map[string]map[string][]string)
			}
			if table[monster][hunger] == nil {
				table[monster][hunger] = make(map[string][]string)
			}
			table[monster][hunger][action] = append(table[monster][hunger][action], text)
		}
	}
}
