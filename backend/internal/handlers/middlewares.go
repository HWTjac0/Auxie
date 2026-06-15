package handlers

import (
	"auxie/backend/internal/clients"
	"auxie/backend/internal/models"
	"auxie/backend/internal/repositories"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"net/http"
)

// SessionAuthMiddleware verifies that a valid session exists.
func SessionAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, err := getSessionUserID(c)
		if err != nil || userID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No active session"})
			c.Abort()
			return
		}
		c.Set("user_id", userID)
		c.Next()
	}
}

// TokenRefreshMiddleware checks if the user is logged in and has a "Registered" status.
// If so, it checks if any of the connected external services' tokens have expired and refreshes them.
// It does NOT block access for Guests or unauthenticated users.
func TokenRefreshMiddleware(
	userRepo repositories.UserRepository,
	spotifyClient *clients.SpotifyClient,
	tidalClient *clients.TidalClient,
	soundCloudClient *clients.SoundCloudClient,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetInt("user_id")
		if userID == 0 {
			// If for some reason SessionAuthMiddleware wasn't used or failed without aborting
			c.Next()
			return
		}

		user, err := userRepo.GetByID(userID)
		if err != nil || user == nil {
			c.Next()
			return
		}

		if user.Type != models.UserTypeRegistered {
			c.Set("user", user)
			c.Next()
			return
		}

		now := time.Now()
		var tokensRefreshed bool

		// Check Spotify
		if user.SpotifyID.Valid && user.SpotifyID.String != "" {
			if user.SpotifyTokenExpiresAt.Valid && user.SpotifyTokenExpiresAt.Time.Before(now) {
				if user.SpotifyRefreshKey.Valid && user.SpotifyRefreshKey.String != "" {
					log.Printf("Refreshing Spotify token for user %d\n", userID)
					tokenResp, err := spotifyClient.TokenRefresh(user.SpotifyRefreshKey.String)
					if err != nil {
						log.Println("Failed to refresh spotify token:", err)
					} else {
						newExpiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
						refreshKey := tokenResp.RefreshToken
						if refreshKey == "" {
							refreshKey = user.SpotifyRefreshKey.String // Keep old refresh key if new one is not returned
						}
						
						err = userRepo.UpdateSpotifyInfo(user.ID, user.SpotifyID.String, tokenResp.AccessToken, refreshKey, newExpiresAt)
						if err != nil {
							log.Println("Failed to save refreshed spotify token:", err)
						} else {
							tokensRefreshed = true
						}
					}
				}
			}
		}

		// Check SoundCloud
		if user.SoundCloudID.Valid && user.SoundCloudID.String != "" {
			if user.SoundCloudTokenExpiresAt.Valid && user.SoundCloudTokenExpiresAt.Time.Before(now) {
				if user.SoundCloudRefreshKey.Valid && user.SoundCloudRefreshKey.String != "" {
					log.Printf("Refreshing SoundCloud token for user %d\n", userID)
					tokenResp, err := soundCloudClient.TokenRefresh(user.SoundCloudRefreshKey.String)
					if err != nil {
						log.Println("Failed to refresh soundcloud token:", err)
					} else {
						newExpiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
						refreshKey := tokenResp.RefreshToken
						if refreshKey == "" {
							refreshKey = user.SoundCloudRefreshKey.String
						}
						
						err = userRepo.UpdateSoundCloudInfo(user.ID, user.SoundCloudID.String, tokenResp.AccessToken, refreshKey, newExpiresAt)
						if err != nil {
							log.Println("Failed to save refreshed soundcloud token:", err)
						} else {
							tokensRefreshed = true
						}
					}
				}
			}
		}

		// Check Tidal
		if user.TidalID.Valid && user.TidalID.String != "" {
			if user.TidalTokenExpiresAt.Valid && user.TidalTokenExpiresAt.Time.Before(now) {
				if user.TidalRefreshKey.Valid && user.TidalRefreshKey.String != "" {
					log.Printf("Refreshing Tidal token for user %d\n", userID)
					tokenResp, err := tidalClient.TokenRefresh(user.TidalRefreshKey.String)
					if err != nil {
						log.Println("Failed to refresh tidal token:", err)
					} else {
						newExpiresAt := now.Add(time.Duration(tokenResp.ExpiresIn) * time.Second)
						refreshKey := tokenResp.RefreshToken
						if refreshKey == "" {
							refreshKey = user.TidalRefreshKey.String
						}
						
						err = userRepo.UpdateTidalInfo(user.ID, user.TidalID.String, tokenResp.AccessToken, refreshKey, newExpiresAt)
						if err != nil {
							log.Println("Failed to save refreshed tidal token:", err)
						} else {
							tokensRefreshed = true
						}
					}
				}
			}
		}

		// Put the updated user object into the context
		if tokensRefreshed {
			user, _ = userRepo.GetByID(userID)
		}
		c.Set("user", user)

		c.Next()
	}
}
