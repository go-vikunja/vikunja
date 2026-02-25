// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/web"

	"github.com/google/uuid"
	"xorm.io/builder"
	"xorm.io/xorm"
)

// Session represents an active user session with a refresh token.
type Session struct {
	// The session UUID. Embedded in JWTs as the `sid` claim.
	ID string `xorm:"varchar(36) not null unique pk" json:"id" param:"session"`
	// The owning user.
	UserID int64 `xorm:"bigint not null index" json:"-"`
	// SHA-256 hash of the refresh token. Used for lookup on refresh.
	TokenHash string `xorm:"varchar(64) not null unique index" json:"-"`
	// The cleartext refresh token. Only populated on session creation, never stored.
	RefreshToken string `xorm:"-" json:"refresh_token,omitempty"`
	// User-Agent string from the login request.
	DeviceInfo string `xorm:"text" json:"device_info"`
	// IP address from the login request.
	IPAddress string `xorm:"varchar(100)" json:"ip_address"`
	// Whether this is a "remember me" session (controls max refresh lifetime).
	IsLongSession bool `xorm:"not null default false" json:"-"`
	// When this session was last refreshed.
	LastActive time.Time `xorm:"not null" json:"last_active"`
	// When this session was created (login time).
	Created time.Time `xorm:"created not null" json:"created"`

	web.Permissions `xorm:"-" json:"-"`
	web.CRUDable    `xorm:"-" json:"-"`
}

func (*Session) TableName() string {
	return "sessions"
}

// HashSessionToken returns the hex-encoded SHA-256 hash of a token string.
// No salt needed because refresh tokens are high-entropy random strings,
// not human passwords — rainbow tables and dictionary attacks don't apply.
func HashSessionToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// generateHashedToken creates a cryptographically random token and returns both
// the raw hex-encoded token (to give to the client) and its SHA-256 hash (to store).
func generateHashedToken() (rawToken, hash string, err error) {
	tokenBytes, err := utils.CryptoRandomBytes(128)
	if err != nil {
		return "", "", err
	}
	rawToken = hex.EncodeToString(tokenBytes)
	return rawToken, HashSessionToken(rawToken), nil
}

// CreateSession creates a new session record and generates a refresh token.
// Returns the session with RefreshToken populated (cleartext, shown only once).
func CreateSession(s *xorm.Session, userID int64, deviceInfo, ipAddress string, isLongSession bool) (*Session, error) {
	rawToken, hash, err := generateHashedToken()
	if err != nil {
		return nil, err
	}

	session := &Session{
		ID:            uuid.New().String(),
		UserID:        userID,
		TokenHash:     hash,
		DeviceInfo:    deviceInfo,
		IPAddress:     ipAddress,
		IsLongSession: isLongSession,
		LastActive:    time.Now(),
	}

	_, err = s.Insert(session)
	if err != nil {
		return nil, err
	}

	session.RefreshToken = rawToken
	return session, nil
}

// GetSessionByRefreshToken finds a session by the SHA-256 hash of the provided token.
func GetSessionByRefreshToken(s *xorm.Session, token string) (*Session, error) {
	hash := HashSessionToken(token)
	session := &Session{}
	has, err := s.Where("token_hash = ?", hash).Get(session)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, &ErrSessionNotFound{}
	}
	return session, nil
}

// GetSessionByID finds a session by its UUID.
func GetSessionByID(s *xorm.Session, id string) (*Session, error) {
	session := &Session{}
	has, err := s.Where("id = ?", id).Get(session)
	if err != nil {
		return nil, err
	}
	if !has {
		return nil, &ErrSessionNotFound{}
	}
	return session, nil
}

// ReadAll returns all sessions for the authenticated user.
func (sess *Session) ReadAll(s *xorm.Session, a web.Auth, _ string, page int, perPage int) (result interface{}, resultCount int, numberOfTotalItems int64, err error) {
	// Link share tokens must not be able to list user sessions.
	if _, is := a.(*LinkSharing); is {
		return nil, 0, 0, ErrGenericForbidden{}
	}

	sessions := []*Session{}

	var where builder.Cond = builder.Eq{"user_id": a.GetID()}

	err = s.
		Where(where).
		OrderBy("last_active DESC").
		Limit(getLimitFromPageIndex(page, perPage)).
		Find(&sessions)
	if err != nil {
		return nil, 0, 0, err
	}

	totalCount, err := s.Where(where).Count(&Session{})
	return sessions, len(sessions), totalCount, err
}

// Delete deletes a session by ID, scoped to the owning user.
func (sess *Session) Delete(s *xorm.Session, a web.Auth) error {
	_, err := s.Where("id = ? AND user_id = ?", sess.ID, a.GetID()).Delete(&Session{})
	return err
}

// UpdateSessionLastActive updates the last_active timestamp of a session.
func UpdateSessionLastActive(s *xorm.Session, sessionID string) error {
	_, err := s.Where("id = ?", sessionID).
		Cols("last_active").
		Update(&Session{LastActive: time.Now()})
	return err
}

// RotateRefreshToken atomically replaces the session's refresh token hash.
// The WHERE clause includes the old hash so that concurrent refreshes with the
// same token cannot both succeed — only the first UPDATE matches a row; the
// second sees 0 affected rows and returns ErrSessionNotFound.
func RotateRefreshToken(s *xorm.Session, session *Session) (newRawToken string, err error) {
	newRawToken, newHash, err := generateHashedToken()
	if err != nil {
		return "", err
	}

	affected, err := s.Where("id = ? AND token_hash = ?", session.ID, session.TokenHash).
		Cols("token_hash").
		Update(&Session{TokenHash: newHash})
	if err != nil {
		return "", err
	}
	if affected == 0 {
		// Another request already rotated this token — reject the replay.
		return "", &ErrSessionNotFound{}
	}
	return newRawToken, nil
}

// DeleteAllUserSessions removes all sessions for a user (e.g., on password change).
func DeleteAllUserSessions(s *xorm.Session, userID int64) error {
	_, err := s.Where("user_id = ?", userID).Delete(&Session{})
	return err
}

// RegisterSessionCleanupCron registers a cron to delete sessions whose refresh
// tokens have expired. Uses is_long_session to pick the right cutoff so short
// sessions don't linger for the full long TTL. Runs hourly.
func RegisterSessionCleanupCron() {
	const logPrefix = "[Session Cleanup Cron] "

	err := cron.Schedule("0 * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		now := time.Now()
		shortMaxAge := time.Duration(config.ServiceJWTTTL.GetInt64()) * time.Second
		longMaxAge := time.Duration(config.ServiceJWTTTLLong.GetInt64()) * time.Second

		// Delete short sessions older than ServiceJWTTTL
		// and long sessions older than ServiceJWTTTLLong
		deleted, err := s.
			Where("(is_long_session = ? AND last_active < ?) OR (is_long_session = ? AND last_active < ?)",
				false, now.Add(-shortMaxAge),
				true, now.Add(-longMaxAge)).
			Delete(&Session{})
		if err != nil {
			log.Errorf(logPrefix+"Error removing stale sessions: %s", err)
			return
		}
		if deleted > 0 {
			log.Debugf(logPrefix+"Deleted %d stale sessions", deleted)
		}

		if err := s.Commit(); err != nil {
			log.Errorf(logPrefix+"Could not commit: %s", err)
		}
	})
	if err != nil {
		log.Fatalf("Could not register session cleanup cron: %s", err)
	}
}
