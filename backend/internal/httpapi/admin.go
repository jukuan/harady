package httpapi

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/jukuan/harady/backend/internal/config"
	"github.com/jukuan/harady/backend/internal/email"
	"github.com/jukuan/harady/backend/internal/models"
	"github.com/jukuan/harady/backend/internal/store"
)

const adminSessionCookie = "harady_admin"

// adminDeps bundles the dependencies the admin routes need.
type adminDeps struct {
	cfg    *config.Config
	cities *store.CityStore
	admins *store.AdminStore
	mailer *email.Sender
	db     *sql.DB
}

// registerAdminRoutes mounts everything under /api/admin/*.
// Auth endpoints live under /api/admin/auth/* (unauthenticated).
// All other routes require either a valid session cookie (browser) or the
// X-Admin-Token header (scripts, backward compat).
func registerAdminRoutes(api *gin.RouterGroup, cfg *config.Config, cities *store.CityStore,
	admins *store.AdminStore, mailer *email.Sender, db *sql.DB) {

	d := &adminDeps{cfg: cfg, cities: cities, admins: admins, mailer: mailer, db: db}
	auth := api.Group("/admin/auth")
	{
		auth.POST("/request", rateLimit(newLimiter(1.0/10.0, 3)), d.requestMagicLink)
		auth.GET("/verify", d.verifyMagicLink)
		auth.GET("/me", d.requireSession(), d.me)
		auth.POST("/logout", d.logout)
	}

	// Everything else is gated.
	gated := api.Group("/admin", d.requireSessionOrToken())

	gated.GET("/cities", d.listCities)
	gated.POST("/cities", d.createCity)
	gated.PUT("/cities/:id", d.updateCity)
	gated.DELETE("/cities/:id", d.deleteCity)

	gated.GET("/tables", d.listTables)
	gated.GET("/tables/:name", d.dumpTable)

	// Stats endpoints (already existed, now behind the same gate).
	gated.GET("/stats/cities", d.statsCities)
	gated.GET("/stats/city/:name", d.statsCity)
	gated.GET("/stats/summary", d.statsSummary)
	gated.GET("/stats/games", d.statsGames)
}

// ---------------------------------------------------------------------------
// auth
// ---------------------------------------------------------------------------

type requestMagicBody struct {
	Email string `json:"email"`
}

func (d *adminDeps) requestMagicLink(c *gin.Context) {
	var body requestMagicBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	emailAddr := strings.ToLower(strings.TrimSpace(body.Email))
	if emailAddr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "email required"})
		return
	}

	if !d.isAllowedEmail(emailAddr) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
		return
	}

	token, err := d.admins.CreateMagicLink(emailAddr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	// Magic link MUST point to the API host, not the frontend host — the
	// verify endpoint exists only on the backend. After verify, we redirect
	// to the frontend /admin page.
	url := d.apiBase(c) + "/api/admin/auth/verify?token=" + token

	subject := "Harady admin — magic link"
	body2 := "Click to log in (expires in 15 minutes):\n\n" + url + "\n\nIf you didn't request this, ignore this email.\n"
	_ = d.mailer.Send(emailAddr, subject, body2)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *adminDeps) verifyMagicLink(c *gin.Context) {
	adminURL := d.adminBase(c) + "/admin"
	token := c.Query("token")
	if token == "" {
		c.Redirect(http.StatusFound, adminURL+"?error=missing_token")
		return
	}
	emailAddr, err := d.admins.ConsumeMagicLink(token)
	if err != nil {
		reason := "expired"
		if errors.Is(err, store.ErrTokenUsed) {
			reason = "used"
		} else if errors.Is(err, store.ErrNotFound) {
			reason = "invalid"
		}
		c.Redirect(http.StatusFound, adminURL+"?error="+reason)
		return
	}

	sess, err := d.admins.CreateSession(emailAddr)
	if err != nil {
		c.Redirect(http.StatusFound, adminURL+"?error=session")
		return
	}

	d.setSessionCookie(c, sess)
	c.Redirect(http.StatusFound, adminURL)
}

