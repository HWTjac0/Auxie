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

	host_id := session.Get("user_id").(int)
	if CheckIfHostHasRoom(host_id) {
		return fmt.Errorf("host already has a room")
	}
	room.HostID = host_id

	room_name := GetRandomRoomName()
	room.Name = room_name

	query := "INSERT INTO rooms (name, host_id) VALUES (?, ?)", room.Name, room.HostID
	query += "UPDATE users SET current_room_id = ?, current_role = ? WHERE id = ?", room.ID, "host", host_id
	err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create room: %w", err)
	}
	return nil
}
