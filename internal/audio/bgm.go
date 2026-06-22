package audio

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"math"
	"path/filepath"
	"strings"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
	"github.com/kivutar/goro/internal/res"
	"github.com/kvark128/minimp3"
)

const defaultSampleRate = 44100

type BGM struct {
	resources  *res.Manager
	context    *ebitenaudio.Context
	player     *ebitenaudio.Player
	table      map[string]string
	current    string
	playerID   int
	enabled    bool
	volume     float64
	sfxPlayers []*ebitenaudio.Player
}

func NewBGM(resources *res.Manager, enabled bool, volume float64) *BGM {
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	return &BGM{
		resources: resources,
		enabled:   enabled,
		volume:    volume,
	}
}

func (b *BGM) PlayMap(mapName string) (string, error) {
	if b == nil || !b.enabled {
		return "", nil
	}
	path := b.ResolveMapBGM(mapName)
	log.Printf("bgm map request map=%s resolved=%s current=%s player=%t", mapName, path, b.current, b.player != nil)
	if path == "" {
		return "", nil
	}
	if sameAudioPath(path, b.current) && b.player != nil {
		if !b.player.IsPlaying() {
			log.Printf("bgm resume existing id=%d path=%s", b.playerID, b.current)
			b.player.Play()
		} else {
			log.Printf("bgm keep existing id=%d path=%s", b.playerID, b.current)
		}
		return b.current, nil
	}
	if err := b.Play(path); err != nil {
		return path, err
	}
	return b.current, nil
}

func (b *BGM) Play(path string) error {
	if b == nil || !b.enabled {
		return nil
	}
	path = normalizeAudioPath(path)
	if path == "" {
		return nil
	}
	data, source, err := readAudioFile(b.resources, path)
	if err != nil {
		return err
	}

	pcm, sourceRate, decoder, err := decodeNativePCM(data)
	if err != nil {
		return fmt.Errorf("decode %s: %w", source, err)
	}
	context := b.ensureContext(sourceRate)
	if sourceRate != context.SampleRate() {
		pcm, err = resamplePCM16Stereo(pcm, sourceRate, context.SampleRate())
		if err != nil {
			return fmt.Errorf("resample %s: %w", source, err)
		}
		decoder = fmt.Sprintf("%s+sinc %d->%d", decoder, sourceRate, context.SampleRate())
	} else {
		decoder = fmt.Sprintf("%s native %d", decoder, sourceRate)
	}
	loop := ebitenaudio.NewInfiniteLoop(bytes.NewReader(pcm), int64(len(pcm)))
	player, err := context.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("player %s: %w", source, err)
	}
	player.SetVolume(b.volume)

	b.stopCurrent()
	b.playerID++
	b.player = player
	b.current = path
	player.Play()
	log.Printf("bgm playing id=%d path=%s source=%s decoder=%s bytes=%d pcm_len=%d sample_rate=%d volume=%.2f", b.playerID, path, source, decoder, len(data), len(pcm), context.SampleRate(), b.volume)
	return nil
}

func (b *BGM) PlaySFX(path string) (string, error) {
	if b == nil {
		return "", nil
	}
	path = normalizeSFXPath(path)
	if path == "" {
		return "", nil
	}
	data, source, err := readSFXFile(b.resources, path)
	if err != nil {
		return "", err
	}
	context := b.ensureContext(defaultSampleRate)
	stream, err := wav.DecodeWithSampleRate(context.SampleRate(), bytes.NewReader(data))
	if err != nil {
		return source, fmt.Errorf("decode sfx %s: %w", source, err)
	}
	player, err := context.NewPlayer(stream)
	if err != nil {
		return source, fmt.Errorf("player sfx %s: %w", source, err)
	}
	player.SetVolume(b.volume)
	b.trimSFXPlayers()
	b.sfxPlayers = append(b.sfxPlayers, player)
	player.Play()
	return source, nil
}

