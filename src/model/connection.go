package model

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func ConnectDatabase() (*sql.DB, error) {
	db, err := sql.Open("postgres", "postgresql://root@gabrielortega:26257/serverinformation?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}

	return db, nil
}
