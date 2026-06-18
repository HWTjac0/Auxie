package handlers

import (
	"auxie/backend/internal/models"
	repositories "auxie/backend/internal/repositories"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"math/rand/v2"
	"net/http"
	"strconv"
	"strings"
	"time"

	"auxie/backend/internal/clients"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomRepo         repositories.RoomRepository
	userRepo         repositories.UserRepository
	trackRepo        repositories.TrackRepository
	hub              *RoomHub
	playbackManagers map[string]*RoomPlaybackManager
	tidalClient      *clients.TidalClient
}

func NewRoomHandler(roomRepo repositories.RoomRepository, userRepo repositories.UserRepository, trackRepo repositories.TrackRepository, tidalClient *clients.TidalClient) *RoomHandler {
	hub := NewRoomHub()
	playbackManagers := make(map[string]*RoomPlaybackManager)

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
					user, _ := userRepo.GetByID(ev.client.UserID)
					role := "Guest"
					if room.HostID == ev.client.UserID {
						role = "Host"
					} else if user != nil && user.CurrentRole != nil && (*user.CurrentRole == "DJ" || *user.CurrentRole == "Host") {
						role = *user.CurrentRole
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
							"id":       ev.client.UserID,
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
		roomRepo:         roomRepo,
		userRepo:         userRepo,
		trackRepo:        trackRepo,
		hub:              hub,
		playbackManagers: playbackManagers,
		tidalClient:      tidalClient,
	}
}

func (h *RoomHandler) getOrCreatePlaybackManager(slug string) *RoomPlaybackManager {
	if m, ok := h.playbackManagers[slug]; ok {
		return m
	}
	m := NewRoomPlaybackManager(h.hub, slug, h.roomRepo, h.trackRepo)
	h.playbackManagers[slug] = m
	return m
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

	callerUser, err := h.userRepo.GetByID(userID)
	if err != nil || callerUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	status := models.TrackStatusQueued
	wsType := "TRACK_ADDED"
	if callerUser.CurrentRole != nil && *callerUser.CurrentRole == "Guest" {
		status = models.TrackStatusProposed
		wsType = "TRACK_PROPOSED"
	}

	// 2. Add to queue
	roomTrack := &models.RoomTrack{
		RoomID:  room.ID,
		TrackID: trackID,
		AddedBy: userID,
		Status:  status,
	}

	if err := h.roomRepo.AddToQueue(roomTrack); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error adding to queue"})
		return
	}

	// 3. Broadcast to WS hub
	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": wsType,
			"payload": gin.H{
				"track_id": trackID,
				"title":    req.Title,
				"artist":   req.Artist,
				"added_by": userID,
			},
		},
	}

	// Ensure playback manager exists and schedule if idle
	mgr := h.getOrCreatePlaybackManager(slug)
	go mgr.StartIfIdle()

	c.JSON(http.StatusOK, gin.H{"message": "Track added", "status": status.String()})
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

	users, err := h.userRepo.GetUsersInRoom(room.ID, &repositories.UserFilter{Fields: []string{"id", "username", "type", "current_role"}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting users"})
		return
	}

	queue, err := h.roomRepo.GetQueue(room.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting queue"})
		return
	}

	proposedQueue, err := h.roomRepo.GetProposedQueue(room.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error while getting proposed queue"})
		return
	}

	session := sessions.Default(c)
	userID := 0
	if uid := session.Get("user_id"); uid != nil {
		if id, ok := uid.(int); ok {
			userID = id
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"room":            room,
		"users":           users,
		"queue":           queue,
		"proposedQueue":   proposedQueue,
		"current_user_id": userID,
	})
}

func (h *RoomHandler) CheckIfHostHasRoom(host_id int) bool {
	return false
}

