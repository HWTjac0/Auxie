package repositories

import database "auxie/backend/internal/db"

type App struct {
	userRepo  *UserRepository
	trackRepo *TrackRepository
	roomRepo  *TrackRepository
}

func NewApp(db *database.DB) *App {
	return &App{
		userRepo:  NewUserSqliteRepo(db),
		trackRepo: NewTrackSqliteRepo(db),
		roomRepo:  NewRoomSqliteRepo(db),
	}
}
