package db

import (
	"testing"
)

func TestGetPgUrl(t *testing.T) {
	got := GetPgUrl(
		"db_user",
		"db_password",
		"localhost",
		5432,
		"db_name",
		"disable",
	)

	want := "postgresql://db_user:db_password@localhost:5432/db_name?sslmode=disable"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
