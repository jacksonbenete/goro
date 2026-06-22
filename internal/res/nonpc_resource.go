package res

import (
	"fmt"
	"strings"
)

const legacyMonsterSpriteRoot = "data\\sprite\\\xB8\xF3\xBD\xBA\xC5\xCD\\"

var npcIdentityLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\npcidentity.lub",
	"data\\lua files\\datainfo\\npcidentity.lub",
	"lua files\\datainfo\\npcidentity.lub",
}

var jobNameLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\jobname.lub",
	"data\\lua files\\datainfo\\jobname.lub",
	"lua files\\datainfo\\jobname.lub",
}

var fallbackJobResourceNames = map[int]string{
	45:   "1_ETC_01",
	46:   "1_ETC_01",
	47:   "1_M_01",
	48:   "1_M_02",
	49:   "1_M_03",
	50:   "1_M_04",
	66:   "1_F_01",
	67:   "1_F_02",
	68:   "1_F_03",
	69:   "1_F_04",
	81:   "4_DOG01",
	82:   "4_KID01",
	83:   "4_M_01",
	84:   "4_M_02",
	85:   "4_M_03",
	86:   "4_M_04",
	1001: "scorpion",
	1002: "poring",
	1004: "hornet",
	1005: "familiar",
	1007: "fabre",
	1008: "pupa",
	1009: "condor",
	1010: "willow",
	1011: "chontchon",
	1013: "wolf",
	1014: "spore",
	1015: "zombie",
	1016: "archer_skeleton",
	1018: "creamie",
	1020: "mandragora",
	1023: "orc_warrior",
	1024: "worm_tail",
	1025: "snake",
	1026: "munak",
	1028: "soldier_skeleton",
}

func (m *Manager) JobResourceName(job int) (string, bool) {
	if !m.jobResourceNamesLoaded {
		m.loadJobResourceNames()
	}
	name, ok := m.jobResourceNames[job]
	return name, ok && name != ""
}

func (m *Manager) loadJobResourceNames() {
	m.jobResourceNamesLoaded = true
	m.jobResourceNames = make(map[int]string, len(fallbackJobResourceNames))
	for job, name := range fallbackJobResourceNames {
		m.jobResourceNames[job] = name
	}

	globals := make(map[string]luaValue)
	for _, candidates := range [][]string{npcIdentityLuaCandidates, jobNameLuaCandidates} {
		_, data, ok := m.ReadFirst(candidates)
		if !ok {
			continue
		}
		if err := executeLua51Bytecode(data, globals); err != nil {
			continue
		}
	}
	table := globals["JobNameTable"]
	if table.kind != luaTable {
		return
	}
	for key, value := range table.table {
		index, ok := key.(int)
		if !ok || value.kind != luaString || value.str == "" {
			continue
		}
		m.jobResourceNames[index] = value.str
	}
}

func NonPCSpriteResourceCandidates(job int, resourceName string, extension string) []string {
	if resourceName == "" {
		return nil
	}
	name := strings.TrimSuffix(resourceName, "."+extension)
	lowerName := strings.ToLower(name)
	seen := make(map[string]struct{})
	var out []string
	add := func(path string) {
		if _, ok := seen[path]; ok {
			return
		}
		seen[path] = struct{}{}
		out = append(out, path)
	}
	addStem := func(root string) {
		add(fmt.Sprintf("%s%s.%s", root, name, extension))
		if lowerName != name {
			add(fmt.Sprintf("%s%s.%s", root, lowerName, extension))
		}
	}

	if job >= 6001 && job <= 6047 {
		if job >= 6017 && job <= 6046 {
			addStem("data\\sprite\\mercenary\\")
		} else {
			addStem("data\\sprite\\homun\\")
		}
		return out
	}
	if job >= 1000 {
		addStem("data\\sprite\\monster\\")
		addStem(legacyMonsterSpriteRoot)
		addStem("data\\sprite\\")
		return out
	}
	addStem("data\\sprite\\NPC\\")
	addStem("data\\sprite\\npc\\")
	return out
}
