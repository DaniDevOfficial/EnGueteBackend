package db

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
)

func InitDB() *sql.DB {
	var err error
	// PostgreSQL connection string format: "postgres://username:password@host:port/dbname?sslmode=disable"
	connStr := "postgres://enguete:root@localhost:5433/enguete?sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to the database:", err)
	}

	return db
}
