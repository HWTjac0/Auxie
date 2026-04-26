package models

import (
	"database/sql"
	"time"
)

type User struct {
	ID                int            `db:"id"`
	Email             string         `db:"email"`    // NOT NULL
	Username          string         `db:"username"` // NOT NULL
	Type              UserType       `db:"type"`
	SpotifyID         sql.NullString `db:"spotify_id"`
	SpotifyAuthKey    sql.NullString `db:"spotify_auth_key"`
	SpotifyRefreshKey sql.NullString `db:"spotify_refresh_key"`
	SoundCloudID      sql.NullString `db:"soundcloud_id"`
	TidalID           sql.NullString `db:"tidal_id"`
	SoundCloudKey     sql.NullString `db:"soundcloud_key"`
	TidalKey          sql.NullString `db:"tidal_key"`
	CurrentRoomID     sql.NullInt64  `db:"current_room_id"`
	CurrentRole       UserRole       `db:"current_role"`
	CreatedAt         time.Time      `db:"created_at"`
}

type Room struct {
	ID                 int           `db:"id"`
	Name               string        `db:"name"`    // NOT NULL
	HostID             int           `db:"host_id"` // NOT NULL
	LastPlayedPosition sql.NullInt64 `db:"last_played_position"`
	CreatedAt          time.Time     `db:"created_at"`
}

type Track struct {
	ID        int            `db:"id"`
	SourceURI string         `db:"source_uri"` // NOT NULL
	Artist    sql.NullString `db:"artist"`
	Title     string         `db:"title"` // NOT NULL
	Album     sql.NullString `db:"album"`
	CoverURL  sql.NullString `db:"cover_url"`
	Platform  sql.NullString `db:"platform"`
}

type RoomTrack struct {
	ID             int          `db:"id"`
	RoomID         int          `db:"room_id"`  // NOT NULL
	TrackID        int          `db:"track_id"` // NOT NULL
	AddedBy        int          `db:"added_by"` // NOT NULL
	Position       int          `db:"position"` // NOT NULL
	Status         TrackStatus  `db:"status"`   // NOT NULL
	StartTimestamp sql.NullTime `db:"start_timestamp"`
	EndTimestamp   sql.NullTime `db:"end_timestamp"`
	LikeCount      int          `db:"like_count"`
	SkipCount      int          `db:"skip_count"`
}

type Archival struct {
	ID        int           `db:"id"`
	RoomID    sql.NullInt64 `db:"room_id"`
	Name      string        `db:"name"` // NOT NULL
	CreatedAt time.Time     `db:"created_at"`
}

type ArchivalTrack struct {
	ArchivalID int `db:"archival_id"`
	TrackID    int `db:"track_id"`
}
