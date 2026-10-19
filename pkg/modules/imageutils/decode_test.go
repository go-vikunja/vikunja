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
	"context"
	"encoding/binary"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/tiff"
)

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, width, height))))
	return buf.Bytes()
}

// x/image/tiff writes the IFD, and with it the dimensions, after the pixel data.
func testTIFF(t *testing.T, width, height int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, tiff.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, width, height)), nil))
	return buf.Bytes()
}

// Google Camera portraits carry a depth map in extended XMP before the image data.
func testJPEGWithMetadata(t *testing.T, width, height, segments int) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, jpeg.Encode(&buf, image.NewNRGBA(image.Rect(0, 0, width, height)), nil))
	encoded := buf.Bytes()

	const maxSegmentPayload = 0xffff - 2
	segment := binary.BigEndian.AppendUint16([]byte{0xff, 0xe1}, 0xffff)
	segment = append(segment, make([]byte, maxSegmentPayload)...)

	out := append([]byte{}, encoded[:2]...)
	for range segments {
		out = append(out, segment...)
	}
	return append(out, encoded[2:]...)
}

type trackedSource struct {
	content []byte
	opened  atomic.Int32
	closed  atomic.Int32
}

func (s *trackedSource) open() (io.ReadCloser, error) {
	s.opened.Add(1)
	return &trackedReader{
		Reader: bytes.NewReader(s.content),
		source: s,
	}, nil
}

type trackedReader struct {
	io.Reader
	source *trackedSource
}

func (r *trackedReader) Close() error {
	r.source.closed.Add(1)
	return nil
}

func noUse(image.Image, string) error { return nil }

func assertBudgetFree(t *testing.T, d *decoder) {
	t.Helper()
	require.True(t, d.slots.TryAcquire(d.budget), "the whole budget must be free")
	d.slots.Release(d.budget)
}

func TestDecode(t *testing.T) {
	t.Run("hands the decoded image and its format to use", func(t *testing.T) {
		var width int
		var format string
		err := Decode(t.Context(), BytesOpener(testPNG(t, 40, 30)), func(img image.Image, f string) error {
			width = img.Bounds().Dx()
			format = f
			return nil
		})
		require.NoError(t, err)
		assert.Equal(t, 40, width)
		assert.Equal(t, "png", format)
	})

	t.Run("returns the error of use", func(t *testing.T) {
		useErr := errors.New("encoding failed")
		err := Decode(t.Context(), BytesOpener(testPNG(t, 4, 4)), func(image.Image, string) error { return useErr })
		require.ErrorIs(t, err, useErr)
		assert.NotErrorIs(t, err, ErrUndecodable)
	})
}

func TestDecode_ClosesEverySourceHandle(t *testing.T) {
	tests := map[string]struct {
		content     []byte
		opened      int32
		undecodable bool
		tooLarge    bool
	}{
		"image": {
			content: testPNG(t, 400, 300),
			opened:  2,
		},
		"image with more than 1 MiB of metadata before the image data": {
			content: testJPEGWithMetadata(t, 400, 300, 24),
			opened:  2,
		},
		"undecodable image": {
			content:     pngWithClaimedDimensions(t, 400, 300),
			opened:      2,
			undecodable: true,
		},
		"image over the pixel limit is never opened for decoding": {
			content:  pngWithClaimedDimensions(t, 10000, 10000),
			opened:   1,
			tooLarge: true,
		},
		"image whose header is past the read limit is never opened for decoding": {
			content:     testTIFF(t, 1024, maxHeaderBytes/(4*1024)+1),
			opened:      1,
			undecodable: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			source := &trackedSource{content: tc.content}

			err := newDecoder(2*MaxPixels, time.Minute).decode(t.Context(), source.open, noUse)

			assert.Equal(t, tc.undecodable, errors.Is(err, ErrUndecodable), "err: %v", err)
			assert.Equal(t, tc.tooLarge, IsErrImageTooLarge(err), "err: %v", err)
			assert.Equal(t, tc.opened, source.opened.Load())
			assert.Equal(t, source.opened.Load(), source.closed.Load())
		})
	}

	t.Run("source that cannot be opened is an error as is", func(t *testing.T) {
		openErr := errors.New("storage unavailable")

		err := newDecoder(2*MaxPixels, time.Minute).decode(t.Context(), func() (io.ReadCloser, error) { return nil, openErr }, noUse)

		require.ErrorIs(t, err, openErr)
		assert.NotErrorIs(t, err, ErrUndecodable)
	})
}

