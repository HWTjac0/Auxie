package handlers

import (
	repositories "auxie/backend/internal/repositories"
)

type RoomHandler struct {
	roomRepo repositories.RoomRepository
}

func NewRoomHandler(roomRepo repositories.RoomRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo}
}

func (h *RoomHandler) GetRandomRoomName() string {
	var randomName string
	return randomName
}

func (h *RoomHandler) CheckIfRoomNameExists(name string) bool {
	return count > 0
}

func CheckIfHostHasRoom(host_id int) bool {
	return false
}

func AddTrackToRoom(room_id int, track_id int, user_id int) error {
	return err
}

func ChangeTrackPosition(room_id int, track_id int, new_position int) error {
	return
}

func NextTrackInRoom(room_id int) (int, error) {

}

