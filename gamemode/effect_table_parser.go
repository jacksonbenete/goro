package gamemode

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var effectObjectFieldPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*:\s*('(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*"|-?\d+(?:\.\d+)?|true|false|null)`)

func parseRobrowserEffectTableSubset(source string) (map[int]worldEffectSpec, error) {
	source = stripJSComments(source)
	out := make(map[int]worldEffectSpec)
	entries, err := parseRobrowserEffectTableEntryArrays(source)
	if err != nil {
		return nil, err
	}
	for id, body := range entries {
		spec := worldEffectSpec{}
		for _, object := range parseJSObjectLiterals(body) {
			component, sfx, ok := parseRobrowserEffectComponent(object)
			if sfx != "" {
				spec.sfx = append(spec.sfx, sfx)
			}
			if !ok {
				continue
			}
			spec.components = append(spec.components, component)
			duration := worldEffectComponentDuration(spec, component)
			if duration > spec.duration {
				spec.duration = duration
			}
		}
		if len(spec.components) > 0 || len(spec.sfx) > 0 {
			out[id] = spec
		}
	}
	return out, nil
}

func parseRobrowserEffectTableEntryIDs(source string) (map[int]struct{}, error) {
	source = stripJSComments(source)
	entries, err := parseRobrowserEffectTableEntryArrays(source)
	if err != nil {
		return nil, err
	}
	out := make(map[int]struct{}, len(entries))
	for id := range entries {
		out[id] = struct{}{}
	}
	return out, nil
}

func parseRobrowserEffectTableEntryArrays(source string) (map[int]string, error) {
	out := make(map[int]string)
	for i := 0; i < len(source); i++ {
		if !isEntryStart(source, i) {
			continue
		}
		start := i
		for i < len(source) && unicode.IsDigit(rune(source[i])) {
			i++
		}
		id, err := strconv.Atoi(source[start:i])
		if err != nil {
			return nil, err
		}
		i = skipJSSpace(source, i)
		if i >= len(source) || source[i] != ':' {
			continue
		}
		i = skipJSSpace(source, i+1)
		if i >= len(source) || source[i] != '[' {
			continue
		}
		end, err := findMatchingJSDelimiter(source, i, '[', ']')
		if err != nil {
			return nil, fmt.Errorf("effect %d: %w", id, err)
		}
		out[id] = source[i+1 : end]
		i = end
	}
	return out, nil
}

func isEntryStart(source string, index int) bool {
	if index < 0 || index >= len(source) || !unicode.IsDigit(rune(source[index])) {
		return false
	}
	if index == 0 {
		return true
	}
	prev := rune(source[index-1])
	return !(unicode.IsDigit(prev) || unicode.IsLetter(prev) || prev == '_')
}

func skipJSSpace(source string, index int) int {
	for index < len(source) && unicode.IsSpace(rune(source[index])) {
		index++
	}
	return index
}

func parseJSObjectLiterals(source string) []string {
	var out []string
	for i := 0; i < len(source); i++ {
		if source[i] != '{' {
			continue
		}
		end, err := findMatchingJSDelimiter(source, i, '{', '}')
		if err != nil {
			return out
		}
		out = append(out, source[i+1:end])
		i = end
	}
	return out
}

func findMatchingJSDelimiter(source string, start int, open, close byte) (int, error) {
	depth := 0
	inString := byte(0)
	escaped := false
	for i := start; i < len(source); i++ {
		c := source[i]
		if inString != 0 {
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == inString {
				inString = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			inString = c
			continue
		}
		if c == open {
			depth++
			continue
		}
		if c == close {
			depth--
			if depth == 0 {
				return i, nil
			}
		}
	}
	return -1, fmt.Errorf("unterminated %c", open)
}

func parseRobrowserEffectComponent(object string) (worldEffectComponent, string, bool) {
	fields := parseJSObjectFields(object)
	componentType := strings.ToUpper(fieldString(fields, "type"))
	sfx := robrowserSFXPath(fieldString(fields, "wav"))
	switch componentType {
	case "STR":
		file := fieldString(fields, "file")
		if file == "" {
			return worldEffectComponent{}, sfx, false
		}
		return worldEffectComponent{
			kind:        effectPrimitiveSTR,
			strFile:     file,
			texturePath: fieldString(fields, "texturePath"),
			duration:    fieldDuration(fields, "duration"),
		}, sfx, true
	case "CYLINDER":
		component := worldEffectComponent{
			kind:             effectPrimitiveCylinder,
			textureName:      fieldString(fields, "textureName"),
			duration:         fieldDuration(fields, "duration"),
			alphaMax:         fieldFloat(fields, "alphaMax"),
			fade:             fieldBool(fields, "fade"),
			fadeIn:           fieldBool(fields, "fadeIn"),
			fadeOut:          fieldBool(fields, "fadeOut"),
			rotate:           fieldBool(fields, "rotate"),
			animation:        fieldInt(fields, "animation"),
			bottomSize:       fieldFloat(fields, "bottomSize"),
			topSize:          fieldFloat(fields, "topSize"),
			height:           fieldFloat(fields, "height"),
			posZ:             fieldFloat(fields, "posZ"),
			totalCircleSides: fieldInt(fields, "totalCircleSides"),
			circleSides:      fieldInt(fields, "circleSides"),
		}
		if component.textureName == "" {
			return worldEffectComponent{}, sfx, false
		}
		if component.totalCircleSides == 0 {
			component.totalCircleSides = 32
		}
		if component.circleSides == 0 {
			component.circleSides = component.totalCircleSides
		}
		return component, sfx, true
	default:
		return worldEffectComponent{}, sfx, false
	}
}

func parseJSObjectFields(object string) map[string]string {
	out := make(map[string]string)
	for _, match := range effectObjectFieldPattern.FindAllStringSubmatch(object, -1) {
		if len(match) != 3 {
			continue
		}
		out[match[1]] = match[2]
	}
	return out
}

func fieldString(fields map[string]string, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	value = strings.TrimSpace(value)
	if len(value) < 2 || (value[0] != '\'' && value[0] != '"') {
		return ""
	}
	inner := value[1 : len(value)-1]
	quoted := `"` + strings.ReplaceAll(inner, `"`, `\"`) + `"`
	out, err := strconv.Unquote(quoted)
	if err != nil {
		return inner
	}
	return out
}

func fieldFloat(fields map[string]string, key string) float64 {
	value, ok := fields[key]
	if !ok {
		return 0
	}
	out, _ := strconv.ParseFloat(value, 64)
	return out
}

func fieldInt(fields map[string]string, key string) int {
	return int(fieldFloat(fields, key))
}

func fieldBool(fields map[string]string, key string) bool {
	return fields[key] == "true"
}

func fieldDuration(fields map[string]string, key string) time.Duration {
	value := fieldFloat(fields, key)
	if value <= 0 {
		return 0
	}
	return time.Duration(value * float64(time.Millisecond))
}

func robrowserSFXPath(wav string) string {
	wav = strings.TrimSpace(wav)
	if wav == "" {
		return ""
	}
	wav = strings.ReplaceAll(wav, "/", "\\")
	if !strings.HasSuffix(strings.ToLower(wav), ".wav") {
		wav += ".wav"
	}
	return wav
}

func stripJSComments(source string) string {
	var out strings.Builder
	inString := byte(0)
	escaped := false
	for i := 0; i < len(source); i++ {
		c := source[i]
		if inString != 0 {
			out.WriteByte(c)
			if escaped {
				escaped = false
				continue
			}
			if c == '\\' {
				escaped = true
				continue
			}
			if c == inString {
				inString = 0
			}
			continue
		}
		if c == '\'' || c == '"' {
			inString = c
			out.WriteByte(c)
			continue
		}
		if c == '/' && i+1 < len(source) {
			switch source[i+1] {
			case '/':
				for i < len(source) && source[i] != '\n' {
					i++
				}
				if i < len(source) {
					out.WriteByte('\n')
				}
				continue
			case '*':
				i += 2
				for i+1 < len(source) && !(source[i] == '*' && source[i+1] == '/') {
					i++
				}
				i++
				continue
			}
		}
		out.WriteByte(c)
	}
	return out.String()
}
