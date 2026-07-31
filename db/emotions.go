package db

import "strings"

type emotionEntry struct {
	id       uint8
	frame    int
	commands []string
}

type Emotion struct {
	ID       uint8
	Frame    int
	Command  string
	Commands []string
}

var emotionEntries = []emotionEntry{
	{0, 0, []string{"!"}},
	{1, 1, []string{"?"}},
	{2, 2, []string{"ho", "delight"}},
	{3, 3, []string{"lv", "heart"}},
	{4, 5, []string{"swt", "sweat"}},
	{5, 6, []string{"ic", "aha"}},
	{6, 7, []string{"an", "fret"}},
	{7, 8, []string{"ag", "anger"}},
	{8, 9, []string{"$", "money"}},
	{9, 10, []string{"..."}},
	{10, 12, []string{"scissors", "gawi"}},
	{11, 11, []string{"rock", "bawi"}},
	{12, 13, []string{"paper", "bo"}},
	{13, 14, nil},
	{14, 4, []string{"lv2"}},
	{15, 15, []string{"thx"}},
	{16, 16, []string{"wah"}},
	{17, 17, []string{"sry", "sorry"}},
	{18, 18, []string{"heh", "smile"}},
	{19, 19, []string{"swt2"}},
	{20, 20, []string{"hmm"}},
	{21, 21, []string{"no1"}},
	{22, 22, []string{"??"}},
	{23, 23, []string{"omg"}},
	{24, 24, []string{"oh", "o"}},
	{25, 25, []string{"x"}},
	{26, 26, []string{"hlp", "help"}},
	{27, 27, []string{"go"}},
	{28, 28, []string{"sob"}},
	{29, 29, []string{"gg"}},
	{30, 30, []string{"kis"}},
	{31, 31, []string{"kis2"}},
	{32, 32, []string{"pif"}},
	{33, 33, []string{"ok"}},
	{34, 1000, nil},
	{35, 34, nil},
	{36, 35, []string{"bzz", "stare", "e1"}},
	{37, 36, []string{"rice", "e2"}},
	{38, 37, []string{"awsm", "cool", "e3"}},
	{39, 38, []string{"meh", "e4"}},
	{40, 39, []string{"shy", "e5"}},
	{41, 40, []string{"pat", "goodboy", "e6"}},
	{42, 41, []string{"mp", "sptime", "e7"}},
	{43, 42, []string{"slur", "e8"}},
	{44, 43, []string{"com", "comeon", "e9"}},
	{45, 44, []string{"yawn", "sleepy", "e10"}},
	{46, 45, []string{"grat", "congrats", "e11"}},
	{47, 46, []string{"hp", "hptime", "e12"}},
	{48, 47, nil},
	{49, 48, nil},
	{50, 49, nil},
	{51, 50, nil},
	{52, 51, []string{"fsh", "e13"}},
	{53, 52, []string{"spin", "e14"}},
	{54, 53, []string{"sigh", "e15"}},
	{55, 54, []string{"dum", "e16"}},
	{56, 55, []string{"crwd", "e17"}},
	{57, 56, []string{"desp", "otl", "e18"}},
	{58, 57, []string{"dice", "e19"}},
	{59, 58, nil},
	{60, 59, nil},
	{61, 60, nil},
	{62, 61, nil},
	{63, 62, nil},
	{64, 63, nil},
	{65, 64, []string{"love", "e20"}},
	{66, 65, nil},
	{67, 66, nil},
	{68, 67, []string{"mobile", "e21"}},
	{69, 68, []string{"mail", "e22"}},
	{70, 69, []string{"antenna0", "e23"}},
	{71, 70, []string{"antenna1", "e24"}},
	{72, 71, []string{"antenna2", "e25"}},
	{73, 72, []string{"antenna3", "e26"}},
	{74, 73, []string{"hum", "e27"}},
	{75, 74, []string{"abs", "e28"}},
	{76, 75, []string{"oops", "e29"}},
	{77, 76, []string{"spit", "e30"}},
	{78, 77, []string{"ene", "e31"}},
	{79, 78, []string{"panic", "e32"}},
	{80, 79, []string{"whisp", "e33"}},
	{81, 80, nil},
	{82, 81, nil},
	{83, 82, nil},
	{84, 83, nil},
	{85, 84, nil},
	{86, 85, nil},
	{87, 86, nil},
}

var emotionCommandIDs = buildEmotionCommandIDs()
var emotionSpriteFrames = buildEmotionSpriteFrames()
var emotionFrameEntries = buildEmotionFrameEntries()

var emotionInterfaceFrames = []int{
	0, 1, 2, 3, 4, 5, 6, 7, 8, 9,
	10, 11, 12, 13, 15, 16, 17, 18, 19, 20,
	21, 22, 23, 24, 25, 26, 27, 28, 29, 30,
	31, 32, 33, 35, 36, 37, 38, 39, 40, 41,
	42, 43, 44, 45, 46, 51, 52, 53, 54, 55,
	56, 64, 67, 68, 70, 71, 72, 73, 74, 75,
	76, 77, 78, 79,
}

func buildEmotionCommandIDs() map[string]uint8 {
	out := make(map[string]uint8)
	for _, entry := range emotionEntries {
		for _, command := range entry.commands {
			out[strings.ToLower(command)] = entry.id
		}
	}
	return out
}

func buildEmotionSpriteFrames() map[uint8]int {
	out := make(map[uint8]int)
	for _, entry := range emotionEntries {
		out[entry.id] = entry.frame
	}
	return out
}

func buildEmotionFrameEntries() map[int]emotionEntry {
	out := make(map[int]emotionEntry)
	for _, entry := range emotionEntries {
		out[entry.frame] = entry
	}
	return out
}

func EmotionCommandID(command string) (uint8, bool) {
	id, ok := emotionCommandIDs[strings.ToLower(strings.TrimSpace(command))]
	return id, ok
}

func EmotionSpriteFrame(id uint8) (int, bool) {
	frame, ok := emotionSpriteFrames[id]
	return frame, ok
}

func EmotionList() []Emotion {
	out := make([]Emotion, 0, len(emotionInterfaceFrames))
	for _, frame := range emotionInterfaceFrames {
		entry, ok := emotionFrameEntries[frame]
		if !ok || len(entry.commands) == 0 {
			continue
		}
		commands := append([]string(nil), entry.commands...)
		out = append(out, Emotion{
			ID:       entry.id,
			Frame:    entry.frame,
			Command:  commands[0],
			Commands: commands,
		})
	}
	return out
}
