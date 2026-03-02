package checkpoint

import (
	"strings"
	"testing"
	"time"
)

func TestParseCheckpointTimestamp(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mode    NamingMode
		want    time.Time
		wantErr bool
	}{
		{"timestamp valid", "checkpoint_2026-03-02_15-30-45.sql", NamingModeTimestamp, time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC), false},
		{"timestamp another", "checkpoint_2025-12-31_23-59-59.sql", NamingModeTimestamp, time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC), false},
		{"timestamp invalid", "checkpoint_3.sql", NamingModeTimestamp, time.Time{}, true},
		{"timestamp garbage", "garbage.sql", NamingModeTimestamp, time.Time{}, true},
		{"compact valid", "checkpoint_20260302T153045.sql", NamingModeCompact, time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC), false},
		{"compact invalid", "checkpoint_3.sql", NamingModeCompact, time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCheckpointTimestamp(tt.input, timestampFormatForMode(tt.mode))
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCheckpointUnix(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    time.Time
		wantErr bool
	}{
		{"valid", "checkpoint_1740934245.sql", time.Unix(1740934245, 0), false},
		{"invalid", "checkpoint_abc.sql", time.Time{}, true},
		{"garbage", "garbage.sql", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCheckpointUnix(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseCheckpointTime(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mode    NamingMode
		want    time.Time
		wantErr bool
	}{
		{"timestamp", "checkpoint_2026-03-02_15-30-45.sql", NamingModeTimestamp, time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC), false},
		{"compact", "checkpoint_20260302T153045.sql", NamingModeCompact, time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC), false},
		{"unix", "checkpoint_1740934245.sql", NamingModeUnix, time.Unix(1740934245, 0), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCheckpointTime(tt.input, tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if !got.Equal(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLatestTimestampCheckpoint(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		mode     NamingMode
		wantFile string
		wantTime time.Time
		wantErr  bool
	}{
		{
			"timestamp basic",
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-03-02_15-30-45.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
			},
			NamingModeTimestamp,
			"checkpoint_2026-03-02_15-30-45.sql",
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			false,
		},
		{
			"compact basic",
			[]string{
				"checkpoint_20260101T100000.sql",
				"checkpoint_20260302T153045.sql",
				"checkpoint_20260215T080000.sql",
			},
			NamingModeCompact,
			"checkpoint_20260302T153045.sql",
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			false,
		},
		{
			"unix basic",
			[]string{
				"checkpoint_1740000000.sql",
				"checkpoint_1740934245.sql",
				"checkpoint_1740500000.sql",
			},
			NamingModeUnix,
			"checkpoint_1740934245.sql",
			time.Unix(1740934245, 0),
			false,
		},
		{
			"timestamp single",
			[]string{"checkpoint_2026-01-01_10-00-00.sql"},
			NamingModeTimestamp,
			"checkpoint_2026-01-01_10-00-00.sql",
			time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			false,
		},
		{
			"error",
			[]string{"checkpoint_bad.sql"},
			NamingModeTimestamp,
			"",
			time.Time{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotTime, err := getLatestTimestampCheckpoint(tt.input, tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if gotFile != tt.wantFile {
				t.Errorf("got file %s, want %s", gotFile, tt.wantFile)
			}
			if !gotTime.Equal(tt.wantTime) {
				t.Errorf("got time %v, want %v", gotTime, tt.wantTime)
			}
		})
	}
}

func TestGetNextTimestampCheckpointFilePath(t *testing.T) {
	modes := []struct {
		name string
		mode NamingMode
	}{
		{"timestamp", NamingModeTimestamp},
		{"compact", NamingModeCompact},
		{"unix", NamingModeUnix},
	}

	for _, tt := range modes {
		t.Run(tt.name, func(t *testing.T) {
			got := getNextTimestampCheckpointFilePath("/tmp/pgcheckpoint", tt.mode)
			dirPrefix := "/tmp/pgcheckpoint/"
			if !strings.HasPrefix(got, dirPrefix) {
				t.Errorf("expected dir prefix %s, got %s", dirPrefix, got)
			}
			if !strings.HasSuffix(got, ".sql") {
				t.Errorf("expected .sql suffix, got %s", got)
			}
			// Verify the filename can be parsed back
			filename := strings.TrimPrefix(got, dirPrefix)
			_, err := parseCheckpointTime(filename, tt.mode)
			if err != nil {
				t.Errorf("generated filename %q does not round-trip parse: %v", filename, err)
			}
		})
	}
}

func TestTimestampCheckpointsToDelete(t *testing.T) {
	tests := []struct {
		name    string
		input   []string
		latest  time.Time
		mode    NamingMode
		want    []string
		wantErr bool
	}{
		{
			"timestamp basic",
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
				"checkpoint_2026-03-02_15-30-45.sql",
			},
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			NamingModeTimestamp,
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
			},
			false,
		},
		{
			"compact basic",
			[]string{
				"checkpoint_20260101T100000.sql",
				"checkpoint_20260302T153045.sql",
			},
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			NamingModeCompact,
			[]string{"checkpoint_20260101T100000.sql"},
			false,
		},
		{
			"unix basic",
			[]string{
				"checkpoint_1740000000.sql",
				"checkpoint_1740934245.sql",
			},
			time.Unix(1740934245, 0),
			NamingModeUnix,
			[]string{"checkpoint_1740000000.sql"},
			false,
		},
		{
			"single latest",
			[]string{"checkpoint_2026-03-02_15-30-45.sql"},
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			NamingModeTimestamp,
			nil,
			false,
		},
		{
			"error",
			[]string{"checkpoint_bad.sql"},
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			NamingModeTimestamp,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timestampCheckpointsToDelete(tt.input, tt.latest, tt.mode)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
				return
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("got %v, want %v", got, tt.want)
					return
				}
			}
		})
	}
}
