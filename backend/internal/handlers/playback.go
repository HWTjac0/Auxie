package handlers

import (
	"context"
	"sync"
	"time"

	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"

	"github.com/gin-gonic/gin"
)

// RoomPlaybackManager coordinates starting tracks for a single room (no buffering).
type RoomPlaybackManager struct {
	roomSlug  string
	roomRepo  repositories.RoomRepository
	trackRepo repositories.TrackRepository
	hub       *RoomHub

	mu      sync.Mutex
	current *models.RoomQueueItem
	cancel  context.CancelFunc
}

func NewRoomPlaybackManager(hub *RoomHub, roomSlug string, roomRepo repositories.RoomRepository, trackRepo repositories.TrackRepository) *RoomPlaybackManager {
	return &RoomPlaybackManager{
		roomSlug:  roomSlug,
		roomRepo:  roomRepo,
		trackRepo: trackRepo,
		hub:       hub,
	}
}

// StartIfIdle checks the queue and if nothing is playing it immediately starts the next queued track.
func (m *RoomPlaybackManager) StartIfIdle() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If something is already playing, ignore
	if m.cancel != nil {
		return
	}

	room, err := m.roomRepo.GetBySlug(m.roomSlug)
	if err != nil || room == nil {
		return
	}

	queue, err := m.roomRepo.GetQueue(room.ID)
	if err != nil || len(queue) == 0 {
		return
	}

	next := queue[0]
	if next.Status.String() == models.TrackStatusPlaying.String() {
		return
	}

	m.cancel = nil
	m.current = &next
	now := time.Now()

	// Immediately broadcast start event
	m.hub.broadcast <- &BroadcastMessage{
		RoomID: m.roomSlug,
		Payload: gin.H{
			"type":       "playback:start",
			"track":      next,
			"started_at": now.Format(time.RFC3339),
		},
	}

	// Update DB: status=playing, start_timestamp=now
	_ = m.roomRepo.UpdateTrackStatus(next.RoomTrackID, "playing")
	_ = m.roomRepo.UpdateTrackTimestamps(next.RoomTrackID, &now, nil)

	// Clear state
	m.cancel = nil
	m.current = nil
}

// Skip marks current as skipped and triggers next.
func (m *RoomPlaybackManager) Skip() {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	if m.current != nil {
		_ = m.roomRepo.UpdateTrackStatus(m.current.RoomTrackID, "skipped")
		now := time.Now()
		_ = m.roomRepo.UpdateTrackTimestamps(m.current.RoomTrackID, nil, &now)
		m.hub.broadcast <- &BroadcastMessage{
			RoomID:  m.roomSlug,
			Payload: gin.H{"type": "playback:skipped", "track": m.current},
		}
		m.current = nil
	}
}

// MarkEnded marks track as played and records end timestamp.
func (m *RoomPlaybackManager) MarkEnded(roomTrackID int) {
	_ = m.roomRepo.UpdateTrackStatus(roomTrackID, "played")
	now := time.Now()
	_ = m.roomRepo.UpdateTrackTimestamps(roomTrackID, nil, &now)
}
