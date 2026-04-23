package main
import (
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"log"
	"os"
	"auxie/backend/internal/database"
    "auxie/backend/internal/router"
)

func main() {
	// W mainie robimy chyba tylko inicjalizacje bazy danych i uruchamiamy serwer
	// wszystko do obsługi bazy danych jest w db.go, a wszystkie structury i funkcje do obsługi danych są w models.go
	// W user_handler.go będą wszystkie funkcje do obsługi zapytań użytkowników (wiesz te wyszukiwarki i takie tam)
	// routes.go no to ścieżki itp. 
	// W sumie to chyba na razie tyle, potem się zobaczy co jeszcze trzeba dodać
}