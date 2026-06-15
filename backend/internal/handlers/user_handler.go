package handlers

import (
	"auxie/backend/internal/clients"
	"auxie/backend/internal/repositories"
	"database/sql"
	"fmt"
	"math/rand/v2"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userRepo         repositories.UserRepository
	roomRepo         repositories.RoomRepository
	spotifyClient    *clients.SpotifyClient
	tidalClient      *clients.TidalClient
	soundCloudClient *clients.SoundCloudClient
}

func NewUserHandler(
	userRepo repositories.UserRepository,
	roomRepo repositories.RoomRepository,
	spotifyClient *clients.SpotifyClient,
	tidalClient *clients.TidalClient,
	soundCloudClient *clients.SoundCloudClient,
) *UserHandler {
	return &UserHandler{
		userRepo:         userRepo,
		roomRepo:         roomRepo,
		spotifyClient:    spotifyClient,
		tidalClient:      tidalClient,
		soundCloudClient: soundCloudClient,
	}
}

func (h *UserHandler) GetRandomUserName(c *gin.Context) {
	adjectives := []string{"Happy", "Lucky", "Funky", "Sonic", "Disco", "Neon", "Groovy", "Velvet", "Retro", "Electric"}
	nouns := []string{"Listener", "Dancer", "Fan", "Maestro", "DJ", "Viber", "Soul", "Star", "Beat", "Explorer"}

	adj := adjectives[rand.IntN(len(adjectives))]
	noun := nouns[rand.IntN(len(nouns))]

	c.JSON(200, gin.H{"name": fmt.Sprintf("%s %s", adj, noun)})
}

func (h *UserHandler) GetUserRooms(c *gin.Context) {
	userID := c.GetInt("user_id")

	room, err := h.roomRepo.GetByHostId(userID)
	if err != nil {
		if err == sql.ErrNoRows {
			c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user rooms"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"rooms": []interface{}{room}})
}

func (h *UserHandler) Logout(c *gin.Context) {
	session := sessions.Default(c)
	session.Clear()
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, "/welcome")
}

func (h *UserHandler) GetMe(c *gin.Context) {
	session := sessions.Default(c)
	name := session.Get("user_name")
	spotifyName := session.Get("spotify_name")
	tidalName := session.Get("tidal_name")

	c.JSON(http.StatusOK, gin.H{
		"name":         name,
		"spotify_name": spotifyName,
		"tidal_name":   tidalName,
		"id":           session.Get("user_id"),
		"image":        session.Get("user_image"),
	})
}

func (h *UserHandler) Search(c *gin.Context) {
	userId := c.GetInt("user_id")
	dbUser, err := h.userRepo.GetByID(userId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to resolve user"})
		return
	}

	keywords := c.Query("search")
	if keywords == "" {
		keywords = c.Query("q")
	}

	results := make(map[string]interface{})

	// Check Spotify
	hasSpotify := dbUser.SpotifyID.Valid && dbUser.SpotifyID.String != ""
	if hasSpotify {
		accessToken := dbUser.SpotifyAuthKey.String
		spotifyRes, err := h.spotifyClient.SearchTrack(accessToken, keywords)
		if err == nil {
			results["spotify"] = spotifyRes
		} else {
			results["spotify_error"] = err.Error()
		}
	} else {
		// Fallback for Guests or users without Spotify to search using client credentials
		token, ccErr := h.spotifyClient.GetClientCredentialsToken()
		if ccErr == nil {
			spotifyRes, err := h.spotifyClient.SearchTrack(token, keywords)
			if err == nil {
				results["spotify"] = spotifyRes
			}
		}
	}

	// Check Tidal
	hasTidal := dbUser.TidalID.Valid && dbUser.TidalID.String != ""
	if hasTidal {
		accessToken := dbUser.TidalKey.String
		tidalRes, err := h.tidalClient.SearchTrack(accessToken, keywords)
		if err == nil {
			results["tidal"] = tidalRes
		} else {
			results["tidal_error"] = err.Error()
		}
	}

	c.JSON(http.StatusOK, results)
}
