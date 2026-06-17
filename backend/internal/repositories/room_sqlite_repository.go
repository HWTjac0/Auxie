package repositories

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/models"
	"log"
	"time"
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
	_, err := r.db.Exec(`UPDATE rooms SET last_played_position = ? WHERE id = ?`, position, roomID)
	return err
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
			COALESCE(t.artist, '') AS artist, 
			COALESCE(t.cover_url, '') AS cover_url, 
			COALESCE(t.platform, '') AS platform,
			COALESCE(t.source_uri, '') AS source_uri
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
			COALESCE(t.artist, '') AS artist, 
			COALESCE(t.cover_url, '') AS cover_url, 
			COALESCE(t.platform, '') AS platform,
			COALESCE(t.source_uri, '') AS source_uri
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
	_, err := r.db.Exec(`DELETE FROM room_tracks WHERE id = ?`, roomTrackID)
	return err
}

func (r *RoomSqliteRepo) UpdateTrackPosition(roomTrackID int, newPosition int) error {
	log.Println("TODO: implement RoomSqliteRepo.UpdateTrackPosition")
	return nil
}

func (r *RoomSqliteRepo) IncrementLikeCount(roomTrackID int) error {
	_, err := r.db.Exec(`UPDATE room_tracks SET like_count = like_count + 1 WHERE id = ?`, roomTrackID)
	return err
}

func (r *RoomSqliteRepo) DecrementLikeCount(roomTrackID int) error {
	_, err := r.db.Exec(`UPDATE room_tracks SET like_count = MAX(0, like_count - 1) WHERE id = ?`, roomTrackID)
	return err
}

func (r *RoomSqliteRepo) GetLikeCount(roomTrackID int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT like_count FROM room_tracks WHERE id = ?`, roomTrackID)
	return count, err
}

func (r *RoomSqliteRepo) IncrementSkipCount(roomTrackID int) error {
	_, err := r.db.Exec(`UPDATE room_tracks SET skip_count = skip_count + 1 WHERE id = ?`, roomTrackID)
	return err
}

func (r *RoomSqliteRepo) GetSkipCount(roomTrackID int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT skip_count FROM room_tracks WHERE id = ?`, roomTrackID)
	return count, err
}

func (r *RoomSqliteRepo) HasUserLiked(roomTrackID int, userID int) (bool, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM room_track_likes WHERE room_track_id = ? AND user_id = ?`, roomTrackID, userID)
	return count > 0, err
}

func (r *RoomSqliteRepo) AddLike(roomTrackID int, userID int) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO room_track_likes (room_track_id, user_id) VALUES (?, ?)`, roomTrackID, userID)
	return err
}

func (r *RoomSqliteRepo) RemoveLike(roomTrackID int, userID int) error {
	_, err := r.db.Exec(`DELETE FROM room_track_likes WHERE room_track_id = ? AND user_id = ?`, roomTrackID, userID)
	return err
}

func (r *RoomSqliteRepo) HasUserVotedSkip(roomTrackID int, userID int) (bool, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM room_track_skip_votes WHERE room_track_id = ? AND user_id = ?`, roomTrackID, userID)
	return count > 0, err
}

func (r *RoomSqliteRepo) AddSkipVote(roomTrackID int, userID int) error {
	_, err := r.db.Exec(`INSERT OR IGNORE INTO room_track_skip_votes (room_track_id, user_id) VALUES (?, ?)`, roomTrackID, userID)
	return err
}

func (r *RoomSqliteRepo) GetSkipVoteCount(roomTrackID int) (int, error) {
	var count int
	err := r.db.Get(&count, `SELECT COUNT(*) FROM room_track_skip_votes WHERE room_track_id = ?`, roomTrackID)
	return count, err
}

func (r *RoomSqliteRepo) UpdateTrackTimestamps(roomTrackID int, startTime *time.Time, endTime *time.Time) error {
	if startTime != nil && endTime != nil {
		_, err := r.db.Exec(`UPDATE room_tracks SET start_timestamp = ?, end_timestamp = ? WHERE id = ?`, startTime, endTime, roomTrackID)
		return err
	} else if startTime != nil {
		_, err := r.db.Exec(`UPDATE room_tracks SET start_timestamp = ? WHERE id = ?`, startTime, roomTrackID)
		return err
	} else if endTime != nil {
		_, err := r.db.Exec(`UPDATE room_tracks SET end_timestamp = ? WHERE id = ?`, endTime, roomTrackID)
		return err
	}
	return nil
}