func decodeNativePCM(data []byte) ([]byte, int, string, error) {
	decoder := minimp3.NewDecoder(bytes.NewReader(data))
	pcm, err := io.ReadAll(decoder)
	if err != nil {
		return nil, 0, "", err
	}
	switch decoder.Channels() {
	case 1:
		pcm, err = monoPCM16ToStereo(pcm)
		if err != nil {
			return nil, 0, "", err
		}
	case 2:
		if len(pcm)%4 != 0 {
			return nil, 0, "", fmt.Errorf("invalid stereo pcm length %d", len(pcm))
		}
	default:
		return nil, 0, "", fmt.Errorf("unsupported mp3 channel count %d", decoder.Channels())
	}
	sampleRate := decoder.SampleRate()
	if sampleRate <= 0 {
		sampleRate = defaultSampleRate
	}
	return pcm, sampleRate, "minimp3", nil
}

func monoPCM16ToStereo(src []byte) ([]byte, error) {
	if len(src)%2 != 0 {
		return nil, fmt.Errorf("invalid mono pcm length %d", len(src))
	}
	dst := make([]byte, len(src)*2)
	for frame := 0; frame < len(src)/2; frame++ {
		sample := int16(uint16(src[frame*2]) | uint16(src[frame*2+1])<<8)
		writePCM16(dst, frame, 0, sample)
		writePCM16(dst, frame, 1, sample)
	}
	return dst, nil
}

func resamplePCM16Stereo(src []byte, srcRate, dstRate int) ([]byte, error) {
	if srcRate <= 0 || dstRate <= 0 {
		return nil, fmt.Errorf("invalid sample rate %d -> %d", srcRate, dstRate)
	}
	if len(src)%4 != 0 {
		return nil, fmt.Errorf("invalid stereo pcm length %d", len(src))
	}
	if srcRate == dstRate {
		return src, nil
	}
	srcFrames := len(src) / 4
	if srcFrames == 0 {
		return nil, fmt.Errorf("empty decoded PCM")
	}
	dstFrames := int(float64(srcFrames) * float64(dstRate) / float64(srcRate))
	if dstFrames <= 0 {
		return nil, fmt.Errorf("empty resampled PCM")
	}
	dst := make([]byte, dstFrames*4)

	const radius = 8.0
	cutoff := 1.0
	if dstRate < srcRate {
		cutoff = float64(dstRate) / float64(srcRate)
	}
	divisor := gcd(srcRate, dstRate)
	phaseCount := dstRate / divisor
	taps := make([][]resampleTap, phaseCount)
	for phase := range taps {
		fraction := float64(phase*divisor) / float64(dstRate)
		taps[phase] = buildResampleTaps(fraction, radius, cutoff)
	}
	for frame := 0; frame < dstFrames; frame++ {
		srcNumerator := int64(frame) * int64(srcRate)
		base := int(srcNumerator / int64(dstRate))
		phase := int(srcNumerator%int64(dstRate)) / divisor
		left := 0.0
		right := 0.0
		weightSum := 0.0
		for _, tap := range taps[phase] {
			sample := base + tap.offset
			if sample < 0 || sample >= srcFrames {
				continue
			}
			weightSum += tap.weight
			left += pcm16ToFloat(readPCM16(src, sample, 0)) * tap.weight
			right += pcm16ToFloat(readPCM16(src, sample, 1)) * tap.weight
		}
		if weightSum != 0 {
			left /= weightSum
			right /= weightSum
		}
		writePCM16(dst, frame, 0, floatToPCM16(left))
		writePCM16(dst, frame, 1, floatToPCM16(right))
	}
	return dst, nil
}

type resampleTap struct {
	offset int
	weight float64
}

func buildResampleTaps(fraction float64, radius, cutoff float64) []resampleTap {
	start := int(math.Floor(fraction - radius))
	end := int(math.Ceil(fraction + radius))
	taps := make([]resampleTap, 0, end-start+1)
	sum := 0.0
	for offset := start; offset <= end; offset++ {
		distance := fraction - float64(offset)
		if math.Abs(distance) > radius {
			continue
		}
		weight := cutoff * sinc(cutoff*distance) * hann(distance/radius)
		if weight == 0 {
			continue
		}
		taps = append(taps, resampleTap{offset: offset, weight: weight})
		sum += weight
	}
	if sum != 0 {
		for i := range taps {
			taps[i].weight /= sum
		}
	}
	return taps
}

func gcd(a, b int) int {
	if a < 0 {
		a = -a
	}
	if b < 0 {
		b = -b
	}
	for b != 0 {
		a, b = b, a%b
	}
	if a == 0 {
		return 1
	}
	return a
}

