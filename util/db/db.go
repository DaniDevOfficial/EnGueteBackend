package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func InitDB(databaseUrl string) *sql.DB {
	db, err := sql.Open("postgres", databaseUrl)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to the database:", err)
	}

	return db
}
