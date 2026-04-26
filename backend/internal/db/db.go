package database

import (
	"context"

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

	if err := SetupSchema(db_conn); err != nil {
		return nil, err
	}

	return &DB{db_conn}, nil
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
		return nil
	}

	return tx.Commit()
}

func SetupSchema(db *sqlx.DB) error {
	_, err := db.Exec(schema)
	return err
}

var schema = `
PRAGMA foreign_keys = OFF;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    type TEXT CHECK(type IN ('Registered', 'Guest')),

    soundcloud_id TEXT,
    soundcloud_key TEXT,
    tidal_id TEXT,
    tidal_key TEXT,

    spotify_id TEXT,
    spotify_auth_key TEXT,
    spotify_refresh_key TEXT,
		spotify_token_expires_at DATETIME,

    current_room_id INTEGER,
    current_role TEXT CHECK(current_role IN ('Host', 'DJ', 'Guest')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (current_room_id) REFERENCES rooms(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS  rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    host_id INTEGER NOT NULL,
    last_played_position INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (host_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS  tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_uri TEXT NOT NULL,
    artist TEXT,
    title TEXT NOT NULL,
    album TEXT,
    cover_url TEXT,
    platform TEXT
);

CREATE TABLE IF NOT EXISTS  room_tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER NOT NULL,
    track_id INTEGER NOT NULL,
    added_by INTEGER NOT NULL,
    position INTEGER NOT NULL,
    status TEXT NOT NULL, -- np. 'playing', 'queued', 'skipped', 'played'
    start_timestamp DATETIME,
    end_timestamp DATETIME,
    like_count INTEGER DEFAULT 0,
    skip_count INTEGER DEFAULT 0,
    
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    FOREIGN KEY (added_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS  archivals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER,
    name TEXT NOT NULL,
    created_at DATE DEFAULT CURRENT_DATE,
    
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS  archival_tracks (
    archival_id INTEGER NOT NULL,
    track_id INTEGER NOT NULL,
    
    FOREIGN KEY (archival_id) REFERENCES archivals(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    PRIMARY KEY (archival_id, track_id) 
);
`
