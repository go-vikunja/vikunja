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

package avatar

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar/botmarble"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/ldap"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/openid"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/user"

	"github.com/gabriel-vasile/mimetype"
	"xorm.io/xorm"
)

var ErrNotAnImage = errors.New("uploaded file is no image")

// Provider defines the avatar provider interface
type Provider interface {
	// GetAvatar is the method used to get an actual avatar for a user
	GetAvatar(ctx context.Context, user *user.User, size int64) (avatar []byte, mimeType string, err error)
	// AsDataURI returns a base64-encoded string representation of the avatar suitable for inline use
	AsDataURI(ctx context.Context, user *user.User, size int64) (inlineData string, err error)
	// FlushCache removes cached avatar data for the user
	FlushCache(u *user.User) error
}

// FlushAllCaches removes cached avatars for the given user for all providers
func FlushAllCaches(u *user.User) {
	providers := []Provider{
		&upload.Provider{},
		&gravatar.Provider{},
		&initials.Provider{},
		&ldap.Provider{},
		&openid.Provider{},
		&marble.Provider{},
		&botmarble.Provider{},
		&empty.Provider{},
	}
	for _, p := range providers {
		if err := p.FlushCache(u); err != nil {
			log.Errorf("Error flushing avatar cache: %v", err)
		}
	}
}

// GetAvatarForUsername resolves and renders the avatar for a username. It is the
// shared core behind both the v1 and v2 avatar endpoints: it looks up the user,
// tolerates an unknown/disabled user (returning the default placeholder rather
// than an error, since avatars are loaded via <img> tags), picks the right
// provider (empty for unknown users, botmarble for bots, otherwise the user's
// configured provider) and normalizes the size.
func GetAvatarForUsername(ctx context.Context, username string, size int64) (data []byte, mime string, err error) {
	u, found, err := avatarUser(username)
	if err != nil {
		return nil, "", err
	}

	provider := GetProvider(u)
	if !found {
		// Unknown user: serve the default placeholder.
		provider = &empty.Provider{}
	}
	if found && u.IsBot() {
		provider = &botmarble.Provider{}
	}

	data, mime, err = provider.GetAvatar(ctx, u, normalizeAvatarSize(size))
	if err != nil {
		if imageutils.IsErrDecodeBusy(err) {
			log.Warningf("Error getting avatar for user %d: %v", u.ID, err)
		} else {
			log.Errorf("Error getting avatar for user %d: %v", u.ID, err)
		}
		return nil, "", err
	}

	return data, mime, nil
}

// Providers cache per size without eviction, so only a fixed set of sizes may reach them.
var avatarSizes = []int64{16, 32, 48, 64, 128, 256, 512, 1024}

const defaultAvatarSize = 250

func normalizeAvatarSize(size int64) int64 {
	maxSize := config.ServiceMaxAvatarSize.GetInt64()
	if size <= 0 {
		size = defaultAvatarSize
	}
	for _, allowed := range avatarSizes {
		if allowed >= size {
			return min(allowed, maxSize)
		}
	}
	return maxSize
}

// The session closes before rendering: providers may queue for decode capacity or call out over HTTP.
func avatarUser(username string) (u *user.User, found bool, err error) {
	s := db.NewSession()
	defer s.Close()

	u, err = user.GetUserWithEmail(s, &user.User{Username: username})
	if err != nil && !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		log.Errorf("Error getting user for avatar: %v", err)
		return nil, false, err
	}
	return u, err == nil || user.IsErrUserStatusError(err), nil
}

// GetProvider returns the appropriate avatar provider for a user
func GetProvider(u *user.User) Provider {
	provider := u.AvatarProvider
	if provider == "" {
		provider = "empty"
	}

	switch provider {
	case "gravatar":
		return &gravatar.Provider{}
	case "initials":
		return &initials.Provider{}
	case "upload":
		return &upload.Provider{}
	case "marble":
		return &marble.Provider{}
	case "ldap":
		return &ldap.Provider{}
	case "openid":
		return &openid.Provider{}
	default:
		return &empty.Provider{}
	}
}

func PrepareUpload(ctx context.Context, src io.Reader) (*upload.EncodedAvatar, error) {
	raw, err := io.ReadAll(src)
	if err != nil {
		return nil, err
	}
	if !strings.HasPrefix(mimetype.Detect(raw).String(), "image") {
		return nil, ErrNotAnImage
	}

	a, err := upload.PrepareAvatar(ctx, raw)
	if errors.Is(err, imageutils.ErrUndecodable) {
		return nil, ErrNotAnImage
	}
	return a, err
}

func SaveUploaded(s *xorm.Session, u *user.User, a *upload.EncodedAvatar) error {
	if err := upload.SaveAvatar(s, u, "upload", a); err != nil {
		return err
	}

	events.OnCommit(s, func() error {
		FlushAllCaches(u)
		return nil
	})

	return nil
}

// Own session: decoding may queue and must not hold the login transaction.
func SyncFromProvider(ctx context.Context, userID int64, provider string, raw []byte) error {
	prepared, err := upload.PrepareAvatar(ctx, raw)
	if err != nil {
		return fmt.Errorf("error decoding avatar: %w", err)
	}

	s := db.NewSession()
	defer s.Close()
	defer events.CleanupPending(s)

	u, err := user.GetUserByID(s, userID)
	if err != nil {
		_ = s.Rollback()
		return fmt.Errorf("error loading user: %w", err)
	}

	if err := upload.SaveAvatar(s, u, provider, prepared); err != nil {
		_ = s.Rollback()
		return fmt.Errorf("error storing avatar: %w", err)
	}

	if err := s.Commit(); err != nil {
		return fmt.Errorf("error committing avatar: %w", err)
	}
	events.DispatchPending(ctx, s)

	FlushAllCaches(u)

	return nil
}
