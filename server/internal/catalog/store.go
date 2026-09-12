package catalog

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

type store struct {
	db *sql.DB
}

func openStore(path string) (*store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite", filepath.ToSlash(path)+"?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS items (
		id TEXT PRIMARY KEY,
		relpath TEXT NOT NULL UNIQUE,
		title TEXT NOT NULL,
		filename TEXT NOT NULL,
		size INTEGER NOT NULL,
		mtime_unix INTEGER NOT NULL
	)`); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &store{db: db}, nil
}

func (s *store) close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *store) replaceAll(files []File) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()
	if _, err := tx.Exec(`DELETE FROM items`); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO items (id, relpath, title, filename, size, mtime_unix) VALUES (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, file := range files {
		if _, err := stmt.Exec(file.Item.ID, file.rel, file.Item.Title, file.Item.Filename, file.size, file.mtime); err != nil {
			return fmt.Errorf("persist %s: %w", file.Item.ID, err)
		}
	}
	return tx.Commit()
}
