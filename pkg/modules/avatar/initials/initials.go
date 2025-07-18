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

package initials

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"strconv"
	"strings"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"code.vikunja.io/api/pkg/user"
	"github.com/disintegration/imaging"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
)

// Provider represents the provider implementation of the initials provider
type Provider struct {
}

// FlushCache removes cached initials avatars for a user
func (p *Provider) FlushCache(u *user.User) error {
	if err := keyvalue.Del(getCacheKey("full", u.ID)); err != nil {
		return err
	}
	return keyvalue.DelPrefix(getCacheKey("resized", u.ID))
}

var (
	avatarBgColors = []*color.RGBA{
		{R: 69, G: 189, B: 243, A: 255},
		{R: 224, G: 143, B: 112, A: 255},
		{R: 77, G: 182, B: 172, A: 255},
		{R: 149, G: 117, B: 205, A: 255},
		{R: 176, G: 133, B: 94, A: 255},
		{R: 240, G: 98, B: 146, A: 255},
		{R: 163, G: 211, B: 108, A: 255},
		{R: 121, G: 134, B: 203, A: 255},
		{R: 241, G: 185, B: 29, A: 255},
	}
)

const (
	dpi         = 72
	defaultSize = 1024
)

func drawImage(text rune, bg *color.RGBA) (img *image.RGBA64, err error) {

	size := defaultSize
	fontSize := float64(size) * 0.8

	// Inspired by https://github.com/holys/initials-avatar

	// Get the font
	f, err := truetype.Parse(goregular.TTF)
	if err != nil {
		return img, err
	}

	// Build the image background
	img = image.NewRGBA64(image.Rect(0, 0, size, size))
	draw.Draw(img, img.Bounds(), &image.Uniform{C: bg}, image.Point{}, draw.Src)

	// Add the text
	drawer := &font.Drawer{
		Dst: img,
		Src: image.White,
		Face: truetype.NewFace(f, &truetype.Options{
			Size:    fontSize,
			DPI:     dpi,
			Hinting: font.HintingNone,
		}),
	}

	// Font Index
	fi := f.Index(text)

	// Glyph example: http://www.freetype.org/freetype2/docs/tutorial/metrics.png
	var gbuf truetype.GlyphBuf
	fsize := fixed.Int26_6(fontSize * dpi * (64.0 / 72.0))
	err = gbuf.Load(f, fsize, fi, font.HintingFull)
	if err != nil {
		drawer.DrawString("")
		return img, err
	}

	// Center
	dY := (size - int(gbuf.Bounds.Max.Y-gbuf.Bounds.Min.Y)>>6) / 2
	dX := (size - int(gbuf.Bounds.Max.X-gbuf.Bounds.Min.X)>>6) / 2
	y := int(gbuf.Bounds.Max.Y>>6) + dY
	x := 0 - int(gbuf.Bounds.Min.X>>6) + dX

	drawer.Dot = fixed.Point26_6{
		X: fixed.I(x),
		Y: fixed.I(y),
	}
	drawer.DrawString(string(text))

	return img, err
}

func getCacheKey(prefix string, keys ...int64) string {
	result := "avatar_initials_" + prefix
	for i, key := range keys {
		result += strconv.Itoa(int(key))
		if i < len(keys) {
			result += "_"
		}
	}
	return result
}

func getAvatarForUser(u *user.User) (fullSizeAvatar *image.RGBA64, err error) {
	cacheKey := getCacheKey("full", u.ID)

	result, err := keyvalue.Remember(cacheKey, func() (any, error) {
		log.Debugf("Initials avatar for user %d not cached, creating...", u.ID)
		avatarText := u.Name
		if avatarText == "" {
			avatarText = u.Username
		}
		firstRune := []rune(strings.ToUpper(avatarText))[0]
		bg := avatarBgColors[int(u.ID)%len(avatarBgColors)] // Random color based on the user id

		res, err := drawImage(firstRune, bg)
		if err != nil {
			return nil, err
		}

		return *res, nil
	})
	if err != nil {
		return nil, err
	}

	aa := result.(image.RGBA64)

	return &aa, nil
}

// CachedAvatar represents a cached avatar with its content and mime type
type CachedAvatar struct {
	Content  []byte
	MimeType string
}

// GetAvatar returns an initials avatar for a user
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {
	cacheKey := getCacheKey("resized", u.ID, size)

	result, err := keyvalue.Remember(cacheKey, func() (any, error) {
		log.Debugf("Initials avatar for user %d and size %d not cached, creating...", u.ID, size)
		fullAvatar, err := getAvatarForUser(u)
		if err != nil {
			return nil, err
		}

		img := imaging.Resize(fullAvatar, int(size), int(size), imaging.Lanczos)
		buf := &bytes.Buffer{}
		err = png.Encode(buf, img)
		if err != nil {
			return nil, err
		}
		avatar := buf.Bytes()
		mimeType := "image/png"

		cachedAvatar := CachedAvatar{
			Content:  avatar,
			MimeType: mimeType,
		}

		return cachedAvatar, nil
	})
	if err != nil {
		return nil, "", err
	}

	cachedAvatar := result.(CachedAvatar)
	return cachedAvatar.Content, cachedAvatar.MimeType, nil
}
