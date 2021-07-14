// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package user

import (
	"time"

	"code.vikunja.io/api/pkg/cron"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
	"xorm.io/xorm"
)

// TokenKind represents a user token kind
type TokenKind int

const (
	TokenUnknown TokenKind = iota
	TokenPasswordReset
	TokenEmailConfirm

	tokenSize = 64
)

// Token is a token a user can use to do things like verify their email or resetting their password
type Token struct {
	ID      int64     `xorm:"bigint autoincr not null unique pk"`
	UserID  int64     `xorm:"not null"`
	Token   string    `xorm:"TEXT not null index"`
	Kind    TokenKind `xorm:"not null"`
	Created time.Time `xorm:"created not null"`
}

// TableName returns the real table name for user tokens
func (t *Token) TableName() string {
	return "user_tokens"
}

func generateNewToken(s *xorm.Session, u *User, kind TokenKind) (token *Token, err error) {
	token = &Token{
		UserID: u.ID,
		Kind:   kind,
		Token:  utils.MakeRandomString(tokenSize),
	}

	_, err = s.Insert(token)
	return
}

func getToken(s *xorm.Session, token string, kind TokenKind) (t *Token, err error) {
	t = &Token{}
	has, err := s.Where("kind = ? AND token = ?", kind, token).
		Get(t)
	if err != nil || !has {
		return nil, err
	}

	return
}

func removeTokens(s *xorm.Session, u *User, kind TokenKind) (err error) {
	_, err = s.Where("user_id = ? AND kind = ?", u.ID, kind).
		Delete(&Token{})
	return
}

// RegisterTokenCleanupCron registers a cron function to clean up all password reset tokens older than 24 hours
func RegisterTokenCleanupCron() {
	const logPrefix = "[User Token Cleanup Cron] "

	err := cron.Schedule("0 * * * *", func() {
		s := db.NewSession()
		defer s.Close()

		deleted, err := s.
			Where("created > ? AND kind = ?", time.Now().Add(time.Hour*24*-1), TokenPasswordReset).
			Delete(&Token{})
		if err != nil {
			log.Errorf(logPrefix+"Error removing old password reset tokens: %s", err)
			return
		}
		if deleted > 0 {
			log.Debugf(logPrefix+"Deleted %d old password reset tokens", deleted)
		}
	})
	if err != nil {
		log.Fatalf("Could not register token cleanup cron: %s", err)
	}
}
