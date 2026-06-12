package handlers

import (
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"errors"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

func getSessionUserID(c *gin.Context) (int, error) {
	session := sessions.Default(c)
	userIDVal := session.Get("user_id")
	if userIDVal == nil {
		return 0, errors.New("no active session user_id")
	}

	switch val := userIDVal.(type) {
	case int:
		return val, nil
	case int64:
		return int(val), nil
	case float64:
		return int(val), nil
	default:
		return 0, errors.New("invalid user ID type in session")
	}
}

func getOrCreateUser(c *gin.Context, userRepo repositories.UserRepository, displayName string) (*models.User, error) {
	if guestUserID, err := getSessionUserID(c); err == nil && guestUserID > 0 {
		return &models.User{ID: guestUserID}, nil
	}

	newUser := &models.User{
		Username:  displayName,
		Type:      models.UserTypeRegistered,
		CreatedAt: time.Now(),
	}
	newID, err := userRepo.Create(newUser)
	if err != nil {
		return nil, err
	}
	return &models.User{ID: int(newID)}, nil
}
