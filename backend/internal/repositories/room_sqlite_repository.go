package repositories

import database "auxie/backend/internal/db"

type RoomSqliteRepo struct {
	db *database.DB
}

func NewRoomSqliteRepo(db *database.DB) *RoomSqliteRepo {
	return &RoomSqliteRepo{db}
}

func AddNewRoom(room *Room) error {
	if session.Get("user_id") == nil {
		return fmt.Errorf("user not logged in")
	}

	

	return nil
}
