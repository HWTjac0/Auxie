package handlers

import (
	"auxie/backend/internal/repositories"
	"github.com/gin-gonic/gin"
)

type RoomHandler struct {
	roomRepo *repositories.RoomSqliteRepo
}

func NewRoomHandler(roomRepo *repositories.RoomSqliteRepo) *RoomHandler {
	host_id := session.Get("user_id")
	name := session.Get("display_name")

	


	return &RoomHandler{roomRepo}
}