package util

import (
	"testing"
)

func TestMinInt(t *testing.T) {
	cases := []struct {
		name string
		x    int
		y    int
		exp  int
	}{
		{name: "same", x: 1, y: 1, exp: 1},
		{name: "x>y", x: 3, y: 2, exp: 2},
		{name: "x<y", x: 4, y: 5, exp: 4},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := MinInt(tc.x, tc.y)
			if got != tc.exp {
				t.Errorf("expected %v; got %v", tc.exp, got)
			}
		})
	}
}
