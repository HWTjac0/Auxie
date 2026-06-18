package repositories

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/models"
	"log"
)

type TrackSqliteRepo struct {
	db *database.DB
}

func NewTrackSqliteRepo(db *database.DB) *TrackSqliteRepo {
	return &TrackSqliteRepo{db}
}

func (r *TrackSqliteRepo) GetByID(id int) (*models.Track, error) {
	query := `SELECT id, source_uri, artist, title, album, cover_url, platform FROM tracks WHERE id = ?`
	var track models.Track
	err := r.db.Unsafe().Get(&track, query, id)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &track, nil
}

func (r *TrackSqliteRepo) GetByURI(uri string) (*models.Track, error) {
	query := `SELECT id, source_uri, artist, title, album, cover_url, platform FROM tracks WHERE source_uri = ?`
	var track models.Track
	err := r.db.Unsafe().Get(&track, query, uri)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return nil, nil
		}
		return nil, err
	}
	return &track, nil
}

func (r *TrackSqliteRepo) Create(track *models.Track) (int, error) {
	query := `INSERT INTO tracks (source_uri, artist, title, album, cover_url, platform) VALUES (?, ?, ?, ?, ?, ?)`
	result, err := r.db.Exec(query, track.SourceURI, track.Artist, track.Title, track.Album, track.CoverURL, track.Platform)
	if err != nil {
		return 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, err
	}
	return int(id), nil
}

func (r *TrackSqliteRepo) AddToRoom(roomTrack *models.RoomTrack) error {
	log.Println("TODO: implement TrackSqliteRepo.AddToRoom")
	return nil
}

func (r *TrackSqliteRepo) GetRoomQueue(roomID int) ([]models.RoomTrack, error) {
	log.Println("TODO: implement TrackSqliteRepo.GetRoomQueue")
	return nil, nil
}

func (r *TrackSqliteRepo) UpdateStatus(id int, status models.TrackStatus) error {
	log.Println("TODO: implement TrackSqliteRepo.UpdateStatus")
	return nil
}

func (r *TrackSqliteRepo) UpdatePosition(id int, newPosition int) error {
	log.Println("TODO: implement TrackSqliteRepo.UpdatePosition")
	return nil
}

func (r *TrackSqliteRepo) DeleteFromRoom(id int) error {
	log.Println("TODO: implement TrackSqliteRepo.DeleteFromRoom")
	return nil
}
