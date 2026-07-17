package res

import "testing"

func TestPetTalkKeysFollowROBrowserLayout(t *testing.T) {
	mob, hunger, action, ok := petTalkKeys(1063010)
	if !ok {
		t.Fatal("petTalkKeys returned !ok")
	}
	if mob != "lunatic" || hunger != "bit_hungry" || action != "feeding" {
		t.Fatalf("keys = %q/%q/%q, want lunatic/bit_hungry/feeding", mob, hunger, action)
	}
}

func TestParsePetTalkTableKeepsRepeatedActions(t *testing.T) {
	xml := []byte(`<?xml version="1.0" encoding="utf-8"?>
<monster_talk_table>
  <lunatic>
    <bit_hungry>
      <feeding>first</feeding>
      <feeding>second</feeding>
    </bit_hungry>
  </lunatic>
</monster_talk_table>`)
	table, err := parsePetTalkTable(xml)
	if err != nil {
		t.Fatal(err)
	}
	got := table["lunatic"]["bit_hungry"]["feeding"]
	if len(got) != 2 || got[0] != "first" || got[1] != "second" {
		t.Fatalf("talks = %#v, want first/second", got)
	}
}
