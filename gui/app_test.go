package gui

import (
	"testing"
	"time"
)

func TestParseTimerInput(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  time.Duration
	}{
		{name: "valid", input: "01:02:03", want: time.Hour + 2*time.Minute + 3*time.Second},
		{name: "invalid format", input: "90", want: 5 * time.Minute},
		{name: "invalid minutes", input: "00:60:00", want: 5 * time.Minute},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := parseTimerInput(test.input); got != test.want {
				t.Fatalf("parseTimerInput(%q) = %s, want %s", test.input, got, test.want)
			}
		})
	}
}
