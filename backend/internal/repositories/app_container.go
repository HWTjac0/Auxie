package repositories

import database "auxie/backend/internal/db"

type App struct {
	UserRepo  UserRepository
	TrackRepo TrackRepository
	RoomRepo  RoomRepository
}

func NewApp(db *database.DB) *App {
	return &App{
		UserRepo:  NewUserSqliteRepo(db),
		TrackRepo: NewTrackSqliteRepo(db),
		RoomRepo:  NewRoomSqliteRepo(db),
	}
}
