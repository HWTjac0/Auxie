package router

import (
	"auxie/backend/internal/app"
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
			auth.GET("/me", a.UserHandler.GetMe)
			auth.GET("/spotify/login", a.SpotifyHandler.Login)
			auth.GET("/spotify/callback", a.SpotifyHandler.Callback)
			auth.GET("/tidal/login", a.TidalHandler.Login)
			auth.GET("/tidal/callback", a.TidalHandler.Callback)

			room := v1.Group("/room")
			room.GET("/random_name", a.RoomHandler.GetRandomRoomName)
			room.GET("/:slug", a.RoomHandler.GetRoomDetails)
			room.POST("/create", a.RoomHandler.CreateRoom)
			room.POST("/join", a.UserHandler.JoinRoom)

			user := v1.Group("/user")
			user.GET("/random_name", a.UserHandler.GetRandomUserName)
			user.GET("/rooms", a.UserHandler.GetUserRooms)
			user.GET("/logout", a.UserHandler.Logout)
		}
		v1.GET("/search", a.SpotifyHandler.SearchTrack)
	}

	return r
}

