package handlers

import (
	"auxie/backend/internal/repositories"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repositories.UserRepository
	roomRepo repositories.RoomRepository
}

func NewUserHandler(userRepo repositories.UserRepository, roomRepo repositories.RoomRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo, roomRepo: roomRepo}
}

func (h *UserHandler) GetRandomUserName(c *gin.Context) {
	adjectives := []string{"Happy", "Lucky", "Funky", "Sonic", "Disco", "Neon", "Groovy", "Velvet", "Retro", "Electric"}
	nouns := []string{"Listener", "Dancer", "Fan", "Maestro", "DJ", "Viber", "Soul", "Star", "Beat", "Explorer"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	c.JSON(200, gin.H{"name": fmt.Sprintf("%s %s", adj, noun)})
}

func (h *UserHandler) GetUserRooms(c *gin.Context) {
	session := sessions.Default(c)
	userIDVal := session.Get("user_id")

	if userIDVal == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No active session"})
		return
	}

	var userID int
	switch val := userIDVal.(type) {
	case int:
		userID = val
	case int64:
		userID = int(val)
	case float64:
		userID = int(val)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID type in session"})
		return
	}

	room, err := h.roomRepo.GetByHostId(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{room}})
}

func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, "/welcome")
}
