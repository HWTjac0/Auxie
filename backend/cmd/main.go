package main

import (
	"auxie/backend/internal/app"
	"auxie/backend/internal/router"
	"log"

	database "auxie/backend/internal/db"

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

	appInstance := app.NewApp(db)

	r := router.SetupRouter(appInstance)
	r.Run("127.0.0.1:8080")
}
