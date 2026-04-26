package repositories

import database "auxie/backend/internal/db"

type UserSqliteRepo struct {
	db *database.DB
}

func NewUserSqliteRepo(db *database.DB) *UserSqliteRepo {
	return &UserSqliteRepo{db}
}