func readPCM16(pcm []byte, frame, channel int) int16 {
	offset := frame*4 + channel*2
	return int16(uint16(pcm[offset]) | uint16(pcm[offset+1])<<8)
}

func writePCM16(pcm []byte, frame, channel int, value int16) {
	offset := frame*4 + channel*2
	u := uint16(value)
	pcm[offset] = byte(u)
	pcm[offset+1] = byte(u >> 8)
}

func pcm16ToFloat(value int16) float64 {
	return float64(value) / 32768
}

func floatToPCM16(value float64) int16 {
	if value > 1 {
		value = 1
	}
	if value < -1 {
		value = -1
	}
	value *= 32767
	if value > 32767 {
		return 32767
	}
	if value < -32768 {
		return -32768
	}
	return int16(value)
}

func sinc(x float64) float64 {
	if math.Abs(x) < 1e-8 {
		return 1
	}
	x *= math.Pi
	return math.Sin(x) / x
}

func hann(x float64) float64 {
	if x < 0 {
		x = -x
	}
	if x >= 1 {
		return 0
	}
	return 0.5 + 0.5*math.Cos(math.Pi*x)
}

func (b *BGM) ensureContext(preferredSampleRate int) *ebitenaudio.Context {
	if b.context != nil {
		return b.context
	}
	if current := ebitenaudio.CurrentContext(); current != nil {
		b.context = current
		log.Printf("bgm using existing audio context sample_rate=%d", current.SampleRate())
		return current
	}
	if preferredSampleRate <= 0 {
		preferredSampleRate = defaultSampleRate
	}
	b.context = ebitenaudio.NewContext(preferredSampleRate)
	log.Printf("bgm created audio context sample_rate=%d", b.context.SampleRate())
	return b.context
}

func (b *BGM) Stop() {
	if b == nil {
		return
	}
	b.stopCurrent()
	b.stopSFX()
	b.current = ""
}

func (b *BGM) stopCurrent() {
	if b.player == nil {
		return
	}
	log.Printf("bgm stopping id=%d path=%s playing=%t", b.playerID, b.current, b.player.IsPlaying())
	b.player.Pause()
	if err := b.player.Close(); err != nil {
		log.Printf("bgm close failed: %v", err)
	}
	b.player = nil
}

func (b *BGM) trimSFXPlayers() {
	if len(b.sfxPlayers) == 0 {
		return
	}
	active := b.sfxPlayers[:0]
	for _, player := range b.sfxPlayers {
		if player == nil {
			continue
		}
		if player.IsPlaying() {
			active = append(active, player)
			continue
		}
		if err := player.Close(); err != nil {
			log.Printf("sfx close failed: %v", err)
		}
	}
	const maxSFXPlayers = 32
	for len(active) > maxSFXPlayers {
		player := active[0]
		active = active[1:]
		if player != nil {
			player.Pause()
			if err := player.Close(); err != nil {
				log.Printf("sfx close failed: %v", err)
			}
		}
	}
	b.sfxPlayers = active
}

func (b *BGM) stopSFX() {
	for _, player := range b.sfxPlayers {
		if player == nil {
			continue
		}
		player.Pause()
		if err := player.Close(); err != nil {
			log.Printf("sfx close failed: %v", err)
		}
	}
	b.sfxPlayers = nil
}

func (b *BGM) ResolveMapBGM(mapName string) string {
	if b == nil {
		return ""
	}
	b.ensureTable()
	key := canonicalMapName(mapName)
	if key == "" {
		return "bgm\\01.mp3"
	}
	if path := b.table[key]; path != "" {
		return path
	}
	return "bgm\\01.mp3"
}

func (b *BGM) ensureTable() {
	if b.table != nil {
		return
	}
	b.table = make(map[string]string)
	if b.resources == nil {
		return
	}
	_, data, ok := b.resources.ReadFirst([]string{
		"data\\mp3nametable.txt",
		"data/mp3nametable.txt",
		"mp3nametable.txt",
	})
	if !ok {
		log.Printf("bgm table not found, falling back to bgm\\01.mp3")
		return
	}
	for key, path := range ParseMP3NameTable(data) {
		b.table[key] = path
	}
	log.Printf("bgm table loaded entries=%d", len(b.table))
}

