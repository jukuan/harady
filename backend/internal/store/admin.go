package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"
)

// ErrTokenExpired / ErrTokenUsed let the HTTP layer return specific errors.
var (
	ErrTokenExpired = errors.New("token expired")
	ErrTokenUsed    = errors.New("token already used")
	ErrEmailInvalid = errors.New("email not allowed")
)

const (
	MagicLinkTTL = 15 * time.Minute
	SessionTTL   = 7 * 24 * time.Hour
)

// AdminStore handles passwordless login tokens.
// Magic links are single-use and short-lived. Sessions are long-lived until
// the cookie expires or the user explicitly logs out.
type AdminStore struct{ db *sql.DB }

func NewAdminStore(db *sql.DB) *AdminStore { return &AdminStore{db: db} }

// CreateMagicLink generates a new single-use token for email and returns the
// plaintext token (the caller must put it into the URL and email it).
// The DB stores only the SHA-256 hash.
func (s *AdminStore) CreateMagicLink(email string) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(MagicLinkTTL).UTC()
	_, err = s.db.Exec(
		`INSERT INTO admin_magic_links (email, token_hash, expires_at) VALUES (?, ?, ?)`,
		email, hash, expires,
	)
	if err != nil {
		return "", err
	}
	return raw, nil
}

// ConsumeMagicLink verifies the token and marks it used. Returns the email
// it was issued for, or an error.
func (s *AdminStore) ConsumeMagicLink(raw string) (string, error) {
	hash := hashToken(raw)
	var (
		id        int64
		email     string
		expiresAt time.Time
		usedAt    sql.NullTime
	)
	err := s.db.QueryRow(
		`SELECT id, email, expires_at, used_at FROM admin_magic_links WHERE token_hash = ?`,
		hash,
	).Scan(&id, &email, &expiresAt, &usedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if usedAt.Valid {
		return "", ErrTokenUsed
	}
	if time.Now().After(expiresAt) {
		return "", ErrTokenExpired
	}
	// Mark used atomically — if another request raced us, RowsAffected will be 0.
	res, err := s.db.Exec(
		`UPDATE admin_magic_links SET used_at = CURRENT_TIMESTAMP WHERE id = ? AND used_at IS NULL`,
		id,
	)
	if err != nil {
		return "", err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return "", ErrTokenUsed
	}
	return email, nil
}

// CreateSession generates a session token for email and returns the plaintext.
func (s *AdminStore) CreateSession(email string) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	expires := time.Now().Add(SessionTTL).UTC()
	_, err = s.db.Exec(
		`INSERT INTO admin_sessions (email, token_hash, expires_at) VALUES (?, ?, ?)`,
		email, hash, expires,
	)
	if err != nil {
		return "", err
	}
	return raw, nil
}

// LookupSession returns the admin email for a valid session token.
func (s *AdminStore) LookupSession(raw string) (string, error) {
	if raw == "" {
		return "", ErrNotFound
	}
	hash := hashToken(raw)
	var (
		email     string
		expiresAt time.Time
	)
	err := s.db.QueryRow(
		`SELECT email, expires_at FROM admin_sessions WHERE token_hash = ?`,
		hash,
	).Scan(&email, &expiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", err
	}
	if time.Now().After(expiresAt) {
		return "", ErrTokenExpired
	}
	return email, nil
}

// DeleteSession revokes one session token (logout).
func (s *AdminStore) DeleteSession(raw string) error {
	hash := hashToken(raw)
	_, err := s.db.Exec(`DELETE FROM admin_sessions WHERE token_hash = ?`, hash)
	return err
}

// CleanupExpired removes tokens that can no longer be used. Called from a
// background ticker in main; not critical for correctness, just hygiene.
func (s *AdminStore) CleanupExpired() (int, error) {
	res1, err := s.db.Exec(
		`DELETE FROM admin_magic_links WHERE expires_at < datetime('now', '-1 hour') OR (used_at IS NOT NULL AND used_at < datetime('now', '-1 day'))`,
	)
	if err != nil {
		return 0, err
	}
	res2, err := s.db.Exec(`DELETE FROM admin_sessions WHERE expires_at < datetime('now')`)
	if err != nil {
		return 0, err
	}
	n1, _ := res1.RowsAffected()
	n2, _ := res2.RowsAffected()
	return int(n1 + n2), nil
}

// newToken returns (plaintext, sha256hex, error). 32 random bytes → 43-char
// base64url string is plenty for a one-off link.
func newToken() (string, string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	raw := base64.RawURLEncoding.EncodeToString(b[:])
	return raw, hashToken(raw), nil
}

func hashToken(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}
