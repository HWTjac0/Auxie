package handlers

import (
	"auxie/backend/internal/models"
	repositories "auxie/backend/internal/repositories"
	"database/sql"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomRepo  repositories.RoomRepository
	userRepo  repositories.UserRepository
	trackRepo repositories.TrackRepository
	hub       *RoomHub
}

func NewRoomHandler(roomRepo repositories.RoomRepository, userRepo repositories.UserRepository, trackRepo repositories.TrackRepository) *RoomHandler {
	hub := NewRoomHub()

	type dbEvent struct {
		isJoin bool
		roomID string
		client *WSClient
	}
	dbQueue := make(chan dbEvent, 256)

	go func() {
		for ev := range dbQueue {
			if ev.isJoin {
				room, err := roomRepo.GetBySlug(ev.roomID)
				if err == nil {
					role := "Guest"
					if room.HostID == ev.client.UserID {
						role = "Host"
					}
					ev.client.Role = role
					if err := userRepo.UpdateRoom(ev.client.UserID, room.ID, &role); err != nil {
						log.Printf("Failed to update user room status on enter: %v", err)
					}
				}

				hub.broadcast <- &BroadcastMessage{
					RoomID: ev.roomID,
					Payload: gin.H{
						"type": "USER_JOINED",
						"payload": gin.H{
							"user_id":  ev.client.UserID,
							"username": ev.client.Username,
							"role":     ev.client.Role,
						},
					},
				}
			} else {
				if err := userRepo.LeaveRoom(ev.client.UserID); err != nil {
					log.Printf("Failed to update user room status on leave: %v", err)
				}

				hub.broadcast <- &BroadcastMessage{
					RoomID: ev.roomID,
					Payload: gin.H{
						"type": "USER_LEFT",
						"payload": gin.H{
							"user_id":  ev.client.UserID,
							"username": ev.client.Username,
						},
					},
				}
			}
		}
	}()

	hub.onUserJoin = func(roomID string, client *WSClient) {
		dbQueue <- dbEvent{isJoin: true, roomID: roomID, client: client}
	}

	hub.onUserLeave = func(roomID string, client *WSClient) {
		dbQueue <- dbEvent{isJoin: false, roomID: roomID, client: client}
	}

	return &RoomHandler{
		roomRepo:  roomRepo,
		userRepo:  userRepo,
		trackRepo: trackRepo,
		hub:       hub,
	}
}

type AddTrackRequest struct {
	SourceURI string `json:"source_uri" binding:"required"`
	Title     string `json:"title" binding:"required"`
	Artist    string `json:"artist"`
	Album     string `json:"album"`
	CoverURL  string `json:"cover_url"`
	Platform  string `json:"platform"`
}

func (h *RoomHandler) AddTrack(c *gin.Context) {
	slug := c.Param("slug")
	userID := c.GetInt("user_id")

	var req AddTrackRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// 1. Get or create track
	track, err := h.trackRepo.GetByURI(req.SourceURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error finding track"})
		return
	}

	var trackID int
	if track == nil {
		newTrack := &models.Track{
			SourceURI: req.SourceURI,
			Title:     req.Title,
			Platform:  sql.NullString{String: req.Platform, Valid: req.Platform != ""},
			Artist:    sql.NullString{String: req.Artist, Valid: req.Artist != ""},
			Album:     sql.NullString{String: req.Album, Valid: req.Album != ""},
			CoverURL:  sql.NullString{String: req.CoverURL, Valid: req.CoverURL != ""},
		}
		trackID, err = h.trackRepo.Create(newTrack)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating track"})
			return
		}
	} else {
		trackID = track.ID
	}

	// 2. Add to queue
	roomTrack := &models.RoomTrack{
		RoomID:  room.ID,
		TrackID: trackID,
		AddedBy: userID,
		Status:  models.TrackStatusQueued,
	}

	if err := h.roomRepo.AddToQueue(roomTrack); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding to queue"})
		return
	}

	// 3. Broadcast to WS hub
	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "TRACK_ADDED",
			"payload": gin.H{
				"track_id": trackID,
				"title":    req.Title,
				"artist":   req.Artist,
				"added_by": userID,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track added to queue"})
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

	queue, err := h.roomRepo.GetQueue(room.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting queue"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"room":  room,
		"users": users,
		"queue": queue,
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

func (h *RoomHandler) HandleWS(c *gin.Context) {
	roomSlug := c.Param("slug")
	userID := c.GetInt("user_id")

	// Verify room exists
	_, err := h.roomRepo.GetBySlug(roomSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	// Fetch user details
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection to websocket: %v", err)
		return
	}

	client := &WSClient{
		Conn:     conn,
		UserID:   user.ID,
		Username: user.Username,
		Send:     make(chan interface{}, 256),
	}

	// Register client in the hub. This will trigger onUserJoin if it's their first connection.
	h.hub.register <- &Subscription{RoomID: roomSlug, Client: client}

	// Start client's write pump
	go client.WritePump()

	// Read loop (to detect disconnection)
	defer func() {
		// Unregister client. This will trigger onUserLeave if it's their last connection.
		h.hub.unregister <- &Subscription{RoomID: roomSlug, Client: client}
	}()

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}
}
