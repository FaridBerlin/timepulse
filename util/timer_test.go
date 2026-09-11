package util

import "testing"

func TestTimerTotalSeconds(t *testing.T) {
	tests := []struct {
		name    string
		hour    string
		minute  string
		second  string
		time    string
		want    int
		wantErr bool
	}{
		{name: "seconds are not scaled", hour: "", minute: "", second: "30", want: 30},
		{name: "classic hour+minute+second", hour: "1", minute: "1", second: "1", want: 3661},
		{name: "no flags defaults to five minutes", want: 300},
		{name: "lazy time flag takes priority over unit flags", hour: "1", minute: "1", second: "1", time: "0:00:05", want: 5},
		{name: "time flag alone", time: "1:20:01", want: 4801},
		{name: "invalid time flag errors", time: "not-a-time", wantErr: true},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := timerTotalSeconds(test.hour, test.minute, test.second, test.time)
			if test.wantErr {
				if err == nil {
					t.Fatalf("timerTotalSeconds(%q, %q, %q, %q) expected an error, got none", test.hour, test.minute, test.second, test.time)
				}
				return
			}
			if err != nil {
				t.Fatalf("timerTotalSeconds(%q, %q, %q, %q) returned unexpected error: %v", test.hour, test.minute, test.second, test.time, err)
			}
			if got != test.want {
				t.Fatalf("timerTotalSeconds(%q, %q, %q, %q) = %d, want %d", test.hour, test.minute, test.second, test.time, got, test.want)
			}
		})
	}
}
