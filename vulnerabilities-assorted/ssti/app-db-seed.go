package main

import (
    "fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

// SQLite3 database schema
var filesSchema = `
CREATE TABLE IF NOT EXISTS posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	title TEXT,
	content TEXT
);
`

func main() {
	db, err := sqlx.Open("sqlite3", "./posts.db")
	if err != nil {
	  panic(err)
	}
	defer db.Close()
  
	// Ping the database to check connectivity
	err = db.Ping()
	if err != nil {
	  panic(err)
	}
  
	// Execute the schema to create the table (if it doesn't exist)
	_, err = db.Exec(filesSchema)
	if err != nil {
	  panic(err)
	}
  
	fmt.Println("Database schema applied successfully!")
  }