package checkpoint

import "testing"

func TestParseNamingMode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    NamingMode
		wantErr bool
	}{
		{"sequential", "sequential", NamingModeSequential, false},
		{"timestamp", "timestamp", NamingModeTimestamp, false},
		{"compact", "compact", NamingModeCompact, false},
		{"unix", "unix", NamingModeUnix, false},
		{"invalid", "foo", NamingMode{}, true},
		{"empty", "", NamingMode{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseNamingMode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNamingModeIsTimestampBased(t *testing.T) {
	tests := []struct {
		name string
		mode NamingMode
		want bool
	}{
		{"sequential", NamingModeSequential, false},
		{"timestamp", NamingModeTimestamp, true},
		{"compact", NamingModeCompact, true},
		{"unix", NamingModeUnix, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.mode.IsTimestampBased()
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
