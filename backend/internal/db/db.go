package database

import (
	"auxie/backend/migrations"
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

// Common interface for sqlx.DB and sqlx.TX
type SQLHandler interface {
	sqlx.Ext
	sqlx.Preparer
	Get(dest interface{}, query string, args ...interface{}) error
	Select(dest interface{}, query string, args ...interface{}) error
	SelectContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type DB struct {
	*sqlx.DB
}

func InitSqliteDB(path string) (*DB, error) {
	db_conn, err := sqlx.Connect("sqlite3", path)
	if err != nil {
		return nil, err
	}

	db_conn.MustExec("PRAGMA journal_mode = WAL;")
	db_conn.MustExec("PRAGMA synchronous = NORMAL;")
	db_conn.MustExec("PRAGMA foreign_keys = ON;")

	db_conn.SetMaxOpenConns(1)

	db := &DB{db_conn}
	if err := db.RunMigrations(); err != nil {
		return nil, fmt.Errorf("migration error: %w", err)
	}

	return db, nil
}

func (db *DB) RunMigrations() error {
	// Tracking migrations
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (version TEXT PRIMARY KEY);`)
	if err != nil {
		return err
	}

	entries, err := migrations.FS.ReadDir(".")
	if err != nil {
		return err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".up.sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)

	// Apply every migration
	for _, file := range files {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM schema_migrations WHERE version = ?", file)
		if err != nil {
			return err
		}

		if count > 0 {
			continue
		}

		fmt.Printf("Running migration: %s\n", file)
		content, err := migrations.FS.ReadFile(file)
		if err != nil {
			return err
		}

		err = db.WithTransaction(context.Background(), func(q SQLHandler) error {
			if _, err := q.Exec(string(content)); err != nil {
				return err
			}
			if _, err := q.Exec("INSERT INTO schema_migrations (version) VALUES (?)", file); err != nil {
				return err
			}
			return nil
		})

		if err != nil {
			return fmt.Errorf("Error in file %s: %w", file, err)
		}
	}

	return nil
}

func (db *DB) WithTransaction(ctx context.Context, fn func(q SQLHandler) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
