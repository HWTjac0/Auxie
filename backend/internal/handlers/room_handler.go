package handlers

import (
	"auxie/backend/internal/repositories"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomRepo *repositories.RoomSqliteRepo
}

func GetRandomRoomName() string {
	var randomName string
	OK_name := false

	for !OK_name {
		randomName = "Room number " + string(rand.Intn(1000))
		if !CheckIfRoomNameExists(randomName) {
			OK_name = true
		}
	}
	return randomName
}

func CheckIfRoomNameExists(name string) bool {
	query := "SELECT COUNT(*) FROM rooms WHERE name = ?"
	var count int
	err := r.db.QueryRow(query, name).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func CheckIfHostHasRoom(host_id int) bool {
	query := "SELECT COUNT(*) FROM rooms WHERE host_id = ?"
	var count int
	err := r.db.QueryRow(query, host_id).Scan(&count)
	if err != nil {
		return false
	}
	return count > 0
}

func AddTrackToRoom(room_id int, track_id int, user_id int) error {
	query := "SELECT COUNT(*) FROM room_tracks WHERE room_id = ?"
	var count int
	err := r.db.QueryRow(query, room_id).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check room queue: %w", err)
	}
	query := "INSERT INTO room_tracks (room_id, track_id, user_id) VALUES (?, ?, ?)"
	_, err := r.db.Exec(query, room_id, track_id, user_id)
	return err
}

func ChangeTrackPosition(room_id int, track_id int, new_position int) error {

};

func NextTrackInRoom(room_id int) (int, error) {

}