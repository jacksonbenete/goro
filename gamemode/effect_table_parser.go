package gamemode

import (
	"fmt"
	"image/color"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode"
)

var effectObjectFieldPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*:\s*('(?:\\.|[^'\\])*'|"(?:\\.|[^"\\])*"|-?\d+(?:\.\d+)?|true|false|null)`)
var effectObjectArrayFieldPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*:\s*\[\s*(-?\d+)\s*,\s*(-?\d+)\s*\]`)

const roBrowserEffectPixelRatio = 1.0 / 35.0

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
		randMin, randMax := fieldIntPair(fields, "rand")
		return worldEffectComponent{
			kind:        effectPrimitiveSTR,
			strFile:     file,
			strRandMin:  randMin,
			strRandMax:  randMax,
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
	case "2D":
		file := fieldString(fields, "file")
		if file == "" {
			return worldEffectComponent{}, sfx, false
		}
		sizeStart := fieldFloat(fields, "sizeStart")
		sizeEnd := fieldFloat(fields, "sizeEnd")
		if size := fieldFloat(fields, "size"); size > 0 {
			if sizeStart <= 0 {
				sizeStart = size
			}
			if sizeEnd <= 0 {
				sizeEnd = size
			}
		}
		if sizeStart <= 0 && sizeEnd > 0 {
			sizeStart = sizeEnd
		}
		if sizeEnd <= 0 && sizeStart > 0 {
			sizeEnd = sizeStart
		}
		angleStart := fieldFloat(fields, "angle")
		angleEnd := angleStart
		if fieldExists(fields, "toAngle") {
			angleEnd = fieldFloat(fields, "toAngle")
		} else if fieldExists(fields, "angleDelta") {
			angleEnd = angleStart + fieldFloat(fields, "angleDelta")
		}
		return worldEffectComponent{
			kind:        effectPrimitive2D,
			textureFile: file,
			duration:    fieldDuration(fields, "duration"),
			alphaMax:    fieldFloat(fields, "alphaMax"),
			fade:        fieldBool(fields, "fade"),
			fadeIn:      fieldBool(fields, "fadeIn"),
			fadeOut:     fieldBool(fields, "fadeOut"),
			posZ:        fieldFloat(fields, "posz"),
			sizeStart:   sizeStart * roBrowserEffectPixelRatio,
			sizeEnd:     sizeEnd * roBrowserEffectPixelRatio,
			angleStart:  angleStart,
			angleEnd:    angleEnd,
		}, sfx, true
	case "3D":
		file := fieldString(fields, "file")
		if file == "" {
			return worldEffectComponent{}, sfx, false
		}
		sizeStart, sizeEnd := effectSizeFields(fields)
		return worldEffectComponent{
			kind:            effectPrimitive3D,
			color:           effectColorFields(fields),
			textureFile:     file,
			duration:        fieldDuration(fields, "duration"),
			delay:           fieldDuration(fields, "delayOffset") + fieldDuration(fields, "delayLate"),
			duplicateDelay:  fieldDuration(fields, "timeBetweenDupli"),
			alphaMax:        fieldFloat(fields, "alphaMax"),
			fade:            fieldBool(fields, "fade"),
			fadeIn:          fieldBool(fields, "fadeIn"),
			fadeOut:         fieldBool(fields, "fadeOut"),
			posX:            fieldFloat(fields, "posx"),
			posY:            fieldFloat(fields, "posy"),
			posZ:            fieldFloat(fields, "posz") + fieldFloat(fields, "poszStart"),
			posXEnd:         fieldFloat(fields, "posxEnd"),
			posYEnd:         fieldFloat(fields, "posyEnd"),
			posZEnd:         fieldFloat(fields, "poszEnd"),
			posXRand:        fieldFloat(fields, "posxRand"),
			posYRand:        fieldFloat(fields, "posyRand"),
			posZStartRand:   fieldFloat(fields, "poszStartRand"),
			posZStartMiddle: fieldFloat(fields, "poszStartRandMiddle"),
			posXEndRand:     fieldFloat(fields, "posxEndRand"),
			posYEndRand:     fieldFloat(fields, "posyEndRand"),
			posZEndRand:     fieldFloat(fields, "poszEndRand"),
			posZEndMiddle:   fieldFloat(fields, "poszEndRandMiddle"),
			sizeStart:       sizeStart * roBrowserEffectPixelRatio,
			sizeEnd:         sizeEnd * roBrowserEffectPixelRatio,
			sizeRand:        fieldFloat(fields, "sizeRand") * roBrowserEffectPixelRatio,
			sizeSmooth:      fieldBool(fields, "sizeSmooth"),
			duplicate:       fieldInt(fields, "duplicate"),
			angleStart:      fieldFloat(fields, "angle"),
			angleEnd:        fieldFloat(fields, "angle") + fieldFloat(fields, "angleDelta"),
		}, sfx, true
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
	for _, match := range effectObjectArrayFieldPattern.FindAllStringSubmatch(object, -1) {
		if len(match) != 4 {
			continue
		}
		out[match[1]] = match[2] + "," + match[3]
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

func fieldExists(fields map[string]string, key string) bool {
	_, ok := fields[key]
	return ok
}

func effectSizeFields(fields map[string]string) (float64, float64) {
	sizeStart := fieldFloat(fields, "sizeStart")
	sizeEnd := fieldFloat(fields, "sizeEnd")
	if size := fieldFloat(fields, "size"); size > 0 {
		if sizeStart <= 0 {
			sizeStart = size
		}
		if sizeEnd <= 0 {
			sizeEnd = size
		}
	}
	if sizeStart <= 0 && sizeEnd > 0 {
		sizeStart = sizeEnd
	}
	if sizeEnd <= 0 && sizeStart > 0 {
		sizeEnd = sizeStart
	}
	return sizeStart, sizeEnd
}

func effectColorFields(fields map[string]string) color.RGBA {
	hasColor := fieldExists(fields, "red") || fieldExists(fields, "green") || fieldExists(fields, "blue")
	if !hasColor {
		return color.RGBA{}
	}
	return color.RGBA{
		R: uint8(clampFloat(fieldColorValue(fields, "red", 1), 0, 1) * 255),
		G: uint8(clampFloat(fieldColorValue(fields, "green", 1), 0, 1) * 255),
		B: uint8(clampFloat(fieldColorValue(fields, "blue", 1), 0, 1) * 255),
		A: 255,
	}
}

func fieldColorValue(fields map[string]string, key string, fallback float64) float64 {
	if !fieldExists(fields, key) {
		return fallback
	}
	return fieldFloat(fields, key)
}

func fieldInt(fields map[string]string, key string) int {
	return int(fieldFloat(fields, key))
}

func fieldIntPair(fields map[string]string, key string) (int, int) {
	value, ok := fields[key]
	if !ok {
		return 0, 0
	}
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0
	}
	first, errFirst := strconv.Atoi(strings.TrimSpace(parts[0]))
	second, errSecond := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errFirst != nil || errSecond != nil {
		return 0, 0
	}
	return first, second
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
