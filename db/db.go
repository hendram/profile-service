package db 

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

var DB *sql.DB

func InitDB() {
	var err error

	// adjust user/password/host if needed
	dsn := "host=localhost port=5432 user=postgres password=postgres dbname=onlineshop sslmode=disable"

	DB, err = sql.Open("postgres", dsn)
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to PostgreSQL")
}
