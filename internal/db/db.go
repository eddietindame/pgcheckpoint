package db

import "fmt"

// GetPgUrl builds a PostgreSQL connection string.
func GetPgUrl(user, password, host string, port int, dbname, sslmode string) string {
	return fmt.Sprintf("postgresql://%s:%s@%s:%d/%s?sslmode=%s",
		user, password, host, port, dbname, sslmode)
}
