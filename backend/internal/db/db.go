package database

import (
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
	_ "log"
)

var DB *sql.DB

// TODO: łączenie się z bazą i możemy wyjebać te paczki od mongodb, bo chyba nie będziemy się piepszyć z noSQL-em, a SQLite jest prosty i lekki, więc chyba będzie idealny do tego projektu. W sumie to może być nawet coś jeszcze lżejszego, ale na razie to chyba wystarczy. No i oczywiście trzeba będzie dodać funkcje do obsługi bazy danych, takie jak dodawanie użytkowników, pokoi, tracków itp. Ale to już w models.go, a tutaj tylko inicjalizacja bazy danych i ewentualnie jakieś funkcje pomocnicze do obsługi połączenia z bazą.
func InitDB() error {
	var err error
	// Jak wygenerujesz bazunie to sie to zrobi
}