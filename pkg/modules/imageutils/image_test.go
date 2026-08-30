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

package imageutils

import (
	"bytes"
	"encoding/binary"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfig(t *testing.T) {
	for _, tc := range []struct {
		name          string
		width, height int
		wantErr       bool
	}{
		{"normal image", 1920, 1080, false},
		{"exactly at the cap", 5000, 10000, false},
		{"skinny image stays bounded", 20000, 10, false},
		{"8000 by 8000 exceeds the cap", 8000, 8000, true},
		{"zero width", 0, 100, true},
		{"negative height", 100, -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateConfig(image.Config{Width: tc.width, Height: tc.height, ColorModel: color.NRGBAModel})
			if tc.wantErr {
				require.Error(t, err)
				assert.True(t, IsErrImageTooLarge(err))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

// The forged IHDR models a tiny decompression-bomb payload.
func pngWithClaimedDimensions(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	buf := &bytes.Buffer{}
	require.NoError(t, png.Encode(buf, img))

	data := buf.Bytes()
	binary.BigEndian.PutUint32(data[16:20], uint32(width))  // #nosec G115 - test constant
	binary.BigEndian.PutUint32(data[20:24], uint32(height)) // #nosec G115 - test constant
	crc := crc32.ChecksumIEEE(data[12:29])
	binary.BigEndian.PutUint32(data[29:33], crc)
	return data
}

func TestValidateReader(t *testing.T) {
	t.Run("normal png passes", func(t *testing.T) {
		cfg, err := ValidateReader(bytes.NewReader(pngWithClaimedDimensions(t, 16, 16)))
		require.NoError(t, err)
		assert.Equal(t, 16, cfg.Width)
		assert.Equal(t, 16, cfg.Height)
	})

	t.Run("decompression bomb is rejected", func(t *testing.T) {
		_, err := ValidateReader(bytes.NewReader(pngWithClaimedDimensions(t, 8000, 8000)))
		require.Error(t, err)
		assert.True(t, IsErrImageTooLarge(err))
	})
}
