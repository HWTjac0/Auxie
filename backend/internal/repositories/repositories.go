package repositories

import "auxie/backend/internal/models"

type UserRepository interface {
	GetByEmail(email string) (*models.User, error)
	Create(user *models.User) (int64, error)
	UpdateRoom(userId int, roomId int, role *string) error
}

type RoomRepository interface {
	Create(room *models.Room) (int64, error)
	GetByID(id int) (*models.Room, error)
	GetActiveByHostID(hostID int) (*models.Room, error)
	UpdateLastPlayedPosition(roomID int, position int) error
	Delete(id int) error

	AddToQueue(track *models.RoomTrack) error
	GetQueue(roomID int) ([]models.RoomTrack, error)
	UpdateTrackStatus(roomTrackID int, status string) error // np. 'playing', 'played'
	RemoveFromQueue(roomTrackID int) error

	UpdateTrackPosition(roomTrackID int, newPosition int) error

	IncrementLikeCount(roomTrackID int) error
	IncrementSkipCount(roomTrackID int) error
}

type TrackRepository interface {
	GetByID(id int) (*models.Track, error)
	GetByURI(uri string) (*models.Track, error) // Zapobiega duplikatom utworów
	Create(track *models.Track) (int, error)    // Zwraca ID nowo stworzonego utworu

	AddToRoom(roomTrack *models.RoomTrack) error
	GetRoomQueue(roomID int) ([]models.RoomTrack, error)
	UpdateStatus(id int, status models.TrackStatus) error
	UpdatePosition(id int, newPosition int) error
	DeleteFromRoom(id int) error
}