func ParseMP3NameTable(data []byte) map[string]string {
	out := make(map[string]string)
	for _, rawLine := range strings.Split(string(data), "\n") {
		line := strings.TrimSpace(strings.TrimSuffix(rawLine, "\r"))
		if line == "" {
			continue
		}
		parts := strings.Split(line, "#")
		if len(parts) < 2 {
			continue
		}
		key := canonicalMapName(parts[0])
		path := normalizeAudioPath(parts[1])
		if key == "" || path == "" {
			continue
		}
		out[key] = path
	}
	return out
}

func canonicalMapName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, "/", "\\")
	name = strings.ToLower(name)
	name = strings.TrimPrefix(name, "data\\")
	name = strings.TrimSuffix(name, ".gat")
	name = strings.TrimSuffix(name, ".rsw")
	if name == "" {
		return ""
	}
	return name + ".rsw"
}

func normalizeAudioPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "\"")
	path = strings.ReplaceAll(path, "/", "\\")
	path = strings.TrimPrefix(path, ".\\")
	path = strings.TrimPrefix(path, "data\\")
	if path == "" {
		return ""
	}
	if !strings.Contains(path, "\\") {
		path = "bgm\\" + path
	}
	return path
}

func readAudioFile(manager *res.Manager, path string) ([]byte, string, error) {
	if manager == nil {
		return nil, "", fmt.Errorf("resource manager is nil")
	}
	for _, candidate := range audioPathCandidates(path) {
		data, err := manager.ReadFile(candidate)
		if err == nil {
			return data, candidate, nil
		}
	}
	return nil, "", fmt.Errorf("bgm not found: %s", path)
}

func readSFXFile(manager *res.Manager, path string) ([]byte, string, error) {
	if manager == nil {
		return nil, "", fmt.Errorf("resource manager is nil")
	}
	for _, candidate := range sfxPathCandidates(path) {
		data, err := manager.ReadFile(candidate)
		if err == nil {
			return data, candidate, nil
		}
	}
	return nil, "", fmt.Errorf("sfx not found: %s", path)
}

func audioPathCandidates(path string) []string {
	normalized := normalizeAudioPath(path)
	if normalized == "" {
		return nil
	}
	slash := strings.ReplaceAll(normalized, "\\", "/")
	candidates := []string{normalized, slash}
	if strings.HasPrefix(strings.ToLower(normalized), "bgm\\") {
		rest := normalized[4:]
		candidates = append(candidates,
			"BGM\\"+rest,
			"BGM/"+strings.ReplaceAll(rest, "\\", "/"),
			"bgm\\"+rest,
			"bgm/"+strings.ReplaceAll(rest, "\\", "/"),
		)
	}
	if filepath.Ext(normalized) == "" {
		candidates = append(candidates, normalized+".mp3", slash+".mp3")
	}
	return uniqueStrings(candidates)
}

func normalizeSFXPath(path string) string {
	path = strings.TrimSpace(path)
	path = strings.Trim(path, "\"")
	path = strings.ReplaceAll(path, "/", "\\")
	path = strings.TrimPrefix(path, ".\\")
	path = strings.TrimPrefix(path, "data\\")
	if path == "" {
		return ""
	}
	if filepath.Ext(path) == "" {
		path += ".wav"
	}
	return path
}

func sfxPathCandidates(path string) []string {
	normalized := normalizeSFXPath(path)
	if normalized == "" {
		return nil
	}
	slash := strings.ReplaceAll(normalized, "\\", "/")
	candidates := []string{normalized, slash}
	lower := strings.ToLower(normalized)
	if strings.HasPrefix(lower, "wav\\") {
		candidates = append(candidates,
			"data\\"+normalized,
			"data/"+slash,
		)
	} else {
		candidates = append(candidates,
			"wav\\"+normalized,
			"wav/"+slash,
			"data\\wav\\"+normalized,
			"data/wav/"+slash,
		)
	}
	return uniqueStrings(candidates)
}

func sameAudioPath(a, b string) bool {
	return strings.EqualFold(normalizeAudioPath(a), normalizeAudioPath(b))
}

func uniqueStrings(values []string) []string {
	out := values[:0]
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}
