package res

import "fmt"

const (
	playerHumanSpriteRoot = "data\\sprite\\\xC0\xCE\xB0\xA3\xC1\xB7\\"
	playerAccessoryRoot   = "data\\sprite\\\xBE\xC7\xBC\xBC\xBB\xE7\xB8\xAE\\"
	playerShieldRoot      = "data\\sprite\\\xB9\xE6\xC6\xD0\\"
	playerPaletteRoot     = "data\\palette\\"
	playerIMFRoot         = "data\\imf\\"
	playerBodyDir         = "\xB8\xF6\xC5\xEB"
	playerHeadDir         = "\xB8\xD3\xB8\xAE\xC5\xEB"
	playerPaletteBodyDir  = "\xB8\xF6"
	playerPaletteHeadDir  = "\xB8\xD3\xB8\xAE"
	playerPaletteHeadFile = "\xB8\xD3\xB8\xAE"
	playerFemaleSex       = "\xBF\xA9"
	playerMaleSex         = "\xB3\xB2"
	weaponLightSuffix     = "\xB0\xCB\xB1\xA4"
)

var accessoryLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\accessoryid.lub",
	"data\\lua files\\datainfo\\accessoryid.lub",
	"lua files\\datainfo\\accessoryid.lub",
}

var accessoryNameLuaCandidates = []string{
	"data\\luafiles514\\lua files\\datainfo\\accname.lub",
	"data\\lua files\\datainfo\\accname.lub",
	"lua files\\datainfo\\accname.lub",
}

var playerJobTokens = map[int]string{
	0:  "\xC3\xCA\xBA\xB8\xC0\xDA",
	1:  "\xB0\xCB\xBB\xE7",
	2:  "\xB8\xB6\xB9\xFD\xBB\xE7",
	3:  "\xB1\xC3\xBC\xF6",
	4:  "\xBC\xBA\xC1\xF7\xC0\xDA",
	5:  "\xBB\xF3\xC0\xCE",
	6:  "\xB5\xB5\xB5\xCF",
	7:  "\xB1\xE2\xBB\xE7",
	8:  "\xC7\xC1\xB8\xAE\xBD\xBA\xC6\xAE",
	9:  "\xC0\xA7\xC0\xFA\xB5\xE5",
	10: "\xC1\xA6\xC3\xB6\xB0\xF8",
	11: "\xC7\xE5\xC5\xCD",
	12: "\xBE\xEE\xBC\xBC\xBD\xC5",
	14: "\xC5\xA9\xB7\xE7\xBC\xBC\xC0\xCC\xB4\xF5",
	15: "\xB8\xF9\xC5\xA9",
	16: "\xBC\xBC\xC0\xCC\xC1\xF6",
	17: "\xB7\xCE\xB1\xD7",
	18: "\xBF\xAC\xB1\xDD\xBC\xFA\xBB\xE7",
	19: "\xB9\xD9\xB5\xE5",
}

func PlayerBodyResourceCandidates(job int, sex byte, extension string) []string {
	sexToken := PlayerSexToken(sex)
	tokens := []string{PlayerJobToken(job)}
	if tokens[0] != playerJobTokens[0] {
		tokens = append(tokens, playerJobTokens[0])
	}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		out = append(out, fmt.Sprintf("%s%s\\%s\\%s_%s.%s", playerHumanSpriteRoot, playerBodyDir, sexToken, token, sexToken, extension))
	}
	return out
}

func PlayerBodyResourcePath(job int, sex byte, extension string) string {
	sexToken := PlayerSexToken(sex)
	return fmt.Sprintf("%s%s\\%s\\%s_%s.%s", playerHumanSpriteRoot, playerBodyDir, sexToken, PlayerJobToken(job), sexToken, extension)
}

func PlayerHeadResourceCandidates(job int, head int, sex byte, extension string) []string {
	sexToken := PlayerSexToken(sex)
	head = NormalizePlayerHead(head, job)
	return []string{
		fmt.Sprintf("%s%s\\%s\\%d_%s.%s", playerHumanSpriteRoot, playerHeadDir, sexToken, head, sexToken, extension),
	}
}

func PlayerAccessoryResourceCandidates(job int, head int, sex byte, viewID int, resourceName string, extension string) []string {
	if viewID == 185 {
		return PlayerHeadResourceCandidates(job, head, sex, extension)
	}
	if viewID <= 0 || resourceName == "" {
		return nil
	}
	sexToken := PlayerSexToken(sex)
	separator := "_"
	if resourceName[0] == '_' {
		separator = ""
	}
	return []string{
		fmt.Sprintf("%s%s\\%s%s%s.%s", playerAccessoryRoot, sexToken, sexToken, separator, resourceName, extension),
	}
}

