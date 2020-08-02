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

package initials

import (
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/user"
	"github.com/disintegration/imaging"
	"strconv"
	"strings"
	"sync"

	"bytes"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/font"
	"golang.org/x/image/font/gofont/goregular"
	"golang.org/x/image/math/fixed"
	"image"
	"image/color"
	"image/draw"
	"image/png"
)

// Provider represents the provider implementation of the initials provider
type Provider struct {
}

var (
	avatarBgColors = []*color.RGBA{
		{69, 189, 243, 255},
		{224, 143, 112, 255},
		{77, 182, 172, 255},
		{149, 117, 205, 255},
		{176, 133, 94, 255},
		{240, 98, 146, 255},
		{163, 211, 108, 255},
		{121, 134, 203, 255},
		{241, 185, 29, 255},
	}

	// Contain the created avatars with a size of defaultSize
	cache            = map[int64]*image.RGBA64{}
	cacheLock        = sync.Mutex{}
	cacheResized     = map[string][]byte{}
	cacheResizedLock = sync.Mutex{}
)

func init() {
	cache = make(map[int64]*image.RGBA64)
	cacheResized = make(map[string][]byte)
}

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

func getAvatarForUser(u *user.User) (fullSizeAvatar *image.RGBA64, err error) {
	var cached bool
	fullSizeAvatar, cached = cache[u.ID]
	if !cached {
		log.Debugf("Initials avatar for user %d not cached, creating...", u.ID)
		firstRune := []rune(strings.ToUpper(u.Username))[0]
		bg := avatarBgColors[int(u.ID)%len(avatarBgColors)] // Random color based on the user id

		fullSizeAvatar, err = drawImage(firstRune, bg)
		if err != nil {
			return nil, err
		}
		cacheLock.Lock()
		cache[u.ID] = fullSizeAvatar
		cacheLock.Unlock()
	}

	return fullSizeAvatar, err
}

// GetAvatar returns an initials avatar for a user
func (p *Provider) GetAvatar(u *user.User, size int64) (avatar []byte, mimeType string, err error) {

	var cached bool
	cacheKey := strconv.Itoa(int(u.ID)) + "_" + strconv.Itoa(int(size))
	avatar, cached = cacheResized[cacheKey]
	if !cached {
		log.Debugf("Initials avatar for user %d and size %d not cached, creating...", u.ID, size)
		fullAvatar, err := getAvatarForUser(u)
		if err != nil {
			return nil, "", err
		}

		img := imaging.Resize(fullAvatar, int(size), int(size), imaging.Lanczos)
		buf := &bytes.Buffer{}
		err = png.Encode(buf, img)
		if err != nil {
			return nil, "", err
		}
		avatar = buf.Bytes()
		cacheResizedLock.Lock()
		cacheResized[cacheKey] = avatar
		cacheResizedLock.Unlock()
	} else {
		log.Debugf("Serving initials avatar for user %d and size %d from cache", u.ID, size)
	}

	return avatar, "image/png", err
}
