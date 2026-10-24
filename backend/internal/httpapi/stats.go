package httpapi

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/store"
)

func registerStatsRoutes(api *gin.RouterGroup, cfg *config.Config, cities *store.CityStore) {
	if cfg.AdminToken == "" {
		return
	}
	admin := api.Group("/admin", requireAdminToken(cfg.AdminToken))

	admin.GET("/stats/cities", func(c *gin.Context) {
		sort := c.DefaultQuery("sort", "top")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		if limit <= 0 || limit > 1000 {
			limit = 50
		}
		var (
			rows []store.CityStat
			err  error
		)
		switch sort {
		case "bottom":
			rows, err = cities.BottomCities(limit)
		default:
			rows, err = cities.TopCities(limit)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sort": sort, "limit": limit, "cities": rows})
	})

	admin.GET("/stats/city/:name", func(c *gin.Context) {
		stat, err := cities.CityStatByName(c.Param("name"))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
			return
		}
		c.JSON(http.StatusOK, stat)
	})

	admin.GET("/stats/summary", func(c *gin.Context) {
		citySum, err := cities.Summary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		gameSum, err := cities.GameSummary()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"cities": citySum, "games": gameSum})
	})

	admin.GET("/stats/games", func(c *gin.Context) {
		sort := c.DefaultQuery("sort", "longest")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		if limit <= 0 || limit > 1000 {
			limit = 20
		}
		var (
			rows []*store.GameStat
			err  error
		)
		switch sort {
		case "recent":
			rows, err = cities.RecentGames(limit)
		default:
			rows, err = cities.LongestGames(limit)
		}
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"sort": sort, "limit": limit, "games": rows})
	})
}

func requireAdminToken(want string) gin.HandlerFunc {
	return func(c *gin.Context) {
		got := c.GetHeader("X-Admin-Token")
		if !constantTimeEqual(got, want) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}

func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := 0; i < len(a); i++ {
		v |= a[i] ^ b[i]
	}
	return v == 0
}
