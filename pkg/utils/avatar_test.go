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
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/modules/imageutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pngBytesForAvatar(t *testing.T, width int, height ...int) []byte {
	t.Helper()
	h := 4
	if len(height) > 0 {
		h = height[0]
	}
	img := image.NewRGBA(image.Rect(0, 0, width, h))
	for x := 0; x < width; x++ {
		for y := 0; y < h; y++ {
			img.Set(x, y, color.RGBA{R: uint8(x), G: uint8(y), B: 42, A: 255})
		}
	}
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))
	return buf.Bytes()
}

// Covers identity-provider image bounds (GHSA-4vh2-39rq-rq8j).
func TestCropAvatarTo1x1(t *testing.T) {
	t.Run("normal non-square image is cropped centered", func(t *testing.T) {
		cropped, err := CropAvatarTo1x1(pngBytesForAvatar(t, 8, 4))
		require.NoError(t, err)

		img, _, err := image.Decode(bytes.NewReader(cropped))
		require.NoError(t, err)
		assert.Equal(t, 4, img.Bounds().Dx())
		assert.Equal(t, 4, img.Bounds().Dy())
	})

	t.Run("already square image is returned as-is", func(t *testing.T) {
		original := pngBytesForAvatar(t, 4, 4)
		cropped, err := CropAvatarTo1x1(original)
		require.NoError(t, err)
		assert.Equal(t, original, cropped)
	})

	t.Run("decompression bomb is rejected", func(t *testing.T) {
		data := pngBytesForAvatar(t, 4, 4)
		binary.BigEndian.PutUint32(data[16:20], uint32(8000))
		binary.BigEndian.PutUint32(data[20:24], uint32(8000))
		binary.BigEndian.PutUint32(data[29:33], crc32.ChecksumIEEE(data[12:29]))

		_, err := CropAvatarTo1x1(data)
		require.Error(t, err)
		assert.True(t, imageutils.IsErrImageTooLarge(err))
	})
}

func TestDownloadImageIsBounded(t *testing.T) {
	// httptest servers live on 127.0.0.1, which the SSRF-safe client refuses
	// unless non-routable destinations are explicitly allowed.
	config.OutgoingRequestsAllowNonRoutableIPs.Set("true")
	t.Cleanup(func() { config.OutgoingRequestsAllowNonRoutableIPs.Set("false") })

	maxBytes := 20 * 1024 * 1024

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(bytes.Repeat([]byte{0x42}, maxBytes+1024))
	}))
	t.Cleanup(srv.Close)

	_, err := DownloadImage(srv.URL)
	require.Error(t, err, "a response larger than the max file size must not be buffered in full")

	small := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, _ = w.Write(pngBytesForAvatar(t, 4, 4))
	}))
	t.Cleanup(small.Close)

	data, err := DownloadImage(small.URL)
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}
