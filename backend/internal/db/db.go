package database

import (
	"github.com/jmoiron/sqlx"
)

var schema = `
PRAGMA foreign_keys = OFF;

CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    email TEXT UNIQUE NOT NULL,
    username TEXT NOT NULL,
    type TEXT CHECK(type IN ('Registered', 'Guest')),
    spotify_id TEXT,
    soundcloud_id TEXT,
    tidal_id TEXT,
    spotify_key TEXT,
    soundcloud_key TEXT,
    tidal_key TEXT,
    current_room_id INTEGER,
    current_role TEXT CHECK(current_role IN ('Host', 'DJ', 'Guest')),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (current_room_id) REFERENCES rooms(id) ON DELETE SET NULL
);

CREATE TABLE rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    host_id INTEGER NOT NULL,
    last_played_position INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    
    FOREIGN KEY (host_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE tracks (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    source_uri TEXT NOT NULL,
    artist TEXT,
    title TEXT NOT NULL,
    album TEXT,
    cover_url TEXT,
    platform TEXT
);

CREATE TABLE room_tracks (
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

CREATE TABLE archivals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER,
    name TEXT NOT NULL,
    created_at DATE DEFAULT CURRENT_DATE,
    
    FOREIGN KEY (room_id) REFERENCES rooms(id) ON DELETE SET NULL
);

CREATE TABLE archival_tracks (
    archival_id INTEGER NOT NULL,
    track_id INTEGER NOT NULL,
    
    FOREIGN KEY (archival_id) REFERENCES archivals(id) ON DELETE CASCADE,
    FOREIGN KEY (track_id) REFERENCES tracks(id) ON DELETE CASCADE,
    PRIMARY KEY (archival_id, track_id) 
);

PRAGMA foreign_keys = ON;
`

type SqliteDB struct {
	db *sqlx.DB
}

func InitSqliteDB(path string) (*SqliteDB, error) {
	db_conn, err := sqlx.Connect("sqlite", path)
	if err != nil {
		return nil, err
	}

	db_conn.MustExec("PRAGMA journal_mode = WAL;")
	db_conn.MustExec("PRAGMA synchronous = NORMAL;")

	db_conn.SetMaxOpenConns(1)

	if _, err := db_conn.Exec(schema); err != nil {
		return nil, err
	}

	return &SqliteDB{db: db_conn}, nil
}

