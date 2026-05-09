package handlers

import (
	"auxie/backend/internal/repositories"
	"fmt"
	"math/rand/v2"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}

func (h *UserHandler) GetRandomUserName() string {
	adjectives := []string{"Happy", "Lucky", "Funky", "Sonic", "Disco", "Neon", "Groovy", "Velvet", "Retro", "Electric"}
	nouns := []string{"Listener", "Dancer", "Fan", "Maestro", "DJ", "Viber", "Soul", "Star", "Beat", "Explorer"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	return fmt.Sprintf("%s %s", adj, noun)
}