type pixelTracker struct {
	gate chan struct{}

	mu       sync.Mutex
	inFlight int64
	peak     int64
}

func newPixelTracker() *pixelTracker {
	return &pixelTracker{gate: make(chan struct{})}
}

func (p *pixelTracker) use(img image.Image, _ string) error {
	pixels := int64(img.Bounds().Dx()) * int64(img.Bounds().Dy())
	p.mu.Lock()
	p.inFlight += pixels
	p.peak = max(p.peak, p.inFlight)
	p.mu.Unlock()

	<-p.gate

	p.mu.Lock()
	p.inFlight -= pixels
	p.mu.Unlock()
	return nil
}

func (p *pixelTracker) pixelsInFlight() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.inFlight
}

func (p *pixelTracker) peakPixels() int64 {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.peak
}

func TestDecode_BoundsPixelsInUse(t *testing.T) {
	const (
		small  = 100 * 100
		budget = 2 * small
	)

	t.Run("distinct images in parallel stay within the budget", func(t *testing.T) {
		d := newDecoder(budget, time.Minute)
		tracker := newPixelTracker()
		content := testPNG(t, 100, 100)

		const callers = 6
		results := make(chan error, callers)
		for range callers {
			go func() {
				results <- d.decode(t.Context(), BytesOpener(content), tracker.use)
			}()
		}

		require.Eventually(t, func() bool { return tracker.pixelsInFlight() >= budget }, 5*time.Second, time.Millisecond)
		assert.Never(t, func() bool { return tracker.peakPixels() > budget }, 200*time.Millisecond, 5*time.Millisecond)

		close(tracker.gate)
		for range callers {
			require.NoError(t, <-results)
		}
		assert.Equal(t, int64(budget), tracker.peakPixels())
		assertBudgetFree(t, d)
	})

	t.Run("a larger image takes more of the budget", func(t *testing.T) {
		d := newDecoder(budget, time.Minute)
		tracker := newPixelTracker()

		results := make(chan error, 2)
		large := BytesOpener(testPNG(t, 200, 100))
		go func() {
			results <- d.decode(t.Context(), large, tracker.use)
		}()
		require.Eventually(t, func() bool { return tracker.pixelsInFlight() == budget }, 5*time.Second, time.Millisecond)

		smallImage := BytesOpener(testPNG(t, 100, 100))
		go func() {
			results <- d.decode(t.Context(), smallImage, tracker.use)
		}()
		assert.Never(t, func() bool { return tracker.peakPixels() > budget }, 200*time.Millisecond, 5*time.Millisecond)

		close(tracker.gate)
		require.NoError(t, <-results)
		require.NoError(t, <-results)
		assert.Equal(t, int64(budget), tracker.peakPixels())
	})

	t.Run("an image larger than the whole budget is rejected without queueing", func(t *testing.T) {
		d := newDecoder(small-1, time.Minute)
		source := &trackedSource{content: testPNG(t, 100, 100)}

		result := make(chan error, 1)
		go func() {
			result <- d.decode(t.Context(), source.open, func(image.Image, string) error {
				t.Error("use must not be called")
				return nil
			})
		}()

		select {
		case err := <-result:
			assert.True(t, IsErrImageTooLarge(err), "err: %v", err)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "the image is queued for a budget it can never get")
		}
		assert.Equal(t, int32(1), source.opened.Load(), "the source must not be opened for decoding")
		assertBudgetFree(t, d)
	})
}