func (h *RoomHandler) SkipTrack(c *gin.Context) {
	slug := c.Param("slug")
	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	callerUser, err := h.userRepo.GetByID(c.GetInt("user_id"))
	if err != nil || callerUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if callerUser.CurrentRole != nil && *callerUser.CurrentRole == "Guest" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Host or DJ can skip tracks"})
		return
	}

	queue, err := h.roomRepo.GetQueue(room.ID)
	if err != nil || len(queue) == 0 {
		// Nothing to skip — treat as success
		c.JSON(http.StatusOK, gin.H{"message": "No track to skip"})
		return
	}

	trackToSkip := queue[0]

	err = h.roomRepo.UpdateTrackStatus(trackToSkip.RoomTrackID, "skipped")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to skip track"})
		return
	}

	now := time.Now()
	_ = h.roomRepo.UpdateTrackTimestamps(trackToSkip.RoomTrackID, nil, &now)

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "TRACK_SKIPPED",
			"payload": gin.H{
				"track_id": trackToSkip.TrackID,
				"title":    trackToSkip.Title,
				"artist":   trackToSkip.Artist,
			},
		},
	}

	// Broadcast queue emptied if no more tracks
	remainingQueue, _ := h.roomRepo.GetQueue(room.ID)
	if len(remainingQueue) == 0 {
		h.hub.broadcast <- &BroadcastMessage{
			RoomID:  slug,
			Payload: gin.H{"type": "playback:queue_empty"},
		}
	}

	// notify playback manager to cancel/start next
	if mgr := h.getOrCreatePlaybackManager(slug); mgr != nil {
		go mgr.Skip()
		go mgr.StartIfIdle()
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track skipped successfully"})
}

func (h *RoomHandler) LikeTrack(c *gin.Context) {
	slug := c.Param("slug")
	roomTrackIDStr := c.Param("track_id")
	userID := c.GetInt("user_id")

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}
	_ = room

	roomTrackID, err := strconv.Atoi(roomTrackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid track ID"})
		return
	}

	alreadyLiked, err := h.roomRepo.HasUserLiked(roomTrackID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	if alreadyLiked {
		// Toggle off
		if err := h.roomRepo.RemoveLike(roomTrackID, userID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove like"})
			return
		}
		if err := h.roomRepo.DecrementLikeCount(roomTrackID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decrement like"})
			return
		}
		likeCount, err := h.roomRepo.GetLikeCount(roomTrackID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get like count"})
			return
		}
		h.hub.broadcast <- &BroadcastMessage{
			RoomID:  slug,
			Payload: gin.H{"type": "TRACK_LIKED", "payload": gin.H{"room_track_id": roomTrackID, "liked": false, "like_count": likeCount}},
		}
		c.JSON(http.StatusOK, gin.H{"liked": false, "like_count": likeCount})
		return
	}

	if err := h.roomRepo.AddLike(roomTrackID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add like"})
		return
	}
	if err := h.roomRepo.IncrementLikeCount(roomTrackID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment like"})
		return
	}
	likeCount, err := h.roomRepo.GetLikeCount(roomTrackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get like count"})
		return
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID:  slug,
		Payload: gin.H{"type": "TRACK_LIKED", "payload": gin.H{"room_track_id": roomTrackID, "liked": true, "like_count": likeCount}},
	}
	c.JSON(http.StatusOK, gin.H{"liked": true, "like_count": likeCount})
}

