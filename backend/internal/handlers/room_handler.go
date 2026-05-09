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

func (h *RoomHandler) CheckIfHostHasRoom(host_id int) bool {
	return false
}

func (h *RoomHandler) AddTrackToRoom(room_id int, track_id int, user_id int) error {
	return err
}

func (h *RoomHandler) ChangeTrackPosition(room_id int, track_id int, new_position int) error {
	return
}

func (h *RoomHandler) NextTrackInRoom(room_id int) (int, error) {

}

