package httpapi

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/game"
	"github.com/jukuan/harady/backend/internal/store"
	"github.com/jukuan/harady/backend/internal/ws"
)

func NewRouter(cfg *config.Config, hub *game.Hub, cities *store.CityStore) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	corsCfg := cors.DefaultConfig()
	corsCfg.AllowOrigins = cfg.CORSOrigins
	corsCfg.AllowCredentials = true
	r.Use(cors.New(corsCfg))

	r.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true, "language": cfg.Language})
	})

	// 1 room per 3 seconds, burst 5 — enough for humans, tough on scripts.
	roomLimiter := newLimiter(1.0/3.0, 5)

	api := r.Group("/api")
	{
		api.POST("/rooms", rateLimit(roomLimiter), func(c *gin.Context) {
			room := hub.CreateRoom()
			c.JSON(http.StatusOK, gin.H{
				"code":   room.Code,
				"ws_url": "/ws/rooms/" + room.Code,
			})
		})

		api.GET("/rooms/:code", func(c *gin.Context) {
			room, err := hub.GetRoom(c.Param("code"))
			if err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
				return
			}
			c.JSON(http.StatusOK, room.Snapshot())
		})

		// Tiny health probe for the cities DB. Deliberately does NOT expose
		// the list (players could cheat by reading answers).
		api.GET("/cities/count", func(c *gin.Context) {
			n, err := cities.Count()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, gin.H{"count": n})
		})
	}

	r.GET("/ws/rooms/:code", ws.ServeRoomWS(hub))
	return r
}
