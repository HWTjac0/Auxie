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
	log.Println("TODO: implement TrackSqliteRepo.GetByID")
	return nil, nil
}

func (r *TrackSqliteRepo) GetByURI(uri string) (*models.Track, error) {
	log.Println("TODO: implement TrackSqliteRepo.GetByURI")
	return nil, nil
}

func (r *TrackSqliteRepo) Create(track *models.Track) (int, error) {
	log.Println("TODO: implement TrackSqliteRepo.Create")
	return 0, nil
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
