package repositories

import (
	"auxie/backend/internal/models"
	"time"
)

type UserFilter struct {
	Role   *string
	Type   *models.UserType
	Fields []string
}

type UserRepository interface {
	GetByEmail(email string) (*models.User, error)
	GetByID(id int) (*models.User, error)
	Create(user *models.User) (int64, error)
	UpdateRoom(userId int, roomId int, role *string) error
	LeaveRoom(userId int) error
	GetUsersInRoom(roomId int, filter *UserFilter) ([]models.User, error)
	GetStreamingID(serviceType models.MusicService, userId int) (string, error)
	GetBySpotifyID(spotifyID string) (*models.User, error)
	UpdateSpotifyInfo(userID int, spotifyID string, authKey string, refreshKey string, expiresAt time.Time) error
	GetByTidalID(tidalID string) (*models.User, error)
	UpdateTidalInfo(userID int, tidalID string, authKey string, refreshKey string, expiresAt time.Time) error
	GetBySoundCloudID(soundCloudID string) (*models.User, error)
	UpdateSoundCloudInfo(userID int, soundCloudID string, authKey string, refreshKey string, expiresAt time.Time) error
	DisconnectSpotify(userID int) error
	DisconnectTidal(userID int) error
	DisconnectSoundCloud(userID int) error
}

type RoomRepository interface {
	Create(room *models.Room) (int64, error)
	GetByID(id int) (*models.Room, error)
	GetActiveByHostID(hostID int) (*models.Room, error)
	GetByHostId(hostID int) (*models.Room, error)
	GetByJoinCode(code string) (*models.Room, error)
	GetBySlug(slug string) (*models.Room, error)
	UpdateLastPlayedPosition(roomID int, position int) error
	Delete(id int) error

	AddToQueue(track *models.RoomTrack) error
	GetQueueItem(roomTrackID int) (*models.RoomQueueItem, error)
	GetRoomTrack(roomTrackID int) (*models.RoomTrack, error)
	GetQueue(roomID int) ([]models.RoomQueueItem, error)
	GetProposedQueue(roomID int) ([]models.RoomQueueItem, error)
	UpdateTrackStatus(roomTrackID int, status string) error // np. 'playing', 'played'
	RemoveFromQueue(roomTrackID int) error
	UpdateTrackTimestamps(roomTrackID int, startTime *time.Time, endTime *time.Time) error

	UpdateTrackPosition(roomTrackID int, newPosition int) error

	IncrementLikeCount(roomTrackID int) error
	DecrementLikeCount(roomTrackID int) error
	GetLikeCount(roomTrackID int) (int, error)
	IncrementSkipCount(roomTrackID int) error
	GetSkipCount(roomTrackID int) (int, error)

	HasUserLiked(roomTrackID int, userID int) (bool, error)
	AddLike(roomTrackID int, userID int) error
	RemoveLike(roomTrackID int, userID int) error

	HasUserVotedSkip(roomTrackID int, userID int) (bool, error)
	AddSkipVote(roomTrackID int, userID int) error
	GetSkipVoteCount(roomTrackID int) (int, error)
	GetRoomHistory(slug string) ([]models.TrackHistory, error)
	CloseRoom(slug string) error
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
