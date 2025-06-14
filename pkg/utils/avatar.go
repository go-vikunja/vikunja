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
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"time"
)

// CropAvatarTo1x1 crops the avatar image to a 1:1 aspect ratio, centered on the image
func CropAvatarTo1x1(imageData []byte) ([]byte, error) {
	if len(imageData) == 0 {
		return nil, errors.New("empty image data")
	}

	// Decode the image
	img, format, err := image.Decode(bytes.NewReader(imageData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Get image dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If already square, return original
	if width == height {
		return imageData, nil
	}

	// Determine the crop size (use the smaller dimension)
	size := width
	if height < width {
		size = height
	}

	// Calculate crop coordinates to center the image
	x0 := (width - size) / 2
	y0 := (height - size) / 2
	x1 := x0 + size
	y1 := y0 + size

	// Create the cropping rectangle
	cropRect := image.Rect(x0, y0, x1, y1)

	// Create a new RGBA image
	croppedImg := image.NewRGBA(image.Rect(0, 0, size, size))

	// Copy the cropped portion
	draw.Draw(croppedImg, croppedImg.Bounds(), img, cropRect.Min, draw.Src)

	// Encode the result
	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, croppedImg, nil)
	default:
		// Default to PNG if format is unknown
		err = png.Encode(&buf, croppedImg)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to encode cropped image: %w", err)
	}

	return buf.Bytes(), nil
}

// DownloadImage downloads an image from a URL and returns the image data
func DownloadImage(url string) ([]byte, error) {
	// 3 seconds is enough for downloading an avatar
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}
