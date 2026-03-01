package checkpoint

import (
	"testing"
)

func TestGetPgUrl(t *testing.T) {
	got := GetPgUrl(5432)
	want := "postgresql://bertie_user_backend:bertie_password_backend@localhost:5432/bertie_db_backend?sslmode=disable"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestParseCheckpointNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"basic", "checkpoint_3.sql", 3, false},
		{"double digit", "checkpoint_12.sql", 12, false},
		{"no suffix", "checkpoint.sql", 0, false},
		{"invalid", "garbage.sql", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseCheckpointNumber(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestGetLatestCheckpoint(t *testing.T) {
	caseBasic := []string{
		"checkpoint_2.sql",
		"checkpoint_1.sql",
		"checkpoint_3.sql",
	}

	caseError1 := []string{
		"checkpoint_1.sql",
		"checkpoint_poop.sql",
	}

	caseNoSuffix := []string{
		"checkpoint.sql",
	}

	tests := []struct {
		name    string
		input   []string
		wantStr string
		wantInt int
		wantErr bool
	}{
		{"basic", caseBasic, "checkpoint_3.sql", 3, false},
		{"error1", caseError1, "", 0, true},
		{"no suffix", caseNoSuffix, "checkpoint_0.sql", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotStr, gotInt, err := getLatestCheckpoint(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("unexpected error: %v", err)
			}
			if gotStr != tt.wantStr {
				t.Errorf("got str %s, want str %s", gotStr, tt.wantStr)
			}
			if gotInt != tt.wantInt {
				t.Errorf("got int %d, want int %d", gotInt, tt.wantInt)
			}
		})
	}
}

func TestCheckpointsToDelete(t *testing.T) {
	tests := []struct {
		name           string
		filenamesInput []string
		latestInput    int
		want           []string
		wantErr        bool
	}{
		{
			"basic",
			[]string{"checkpoint_1.sql", "checkpoint_2.sql", "checkpoint_3.sql"},
			3,
			[]string{"checkpoint_1.sql", "checkpoint_2.sql"},
			false,
		},
		{
			"single",
			[]string{"checkpoint_1.sql"},
			1,
			[]string{},
			false,
		},
		{
			"weird order",
			[]string{"checkpoint_1.sql", "checkpoint_3.sql"},
			3,
			[]string{"checkpoint_1.sql"},
			false,
		},
		{
			"error",
			[]string{"checkpoint_1.sql", "checkpointerror.sql"},
			3,
			[]string{},
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := checkpointsToDelete(tt.filenamesInput, tt.latestInput)
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

func TestGetNextCheckpointFilePath(t *testing.T) {
	tests := []struct {
		name    string
		largest int
		dir     string
		want    string
	}{
		{"first", 0, "/tmp/pgcheckpoint", "/tmp/pgcheckpoint/checkpoint_1"},
		{"third", 2, "/tmp/pgcheckpoint", "/tmp/pgcheckpoint/checkpoint_3"},
		{"custom dir", 5, "/home/user/dumps", "/home/user/dumps/checkpoint_6"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getNextCheckpointFilePath(tt.largest, tt.dir)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
