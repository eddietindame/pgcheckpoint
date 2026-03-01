package checkpoint

import "testing"

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
