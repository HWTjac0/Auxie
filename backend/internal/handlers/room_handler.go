package handlers

import (
	repositories "auxie/backend/internal/repositories"
	"fmt"
	"math/rand/v2"
)

type RoomHandler struct {
	roomRepo repositories.RoomRepository
}

func NewRoomHandler(roomRepo repositories.RoomRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo}
}

func (h *RoomHandler) GetRandomRoomName() string {
	adjectives := []string{"Awesome", "Cool", "Epic", "Groovy", "Funky", "Wild", "Chill", "Magic", "Hyper", "Vibey", "Dazzling", "Electric"}
	nouns := []string{"Party", "Room", "Lounge", "Club", "Session", "Basement", "Vibe", "Station", "Hub", "Zone", "Cave", "Arena"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
}

func (h *RoomHandler) CheckIfHostHasRoom(host_id int) bool {
	return false
}

func (h *RoomHandler) AddTrackToRoom(room_id int, track_id int, user_id int) error {
	return nil
}

func (h *RoomHandler) ChangeTrackPosition(room_id int, track_id int, new_position int) error {
	return nil
}

func (h *RoomHandler) NextTrackInRoom(room_id int) (int, error) {
	return 0, nil
}

