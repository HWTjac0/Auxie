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
	log.Println("TODO: implement UserSqliteRepo.Create")
	return 0, nil
}

func (r *UserSqliteRepo) UpdateRoom(userId int, roomId int, role *string) error {
	log.Println("TODO: implement UserSqliteRepo.UpdateRoom")
	return nil
}
