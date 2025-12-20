package db

import (
	"database/sql"
	"log"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func InitDB(dataSourceName string) {
	var err error
	// Enable WAL mode and set a busy timeout to handle concurrent writes
	// _busy_timeout=5000: Wait up to 5000ms before erroring with SQLITE_BUSY
	// _journal_mode=WAL: Write-Ahead Logging allows better concurrency
	DB, err = sql.Open("sqlite", dataSourceName+"?_busy_timeout=5000&_journal_mode=WAL")
	if err != nil {
		log.Fatal(err)
	}

	if err = DB.Ping(); err != nil {
		log.Fatal(err)
	}

	createTables()
}

func createTables() {
	createFreezerTable := `
	CREATE TABLE IF NOT EXISTS freezers (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createCategoryTable := `
	CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL UNIQUE
	);`

	createItemTable := `
	CREATE TABLE IF NOT EXISTS items (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT NOT NULL,
		category_id INTEGER,
		freezer_id INTEGER NOT NULL,
		weight TEXT,
		frozen_date DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY(category_id) REFERENCES categories(id),
		FOREIGN KEY(freezer_id) REFERENCES freezers(id)
	);`

	_, err := DB.Exec(createFreezerTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createCategoryTable)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(createItemTable)
	if err != nil {
		log.Fatal(err)
	}

	seedData()
}

func seedData() {
	var count int
	row := DB.QueryRow("SELECT COUNT(*) FROM categories")
	if err := row.Scan(&count); err != nil {
		log.Println("Error checking categories:", err)
		return
	}
	if count == 0 {
		_, err := DB.Exec("INSERT INTO categories (name) VALUES ('Uncategorized')")
		if err != nil {
			log.Println("Error seeding categories:", err)
		} else {
			log.Println("Seeded 'Uncategorized' category")
		}
	}
}
