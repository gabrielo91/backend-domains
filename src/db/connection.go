package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func ConnectDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgresql://root@gabrielortega:26257/serverinformation?sslmode=disable")
	if err != nil {
		return db, err
	}
	return db, nil
}
