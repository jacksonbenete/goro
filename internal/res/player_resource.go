package res

import "fmt"

const (
	playerHumanSpriteRoot = "data\\sprite\\\xC0\xCE\xB0\xA3\xC1\xB7\\"
	playerBodyDir         = "\xB8\xF6\xC5\xEB"
	playerHeadDir         = "\xB8\xD3\xB8\xAE\xC5\xEB"
	playerFemaleSex       = "\xBF\xA9"
	playerMaleSex         = "\xB3\xB2"
)

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

func PlayerHeadResourceCandidates(job int, head int, sex byte, extension string) []string {
	sexToken := PlayerSexToken(sex)
	head = NormalizePlayerHead(head, job)
	return []string{
		fmt.Sprintf("%s%s\\%s\\%d_%s.%s", playerHumanSpriteRoot, playerHeadDir, sexToken, head, sexToken, extension),
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
