package router

import (
	"auxie/backend/internal/app"
	"auxie/backend/internal/handlers"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func SetupRouter(a *app.App) *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	secret := os.Getenv("COOKIE_SECRET_KEY")
	store := cookie.NewStore([]byte(secret))

	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   3600 * 24, // 24h
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	r.Use(sessions.Sessions("auxie-session", store))

	api := r.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			auth := v1.Group("/auth")
			auth.GET("/me", handlers.SessionAuthMiddleware(), a.UserHandler.GetMe)
			auth.GET("/spotify/login", a.SpotifyHandler.Login)
			auth.GET("/spotify/callback", a.SpotifyHandler.Callback)
			auth.GET("/tidal/login", a.TidalHandler.Login)
			auth.GET("/tidal/callback", a.TidalHandler.Callback)
			auth.GET("/soundcloud/login", a.SoundCloudHandler.SoundCloudLogin)
			auth.GET("/soundcloud/callback", a.SoundCloudHandler.SoundCloudCallback)

			room := v1.Group("/room")
			room.GET("/random_name", a.RoomHandler.GetRandomRoomName)
			room.GET("/:slug", a.RoomHandler.GetRoomDetails)
			room.POST("/join", a.UserHandler.JoinRoom)

			stream := v1.Group("/stream")
			stream.GET("/spotify/:room_track_id", a.RoomHandler.StreamSpotify)
			stream.GET("/tidal/:room_track_id", a.RoomHandler.StreamTidal)
			stream.GET("/soundcloud/:room_track_id", a.RoomHandler.StreamSoundCloud)

			user := v1.Group("/user")
			user.GET("/random_name", a.UserHandler.GetRandomUserName)
			user.GET("/logout", a.UserHandler.Logout)

			protected := v1.Group("/")
			protected.Use(handlers.SessionAuthMiddleware())
			protected.Use(handlers.TokenRefreshMiddleware(
				a.UserRepo,
				a.SpotifyClient,
				a.TidalClient,
				a.SoundCloudClient,
			))

			protected.POST("/room/create", a.RoomHandler.CreateRoom)
			protected.POST("/room/:slug/track", a.RoomHandler.AddTrack)
			protected.GET("/playback/token", a.RoomHandler.GetPlaybackToken)
			protected.POST("/room/:slug/skip", a.RoomHandler.SkipTrack)
			protected.POST("/room/:slug/vote-skip", a.RoomHandler.VoteSkip)
			protected.POST("/room/:slug/track/:track_id/like", a.RoomHandler.LikeTrack)
			protected.POST("/room/:slug/user/:username/role", a.RoomHandler.ChangeUserRole)
			protected.DELETE("/room/:slug/user/:username", a.RoomHandler.KickUser)
			protected.POST("/room/:slug/proposed/:track_id/approve", a.RoomHandler.ApproveTrack)
			protected.POST("/room/:slug/proposed/:track_id/reject", a.RoomHandler.RejectTrack)
			protected.GET("/user/rooms", a.UserHandler.GetUserRooms)
			protected.GET("/search", a.UserHandler.Search)
			protected.GET("/room/:slug/ws", a.RoomHandler.HandleWS)
		}
	}

	return r
}
