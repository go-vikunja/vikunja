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

package models

// Spike only: what registering the AVIF decoder changes for uploads. Not meant to be merged.

import (
	"bytes"
	"encoding/binary"
	"image"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/modules/imageutils"

	"github.com/gen2brain/gav1d/avif"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// spikeClaimDimensions rewrites the ispe box, which is where DecodeConfig reads the dimensions from.
func spikeClaimDimensions(t *testing.T, encoded []byte, width, height uint32) []byte {
	t.Helper()
	at := bytes.Index(encoded, []byte("ispe"))
	require.Positive(t, at)
	patched := bytes.Clone(encoded)
	binary.BigEndian.PutUint32(patched[at+8:], width)
	binary.BigEndian.PutUint32(patched[at+12:], height)
	return patched
}

const spikeLargeAVIFEnv = "VIKUNJA_PREVIEW_SPIKE_LARGE_AVIF"

func TestPreviewSpikeDecoderSideEffects(t *testing.T) {
	if os.Getenv(spikeEnv) == "" {
		t.Skipf("set %s=1 to run the preview format comparison", spikeEnv)
	}

	t.Run("avif uploads decode wherever image.Decode is used", func(t *testing.T) {
		encoded := &bytes.Buffer{}
		require.NoError(t, avif.Encode(encoded, opaqueImage(64, 48)))

		cfg, err := imageutils.ValidateReader(bytes.NewReader(encoded.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, 64, cfg.Width)

		_, format, err := image.Decode(bytes.NewReader(encoded.Bytes()))
		require.NoError(t, err)
		assert.Equal(t, "avif", format)
	})

	claimedPixels := func(t *testing.T, content []byte) int64 {
		t.Helper()
		pixels, err := sourcePixels(func() (io.ReadCloser, error) { return io.NopCloser(bytes.NewReader(content)), nil })
		require.NoError(t, err)
		return pixels
	}
	render := func(t *testing.T, content []byte) (allocated uint64) {
		t.Helper()
		runtime.GC()
		var before, after runtime.MemStats
		runtime.ReadMemStats(&before)
		started := time.Now()
		preview, err := renderPreview(bytes.NewReader(content), PreviewSmall)
		took := time.Since(started)
		runtime.ReadMemStats(&after)
		require.NoError(t, err)
		allocated = after.TotalAlloc - before.TotalAlloc
		t.Logf("rendering a %d byte file allocated %d bytes in %s for a %d byte preview", len(content), allocated, took, len(preview))
		return allocated
	}

	t.Run("a frame larger than its ispe box is decoded in full", func(t *testing.T) {
		const side = 2000
		encoded := &bytes.Buffer{}
		require.NoError(t, avif.Encode(encoded, image.NewGray(image.Rect(0, 0, side, side)), avif.EncodeOptions{Speed: 10}))

		hostile := spikeClaimDimensions(t, encoded.Bytes(), 16, 16)
		assert.Equal(t, int64(16*16), claimedPixels(t, hostile), "the semaphore is charged for the claimed size")
		assert.Greater(t, render(t, hostile), uint64(side*side))
	})

	// The library's own encoder cannot write a frame this large, see the next test.
	// ffmpeg -f lavfi -i testsrc2=s=8000x8000:d=1 -frames:v 1 -pix_fmt yuv420p -c:v libaom-av1 -still-picture 1 -cpu-used 8 -crf 50 large.avif
	t.Run("a frame over the pixel limit passes the guard behind a small ispe box", func(t *testing.T) {
		path := os.Getenv(spikeLargeAVIFEnv)
		if path == "" {
			t.Skipf("set %s to an AVIF larger than %d pixels", spikeLargeAVIFEnv, imageutils.MaxPixels)
		}
		honest, err := os.ReadFile(path) // #nosec G703 -- path from the developer's environment
		require.NoError(t, err)

		_, err = imageutils.ValidateReader(bytes.NewReader(honest))
		require.Error(t, err)
		assert.True(t, imageutils.IsErrImageTooLarge(err), "err: %v", err)

		hostile := spikeClaimDimensions(t, honest, 16, 16)
		assert.Equal(t, int64(16*16), claimedPixels(t, hostile))
		assert.Greater(t, render(t, hostile), uint64(imageutils.MaxPixels))
	})

	// AV1 limits a tile to 4096 pixels in width and 4096x2304 in area, and the encoder only writes one.
	t.Run("the encoder writes an undecodable file above the size of one tile", func(t *testing.T) {
		roundtrip := func(width, height int) error {
			encoded := &bytes.Buffer{}
			require.NoError(t, avif.Encode(encoded, opaqueImage(width, height), avif.EncodeOptions{Speed: 10}))
			_, err := avif.Decode(bytes.NewReader(encoded.Bytes()))
			return err
		}

		require.NoError(t, roundtrip(800, 11000))
		require.Error(t, roundtrip(800, 12000))
		require.Error(t, roundtrip(4097, 100))
	})
}