func (d *adminDeps) me(c *gin.Context) {
	// requireSession() already validated and stashed the email.
	c.JSON(http.StatusOK, gin.H{"email": c.GetString("admin_email")})
}

func (d *adminDeps) logout(c *gin.Context) {
	if raw, err := c.Cookie(adminSessionCookie); err == nil && raw != "" {
		_ = d.admins.DeleteSession(raw)
	}
	d.clearSessionCookie(c)
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *adminDeps) isAllowedEmail(emailAddr string) bool {
	for _, e := range d.cfg.AdminEmails {
		if strings.EqualFold(strings.TrimSpace(e), emailAddr) {
			return true
		}
	}
	return false
}

func (d *adminDeps) setSessionCookie(c *gin.Context, token string) {
	// SameSite=Lax so normal top-level navigation from the magic-link email
	// carries the cookie; also lets the frontend call the API with it.
	// Secure only when not on localhost — otherwise dev login breaks.
	adminURL := d.adminBase(c)
	secure := strings.HasPrefix(adminURL, "https://")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(
		adminSessionCookie, token,
		int(store.SessionTTL/time.Second),
		"/", "", secure, true,
	)
}

func (d *adminDeps) clearSessionCookie(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(adminSessionCookie, "", -1, "/", "", false, true)
}

// requireSession validates the session cookie and sets c["admin_email"].
func (d *adminDeps) requireSession() gin.HandlerFunc {
	return func(c *gin.Context) {
		raw, err := c.Cookie(adminSessionCookie)
		if err != nil || raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		emailAddr, err := d.admins.LookupSession(raw)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Set("admin_email", emailAddr)
		c.Next()
	}
}

// requireSessionOrToken accepts either the session cookie (browser) OR the
// X-Admin-Token header (scripts). This preserves existing automation.
func (d *adminDeps) requireSessionOrToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Session cookie?
		if raw, err := c.Cookie(adminSessionCookie); err == nil && raw != "" {
			if _, err := d.admins.LookupSession(raw); err == nil {
				c.Next()
				return
			}
		}
		// 2. Header?
		if d.cfg.AdminToken != "" {
			if constantTimeEqual(c.GetHeader("X-Admin-Token"), d.cfg.AdminToken) {
				c.Next()
				return
			}
		}
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
	}
}

// ---------------------------------------------------------------------------
// cities CRUD
// ---------------------------------------------------------------------------

func (d *adminDeps) listCities(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	all, err := d.cities.List()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if q != "" {
		low := strings.ToLower(q)
		filtered := all[:0]
		for _, city := range all {
			if strings.Contains(strings.ToLower(city.Name), low) ||
				strings.Contains(strings.ToLower(city.Region), low) {
				filtered = append(filtered, city)
			}
			if len(filtered) >= limit {
				break
			}
		}
		all = filtered
	} else if len(all) > limit {
		all = all[:limit]
	}
	if all == nil {
		all = []*models.City{}
	}
	c.JSON(http.StatusOK, gin.H{"cities": all, "total": len(all)})
}

type cityBody struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func (d *adminDeps) createCity(c *gin.Context) {
	var body cityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	body.Name = strings.TrimSpace(body.Name)
	body.Region = strings.TrimSpace(body.Region)
	if body.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name required"})
		return
	}
	id, err := d.cities.Create(&models.City{Name: body.Name, Region: body.Region})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

func (d *adminDeps) updateCity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	var body cityBody
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	city, err := d.cities.GetByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if v := strings.TrimSpace(body.Name); v != "" {
		city.Name = v
	}
	city.Region = strings.TrimSpace(body.Region)
	if err := d.cities.Update(city); err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

func (d *adminDeps) deleteCity(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := d.cities.Delete(id); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"ok": true})
}

// ---------------------------------------------------------------------------
// table viewer
// ---------------------------------------------------------------------------

// Only these tables are exposed. Admin auth tables (magic links / sessions)
// are deliberately absent — their rows contain hashed tokens.
var visibleTables = []string{"cities", "city_submissions", "games", "schema_migrations"}