func (h *RoomHandler) VoteSkip(c *gin.Context) {
	slug := c.Param("slug")
	userID := c.GetInt("user_id")

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	queue, err := h.roomRepo.GetQueue(room.ID)
	if err != nil || len(queue) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No track currently playing"})
		return
	}

	currentTrack := queue[0]
	roomTrackID := currentTrack.RoomTrackID

	alreadyVoted, err := h.roomRepo.HasUserVotedSkip(roomTrackID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}
	if alreadyVoted {
		c.JSON(http.StatusConflict, gin.H{"error": "Already voted to skip"})
		return
	}

	if err := h.roomRepo.AddSkipVote(roomTrackID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to register vote"})
		return
	}
	if err := h.roomRepo.IncrementSkipCount(roomTrackID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to increment skip count"})
		return
	}

	voteCount, err := h.roomRepo.GetSkipVoteCount(roomTrackID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "DB error"})
		return
	}

	// Count users in room
	users, err := h.userRepo.GetUsersInRoom(room.ID, nil)
	usersInRoom := len(users)
	if err != nil || usersInRoom == 0 {
		usersInRoom = 1
	}

	// Threshold: majority (>50%), minimum 2 votes
	threshold := usersInRoom/2 + 1
	if threshold < 2 {
		threshold = 2
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "SKIP_VOTE",
			"payload": gin.H{
				"room_track_id": roomTrackID,
				"votes":         voteCount,
				"threshold":     threshold,
			},
		},
	}

	if voteCount >= threshold {
		// Auto-skip
		_ = h.roomRepo.UpdateTrackStatus(roomTrackID, "skipped")
		now := time.Now()
		_ = h.roomRepo.UpdateTrackTimestamps(roomTrackID, nil, &now)

		h.hub.broadcast <- &BroadcastMessage{
			RoomID: slug,
			Payload: gin.H{
				"type": "TRACK_SKIPPED",
				"payload": gin.H{
					"track_id": currentTrack.TrackID,
					"title":    currentTrack.Title,
					"artist":   currentTrack.Artist,
					"by_vote":  true,
				},
			},
		}

		remainingQueue, _ := h.roomRepo.GetQueue(room.ID)
		if len(remainingQueue) == 0 {
			h.hub.broadcast <- &BroadcastMessage{
				RoomID:  slug,
				Payload: gin.H{"type": "playback:queue_empty"},
			}
		}

		if mgr := h.getOrCreatePlaybackManager(slug); mgr != nil {
			go mgr.Skip()
			go mgr.StartIfIdle()
		}

		c.JSON(http.StatusOK, gin.H{"skipped": true, "votes": voteCount, "threshold": threshold})
		return
	}

	c.JSON(http.StatusOK, gin.H{"skipped": false, "votes": voteCount, "threshold": threshold})
}

func (h *RoomHandler) ApproveTrack(c *gin.Context) {
	slug := c.Param("slug")
	roomTrackIDStr := c.Param("track_id")

	_, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	callerUser, err := h.userRepo.GetByID(c.GetInt("user_id"))
	if err != nil || callerUser == nil || (callerUser.CurrentRole != nil && *callerUser.CurrentRole == "Guest") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Host or DJ can approve tracks"})
		return
	}

	// In a real scenario we might want to check if the track belongs to the room and is proposed.
	// For now we'll trust the roomTrackID and just update the status.
	roomTrackID, _ := strconv.Atoi(roomTrackIDStr)
	err = h.roomRepo.UpdateTrackStatus(roomTrackID, "queued")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to approve track"})
		return
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "TRACK_APPROVED",
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track approved"})
}

func (h *RoomHandler) RejectTrack(c *gin.Context) {
	slug := c.Param("slug")
	roomTrackIDStr := c.Param("track_id")

	_, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	callerUser, err := h.userRepo.GetByID(c.GetInt("user_id"))
	if err != nil || callerUser == nil || (callerUser.CurrentRole != nil && *callerUser.CurrentRole == "Guest") {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only Host or DJ can reject tracks"})
		return
	}

	roomTrackID, _ := strconv.Atoi(roomTrackIDStr)
	err = h.roomRepo.UpdateTrackStatus(roomTrackID, "skipped") // Or delete it
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reject track"})
		return
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "TRACK_REJECTED",
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Track rejected"})
}

