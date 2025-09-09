package store

import (
	"context"
	"database/sql"
	"os"

	_ "modernc.org/sqlite"
)

func OpenSQLite(path string) (*sql.DB, error) {
	// enable WAL mode for better concurrency
	dsn := path + "?_pragma=journal_mode(WAL)"
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	return db, db.Ping()
}

func RunMigrations(db *sql.DB, file string) error {
	b, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	_, err = db.Exec(string(b))
	return err
}

func UpsertDocument(ctx context.Context, db *sql.DB, url, title, content string) error {
	_, err := db.ExecContext(ctx, `
	INSERT INTO documents (url, title, content)
	VALUES (?, ?, ?)
	ON CONFLICT(url) DO UPDATE SET
	  title=excluded.title,
	  content=excluded.content
	`, url, title, content)
	return err
}
