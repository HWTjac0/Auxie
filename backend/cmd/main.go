package main

import (
	"auxie/backend/internal/handlers"
	"log"
	"net/http"
	"os"
	"time"

	database "auxie/backend/internal/db"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := database.InitSqliteDB("auxie.db")
	if err != nil {
		log.Fatalf("Database couldn't be initialized: %v", err)
	}

	defer db.Close()

	router := gin.Default()
	router.Use(cors.New(cors.Config{
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
	router.Use(sessions.Sessions("auxie-session", store))

	api := router.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			auth := v1.Group("/auth")
			auth.GET("/me", handlers.GetMe)
			auth.GET("/spotify/login", handlers.SpotifyLogin)
			auth.GET("/spotify/callback", handlers.SpotifyCallback)
		}
	}
	router.Run("127.0.0.1:8080")
}
