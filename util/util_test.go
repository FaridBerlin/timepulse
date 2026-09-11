package util

import (
	"testing"
	"time"

	"github.com/nsf/termbox-go"
)

func TestStopwatchFormatTime(t *testing.T) {
	tests := []struct {
		name               string
		d                  time.Duration
		disableMillisecond bool
		want               string
	}{
		{name: "with milliseconds", d: time.Hour + 2*time.Minute + 3*time.Second + 456*time.Millisecond, disableMillisecond: false, want: "01:02:03.456"},
		{name: "without milliseconds", d: time.Hour + 2*time.Minute + 3*time.Second + 456*time.Millisecond, disableMillisecond: true, want: "01:02:03"},
		{name: "zero", d: 0, disableMillisecond: false, want: "00:00:00.000"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StopwatchFormatTime(test.d, test.disableMillisecond); got != test.want {
				t.Fatalf("StopwatchFormatTime(%s, %v) = %q, want %q", test.d, test.disableMillisecond, got, test.want)
			}
		})
	}
}

func TestStopwatchFormatTimeWihtoutHour(t *testing.T) {
	tests := []struct {
		name               string
		d                  time.Duration
		disableMillisecond bool
		want               string
	}{
		{name: "with milliseconds", d: 2*time.Minute + 3*time.Second + 456*time.Millisecond, disableMillisecond: false, want: "02:03.456"},
		{name: "without milliseconds", d: 2*time.Minute + 3*time.Second + 456*time.Millisecond, disableMillisecond: true, want: "02:03"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := StopwatchFormatTimeWihtoutHour(test.d, test.disableMillisecond); got != test.want {
				t.Fatalf("StopwatchFormatTimeWihtoutHour(%s, %v) = %q, want %q", test.d, test.disableMillisecond, got, test.want)
			}
		})
	}
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name            string
		hour            int
		use12HourFormat bool
		want            string
	}{
		{name: "24h midnight", hour: 0, use12HourFormat: false, want: "00:30:15"},
		{name: "24h afternoon", hour: 15, use12HourFormat: false, want: "15:30:15"},
		{name: "12h midnight becomes 12", hour: 0, use12HourFormat: true, want: "12:30:15"},
		{name: "12h afternoon wraps", hour: 15, use12HourFormat: true, want: "03:30:15"},
		{name: "12h noon stays 12", hour: 12, use12HourFormat: true, want: "12:30:15"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			d := time.Date(2024, 1, 1, test.hour, 30, 15, 0, time.UTC)
			if got := formatTime(d, test.use12HourFormat); got != test.want {
				t.Fatalf("formatTime(hour=%d, 12h=%v) = %q, want %q", test.hour, test.use12HourFormat, got, test.want)
			}
		})
	}
}

func TestFlagColor(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  termbox.Attribute
	}{
		{name: "named color", input: "red", want: termbox.ColorRed},
		{name: "empty defaults to green", input: "", want: termbox.ColorGreen},
		{name: "numeric code", input: "42", want: termbox.Attribute(42)},
		{name: "out of range numeric falls back to green", input: "999", want: termbox.ColorGreen},
		{name: "unknown name falls back to green", input: "not-a-color", want: termbox.ColorGreen},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := FlagColor(test.input); got != test.want {
				t.Fatalf("FlagColor(%q) = %v, want %v", test.input, got, test.want)
			}
		})
	}
}

func TestTimeStringToSeconds(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{name: "valid", input: "01:02:03", want: 3723},
		{name: "wrong number of parts", input: "01:02", wantErr: true},
		{name: "non-numeric hour", input: "aa:02:03", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := timeStringToSeconds(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("timeStringToSeconds(%q) expected an error, got none", test.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("timeStringToSeconds(%q) returned unexpected error: %v", test.input, err)
			}
			if got != test.want {
				t.Fatalf("timeStringToSeconds(%q) = %d, want %d", test.input, got, test.want)
			}
		})
	}
}