func (d *adminDeps) listTables(c *gin.Context) {
	type info struct {
		Name    string `json:"name"`
		Rows    int    `json:"rows"`
		Columns int    `json:"columns"`
	}
	out := make([]info, 0, len(visibleTables))
	for _, t := range visibleTables {
		var n int
		_ = d.db.QueryRow("SELECT COUNT(*) FROM " + t).Scan(&n)
		cols, _ := d.db.Query("SELECT * FROM " + t + " LIMIT 0")
		colCount := 0
		if cols != nil {
			colCount = len(mustColumns(cols))
			cols.Close()
		}
		out = append(out, info{Name: t, Rows: n, Columns: colCount})
	}
	c.JSON(http.StatusOK, gin.H{"tables": out})
}

func (d *adminDeps) dumpTable(c *gin.Context) {
	name := c.Param("name")
	if !isVisibleTable(name) {
		c.JSON(http.StatusNotFound, gin.H{"error": "table not visible"})
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "100"))
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	// Table name comes from the whitelist, never from user input directly.
	q := "SELECT * FROM " + name + " LIMIT ? OFFSET ?"
	rows, err := d.db.Query(q, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	cols := mustColumns(rows)
	data := make([][]any, 0, limit)
	for rows.Next() {
		row := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range row {
			ptrs[i] = &row[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		data = append(data, row)
	}
	if err := rows.Err(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Also return the CREATE TABLE schema, useful for understanding columns.
	var schema string
	_ = d.db.QueryRow(
		`SELECT sql FROM sqlite_master WHERE type = 'table' AND name = ?`, name,
	).Scan(&schema)

	c.JSON(http.StatusOK, gin.H{
		"table":   name,
		"columns": cols,
		"rows":    data,
		"limit":   limit,
		"offset":  offset,
		"schema":  schema,
	})
}

func mustColumns(rows *sql.Rows) []string {
	cols, err := rows.Columns()
	if err != nil {
		return nil
	}
	return cols
}

func isVisibleTable(name string) bool {
	for _, t := range visibleTables {
		if t == name {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// stats handlers (moved here so they can use the same deps)
// ---------------------------------------------------------------------------

func (d *adminDeps) statsCities(c *gin.Context) {
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
		rows, err = d.cities.BottomCities(limit)
	default:
		rows, err = d.cities.TopCities(limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sort": sort, "limit": limit, "cities": rows})
}

func (d *adminDeps) statsCity(c *gin.Context) {
	stat, err := d.cities.CityStatByName(c.Param("name"))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "city not found"})
		return
	}
	c.JSON(http.StatusOK, stat)
}

func (d *adminDeps) statsSummary(c *gin.Context) {
	citySum, err := d.cities.Summary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	gameSum, err := d.cities.GameSummary()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"cities": citySum, "games": gameSum})
}

func (d *adminDeps) statsGames(c *gin.Context) {
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
		rows, err = d.cities.RecentGames(limit)
	default:
		rows, err = d.cities.LongestGames(limit)
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sort": sort, "limit": limit, "games": rows})
}

// Note: registerStatsRoutes and requireAdminToken from the previous
// httpapi/stats.go are no longer used — registerAdminRoutes covers both.
// Delete stats.go if it still exists (the script does this).

// constantTimeEqual compares two strings in constant time to avoid leaking
// via response-timing which prefix (if any) is correct.
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

// apiBase returns the base URL where the API is reachable from a browser.
// Priority: explicit config → inferred from the incoming request.
func (d *adminDeps) apiBase(c *gin.Context) string {
	if v := strings.TrimRight(d.cfg.APIBaseURL, "/"); v != "" {
		return v
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
}

// adminBase returns the frontend URL where /admin lives.
// Priority: explicit config → first CORS origin → apiBase (fallback).
func (d *adminDeps) adminBase(c *gin.Context) string {
	if v := strings.TrimRight(d.cfg.AdminBaseURL, "/"); v != "" {
		return v
	}
	if len(d.cfg.CORSOrigins) > 0 {
		return strings.TrimRight(d.cfg.CORSOrigins[0], "/")
	}
	return d.apiBase(c)
}
