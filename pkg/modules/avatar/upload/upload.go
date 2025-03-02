// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
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

package upload

import (
	"bytes"
	"image"
	"image/png"
	"io"
	"strconv"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"

	"github.com/disintegration/imaging"
)

// Provider represents the upload avatar provider
type Provider struct {
}

// CachedAvatar represents a cached avatar with its content and mime type
type CachedAvatar struct {
	Content  []byte
	MimeType string
}

// GetAvatar returns an uploaded user avatar
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {

	cacheKey := "avatar_upload_" + strconv.Itoa(int(u.ID))

	var cached map[int64]*CachedAvatar
	exists, err := keyvalue.GetWithValue(cacheKey, &cached)
	if err != nil {
		return nil, "", err
	}

	if !exists {
		// Nothing ever cached for this user so we need to create the size map to avoid panics
		cached = make(map[int64]*CachedAvatar)
	} else {
		a := cached
		if a != nil && a[size] != nil {
			log.Debugf("Serving uploaded avatar for user %d and size %d from cache.", u.ID, size)
			return a[size].Content, a[size].MimeType, nil
		}
		// This means we have a map for the user, but nothing in it.
		if a == nil {
			cached = make(map[int64]*CachedAvatar)
		}
	}

	log.Debugf("Uploaded avatar for user %d and size %d not cached, resizing and caching.", u.ID, size)

	// If we get this far, the avatar is either not cached at all or not in this size
	f := &files.File{ID: u.AvatarFileID}
	if err := f.LoadFileByID(); err != nil {
		return nil, "", err
	}

	if err := f.LoadFileMetaByID(); err != nil {
		return nil, "", err
	}

	img, _, err := image.Decode(f.File)
	if err != nil {
		return nil, "", err
	}
	resizedImg := imaging.Resize(img, 0, int(size), imaging.Lanczos)
	buf := &bytes.Buffer{}
	if err := png.Encode(buf, resizedImg); err != nil {
		return nil, "", err
	}

	avatar, err = io.ReadAll(buf)
	if err != nil {
		return nil, "", err
	}

	// Always use image/png for resized avatars since we're encoding with png
	mimeType = "image/png"
	cached[size] = &CachedAvatar{
		Content:  avatar,
		MimeType: mimeType,
	}

	err = keyvalue.Put(cacheKey, cached)
	return avatar, mimeType, err
}

// InvalidateCache invalidates the avatar cache for a user
func InvalidateCache(u *user.User) {
	if err := keyvalue.Del("avatar_upload_" + strconv.Itoa(int(u.ID))); err != nil {
		log.Errorf("Could not invalidate upload avatar cache for user %d, error was %s", u.ID, err)
	}
}
