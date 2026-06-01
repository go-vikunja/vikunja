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
	"errors"
	"io"
	"strings"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/avatar/botmarble"
	"code.vikunja.io/api/pkg/modules/avatar/empty"
	"code.vikunja.io/api/pkg/modules/avatar/gravatar"
	"code.vikunja.io/api/pkg/modules/avatar/initials"
	"code.vikunja.io/api/pkg/modules/avatar/ldap"
	"code.vikunja.io/api/pkg/modules/avatar/marble"
	"code.vikunja.io/api/pkg/modules/avatar/openid"
	"code.vikunja.io/api/pkg/modules/avatar/upload"
	"code.vikunja.io/api/pkg/user"

	"github.com/gabriel-vasile/mimetype"
	"xorm.io/xorm"
)

// ErrNotAnImage is returned by StoreUploadedAvatar when the uploaded file is not an image.
var ErrNotAnImage = errors.New("uploaded file is no image")

// Provider defines the avatar provider interface
type Provider interface {
	// GetAvatar is the method used to get an actual avatar for a user
	GetAvatar(user *user.User, size int64) (avatar []byte, mimeType string, err error)
	// AsDataURI returns a base64-encoded string representation of the avatar suitable for inline use
	AsDataURI(user *user.User, size int64) (inlineData string, err error)
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
// configured provider) and clamps the size to the server's configured maximum.
func GetAvatarForUsername(s *xorm.Session, username string, size int64) (data []byte, mime string, err error) {
	u, err := user.GetUserWithEmail(s, &user.User{Username: username})
	if err != nil && !user.IsErrUserDoesNotExist(err) && !user.IsErrUserStatusError(err) {
		log.Errorf("Error getting user for avatar: %v", err)
		return nil, "", err
	}

	found := err == nil || user.IsErrUserStatusError(err)

	provider := GetProvider(u)
	if !found {
		// Unknown user: serve the default placeholder.
		provider = &empty.Provider{}
	}
	if found && u.IsBot() {
		provider = &botmarble.Provider{}
	}

	if size > config.ServiceMaxAvatarSize.GetInt64() {
		size = config.ServiceMaxAvatarSize.GetInt64()
	}

	data, mime, err = provider.GetAvatar(u, size)
	if err != nil {
		log.Errorf("Error getting avatar for user %d: %v", u.ID, err)
		return nil, "", err
	}

	return data, mime, nil
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

// StoreUploadedAvatar validates that src is an image, switches the user's avatar
// provider to "upload", stores the image as the user's avatar and flushes all
// cached avatars for the user. It returns ErrNotAnImage if src is not an image.
func StoreUploadedAvatar(s *xorm.Session, u *user.User, src io.ReadSeeker) error {
	mime, err := mimetype.DetectReader(src)
	if err != nil {
		return err
	}
	if !strings.HasPrefix(mime.String(), "image") {
		return ErrNotAnImage
	}
	if _, err := src.Seek(0, io.SeekStart); err != nil {
		return err
	}

	u.AvatarProvider = "upload"
	if err := upload.StoreAvatarFile(s, u, src); err != nil {
		return err
	}

	FlushAllCaches(u)

	return nil
}
