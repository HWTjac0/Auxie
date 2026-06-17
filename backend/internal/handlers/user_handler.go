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
	soundcloudName := session.Get("soundcloud_name")

	c.JSON(http.StatusOK, gin.H{
		"name":            name,
		"spotify_name":    spotifyName,
		"tidal_name":      tidalName,
		"soundcloud_name": soundcloudName,
		"id":              session.Get("user_id"),
		"image":           session.Get("user_image"),
	})
}

func (h *UserHandler) DisconnectService(c *gin.Context) {
	userID := c.GetInt("user_id")
	service := c.Param("service")

	var err error
	session := sessions.Default(c)

	switch service {
	case "spotify":
		err = h.userRepo.DisconnectSpotify(userID)
		session.Delete("spotify_name")
	case "tidal":
		err = h.userRepo.DisconnectTidal(userID)
		session.Delete("tidal_name")
	case "soundcloud":
		err = h.userRepo.DisconnectSoundCloud(userID)
		session.Delete("soundcloud_name")
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid service"})
		return
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to disconnect service"})
		return
	}

	_ = session.Save()
	c.JSON(http.StatusOK, gin.H{"message": fmt.Sprintf("%s disconnected successfully", service)})
}

func (h *UserHandler) RefreshMe(c *gin.Context) {
	userID := c.GetInt("user_id")
	dbUser, err := h.userRepo.GetByID(userID)
	if err != nil || dbUser == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}

	session := sessions.Default(c)
	session.Set("user_id", dbUser.ID)
	session.Set("user_name", dbUser.Username)

	if dbUser.SpotifyID.Valid && dbUser.SpotifyID.String != "" {
		// If we don't have a specific name in session, we use db user's name
		if session.Get("spotify_name") == nil {
			session.Set("spotify_name", dbUser.Username)
		}
	} else {
		session.Delete("spotify_name")
	}

	if dbUser.TidalID.Valid && dbUser.TidalID.String != "" {
		if session.Get("tidal_name") == nil {
			session.Set("tidal_name", dbUser.Username)
		}
	} else {
		session.Delete("tidal_name")
	}

	if dbUser.SoundCloudID.Valid && dbUser.SoundCloudID.String != "" {
		if session.Get("soundcloud_name") == nil {
			session.Set("soundcloud_name", dbUser.Username)
		}
	} else {
		session.Delete("soundcloud_name")
	}

	_ = session.Save()

	c.JSON(http.StatusOK, gin.H{
		"name":            dbUser.Username,
		"spotify_name":    session.Get("spotify_name"),
		"tidal_name":      session.Get("tidal_name"),
		"soundcloud_name": session.Get("soundcloud_name"),
		"id":              dbUser.ID,
		"image":           session.Get("user_image"),
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
