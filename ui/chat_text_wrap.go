package ui

import (
	"image/color"
	"strings"

	"github.com/gogpu/ui/widget"
)

type wrappedConsoleLine struct {
	text  string
	color color.RGBA
}

type wrappedChatRoomLine struct {
	text  string
	color widget.Color
}

func wrapConsoleMessages(lines []ConsoleMessage, maxWidth int) []wrappedConsoleLine {
	out := make([]wrappedConsoleLine, 0, len(lines))
	for _, line := range lines {
		for _, text := range wrapChatText(line.Text, maxWidth) {
			out = append(out, wrappedConsoleLine{text: text, color: line.Color})
		}
	}
	return out
}

func wrapChatRoomLines(lines []chatRoomLine, maxWidth int) []wrappedChatRoomLine {
	out := make([]wrappedChatRoomLine, 0, len(lines))
	for _, line := range lines {
		for _, text := range wrapChatText(line.text, maxWidth) {
			out = append(out, wrappedChatRoomLine{text: text, color: line.color})
		}
	}
	return out
}

func wrapChatText(text string, maxWidth int) []string {
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	paragraphs := strings.Split(text, "\n")
	lines := make([]string, 0, len(paragraphs))
	for _, paragraph := range paragraphs {
		lines = appendWrappedParagraph(lines, paragraph, maxWidth)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func appendWrappedParagraph(lines []string, paragraph string, maxWidth int) []string {
	words := strings.Fields(paragraph)
	if len(words) == 0 {
		return append(lines, "")
	}
	if maxWidth <= 0 {
		return append(lines, strings.Join(words, " "))
	}
	current := ""
	for _, word := range words {
		if current == "" {
			chunks := splitChatWord(word, maxWidth)
			for i, chunk := range chunks {
				if i == len(chunks)-1 {
					current = chunk
					continue
				}
				lines = append(lines, chunk)
			}
			continue
		}
		candidate := current + " " + word
		if chatTextWidth(candidate) <= maxWidth {
			current = candidate
			continue
		}
		lines = append(lines, current)
		current = ""
		chunks := splitChatWord(word, maxWidth)
		for i, chunk := range chunks {
			if i == len(chunks)-1 {
				current = chunk
				continue
			}
			lines = append(lines, chunk)
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	return lines
}

func splitChatWord(word string, maxWidth int) []string {
	if word == "" || maxWidth <= 0 || chatTextWidth(word) <= maxWidth {
		return []string{word}
	}
	chunks := make([]string, 0, 2)
	var current strings.Builder
	currentWidth := 0
	for _, r := range word {
		runeWidth := chatRuneWidth(r)
		if current.Len() > 0 && currentWidth+runeWidth > maxWidth {
			chunks = append(chunks, current.String())
			current.Reset()
			currentWidth = 0
		}
		current.WriteRune(r)
		currentWidth += runeWidth
	}
	if current.Len() > 0 {
		chunks = append(chunks, current.String())
	}
	return chunks
}

func chatTextWidth(text string) int {
	width := 0
	for _, r := range text {
		width += chatRuneWidth(r)
	}
	return width
}

func chatRuneWidth(r rune) int {
	switch r {
	case ' ', '\t':
		return 4
	case 'i', 'j', 'l', 'I', '.', ',', ':', ';', '!', '\'', '`', '|':
		return 3
	case 'm', 'w', 'M', 'W', '@', '#', '%', '&':
		return 9
	}
	if r >= 0x1100 {
		return 11
	}
	return 7
}
