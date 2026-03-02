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
		want    time.Time
		wantErr bool
	}{
		{"valid", "checkpoint_2026-03-02_15-30-45.sql", time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC), false},
		{"another valid", "checkpoint_2025-12-31_23-59-59.sql", time.Date(2025, 12, 31, 23, 59, 59, 0, time.UTC), false},
		{"invalid format", "checkpoint_3.sql", time.Time{}, true},
		{"garbage", "garbage.sql", time.Time{}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCheckpointTimestamp(tt.input)
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
		wantFile string
		wantTime time.Time
		wantErr  bool
	}{
		{
			"basic",
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-03-02_15-30-45.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
			},
			"checkpoint_2026-03-02_15-30-45.sql",
			time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC),
			false,
		},
		{
			"single",
			[]string{"checkpoint_2026-01-01_10-00-00.sql"},
			"checkpoint_2026-01-01_10-00-00.sql",
			time.Date(2026, 1, 1, 10, 0, 0, 0, time.UTC),
			false,
		},
		{
			"error",
			[]string{"checkpoint_bad.sql"},
			"",
			time.Time{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotFile, gotTime, err := getLatestTimestampCheckpoint(tt.input)
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
	got := getNextTimestampCheckpointFilePath("/tmp/pgcheckpoint")
	prefix := "/tmp/pgcheckpoint/checkpoint_"
	suffix := ".sql"
	if !strings.HasPrefix(got, prefix) {
		t.Errorf("expected prefix %s, got %s", prefix, got)
	}
	if !strings.HasSuffix(got, suffix) {
		t.Errorf("expected suffix %s, got %s", suffix, got)
	}
	// Extract timestamp portion and verify it parses
	tsStr := strings.TrimSuffix(strings.TrimPrefix(got, prefix), suffix)
	if _, err := time.Parse(timestampFormat, tsStr); err != nil {
		t.Errorf("timestamp portion %q does not parse: %v", tsStr, err)
	}
}

func TestTimestampCheckpointsToDelete(t *testing.T) {
	latest := time.Date(2026, 3, 2, 15, 30, 45, 0, time.UTC)
	tests := []struct {
		name    string
		input   []string
		latest  time.Time
		want    []string
		wantErr bool
	}{
		{
			"basic",
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
				"checkpoint_2026-03-02_15-30-45.sql",
			},
			latest,
			[]string{
				"checkpoint_2026-01-01_10-00-00.sql",
				"checkpoint_2026-02-15_08-00-00.sql",
			},
			false,
		},
		{
			"single latest",
			[]string{"checkpoint_2026-03-02_15-30-45.sql"},
			latest,
			nil,
			false,
		},
		{
			"error",
			[]string{"checkpoint_bad.sql"},
			latest,
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := timestampCheckpointsToDelete(tt.input, tt.latest)
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
