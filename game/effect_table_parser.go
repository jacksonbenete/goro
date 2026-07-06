package game

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
var effectObjectIdentifierFieldPattern = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_]*)\s*:\s*([A-Za-z_][A-Za-z0-9_]*)`)

const effectPixelRatio = 1.0 / 35.0

func parseReferenceEffectTableSubset(source string) (map[int]worldEffectSpec, error) {
	source = stripJSComments(source)
	out := make(map[int]worldEffectSpec)
	entries, err := parseReferenceEffectTableEntryArrays(source)
	if err != nil {
		return nil, err
	}
	for id, body := range entries {
		spec := worldEffectSpec{}
		for _, object := range parseJSObjectLiterals(body) {
			component, sfx, ok := parseReferenceEffectComponent(object)
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

func parseReferenceEffectTableEntryIDs(source string) (map[int]struct{}, error) {
	source = stripJSComments(source)
	entries, err := parseReferenceEffectTableEntryArrays(source)
	if err != nil {
		return nil, err
	}
	out := make(map[int]struct{}, len(entries))
	for id := range entries {
		out[id] = struct{}{}
	}
	return out, nil
}

func parseReferenceEffectTableEntryArrays(source string) (map[int]string, error) {
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

func parseReferenceEffectComponent(object string) (worldEffectComponent, string, bool) {
	fields := parseJSObjectFields(object)
	componentType := strings.ToUpper(fieldString(fields, "type"))
	sfx := effectTableSFXPath(fieldString(fields, "wav"))
	switch componentType {
	case "STR":
		file := fieldString(fields, "file")
		if file == "" {
			return worldEffectComponent{}, sfx, false
		}
		randMin, randMax := fieldIntPair(fields, "rand")
		return worldEffectComponent{
			kind:           effectComponentSTR,
			strFile:        file,
			strMinFile:     fieldString(fields, "min"),
			strRandMin:     randMin,
			strRandMax:     randMax,
			attachedEntity: fieldBool(fields, "attachedEntity"),
			spriteHead:     fieldBool(fields, "head"),
			texturePath:    fieldString(fields, "texturePath"),
			duration:       fieldDuration(fields, "duration"),
		}, sfx, true
	case "CYLINDER":
		component := worldEffectComponent{
			kind:             effectComponentCylinder,
			color:            effectColorFields(fields),
			textureName:      fieldString(fields, "textureName"),
			duration:         fieldDuration(fields, "duration"),
			repeat:           fieldBool(fields, "repeat"),
			repeatDelay:      fieldSignedDuration(fields, "repeatDelay"),
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
			blendAdditive:    fieldInt(fields, "blendMode") == 2,
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
			kind:        effectComponent2D,
			textureFile: file,
			duration:    fieldDuration(fields, "duration"),
			repeat:      fieldBool(fields, "repeat"),
			repeatDelay: fieldSignedDuration(fields, "repeatDelay"),
			alphaMax:    fieldFloat(fields, "alphaMax"),
			fade:        fieldBool(fields, "fade"),
			fadeIn:      fieldBool(fields, "fadeIn"),
			fadeOut:     fieldBool(fields, "fadeOut"),
			posZ:        fieldFloat(fields, "posz"),
			sizeStart:   sizeStart * effectPixelRatio,
			sizeEnd:     sizeEnd * effectPixelRatio,
			angleStart:  angleStart,
			angleEnd:    angleEnd,
		}, sfx, true
	case "3D":
		file := fieldString(fields, "file")
		spriteFile := fieldString(fields, "absoluteSpriteName")
		if spriteFile == "" {
			spriteFile = fieldString(fields, "spriteName")
		}
		shadowTexture := fieldBool(fields, "shadowTexture")
		if spriteFile == "" && shadowTexture {
			spriteFile = "data\\sprite\\shadow"
		}
		if file == "" && spriteFile == "" {
			return worldEffectComponent{}, sfx, false
		}
		sizeStart, sizeEnd := effectSizeFields(fields)
		duplicateDelay := fieldDuration(fields, "timeBetweenDupli")
		if fieldInt(fields, "duplicate") > 1 && !fieldExists(fields, "timeBetweenDupli") {
			duplicateDelay = 200 * time.Millisecond
		}
		return worldEffectComponent{
			kind:             effectComponent3D,
			color:            effectColorFields(fields),
			textureFile:      file,
			spriteFile:       spriteFile,
			shadowTexture:    shadowTexture,
			spriteRepeat:     fieldBool(fields, "playSprite"),
			spriteDelay:      fieldDuration(fields, "sprDelay"),
			fromSrc:          fieldBool(fields, "fromSrc"),
			toSrc:            fieldBool(fields, "toSrc"),
			arc:              fieldFloat(fields, "arc"),
			retreat:          fieldFloat(fields, "retreat"),
			duration:         fieldDuration(fields, "duration"),
			delay:            fieldDuration(fields, "delayOffset") + fieldDuration(fields, "delayLate"),
			duplicateDelay:   duplicateDelay,
			repeat:           fieldBool(fields, "repeat"),
			repeatDelay:      fieldSignedDuration(fields, "repeatDelay"),
			alphaMax:         fieldFloat(fields, "alphaMax"),
			alphaMaxDelta:    fieldFloat(fields, "alphaMaxDelta"),
			sparkling:        fieldBool(fields, "sparkling"),
			sparkNumber:      fieldInt(fields, "sparkNumber"),
			fade:             fieldBool(fields, "fade"),
			fadeIn:           fieldBool(fields, "fadeIn"),
			fadeOut:          fieldBool(fields, "fadeOut"),
			rotate:           fieldBool(fields, "rotate"),
			rotateWithCamera: fieldBool(fields, "rotateWithCamera"),
			rotateToTarget:   fieldBool(fields, "rotateToTarget"),
			posX:             fieldFloat(fields, "posx"),
			posY:             fieldFloat(fields, "posy"),
			posZ:             fieldFloat(fields, "posz") + fieldFloat(fields, "poszStart") + fieldFloat(fields, "zOffset") + fieldFloat(fields, "zOffsetStart"),
			posXEnd:          fieldFloat(fields, "posxEnd"),
			posYEnd:          fieldFloat(fields, "posyEnd"),
			posZEnd:          fieldFloat(fields, "poszEnd"),
			posXRand:         fieldFloat(fields, "posxRand"),
			posYRand:         fieldFloat(fields, "posyRand"),
			posXStartRand:    fieldFloat(fields, "posxStartRand"),
			posYStartRand:    fieldFloat(fields, "posyStartRand"),
			posZStartRand:    fieldFloat(fields, "poszStartRand"),
			posXStartMiddle:  fieldFloat(fields, "posxStartRandMiddle"),
			posYStartMiddle:  fieldFloat(fields, "posyStartRandMiddle"),
			posZStartMiddle:  fieldFloat(fields, "poszStartRandMiddle"),
			posXEndRand:      fieldFloat(fields, "posxEndRand"),
			posYEndRand:      fieldFloat(fields, "posyEndRand"),
			posZEndRand:      fieldFloat(fields, "poszEndRand"),
			posXEndMiddle:    fieldFloat(fields, "posxEndRandMiddle"),
			posYEndMiddle:    fieldFloat(fields, "posyEndRandMiddle"),
			posZEndMiddle:    fieldFloat(fields, "poszEndRandMiddle"),
			posXSmooth:       fieldBool(fields, "posxSmooth"),
			posYSmooth:       fieldBool(fields, "posySmooth"),
			posZSmooth:       fieldBool(fields, "poszSmooth"),
			sizeStart:        sizeStart * effectPixelRatio,
			sizeEnd:          sizeEnd * effectPixelRatio,
			sizeRand:         fieldFloat(fields, "sizeRand") * effectPixelRatio,
			sizeStartX:       effectSizeAxisField(fields, "sizeX", "sizeStartX") * effectPixelRatio,
			sizeStartY:       effectSizeAxisField(fields, "sizeY", "sizeStartY") * effectPixelRatio,
			sizeEndX:         effectSizeAxisField(fields, "sizeX", "sizeEndX") * effectPixelRatio,
			sizeEndY:         effectSizeAxisField(fields, "sizeY", "sizeEndY") * effectPixelRatio,
			sizeRandX:        fieldFloat(fields, "sizeRandX") * effectPixelRatio,
			sizeRandY:        fieldFloat(fields, "sizeRandY") * effectPixelRatio,
			sizeRandXMiddle:  fieldFloat(fields, "sizeRandXMiddle") * effectPixelRatio,
			sizeRandYMiddle:  fieldFloat(fields, "sizeRandYMiddle") * effectPixelRatio,
			sizeDelta:        fieldFloat(fields, "sizeDelta"),
			sizeSmooth:       fieldBool(fields, "sizeSmooth"),
			duplicate:        fieldInt(fields, "duplicate"),
			angleStart:       fieldFloat(fields, "angle"),
			angleEnd:         effectAngleEnd(fields),
			orbitRadiusX:     fieldFloat(fields, "rotatePosX"),
			orbitRadiusY:     fieldFloat(fields, "rotatePosY"),
			orbitRadiusZ:     fieldFloat(fields, "rotatePosZ"),
			orbitRotations:   fieldFloat(fields, "nbOfRotation"),
			orbitPhase:       fieldFloat(fields, "rotateLate"),
			orbitPhaseDelta:  fieldFloat(fields, "rotateLateDelta"),
			orbitClockwise:   fieldBool(fields, "rotationClockwise"),
			blendAdditive:    fieldInt(fields, "blendMode") == 2,
			overlay:          fieldBool(fields, "overlay"),
		}, sfx, true
	case "SPR":
		file := fieldString(fields, "file")
		if file == "" {
			return worldEffectComponent{}, sfx, false
		}
		return worldEffectComponent{
			kind:             effectComponentSPR,
			spriteFile:       file,
			duration:         fieldDuration(fields, "duration"),
			delay:            fieldDuration(fields, "delayOffset") + fieldDuration(fields, "delayLate"),
			spriteHead:       fieldBool(fields, "head"),
			spriteDirection:  fieldBool(fields, "direction"),
			spriteRepeat:     fieldBool(fields, "repeat"),
			spriteStopAtEnd:  fieldBool(fields, "stopAtEnd"),
			spriteFrame:      fieldInt(fields, "frame"),
			spriteDelay:      fieldDuration(fields, "delayFrame"),
			spriteXOffset:    fieldFloat(fields, "xOffset"),
			spriteYOffset:    fieldFloat(fields, "yOffset"),
			worldSizedSprite: true,
		}, sfx, true
	case "FUNC":
		funcName, adapter := effectTableFuncAdapter(object, fields)
		return worldEffectComponent{
			kind:           effectComponentFUNC,
			funcAdapter:    adapter,
			funcName:       funcName,
			attachedEntity: fieldBool(fields, "attachedEntity"),
			duration:       fieldDuration(fields, "duration"),
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
	for _, match := range effectObjectIdentifierFieldPattern.FindAllStringSubmatch(object, -1) {
		if len(match) != 3 {
			continue
		}
		if _, exists := out[match[1]]; exists {
			continue
		}
		out[match[1]] = match[2]
	}
	return out
}

func effectTableFuncAdapter(object string, fields map[string]string) (string, effectFuncAdapter) {
	funcName := fieldIdentifier(fields, "func")
	if strings.Contains(object, "MagicTarget") {
		return "MagicTarget", effectFuncGroundSample
	}
	switch funcName {
	case "MagicTarget":
		return funcName, effectFuncGroundSample
	default:
		return funcName, effectFuncUnknown
	}
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

func fieldIdentifier(fields map[string]string, key string) string {
	value, ok := fields[key]
	if !ok {
		return ""
	}
	if decoded := fieldString(fields, key); decoded != "" {
		return decoded
	}
	value = strings.TrimSpace(value)
	if value == "function" || value == "null" || value == "true" || value == "false" {
		return ""
	}
	return value
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

func effectSizeAxisField(fields map[string]string, fixedKey, endpointKey string) float64 {
	if size := fieldFloat(fields, fixedKey); size > 0 {
		return size
	}
	return fieldFloat(fields, endpointKey)
}

func effectAngleEnd(fields map[string]string) float64 {
	angle := fieldFloat(fields, "angle")
	if fieldExists(fields, "toAngle") {
		return fieldFloat(fields, "toAngle")
	}
	return angle + fieldFloat(fields, "angleDelta")
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

func fieldSignedDuration(fields map[string]string, key string) time.Duration {
	value := fieldFloat(fields, key)
	if value == 0 {
		return 0
	}
	return time.Duration(value * float64(time.Millisecond))
}

func effectTableSFXPath(wav string) string {
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
