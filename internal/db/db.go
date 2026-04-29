package db

import "fmt"

// GetPgUrl builds a PostgreSQL connection string from individual components.
func GetPgUrl(user, password, host string, port int, dbname, sslmode string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}

// ResolvePgUrl returns url if non-empty, otherwise builds one from the
// individual components. This lets callers prefer a full DATABASE_URL-style
// connection string when provided and fall back to the per-field flags.
func ResolvePgUrl(url, user, password, host string, port int, dbname, sslmode string) string {
	if url != "" {
		return url
	}
	return GetPgUrl(user, password, host, port, dbname, sslmode)
}
