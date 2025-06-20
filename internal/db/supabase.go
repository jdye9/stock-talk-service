package db

import (
	"database/sql"

	_ "github.com/lib/pq"
)

func InitSupabase(connString string) (*sql.DB, error) {

    db, err := sql.Open("postgres", connString)
    if err != nil {
        return nil, err
    }
    // Optionally ping to verify connection
    if err := db.Ping(); err != nil {
        return nil, err
    }
    return db, nil
}