package ui

import (
	"reflect"
	"testing"

	"github.com/kivutar/goro/client"
	"github.com/kivutar/goro/session"
)

func TestChangeCartChoicesFollowReferenceThresholds(t *testing.T) {
	tests := []struct {
		level int
		want  []int
	}{
		{1, []int{1}},
		{40, []int{1}},
		{41, []int{1, 2}},
		{66, []int{1, 2, 3}},
		{81, []int{1, 2, 3, 4}},
		{91, []int{1, 2, 3, 4, 5}},
		{101, []int{1, 2, 3, 4, 5, 6}},
		{111, []int{1, 2, 3, 4, 5, 6, 7}},
		{121, []int{1, 2, 3, 4, 5, 6, 7, 8}},
		{131, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}
	for _, test := range tests {
		if got := changeCartChoices(test.level); !reflect.DeepEqual(got, test.want) {
			t.Fatalf("level %d choices = %v, want %v", test.level, got, test.want)
		}
	}
}

func TestBaseLevelForChangeCartFallsBackToSelectedCharacter(t *testing.T) {
	ctx := client.Context{Session: &session.Session{Selected: session.Character{Level: 72}}}
	if got := baseLevelForChangeCart(ctx); got != 72 {
		t.Fatalf("base level = %d, want 72", got)
	}
	ctx.Session.Progress.BaseLevel = 80
	if got := baseLevelForChangeCart(ctx); got != 80 {
		t.Fatalf("progress base level = %d, want 80", got)
	}
}
