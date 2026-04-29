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

func TestResolvePgUrl_PrefersUrl(t *testing.T) {
	want := "postgres://override:secret@example.com:5433/foo?sslmode=require"
	got := ResolvePgUrl(
		want,
		"db_user", "db_password", "localhost", 5432, "db_name", "disable",
	)

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}

func TestResolvePgUrl_FallsBackToComponents(t *testing.T) {
	got := ResolvePgUrl(
		"",
		"db_user", "db_password", "localhost", 5432, "db_name", "disable",
	)

	want := "postgresql://db_user:db_password@localhost:5432/db_name?sslmode=disable"

	if got != want {
		t.Errorf("got %s, want %s", got, want)
	}
}
