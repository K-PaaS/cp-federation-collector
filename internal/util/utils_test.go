package util

import "testing"

func TestRound(t *testing.T) {
	type testCase struct {
		name     string
		input    float64
		decimals int
		want     float64
	}

	tests := []testCase{
		{
			name:     "round down",
			input:    1.2345,
			decimals: 2,
			want:     1.23,
		},
		{
			name:     "round up",
			input:    1.235,
			decimals: 2,
			want:     1.24,
		},
		{
			name:     "no decimal places round down",
			input:    1.4,
			decimals: 0,
			want:     1,
		},
		{
			name:     "no decimal places round up",
			input:    1.5,
			decimals: 0,
			want:     2,
		},
		{
			name:     "negative number",
			input:    -1.234,
			decimals: 2,
			want:     -1.23,
		},
		{
			name:     "many decimals",
			input:    0.123456,
			decimals: 4,
			want:     0.1235,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := Round(tt.input, tt.decimals)
			if got != tt.want {
				t.Fatalf("Round(%v, %d) = %v, want %v", tt.input, tt.decimals, got, tt.want)
			}
		})
	}
}
