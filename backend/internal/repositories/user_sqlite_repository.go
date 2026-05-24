package repositories

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/models"
	"fmt"
	"log"
	"strings"
)

type UserSqliteRepo struct {
	db *database.DB
}

func NewUserSqliteRepo(db *database.DB) *UserSqliteRepo {
	return &UserSqliteRepo{db}
}

func (r *UserSqliteRepo) GetByEmail(email string) (*models.User, error) {
	log.Println("TODO: implement UserSqliteRepo.GetByEmail")
	return nil, nil
}

func (r *UserSqliteRepo) Create(user *models.User) (int64, error) {
	query := `INSERT INTO users (email, username, type, created_at) VALUES (?, ?, ?, ?)`

	result, err := r.db.Exec(query, user.Email, user.Username, user.Type, user.CreatedAt)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (r *UserSqliteRepo) UpdateRoom(userId int, roomId int, role *string) error {
	query := `UPDATE users SET current_room_id = ?, current_role = ? WHERE id = ?`

	_, err := r.db.Exec(query, roomId, role, userId)
	return err
}

func (r *UserSqliteRepo) GetUsersInRoom(roomId int, filter *UserFilter) ([]models.User, error) {
	var users []models.User

	selectFields := "username, type"
	if filter != nil && len(filter.Fields) > 0 {
		allowedFields := map[string]bool{
			"id": true, "email": true, "username": true, "type": true,
			"soundcloud_id": true, "soundcloud_key": true, "tidal_id": true, "tidal_key": true,
			"spotify_id": true, "spotify_auth_key": true, "spotify_refresh_key": true,
			"spotify_token_expires_at": true, "current_room_id": true, "current_role": true,
			"created_at": true, "updated_at": true,
		}
		var validatedFields []string
		for _, f := range filter.Fields {
			if allowedFields[f] {
				validatedFields = append(validatedFields, f)
			}
		}
		if len(validatedFields) > 0 {
			selectFields = strings.Join(validatedFields, ", ")
		}
	}

	query := fmt.Sprintf(`SELECT %s FROM users WHERE current_room_id = ?`, selectFields)
	args := []interface{}{roomId}

	if filter != nil {
		if filter.Role != nil {
			query += ` AND current_role = ?`
			args = append(args, *filter.Role)
		}
		if filter.Type != nil {
			query += ` AND type = ?`
			args = append(args, *filter.Type)
		}
	}

	err := r.db.Unsafe().Select(&users, query, args...)

	if err != nil {
		return nil, err
	}
	return users, nil
}
