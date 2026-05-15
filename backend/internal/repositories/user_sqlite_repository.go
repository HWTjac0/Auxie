package repositories

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/models"
	"log"
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
