// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package upload

import (
	"bytes"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	e "code.vikunja.io/api/pkg/modules/keyvalue/error"
	"code.vikunja.io/api/pkg/user"
	"github.com/disintegration/imaging"
	"image"
	"image/png"
	"io/ioutil"
	"strconv"
)

// Provider represents the upload avatar provider
type Provider struct {
}

// GetAvatar returns an uploaded user avatar
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {

	cacheKey := "avatar_upload_" + strconv.Itoa(int(u.ID))

	ai, err := keyvalue.Get(cacheKey)
	if err != nil && !e.IsErrValueNotFoundForKey(err) {
		return nil, "", err
	}

	var cached map[int64][]byte

	if ai != nil {
		cached = ai.(map[int64][]byte)
	}

	if err != nil && e.IsErrValueNotFoundForKey(err) {
		// Nothing ever cached for this user so we need to create the size map to avoid panics
		cached = make(map[int64][]byte)
	} else {
		a := ai.(map[int64][]byte)
		if a != nil && a[size] != nil {
			log.Debugf("Serving uploaded avatar for user %d and size %d from cache.", u.ID, size)
			return a[size], "", nil
		}
		// This means we have a map for the user, but nothing in it.
		if a == nil {
			cached = make(map[int64][]byte)
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

	avatar, err = ioutil.ReadAll(buf)
	if err != nil {
		return nil, "", err
	}
	cached[size] = avatar
	err = keyvalue.Put(cacheKey, cached)
	return avatar, f.Mime, err
}

// InvalidateCache invalidates the avatar cache for a user
func InvalidateCache(u *user.User) {
	if err := keyvalue.Del("avatar_upload_" + strconv.Itoa(int(u.ID))); err != nil {
		log.Errorf("Could not invalidate upload avatar cache for user %d, error was %s", u.ID, err)
	}
}