func (m *Manager) AccessoryResourceName(viewID int) (string, bool) {
	if viewID <= 0 {
		return "", false
	}
	if !m.accessoryNamesLoaded {
		m.loadAccessoryResourceNames()
	}
	name, ok := m.accessoryNames[viewID]
	return name, ok && name != ""
}

func (m *Manager) loadAccessoryResourceNames() {
	m.accessoryNamesLoaded = true
	m.accessoryNames = make(map[int]string)
	globals := make(map[string]luaValue)
	for _, candidates := range [][]string{accessoryLuaCandidates, accessoryNameLuaCandidates} {
		_, data, ok := m.ReadFirst(candidates)
		if !ok {
			return
		}
		if err := executeLua51Bytecode(data, globals); err != nil {
			return
		}
	}
	table := globals["AccNameTable"]
	if table.kind != luaTable {
		return
	}
	for key, value := range table.table {
		index, ok := key.(int)
		if !ok || value.kind != luaString || value.str == "" {
			continue
		}
		m.accessoryNames[index] = value.str
	}
}

func PlayerIMFResourceCandidates(job int, sex byte) []string {
	sexToken := PlayerSexToken(sex)
	tokens := []string{PlayerJobToken(job)}
	if tokens[0] != playerJobTokens[0] {
		tokens = append(tokens, playerJobTokens[0])
	}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		out = append(out, fmt.Sprintf("%s%s_%s.imf", playerIMFRoot, token, sexToken))
	}
	return out
}

