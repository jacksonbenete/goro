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
	playerWeaponTypeMax   = 103
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
	weaponType := PlayerWeaponOverlayTypeForJob(job, PlayerWeaponType(weaponValue), false)
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
	out := make([]string, 0, 2)
	if secondLayer {
		out = append(out,
			fmt.Sprintf("%s%s\\%s_%s_%d_%s.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, weaponValue&0xFFFF, weaponLightSuffix, extension),
			fmt.Sprintf("%s%s\\%s_%s_%s_%s.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, token, weaponLightSuffix, extension),
		)
		return out
	}
	out = append(out,
		fmt.Sprintf("%s%s\\%s_%s_%d.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, weaponValue&0xFFFF, extension),
		fmt.Sprintf("%s%s\\%s_%s_%s.%s", playerHumanSpriteRoot, jobToken, jobToken, sexToken, token, extension),
	)
	return out
}

func (m *Manager) PlayerWeaponViewID(weaponValue int) int {
	return PlayerWeaponViewID(m, weaponValue)
}

func PlayerWeaponViewID(manager *Manager, weaponValue int) int {
	if weaponValue <= 0 {
		return 0
	}
	if weaponValue < playerWeaponTypeMax {
		return weaponValue
	}
	if manager != nil {
		if classNum, ok := manager.ItemClassNum(weaponValue); ok {
			return classNum
		}
	}
	return PlayerWeaponType(weaponValue)
}

func NormalizePlayerWeaponShield(weapon, shield int) (int, int) {
	if weapon <= 0 && shield > 0 && PlayerWeaponType(shield) > 0 && PlayerShieldToken(shield) == "" {
		return shield, 0
	}
	return weapon, shield
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
	if weaponType, ok := playerWeaponTypeExpansion[weaponValue]; ok {
		return weaponType
	}
	if weaponValue < playerWeaponTypeMax {
		return weaponValue
	}
	switch {
	case weaponValue >= 1100 && weaponValue <= 1149:
		return 2
	case weaponValue >= 1150 && weaponValue <= 1199:
		return 3
	case weaponValue >= 1200 && weaponValue <= 1249:
		return 1
	case weaponValue >= 1250 && weaponValue <= 1299:
		return 16
	case weaponValue >= 1300 && weaponValue <= 1349:
		return 6
	case weaponValue >= 1350 && weaponValue <= 1399:
		return 7
	case weaponValue >= 1400 && weaponValue <= 1449:
		return 4
	case weaponValue >= 1450 && weaponValue <= 1499:
		return 5
	case weaponValue >= 1500 && weaponValue <= 1549:
		return 8
	case weaponValue >= 1550 && weaponValue <= 1599:
		return 15
	case weaponValue >= 1600 && weaponValue <= 1699:
		return 10
	case weaponValue >= 1700 && weaponValue <= 1749:
		return 11
	case weaponValue >= 1800 && weaponValue <= 1849:
		return 12
	case weaponValue >= 1900 && weaponValue <= 1949:
		return 13
	case weaponValue >= 1950 && weaponValue <= 1999:
		return 14
	case weaponValue >= 13000 && weaponValue <= 13099:
		return 1
	case weaponValue >= 13100 && weaponValue <= 13149:
		return 17
	case weaponValue >= 13150 && weaponValue <= 13199:
		return 18
	case weaponValue >= 13300 && weaponValue <= 13399:
		return 22
	case weaponValue >= 13400 && weaponValue <= 13499:
		return 2
	case weaponValue >= 18100 && weaponValue <= 18499:
		return 11
	case weaponValue >= 20000 && weaponValue <= 20999:
		return 23
	case weaponValue >= 21000 && weaponValue <= 21999:
		return 3
	default:
		return 0
	}
}

var playerWeaponTypeExpansion = map[int]int{
	31:  1,
	32:  1,
	33:  1,
	34:  1,
	35:  1,
	36:  1,
	37:  1,
	38:  1,
	39:  2,
	40:  2,
	41:  2,
	42:  2,
	43:  2,
	44:  2,
	45:  2,
	46:  2,
	47:  2,
	48:  3,
	49:  3,
	50:  3,
	51:  3,
	52:  4,
	53:  4,
	54:  4,
	55:  4,
	56:  4,
	57:  4,
	58:  6,
	59:  6,
	60:  6,
	61:  6,
	62:  8,
	63:  8,
	64:  8,
	65:  8,
	66:  8,
	67:  8,
	68:  8,
	69:  10,
	70:  10,
	71:  10,
	72:  10,
	73:  11,
	74:  11,
	75:  11,
	76:  11,
	77:  11,
	78:  12,
	79:  12,
	80:  12,
	81:  12,
	82:  12,
	83:  12,
	84:  12,
	85:  12,
	86:  14,
	87:  14,
	88:  14,
	89:  15,
	90:  15,
	91:  15,
	92:  15,
	93:  15,
	94:  15,
	95:  15,
	96:  23,
	97:  23,
	98:  8,
	99:  10,
	100: 10,
	101: 10,
	102: 10,
}

func PlayerWeaponOverlayTypeForJob(job int, weaponType int, dualWeapon bool) int {
	if weaponType <= 0 {
		if dualWeapon && IsDualWeaponPlayerJob(job) {
			return 2
		}
		return 0
	}
	if dualWeapon {
		if IsDualWeaponPlayerJob(job) && (weaponType == 7 || weaponType == 8) {
			return 2
		}
		return weaponType
	}
	switch {
	case (job == 6 || job == 4007 || job == 4029) && (weaponType == 6 || weaponType == 7 || weaponType == 8):
		return 2
	case (job == 12 || job == 17 || job == 4013 || job == 4018 || job == 4035 || job == 4040) && (weaponType == 7 || weaponType == 8):
		return 2
	case (job == 2 || job == 9 || job == 16 || job == 4003 || job == 4010 || job == 4017 || job == 4025 || job == 4032 || job == 4039) && weaponType == 5:
		return 10
	case (job == 17 || job == 4018 || job == 4040) && weaponType == 6:
		return 1
	case (job == 8 || job == 4009 || job == 4031) && weaponType == 12:
		return 0
	default:
		return weaponType
	}
}

func IsDualWeaponPlayerJob(job int) bool {
	switch job {
	case 12, 4013, 4035, 4060, 4079:
		return true
	default:
		return false
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
		return "\xB7\xD4\xB5\xE5"
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
