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

package utils

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/imageutils"

	"github.com/disintegration/imaging"
)

const avatarMaxSide = 1024

// CropAvatarTo1x1 crops the avatar image to a 1:1 aspect ratio, centered on the image
func CropAvatarTo1x1(ctx context.Context, imageData []byte) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, errors.New("empty image data")
	}

	var cropped []byte
	err := imageutils.Decode(ctx, imageutils.BytesOpener(imageData), func(img image.Image, format string) error {
		var err error
		cropped, err = cropTo1x1(imageData, img, format)
		return err
	})
	if err != nil {
		return nil, err
	}
	return cropped, nil
}

func cropTo1x1(imageData []byte, img image.Image, format string) ([]byte, error) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == height {
		return imageData, nil
	}

	// Scaling here avoids a full-resolution copy outside the decode budget.
	side := min(width, height, avatarMaxSide)
	square := imaging.Fill(img, side, side, imaging.Center, imaging.Lanczos)

	var buf bytes.Buffer
	var err error
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, square, nil)
	default:
		err = png.Encode(&buf, square)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to encode cropped image: %w", err)
	}

	return buf.Bytes(), nil
}

// DownloadImage downloads an image from a URL and returns the image data
func DownloadImage(ctx context.Context, url string) ([]byte, error) {
	// 3 seconds is enough for downloading an avatar
	ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := NewSSRFSafeHTTPClient().Do(req) //nolint:gosec // SSRF protection is handled by the SSRF-safe client
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	// Bound the remote stream before buffering it.
	// #nosec G115 - the configured size is far below int64 max.
	maxBytes := int64(config.GetMaxFileSizeInMBytes()) * 1024 * 1024
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, fmt.Errorf("downloaded image exceeds the maximum file size of %d bytes", maxBytes)
	}

	return data, nil
}
