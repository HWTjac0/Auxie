package app

import (
	"auxie/backend/internal/clients"
	database "auxie/backend/internal/db"
	"auxie/backend/internal/handlers"
	"auxie/backend/internal/repositories"
	"os"
)

type App struct {
	UserRepo  repositories.UserRepository
	TrackRepo repositories.TrackRepository
	RoomRepo  repositories.RoomRepository

	UserHandler       *handlers.UserHandler
	RoomHandler       *handlers.RoomHandler
	SpotifyHandler    *handlers.SpotifyHandler
	TidalHandler      *handlers.TidalHandler
	SoundCloudHandler *handlers.SoundCloudHandler
	
	SpotifyClient    *clients.SpotifyClient
	TidalClient      *clients.TidalClient
	SoundCloudClient *clients.SoundCloudClient
}

func NewApp(db *database.DB) *App {
	userRepo := repositories.NewUserSqliteRepo(db)
	trackRepo := repositories.NewTrackSqliteRepo(db)
	roomRepo := repositories.NewRoomSqliteRepo(db)

	spotifyClient := clients.NewSpotifyClient(
		os.Getenv("SPOTIFY_CLIENT_ID"),
		os.Getenv("SPOTIFY_CLIENT_SECRET"),
		"http://127.0.0.1:8080/api/v1/auth/spotify/callback",
	)

	tidalClient := clients.NewTidalClient(
		os.Getenv("TIDAL_CLIENT_ID"),
		"http://127.0.0.1:8080/api/v1/auth/tidal/callback",
	)

	soundCloudClient := clients.NewSoundCloudClient(
		os.Getenv("SOUNDCLOUD_CLIENT_ID"),
		os.Getenv("SOUNDCLOUD_CLIENT_SECRET"),
		"http://127.0.0.1:8080/api/v1/auth/soundcloud/callback",
	)

	return &App{
		UserRepo:  userRepo,
		TrackRepo: trackRepo,
		RoomRepo:  roomRepo,

		UserHandler:       handlers.NewUserHandler(userRepo, roomRepo, spotifyClient, tidalClient, soundCloudClient),
		RoomHandler:       handlers.NewRoomHandler(roomRepo, userRepo),
		SpotifyHandler:    handlers.NewSpotifyHandler(userRepo, spotifyClient),
		TidalHandler:      handlers.NewTidalHandler(userRepo, tidalClient),
		SoundCloudHandler: handlers.NewSoundCloudHandler(userRepo),

		SpotifyClient:    spotifyClient,
		TidalClient:      tidalClient,
		SoundCloudClient: soundCloudClient,
	}
}
