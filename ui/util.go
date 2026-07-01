package ui

const (
	CursorActionDefault = 0
	CursorActionClick   = 2
)

func pointInRect(px, py, x, y, w, h int) bool {
	return px >= x && py >= y && px < x+w && py < y+h
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func trimRunes(text string, maxRunes int) string {
	runes := []rune(text)
	if len(runes) <= maxRunes {
		return text
	}
	if maxRunes <= 0 {
		return ""
	}
	if maxRunes <= 3 {
		return string(runes[:maxRunes])
	}
	return string(runes[:maxRunes-3]) + "..."
}

func rectArray(x, y, w, h int) [4]int {
	return [4]int{x, y, w, h}
}
