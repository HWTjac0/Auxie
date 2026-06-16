package repositories

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/models"
	"log"
)

type RoomSqliteRepo struct {
	db *database.DB
}

func NewRoomSqliteRepo(db *database.DB) *RoomSqliteRepo {
	return &RoomSqliteRepo{db}
}

func (r *RoomSqliteRepo) Create(room *models.Room) (int64, error) {
	query := `INSERT INTO rooms (name, join_code, slug, host_id, created_at) VALUES (?, ?, ?, ?, ?)`

	result, err := r.db.Exec(query, room.Name, room.JoinCode, room.Slug, room.HostID, room.CreatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *RoomSqliteRepo) GetByID(id int) (*models.Room, error) {
	log.Println("TODO: implement RoomSqliteRepo.GetByID")
	return nil, nil
}

func (r *RoomSqliteRepo) GetActiveByHostID(hostID int) (*models.Room, error) {
	log.Println("TODO: implement RoomSqliteRepo.GetActiveByHostID")
	return nil, nil
}

func (r *RoomSqliteRepo) GetByHostId(hostID int) (*models.Room, error) {
	var room models.Room
	query := `SELECT * FROM rooms WHERE host_id = ? LIMIT 1`
	err := r.db.Get(&room, query, hostID)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomSqliteRepo) GetByJoinCode(code string) (*models.Room, error) {
	var room models.Room
	query := `SELECT * FROM rooms WHERE join_code = ? LIMIT 1`
	err := r.db.Get(&room, query, code)
	if err != nil {
		return nil, err
	}
	return &room, nil
}
func (r *RoomSqliteRepo) GetBySlug(slug string) (*models.Room, error) {
	var room models.Room
	query := `SELECT * FROM rooms WHERE slug = ? LIMIT 1`
	err := r.db.Get(&room, query, slug)
	if err != nil {
		return nil, err
	}
	return &room, nil
}

func (r *RoomSqliteRepo) UpdateLastPlayedPosition(roomID int, position int) error {
	log.Println("TODO: implement RoomSqliteRepo.UpdateLastPlayedPosition")
	return nil
}

func (r *RoomSqliteRepo) Delete(id int) error {
	log.Println("TODO: implement RoomSqliteRepo.Delete")
	return nil
}

func (r *RoomSqliteRepo) AddToQueue(track *models.RoomTrack) error {
	query := `INSERT INTO room_tracks (room_id, track_id, added_by, position, status) 
	          VALUES (?, ?, ?, COALESCE((SELECT MAX(position) FROM room_tracks WHERE room_id = ?), 0) + 1, ?)`

	_, err := r.db.Exec(query, track.RoomID, track.TrackID, track.AddedBy, track.RoomID, track.Status)
	return err
}

func (r *RoomSqliteRepo) GetQueue(roomID int) ([]models.RoomQueueItem, error) {
	query := `
		SELECT 
			rt.id AS room_track_id, 
			rt.position, 
			rt.status, 
			rt.added_by, 
			rt.like_count, 
			t.id AS track_id, 
			t.title, 
			t.artist, 
			t.cover_url, 
			t.platform
		FROM room_tracks rt
		JOIN tracks t ON rt.track_id = t.id
		WHERE rt.room_id = ? AND rt.status IN ('queued', 'playing')
		ORDER BY rt.position ASC
	`
	
	var queue []models.RoomQueueItem
	err := r.db.Select(&queue, query, roomID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return []models.RoomQueueItem{}, nil
		}
		return nil, err
	}
	
	return queue, nil
}

func (r *RoomSqliteRepo) GetProposedQueue(roomID int) ([]models.RoomQueueItem, error) {
	query := `
		SELECT 
			rt.id AS room_track_id, 
			rt.position, 
			rt.status, 
			rt.added_by, 
			rt.like_count, 
			t.id AS track_id, 
			t.title, 
			t.artist, 
			t.cover_url, 
			t.platform
		FROM room_tracks rt
		JOIN tracks t ON rt.track_id = t.id
		WHERE rt.room_id = ? AND rt.status = 'proposed'
		ORDER BY rt.position ASC
	`
	
	var queue []models.RoomQueueItem
	err := r.db.Select(&queue, query, roomID)
	if err != nil {
		if err.Error() == "sql: no rows in result set" {
			return []models.RoomQueueItem{}, nil
		}
		return nil, err
	}
	
	return queue, nil
}

func (r *RoomSqliteRepo) UpdateTrackStatus(roomTrackID int, status string) error {
	_, err := r.db.Exec(`UPDATE room_tracks SET status = ? WHERE id = ?`, status, roomTrackID)
	return err
}

func (r *RoomSqliteRepo) RemoveFromQueue(roomTrackID int) error {
	log.Println("TODO: implement RoomSqliteRepo.RemoveFromQueue")
	return nil
}

func (r *RoomSqliteRepo) UpdateTrackPosition(roomTrackID int, newPosition int) error {
	log.Println("TODO: implement RoomSqliteRepo.UpdateTrackPosition")
	return nil
}

func (r *RoomSqliteRepo) IncrementLikeCount(roomTrackID int) error {
	log.Println("TODO: implement RoomSqliteRepo.IncrementLikeCount")
	return nil
}

func (r *RoomSqliteRepo) IncrementSkipCount(roomTrackID int) error {
	log.Println("TODO: implement RoomSqliteRepo.IncrementSkipCount")
	return nil
}
