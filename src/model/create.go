package model

import (
	"database/sql"
	"time"

	_ "github.com/lib/pq"
)

func CreateRowDatabase(db *sql.DB, id string, requestDomain string, requestHash string, previousGrade string) {
	var timeNow int64 = time.Now().UnixNano() / int64(time.Millisecond)
	sqlStatement := `INSERT INTO domainsV2 (id, requestDomain, requestHash, previousGrade, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := db.Exec(sqlStatement, id, requestDomain, requestHash, previousGrade, timeNow, timeNow)
	if err != nil {
		panic(err)
	}
}
