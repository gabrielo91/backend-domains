package model

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

type DomainsRequests struct {
	Item []string
}

func GetAllIdDatabase(db *sql.DB) DomainsRequests {

	sqlStatement := `select id FROM domainsV2;`

	rows, err := db.Query(sqlStatement)
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()

	var domainsRequests DomainsRequests
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			log.Fatal(err)
		}
		//fmt.Printf("%s", id)
		domainsRequests.Item = append(domainsRequests.Item, id)
	}
	return domainsRequests
}

func GetRowDatabase(db *sql.DB, id string) (string, string, int64) {

	sqlStatement := `select requestHash, previousGrade, updated_at FROM domainsV2 WHERE id = $1;`
	var requestHash string
	var previousGrade string
	var updated_at int64
	row := db.QueryRow(sqlStatement, id)
	switch err := row.Scan(&requestHash, &previousGrade, &updated_at); err {
	case sql.ErrNoRows:
		return "", "", time.Now().UnixNano() % 1e6 / 1e3
	case nil:
		return requestHash, previousGrade, updated_at
	default:
		panic(err)
	}
}
