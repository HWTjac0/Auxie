package repositories

import database "auxie/backend/internal/db"

type TrackSqliteRepo struct {
	db *database.DB
}

func NewTrackSqliteRepo(db *database.DB) *TrackSqliteRepo {
	return &TrackSqliteRepo{db}
}
