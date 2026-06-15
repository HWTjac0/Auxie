package handlers

import (
	"auxie/backend/internal/models"
	repositories "auxie/backend/internal/repositories"
	"fmt"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
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

	var userID int
	session := sessions.Default(c)
	existingUserIDVal := session.Get("user_id")

	if existingUserIDVal != nil {
		var tempID int
		switch val := existingUserIDVal.(type) {
		case int:
			tempID = val
		case int64:
			tempID = int(val)
		case float64:
			tempID = int(val)
		}
		if tempID > 0 {
			if existingUser, err := h.userRepo.GetByID(tempID); err == nil && existingUser != nil {
				userID = existingUser.ID
			}
		}
	}

	if userID == 0 {
		// Create a new guest user
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
		userID = int(guestID)
	}

	session.Set("user_name", req.Username)
	session.Set("user_id", userID)
	if img := session.Get("user_image"); img == nil || img == "" {
		session.Set("user_image", "")
	}
	session.Save()

	joinCode := generateJoinCode(6)
	slug := generateSlug(req.RoomName, joinCode)

	room := &models.Room{
		Name:      req.RoomName,
		HostID:    userID,
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
	_ = h.userRepo.UpdateRoom(userID, int(roomID), &role)

	c.JSON(http.StatusCreated, gin.H{
		"room":    room,
		"user_id": userID,
	})
}

type JoinRoomRequest struct {
	JoinCode string `json:"join_code" binding:"required"`
	UserName string `json:"username" binding:"required"`
}

func (h *UserHandler) JoinRoom(c *gin.Context) {
	var req JoinRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomRepo.GetByJoinCode(req.JoinCode)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	var userID int
	session := sessions.Default(c)
	existingUserIDVal := session.Get("user_id")

	if existingUserIDVal != nil {
		var tempID int
		switch val := existingUserIDVal.(type) {
		case int:
			tempID = val
		case int64:
			tempID = int(val)
		case float64:
			tempID = int(val)
		}
		if tempID > 0 {
			if existingUser, err := h.userRepo.GetByID(tempID); err == nil && existingUser != nil {
				userID = existingUser.ID
			}
		}
	}

	if userID == 0 {
		// Create a new guest user
		guest := &models.User{
			Username:  req.UserName,
			Type:      models.UserTypeGuest,
			CreatedAt: time.Now(),
		}

		guestID, err := h.userRepo.Create(guest)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create guest user"})
			return
		}
		userID = int(guestID)
	}

	role := "Guest"
	if room.HostID == userID {
		role = "Host"
	}
	err = h.userRepo.UpdateRoom(userID, room.ID, &role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join room"})
		return
	}

	session.Set("user_name", req.UserName)
	session.Set("user_id", userID)
	if img := session.Get("user_image"); img == nil || img == "" {
		session.Set("user_image", "")
	}
	session.Save()

	c.JSON(http.StatusOK, gin.H{
		"room":    room,
		"user_id": userID,
	})
}

func (h *RoomHandler) GetRoomDetails(c *gin.Context) {
	slug := c.Param("slug")
	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	users, err := h.userRepo.GetUsersInRoom(room.ID, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room":  room,
		"users": users,
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
