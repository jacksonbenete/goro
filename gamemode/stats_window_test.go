package gamemode

import "testing"

func TestStatPointCostPreRenewal(t *testing.T) {
	tests := []struct {
		current int
		want    int
	}{
		{current: 1, want: 2},
		{current: 10, want: 2},
		{current: 11, want: 3},
		{current: 20, want: 3},
		{current: 91, want: 11},
	}
	for _, test := range tests {
		if got := statPointCost(test.current); got != test.want {
			t.Fatalf("statPointCost(%d) = %d, want %d", test.current, got, test.want)
		}
	}
}

func TestStatCostPrefersServerValue(t *testing.T) {
	row := statRow{value: 31, cost: 8}
	if got := statCost(row); got != 8 {
		t.Fatalf("statCost() = %d, want 8", got)
	}
}
