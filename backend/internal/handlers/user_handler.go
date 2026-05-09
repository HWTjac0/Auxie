package handlers

import "auxie/backend/internal/repositories"

type UserHandler struct {
	userRepo repositories.UserRepository
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{userRepo: userRepo}
}