func (h *RoomHandler) ChangeUserRole(c *gin.Context) {
	slug := c.Param("slug")
	targetUsername := c.Param("username")

	var req struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	callerUser, err := h.userRepo.GetByID(c.GetInt("user_id"))
	if err != nil || callerUser == nil || callerUser.CurrentRole == nil || *callerUser.CurrentRole != "Host" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the Host can manage roles"})
		return
	}

	users, err := h.userRepo.GetUsersInRoom(room.ID, &repositories.UserFilter{Fields: []string{"id", "username"}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users"})
		return
	}

	var targetUserID int = 0
	for _, u := range users {
		if u.Username == targetUsername {
			targetUserID = u.ID
			break
		}
	}

	if targetUserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in this room"})
		return
	}

	err = h.userRepo.UpdateRoom(targetUserID, room.ID, &req.Role)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update role"})
		return
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "USER_ROLE_CHANGED",
			"payload": gin.H{
				"username": targetUsername,
				"role":     req.Role,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "Role updated"})
}

func (h *RoomHandler) KickUser(c *gin.Context) {
	slug := c.Param("slug")
	targetUsername := c.Param("username")

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	callerUser, err := h.userRepo.GetByID(c.GetInt("user_id"))
	if err != nil || callerUser == nil || callerUser.CurrentRole == nil || *callerUser.CurrentRole != "Host" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only the Host can kick users"})
		return
	}

	users, err := h.userRepo.GetUsersInRoom(room.ID, &repositories.UserFilter{Fields: []string{"id", "username"}})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Error getting users"})
		return
	}

	var targetUserID int = 0
	for _, u := range users {
		if u.Username == targetUsername {
			targetUserID = u.ID
			break
		}
	}

	if targetUserID == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found in this room"})
		return
	}

	err = h.userRepo.LeaveRoom(targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to kick user"})
		return
	}

	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "USER_KICKED",
			"payload": gin.H{
				"username": targetUsername,
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"message": "User kicked"})
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
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// Expect JSON messages with a `type` field
		var payload map[string]interface{}
		if err := json.Unmarshal(msg, &payload); err != nil {
			// ignore malformed
			continue
		}

		typ, _ := payload["type"].(string)
		switch typ {
		case "playback:ended":
			// Client notifies that the currently playing track finished
			// Prefer room_track_id if present
			var roomTrackID int
			if v, ok := payload["room_track_id"].(float64); ok {
				roomTrackID = int(v)
			} else {
				// fallback: mark first playing track as played
				roomObj, _ := h.roomRepo.GetBySlug(roomSlug)
				if roomObj != nil {
					q, _ := h.roomRepo.GetQueue(roomObj.ID)
					if len(q) > 0 {
						roomTrackID = q[0].RoomTrackID
					}
				}
			}

			if roomTrackID != 0 {
				mgr := h.getOrCreatePlaybackManager(roomSlug)
				mgr.MarkEnded(roomTrackID)
				// Broadcast to other clients
				h.hub.broadcast <- &BroadcastMessage{RoomID: roomSlug, Payload: gin.H{"type": "playback:ended", "room_track_id": roomTrackID}}
				// schedule next
				go mgr.StartIfIdle()
			}

		case "playback:position":
			// client sends position update (seconds)
			if v, ok := payload["position"].(float64); ok {
				pos := int(v)
				roomObj, _ := h.roomRepo.GetBySlug(roomSlug)
				if roomObj != nil {
					_ = h.roomRepo.UpdateLastPlayedPosition(roomObj.ID, pos)
				}
			}

		case "playback:ready":
			// client signals readiness - for now, just broadcast ready state
			h.hub.broadcast <- &BroadcastMessage{RoomID: roomSlug, Payload: gin.H{"type": "playback:ready", "user_id": client.UserID}}

		default:
			// ignore other message types for now
		}
	}
}

// StreamSpotify returns audio stream for a Spotify track
func (h *RoomHandler) StreamSpotify(c *gin.Context) {
	// TODO: Implement actual Spotify audio streaming
	// For now, return a placeholder error
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "Spotify audio streaming not yet implemented",
		"message": "Please use the Spotify app to listen to this track",
	})
}