func TestDecode_ReleasesTheBudget(t *testing.T) {
	const pixels = 100 * 100

	t.Run("when decoding fails", func(t *testing.T) {
		d := newDecoder(pixels, time.Minute)

		err := d.decode(t.Context(), BytesOpener(pngWithClaimedDimensions(t, 100, 100)), noUse)

		require.ErrorIs(t, err, ErrUndecodable)
		assertBudgetFree(t, d)
	})

	t.Run("when use fails", func(t *testing.T) {
		d := newDecoder(pixels, time.Minute)
		useErr := errors.New("encoding failed")

		err := d.decode(t.Context(), BytesOpener(testPNG(t, 100, 100)), func(image.Image, string) error { return useErr })

		require.ErrorIs(t, err, useErr)
		assertBudgetFree(t, d)
	})

	t.Run("when use panics", func(t *testing.T) {
		d := newDecoder(pixels, time.Minute)
		source := &trackedSource{content: testPNG(t, 100, 100)}

		assert.PanicsWithValue(t, "broken encoder", func() {
			_ = d.decode(t.Context(), source.open, func(image.Image, string) error { panic("broken encoder") })
		})
		assertBudgetFree(t, d)
		assert.Equal(t, source.opened.Load(), source.closed.Load())
	})

	t.Run("when decoding panics", func(t *testing.T) {
		d := newDecoder(pixels, time.Minute)
		source := &panickingSource{content: testPNG(t, 100, 100)}

		assert.PanicsWithValue(t, "broken decoder", func() {
			_ = d.decode(t.Context(), source.open, noUse)
		})
		assertBudgetFree(t, d)
	})
}

type panickingSource struct {
	content []byte
	opened  int
}

func (s *panickingSource) open() (io.ReadCloser, error) {
	s.opened++
	if s.opened == 1 {
		return io.NopCloser(bytes.NewReader(s.content)), nil
	}
	return io.NopCloser(panickingReader{}), nil
}

type panickingReader struct{}

func (panickingReader) Read([]byte) (int, error) { panic("broken decoder") }

func TestDecode_Queue(t *testing.T) {
	const pixels = 100 * 100

	t.Run("cancelled caller leaves the queue with an error and holds nothing", func(t *testing.T) {
		d := newDecoder(pixels, time.Minute)
		require.True(t, d.slots.TryAcquire(pixels))
		source := &trackedSource{content: testPNG(t, 100, 100)}
		ctx, cancel := context.WithCancel(t.Context())

		result := make(chan error, 1)
		go func() {
			result <- d.decode(ctx, source.open, noUse)
		}()
		require.Eventually(t, func() bool { return source.opened.Load() == 1 }, 5*time.Second, time.Millisecond)
		cancel()

		select {
		case err := <-result:
			require.ErrorIs(t, err, context.Canceled)
			assert.False(t, IsErrDecodeBusy(err))
		case <-time.After(5 * time.Second):
			require.FailNow(t, "cancelled caller is still queued")
		}
		assert.Equal(t, int32(1), source.opened.Load(), "the source must not be opened for decoding")

		d.slots.Release(pixels)
		assertBudgetFree(t, d)
	})

	t.Run("gives up with a 503 when no capacity frees up", func(t *testing.T) {
		d := newDecoder(pixels, 10*time.Millisecond)
		require.True(t, d.slots.TryAcquire(pixels))
		source := &trackedSource{content: testPNG(t, 100, 100)}

		err := d.decode(t.Context(), source.open, noUse)

		var busy ErrDecodeBusy
		require.ErrorAs(t, err, &busy)
		assert.Equal(t, http.StatusServiceUnavailable, busy.HTTPError().HTTPCode)
		assert.Equal(t, ErrCodeDecodeBusy, busy.HTTPError().Code)
		assert.Equal(t, int32(1), source.opened.Load(), "the source must not be opened for decoding")

		d.slots.Release(pixels)
		assertBudgetFree(t, d)
	})
}
