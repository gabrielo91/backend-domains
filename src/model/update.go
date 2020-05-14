package model

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"log"
)

func UpdateRowDatabase(db *sql.DB, id string, requestHash string, update_at int64, grade string) {
	stmt, err := db.Prepare("UPDATE domainsV2 SET requestHash = $1, updated_at = $2, previousGrade =$3 WHERE id=$4")
	if err != nil {
		log.Fatal(err)
	}

	defer stmt.Close()

	res, err := stmt.Exec(requestHash, update_at, grade, id)
	if err != nil {
		log.Fatal(err)
	}

	affect, err := res.RowsAffected()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(affect, "rows changed")
}
