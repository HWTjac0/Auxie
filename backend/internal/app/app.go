package app

import (
	database "auxie/backend/internal/db"
	"auxie/backend/internal/handlers"
	"auxie/backend/internal/repositories"
)

type App struct {
	UserRepo  repositories.UserRepository
	TrackRepo repositories.TrackRepository
	RoomRepo  repositories.RoomRepository

	UserHandler    *handlers.UserHandler
	RoomHandler    *handlers.RoomHandler
	SpotifyHandler *handlers.SpotifyHandler
	TidalHandler   *handlers.TidalHandler
}

func NewApp(db *database.DB) *App {
	userRepo := repositories.NewUserSqliteRepo(db)
	trackRepo := repositories.NewTrackSqliteRepo(db)
	roomRepo := repositories.NewRoomSqliteRepo(db)
	return &App{
		UserRepo:  userRepo,
		TrackRepo: trackRepo,
		RoomRepo:  roomRepo,

		UserHandler:    handlers.NewUserHandler(userRepo, roomRepo),
		RoomHandler:    handlers.NewRoomHandler(roomRepo, userRepo),
		SpotifyHandler: handlers.NewSpotifyHandler(userRepo),
		TidalHandler:   handlers.NewTidalHandler(userRepo),
	}
}
