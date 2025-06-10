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
	TokenAccountDeletion
	TokenCaldavAuth

	tokenSize = 64
)

// Token is a token a user can use to do things like verify their email or resetting their password
type Token struct {
	ID             int64     `xorm:"bigint autoincr not null unique pk" json:"id"`
	UserID         int64     `xorm:"not null" json:"-"`
	Token          string    `xorm:"varchar(450) not null index" json:"-"`
	ClearTextToken string    `xorm:"-" json:"token"`
	Kind           TokenKind `xorm:"not null" json:"-"`
	Created        time.Time `xorm:"created not null" json:"created"`
}

// TableName returns the real table name for user tokens
func (t *Token) TableName() string {
	return "user_tokens"
}

func genToken(u *User, kind TokenKind) (*Token, error) {
	tokenStr, err := utils.CryptoRandomString(tokenSize)
	if err != nil {
		return nil, err
	}
	return &Token{
		UserID: u.ID,
		Kind:   kind,
		Token:  tokenStr,
	}, nil
}

func generateToken(s *xorm.Session, u *User, kind TokenKind) (token *Token, err error) {
	token, err = genToken(u, kind)
	if err != nil {
		return nil, err
	}

	_, err = s.Insert(token)
	return
}

func generateHashedToken(s *xorm.Session, u *User, kind TokenKind) (token *Token, err error) {
	token, err = genToken(u, kind)
	if err != nil {
		return nil, err
	}
	token.ClearTextToken = token.Token
	token.Token, err = HashPassword(token.ClearTextToken)
	if err != nil {
		return nil, err
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

func getTokensForKind(s *xorm.Session, u *User, kind TokenKind) (tokens []*Token, err error) {
	tokens = []*Token{}

	err = s.Where("kind = ? AND user_id = ?", kind, u.ID).
		Find(&tokens)
	return
}

func removeTokens(s *xorm.Session, u *User, kind TokenKind) (err error) {
	_, err = s.Where("user_id = ? AND kind = ?", u.ID, kind).
		Delete(&Token{})
	return
}

func removeTokenByID(s *xorm.Session, u *User, kind TokenKind, id int64) (err error) {
	_, err = s.Where("id = ? AND user_id = ? AND kind = ?", id, u.ID, kind).
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
			Where("created > ? AND (kind = ? OR kind = ?)", time.Now().Add(time.Hour*24*-1), TokenPasswordReset, TokenAccountDeletion).
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