// StreamTidal returns audio stream for a Tidal track
func (h *RoomHandler) StreamTidal(c *gin.Context) {
	roomTrackIDStr := c.Param("room_track_id")
	roomTrackID, err := strconv.Atoi(roomTrackIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid room_track_id"})
		return
	}

	// We need to find the room and host to get the host's Tidal token
	// This could be optimized, but let's query the room track to find the room id
	roomTrack, err := h.roomRepo.GetRoomTrack(roomTrackID)
	if err != nil || roomTrack == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track not found in room"})
		return
	}

	actualTrack, err := h.trackRepo.GetByID(roomTrack.TrackID)
	if err != nil || actualTrack == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Track details not found"})
		return
	}

	room, err := h.roomRepo.GetByID(roomTrack.RoomID)
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	hostUser, err := h.userRepo.GetByID(room.HostID)
	if err != nil || !hostUser.TidalID.Valid || !hostUser.TidalKey.Valid {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Host does not have Tidal connected"})
		return
	}

	streamURL, err := h.tidalClient.GetStreamURL(hostUser.TidalKey.String, actualTrack.SourceURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Tidal stream URL", "details": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, streamURL)
}

// StreamSoundCloud returns audio stream for a SoundCloud track
func (h *RoomHandler) StreamSoundCloud(c *gin.Context) {
	// TODO: Implement actual SoundCloud audio streaming
	// For now, return a placeholder error
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "SoundCloud audio streaming not yet implemented",
		"message": "Please use the SoundCloud app to listen to this track",
	})
}

// GetPlaybackToken returns the Spotify access token for the current user for Web Playback SDK
func (h *RoomHandler) GetPlaybackToken(c *gin.Context) {
	userID := c.GetInt("user_id")
	user, err := h.userRepo.GetByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	if user.SpotifyAuthKey.String == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not connected to Spotify"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"access_token": user.SpotifyAuthKey.String,
		"token_type":   "Bearer",
	})
}

// GetHistory pobiera historię dla okienka po zamknięciu/wyjściu z pokoju
func (h *RoomHandler) GetHistory(c *gin.Context) {
	slug := c.Param("slug")

	history, err := h.roomRepo.GetRoomHistory(slug)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load history"})
		return
	}

	c.JSON(http.StatusOK, history)
}

// LeaveRoom służy do opuszczenia pokoju przez "zwykłego" gościa
func (h *RoomHandler) LeaveRoom(c *gin.Context) {
	userID := c.GetInt("user_id")

	// Widziałem w Twoim kodzie, że masz już userRepo.LeaveRoom(userID)
	if err := h.userRepo.LeaveRoom(userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave room"})
		return
	}

	// UWAGA: Powiadomienie USER_LEFT wysyła się u Ciebie samoistnie po przerwaniu
	// połączenia WS (zobacz dbQueue w NewRoomHandler), więc nie musimy tego tu robić ręcznie!

	c.JSON(http.StatusOK, gin.H{"message": "Left room successfully"})
}

// DeleteRoom zamyka pokój i wymusza wysłanie historii wszystkim gościom
func (h *RoomHandler) DeleteRoom(c *gin.Context) {
	slug := c.Param("slug")
	userID := c.GetInt("user_id")

	room, err := h.roomRepo.GetBySlug(slug)
	if err != nil || room == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
		return
	}

	if room.HostID != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only host can close the room"})
		return
	}

	// 1. ZANIM usuniemy pokój, pobieramy z bazy pełną historię
	history, err := h.roomRepo.GetRoomHistory(slug)
	if err != nil {
		history = []models.TrackHistory{}
	}

	// 2. Rozgłaszamy event "ROOM_CLOSED" do wszystkich i DOŁĄCZAMY pobraną historię!
	h.hub.broadcast <- &BroadcastMessage{
		RoomID: slug,
		Payload: gin.H{
			"type": "ROOM_CLOSED",
			"payload": gin.H{
				"history": history,
			},
		},
	}

	if mgr, ok := h.playbackManagers[slug]; ok {
		_ = mgr
	}

	// 3. Dopiero teraz bezpiecznie usuwamy pokój
	if err := h.roomRepo.CloseRoom(slug); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to close room in DB"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Room closed successfully"})
}
