package res

import "testing"

func TestParseSkillNameTable(t *testing.T) {
	nameToID := map[string]int{"sm_bash": 5}
	got := parseSkillNameTable([]byte("// comment\nSM_BASH#Bash#\nUNKNOWN#Nope#\n"), nameToID)
	if got[5] != "Bash" {
		t.Fatalf("skill name = %q", got[5])
	}
	if _, ok := got[0]; ok {
		t.Fatal("zero skill id should not be populated")
	}
}

func TestParseSkillDescriptionTable(t *testing.T) {
	nameToID := map[string]int{"sm_bash": 5}
	names, descriptions := parseSkillDescriptionTable([]byte("SM_BASH#\nBash\nStrike an enemy.\nConsumes SP.\n#\n"), nameToID)
	if names[5] != "Bash" {
		t.Fatalf("skill title = %q", names[5])
	}
	lines := descriptions[5]
	if len(lines) != 2 || lines[0] != "Strike an enemy." || lines[1] != "Consumes SP." {
		t.Fatalf("description = %#v", lines)
	}
}

func TestSkillMetadataLookupCopiesDescription(t *testing.T) {
	manager := &Manager{
		skillMetadataLoaded: true,
		skillDisplayNames: map[int]string{
			5: "Bash",
		},
		skillDescriptions: map[int][]string{
			5: []string{"Strike an enemy."},
		},
	}
	if got, ok := manager.SkillDisplayName(5); !ok || got != "Bash" {
		t.Fatalf("display name = %q ok=%v", got, ok)
	}
	lines, ok := manager.SkillDescription(5)
	if !ok || len(lines) != 1 || lines[0] != "Strike an enemy." {
		t.Fatalf("description = %#v ok=%v", lines, ok)
	}
	lines[0] = "mutated"
	if lines, _ := manager.SkillDescription(5); lines[0] != "Strike an enemy." {
		t.Fatalf("description was not copied: %#v", lines)
	}
}