func PlayerWeaponOverlayResourceCandidates(job int, sex byte, weaponValue int, secondLayer bool, extension string) []string {
	if weaponValue <= 0 {
		return nil
	}
	weaponType := PlayerWeaponType(weaponValue)
	if weaponType <= 0 {
		return nil
	}
	if secondLayer && !PlayerWeaponOverlaySupportsSecondLayer(weaponType) {
		return nil
	}
	token := PlayerWeaponOverlayToken(weaponType)
	if token == "" {
		return nil
	}
	sexToken := PlayerSexToken(sex)
	jobToken := PlayerJobToken(job)
	if secondLayer {
		return []string{
			fmt.Sprintf("%s%s\\%s_%s_%s_%s.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, token, weaponLightSuffix, extension),
		}
	}
	return []string{
		fmt.Sprintf("%s%s\\%s_%s_%s.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, token, extension),
	}
}

func PlayerShieldOverlayResourceCandidates(job int, sex byte, shield int, extension string) []string {
	token := PlayerShieldToken(shield)
	if token == "" {
		return nil
	}
	sexToken := PlayerSexToken(sex)
	jobToken := PlayerJobToken(job)
	return []string{
		fmt.Sprintf("%s%s\\%s_%s_%s.%s", playerShieldRoot, jobToken, jobToken, sexToken, token, extension),
	}
}

func PlayerBodyPaletteResourceCandidates(job int, sex byte, palette int, extension string) []string {
	if palette <= 0 {
		return nil
	}
	sexToken := PlayerSexToken(sex)
	tokens := []string{PlayerJobToken(job)}
	if tokens[0] != playerJobTokens[0] {
		tokens = append(tokens, playerJobTokens[0])
	}
	out := make([]string, 0, len(tokens))
	for _, token := range tokens {
		out = append(out, fmt.Sprintf("%s%s\\%s_%s_%d.%s", playerPaletteRoot, playerPaletteBodyDir, token, sexToken, palette, extension))
	}
	return out
}

func PlayerHeadPaletteResourceCandidates(job int, head int, sex byte, palette int, extension string) []string {
	if palette <= 0 {
		return nil
	}
	sexToken := PlayerSexToken(sex)
	head = NormalizePlayerHead(head, job)
	return []string{
		fmt.Sprintf("%s%s\\%s%d_%s_%d.%s", playerPaletteRoot, playerPaletteHeadDir, playerPaletteHeadFile, head, sexToken, palette, extension),
	}
}

func PlayerSexToken(sex byte) string {
	if sex != 0 {
		return playerMaleSex
	}
	return playerFemaleSex
}

func PlayerSexLabel(sex byte) string {
	if sex != 0 {
		return "male"
	}
	return "female"
}

func PlayerJobToken(job int) string {
	if token, ok := playerJobTokens[job]; ok {
		return token
	}
	return playerJobTokens[0]
}

func HasPlayerJobToken(job int) bool {
	_, ok := playerJobTokens[job]
	return ok
}

func NormalizePlayerHead(head int, job int) int {
	if head == 0 {
		switch job {
		case 1:
			return 2
		case 2:
			return 3
		case 3:
			return 4
		case 4:
			return 5
		case 5:
			return 6
		case 6:
			return 7
		default:
			return 1
		}
	}
	if head < 1 || head > 25 {
		return 13
	}
	return head
}

func PlayerWeaponType(weaponValue int) int {
	if weaponValue <= 0 {
		return 0
	}
	if weaponValue <= 31 {
		return weaponValue
	}
	switch {
	case weaponValue >= 1101 && weaponValue <= 1119:
		return 2
	case weaponValue >= 1201 && weaponValue <= 1250:
		return 1
	case weaponValue >= 1301 && weaponValue <= 1399:
		return 9
	case weaponValue >= 1401 && weaponValue <= 1499:
		return 4
	case weaponValue >= 1501 && weaponValue <= 1599:
		return 8
	case weaponValue >= 1601 && weaponValue <= 1699:
		return 10
	case weaponValue >= 1701 && weaponValue <= 1799:
		return 11
	case weaponValue >= 1801 && weaponValue <= 1899:
		return 12
	case weaponValue >= 1901 && weaponValue <= 1949:
		return 13
	case weaponValue >= 1950 && weaponValue <= 1999:
		return 14
	default:
		return 0
	}
}

func PlayerWeaponOverlaySupportsSecondLayer(weaponType int) bool {
	switch weaponType {
	case 1, 2, 3, 4, 5, 6, 7, 16, 17, 18, 25, 26, 27, 28, 29, 30:
		return true
	default:
		return false
	}
}

func PlayerWeaponOverlayToken(weaponType int) string {
	switch weaponType {
	case 1:
		return "\xB4\xDC\xB0\xCB"
	case 2, 3:
		return "\xB0\xCB"
	case 4, 5:
		return "\xC3\xA2"
	case 6, 7:
		return "\xB5\xB5\xB3\xA2"
	case 8, 9:
		return "\xC5\xAC\xB7\xB4"
	case 10, 23:
		return "\xB7\xCE\xB5\xE5"
	case 11:
		return "\xC8\xB0"
	case 12:
		return "\xB3\xCA\xC5\xAC"
	case 13:
		return "\xBE\xC7\xB1\xE2"
	case 14:
		return "\xC3\xA4\xC2\xEF"
	case 15:
		return "\xC3\xA5"
	case 16:
		return "\xC4\xAB\xC5\xB8\xB8\xA3_\xC4\xAB\xC5\xB8\xB8\xA3"
	case 17:
		return "\xB1\xC7\xC3\xD1"
	case 18:
		return "\xB6\xF3\xC0\xCC\xC7\xC3"
	case 19:
		return "\xB1\xE2\xB0\xFC\xC3\xD1"
	case 20:
		return "\xBC\xA6\xB0\xC7"
	case 22:
		return "\xBC\xF6\xB8\xAE\xB0\xCB"
	case 25:
		return "\xB4\xDC\xB0\xCB_\xB4\xDC\xB0\xCB"
	case 26:
		return "\xB0\xCB_\xB0\xCB"
	case 27:
		return "\xB5\xB5\xB3\xA2_\xB5\xB5\xB3\xA2"
	case 28:
		return "\xB4\xDC\xB0\xCB_\xB0\xCB"
	case 29:
		return "\xB4\xDC\xB0\xCB_\xB5\xB5\xB3\xA2"
	case 30:
		return "\xB0\xCB_\xB5\xB5\xB3\xA2"
	default:
		return ""
	}
}

func PlayerShieldToken(shield int) string {
	if shield <= 0 {
		return ""
	}
	viewID := shield
	if shield > 4 {
		switch shield {
		case 2101, 2102:
			viewID = 1
		case 2103, 2104:
			viewID = 2
		case 2105, 2106:
			viewID = 3
		case 2107, 2108, 2110, 2111:
			viewID = 4
		default:
			return ""
		}
	}
	switch viewID {
	case 1:
		return "guard"
	case 2:
		return "buckler"
	case 3:
		return "shield"
	case 4:
		return "mirrorshield"
	default:
		return ""
	}
}
