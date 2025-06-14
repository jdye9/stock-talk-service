package db

import (
	"database/sql"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

// Schema definitions for stocks.db
var StockSchemas = map[string]string{
	"nasdaq": "symbol TEXT PRIMARY KEY, name TEXT NOT NULL",
	"other":   "symbol TEXT PRIMARY KEY, name TEXT NOT NULL",
}

// Schema definitions for crypto.db
var CryptoSchemas = map[string]string{
	"crypto": "uid TEXT PRIMARY KEY, symbol TEXT NOT NULL, name TEXT NOT NULL, id TEXT NOT NULL",
}


func InitSQLite(path string, tableSchemas map[string]string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db, tableSchemas); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func migrate(db *sql.DB, tableSchemas map[string]string) error {
	for tableName, schema := range tableSchemas {
		query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (%s);", tableName, schema)
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to create table %s: %w", tableName, err)
		}
	}
	return nil
}
