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
	"context"
	"encoding/base64"
	"encoding/gob"
	"fmt"
	"image"
	"image/png"
	"io"
	"math"
	"strconv"

	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/imageutils"
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

const maxAvatarSide = 1024

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
	// size is converted to int below for imaging.Resize; reject values that
	// would be negative or overflow on platforms where int is 32 bits.
	if size <= 0 || size > math.MaxInt32 {
		return nil, "", fmt.Errorf("invalid avatar size %d", size)
	}

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

		// Validation and decoding both need the non-seekable file.
		data, err := io.ReadAll(f.File)
		if err != nil {
			return CachedAvatar{}, err
		}

		buf := &bytes.Buffer{}
		err = imageutils.Decode(context.Background(), imageutils.BytesOpener(data), func(img image.Image, _ string) error {
			return png.Encode(buf, imaging.Fit(img, int(size), int(size), imaging.Lanczos))
		})
		if err != nil {
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

type EncodedAvatar struct {
	png []byte
}

// May queue for decode capacity; call outside a transaction.
func PrepareAvatar(ctx context.Context, raw []byte) (*EncodedAvatar, error) {
	buf := &bytes.Buffer{}
	err := imageutils.Decode(ctx, imageutils.BytesOpener(raw), func(img image.Image, _ string) error {
		return png.Encode(buf, imaging.Fit(centeredSquare(img), maxAvatarSide, maxAvatarSide, imaging.Lanczos))
	})
	if err != nil {
		return nil, err
	}
	return &EncodedAvatar{png: buf.Bytes()}, nil
}

// Zero-copy where possible: a full-resolution copy would escape the decode budget.
func centeredSquare(img image.Image) image.Image {
	b := img.Bounds()
	side := min(b.Dx(), b.Dy())
	sub, ok := img.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		return imaging.CropCenter(img, side, side)
	}
	topLeft := b.Min.Add(image.Pt((b.Dx()-side)/2, (b.Dy()-side)/2))
	return sub.SubImage(image.Rectangle{Min: topLeft, Max: topLeft.Add(image.Pt(side, side))})
}

// SaveAvatar removes the replaced avatar blob in events.DispatchPending(s).
func SaveAvatar(s *xorm.Session, u *user.User, provider string, a *EncodedAvatar) (err error) {
	oldAvatarFileID := u.AvatarFileID
	ownsOldAvatar := u.OwnsAvatarFile()
	oldProvider := u.AvatarProvider
	defer func() {
		if err != nil {
			u.AvatarFileID = oldAvatarFileID
			u.AvatarProvider = oldProvider
		}
	}()

	err = (&Provider{}).FlushCache(u)
	if err != nil {
		log.Errorf("Could not invalidate upload avatar cache for user %d, error was %s", u.ID, err)
	}

	f, err := files.CreateWithMimeAndSession(s, bytes.NewReader(a.png), "avatar.png", uint64(len(a.png)), u, "image/png", true)
	if err != nil {
		return err
	}

	u.AvatarFileID = f.ID
	u.AvatarProvider = provider

	_, err = user.UpdateUser(s, u, false)
	if err != nil {
		return err
	}

	// Log only: the new avatar is already persisted.
	if ownsOldAvatar {
		old := &files.File{ID: oldAvatarFileID}
		if delErr := old.DeleteOnCommit(s); delErr != nil && !files.IsErrFileDoesNotExist(delErr) {
			log.Errorf("Could not delete old avatar file %d of user %d: %s", oldAvatarFileID, u.ID, delErr)
		}
	}

	return nil
}
