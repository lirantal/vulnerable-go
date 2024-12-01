package main

import (
    "fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/jmoiron/sqlx"
)

// SQLite3 database schema
var filesSchema = `
CREATE TABLE IF NOT EXISTS files (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	filename TEXT NOT NULL,
	signature TEXT NOT NULL,
	tenant_id TEXT NOT NULL,
	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
`

var usersSchema = `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL,
	email TEXT NOT NULL	
);`

func main() {
	db, err := sqlx.Open("sqlite3", "./mydb.db")
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

	_, err = db.Exec(usersSchema)
	if err != nil {
	  panic(err)
	}
  
	fmt.Println("Database schema applied successfully!")
  }