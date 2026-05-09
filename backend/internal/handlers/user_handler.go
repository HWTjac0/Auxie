package handlers

import (
	"auxie/backend/internal/repositories"
	"fmt"
	"math/rand/v2"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetRandomUserName(c *gin.Context) {
	adjectives := []string{"Happy", "Lucky", "Funky", "Sonic", "Disco", "Neon", "Groovy", "Velvet", "Retro", "Electric"}
	nouns := []string{"Listener", "Dancer", "Fan", "Maestro", "DJ", "Viber", "Soul", "Star", "Beat", "Explorer"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	c.JSON(200, gin.H{"name": fmt.Sprintf("%s %s", adj, noun)})
}
