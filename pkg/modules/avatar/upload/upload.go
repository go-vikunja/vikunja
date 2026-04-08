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

package upload

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"image"
	"image/png"
	"io"
	"strconv"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/disintegration/imaging"
	"xorm.io/xorm"
)

func init() {
	gob.Register(CachedAvatar{})
}

// Provider represents the upload avatar provider
type Provider struct {
}

const CacheKeyPrefix = "avatar_upload_"

// FlushCache removes cached avatars for a user
func (p *Provider) FlushCache(u *user.User) error {
	return keyvalue.DelPrefix(CacheKeyPrefix + strconv.Itoa(int(u.ID)) + "_")
}

// CachedAvatar represents a cached avatar with its content and mime type
type CachedAvatar struct {
	Content  []byte
	MimeType string
}

// GetAvatar returns an uploaded user avatar
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {
	cacheKey := CacheKeyPrefix + strconv.Itoa(int(u.ID)) + "_" + strconv.FormatInt(size, 10)

	cachedAvatar, err := keyvalue.RememberValue(cacheKey, func() (CachedAvatar, error) {
		log.Debugf("Uploaded avatar for user %d and size %d not cached, resizing and caching.", u.ID, size)

		if u.AvatarFileID == 0 {
			return CachedAvatar{}, fmt.Errorf("user %d has no avatar file", u.ID)
		}

		f := &files.File{ID: u.AvatarFileID}
		if err := f.LoadFileByID(); err != nil {
			return CachedAvatar{}, err
		}

		if err := f.LoadFileMetaByID(); err != nil {
			return CachedAvatar{}, err
		}

		img, _, err := image.Decode(f.File)
		if err != nil {
			return CachedAvatar{}, err
		}
		resizedImg := imaging.Resize(img, 0, int(size), imaging.Lanczos)
		buf := &bytes.Buffer{}
		if err := png.Encode(buf, resizedImg); err != nil {
			return CachedAvatar{}, err
		}

		avatarBytes, err := io.ReadAll(buf)
		if err != nil {
			return CachedAvatar{}, err
		}

		return CachedAvatar{
			Content:  avatarBytes,
			MimeType: "image/png",
		}, nil
	})
	if err != nil {
		return nil, "", err
	}

	return cachedAvatar.Content, cachedAvatar.MimeType, nil
}

// AsDataURI returns a base64 encoded data URI for the uploaded avatar
func (p *Provider) AsDataURI(u *user.User, size int64) (string, error) {
	avatarData, mimeType, err := p.GetAvatar(u, size)
	if err != nil {
		return "", err
	}

	// Encode the avatar data as base64 and create a data URI
	base64Data := base64.StdEncoding.EncodeToString(avatarData)
	dataURI := fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)

	return dataURI, nil
}

func StoreAvatarFile(s *xorm.Session, u *user.User, src io.Reader) (err error) {

	// Remove the old file if one exists
	if u.AvatarFileID != 0 {
		f := &files.File{ID: u.AvatarFileID}
		if err := f.Delete(s); err != nil {
			if !files.IsErrFileDoesNotExist(err) {
				return err
			}
		}
		u.AvatarFileID = 0
	}

	// Resize the new file to a max height of 1024
	img, _, err := image.Decode(src)
	if err != nil {
		return
	}
	resizedImg := imaging.Resize(img, 0, 1024, imaging.Lanczos)
	buf := &bytes.Buffer{}
	err = png.Encode(buf, resizedImg)
	if err != nil {
		return
	}

	err = (&Provider{}).FlushCache(u)
	if err != nil {
		log.Errorf("Could not invalidate upload avatar cache for user %d, error was %s", u.ID, err)
	}

	// Save the file
	f, err := files.CreateWithMimeAndSession(s, bytes.NewReader(buf.Bytes()), "avatar.png", uint64(buf.Len()), u, "image/png", true)
	if err != nil {
		return err
	}

	u.AvatarFileID = f.ID

	_, err = user.UpdateUser(s, u, false)
	return
}
