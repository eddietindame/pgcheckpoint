package checkpoint

import (
	"testing"
)

func TestSanitizeName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"spaces to hyphens", "before migration", "before-migration"},
		{"underscores to hyphens", "before_migration", "before-migration"},
		{"uppercase", "Before Migration", "before-migration"},
		{"special chars", "my!!checkpoint@v2", "mycheckpointv2"},
		{"consecutive hyphens", "a--b---c", "a-b-c"},
		{"leading trailing hyphens", "-abc-", "abc"},
		{"mixed", " My_Cool Checkpoint!! ", "my-cool-checkpoint"},
		{"already clean", "pre-deploy", "pre-deploy"},
		{"empty", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeName(tt.input)
			if got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
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
		{"named", "checkpoint_3_before-migration.sql", 3, false},
		{"named double digit", "checkpoint_12_schema-v2.sql", 12, false},
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
		{"no suffix", caseNoSuffix, "checkpoint.sql", 0, false},
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
		cpName  string
		want    string
	}{
		{"first", 0, "/tmp/pgcheckpoint", "", "/tmp/pgcheckpoint/checkpoint_1.sql"},
		{"third", 2, "/tmp/pgcheckpoint", "", "/tmp/pgcheckpoint/checkpoint_3.sql"},
		{"custom dir", 5, "/home/user/dumps", "", "/home/user/dumps/checkpoint_6.sql"},
		{"with name", 2, "/tmp/pgcheckpoint", "before-migration", "/tmp/pgcheckpoint/checkpoint_3_before-migration.sql"},
		{"with unsanitized name", 0, "/tmp/pgcheckpoint", "Before Migration!", "/tmp/pgcheckpoint/checkpoint_1_before-migration.sql"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getNextCheckpointFilePath(tt.largest, tt.dir, tt.cpName)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
