package handlers

import (
	"auxie/backend/internal/models"
	repositories "auxie/backend/internal/repositories"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomRepo repositories.RoomRepository
	userRepo repositories.UserRepository
}

func NewRoomHandler(roomRepo repositories.RoomRepository, userRepo repositories.UserRepository) *RoomHandler {
	return &RoomHandler{roomRepo: roomRepo, userRepo: userRepo}
}

func (h *RoomHandler) GetRandomRoomName(c *gin.Context) {
	adjectives := []string{"Awesome", "Cool", "Epic", "Groovy", "Funky", "Wild", "Chill", "Magic", "Hyper", "Vibey", "Dazzling", "Electric"}
	nouns := []string{"Party", "Room", "Lounge", "Club", "Session", "Basement", "Vibe", "Station", "Hub", "Zone", "Cave", "Arena"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	c.JSON(200, gin.H{"name": fmt.Sprintf("%s %s", adj, noun)})
}

const charset = "ABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func generateJoinCode(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.IntN(len(charset))]
	}
	return string(b)
}

func generateSlug(name string, code string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	return fmt.Sprintf("%s-%s", slug, strings.ToLower(code))
}

type CreateRoomRequest struct {
	RoomName string `json:"room_name" binding:"required"`
	Username string `json:"username" binding:"required"`
}

func (h *RoomHandler) CreateRoom(c *gin.Context) {
	var req CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	guest := &models.User{
		Username:  req.Username,
		Type:      models.UserTypeGuest,
		CreatedAt: time.Now(),
	}

	guestID, err := h.userRepo.Create(guest)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest user"})
		return
	}

	joinCode := generateJoinCode(6)
	slug := generateSlug(req.RoomName, joinCode)

	room := &models.Room{
		Name:      req.RoomName,
		HostID:    int(guestID),
		JoinCode:  joinCode,
		Slug:      slug,
		CreatedAt: time.Now(),
	}

	roomID, err := h.roomRepo.Create(room)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}

	role := "Host"
	_ = h.userRepo.UpdateRoom(int(guestID), int(roomID), &role)

	c.JSON(http.StatusCreated, gin.H{
		"room":    room,
		"user_id": guestID,
	})
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
