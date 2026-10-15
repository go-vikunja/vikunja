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

import (
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/image/tiff"
)

func TestTaskAttachment_ReadOne(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("Normal File", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 1,
		}
		err := ta.ReadOne(s, u)
		require.NoError(t, err)
		assert.NotNil(t, ta.File)
		assert.True(t, ta.File.ID == ta.FileID && ta.FileID != 0)

		// Load the actual attachment file and check its content
		err = ta.File.LoadFileByID()
		require.NoError(t, err)

		// Validate the file exists at the expected storage path
		stat, err := files.FileStat(ta.File)
		require.NoError(t, err)
		assert.NotNil(t, stat)

		content := make([]byte, 9)
		read, err := ta.File.File.Read(content)
		require.NoError(t, err)
		assert.Equal(t, 9, read)
		assert.Equal(t, []byte("testfile1"), content)
	})
	t.Run("Nonexisting Attachment", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 9999,
		}
		err := ta.ReadOne(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing Attachment, Nonexisting File", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		files.InitTestFileFixtures(t)
		ta := &TaskAttachment{
			ID: 2,
		}
		err := ta.ReadOne(s, u)
		require.Error(t, err)
		require.EqualError(t, err, "file 9999 does not exist")
	})
}

func TestTaskAttachment_NewAttachment(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	files.InitTestFileFixtures(t)
	// Assert the file is being stored correctly
	ta := TaskAttachment{
		TaskID: 1,
	}
	testuser := &user.User{ID: 1}

	content := []byte("testingstuff")
	err := ta.NewAttachment(s, bytes.NewReader(content), "testfile", uint64(len(content)), testuser)
	require.NoError(t, err)
	assert.NotEqual(t, 0, ta.FileID)
	_, err = files.FileStat(ta.File)
	require.NoError(t, err)
	assert.False(t, os.IsNotExist(err))
	assert.Equal(t, testuser.ID, ta.CreatedByID)

	err = s.Commit()
	require.NoError(t, err)

	s2 := db.NewSession()
	defer s2.Close()

	// Check the file was inserted correctly
	ta.File = &files.File{ID: ta.FileID}
	err = ta.File.LoadFileMetaByID(s2)
	require.NoError(t, err)
	assert.Equal(t, testuser.ID, ta.File.CreatedByID)
	assert.Equal(t, "testfile", ta.File.Name)
	assert.Equal(t, uint64(len(content)), ta.File.Size)
	assert.NotEmpty(t, ta.File.Mime, "mime type should be detected and stored")

	// Extra test for max size test
}

func TestTaskAttachment_ReadAll(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	s := db.NewSession()
	defer s.Close()

	files.InitTestFileFixtures(t)
	ta := &TaskAttachment{TaskID: 1}
	as, _, _, err := ta.ReadAll(s, &user.User{ID: 1}, "", 0, 50)
	attachments, _ := as.([]*TaskAttachment)
	require.NoError(t, err)
	assert.Len(t, attachments, 3)
	assert.Equal(t, "test", attachments[0].File.Name)
	for _, a := range attachments {
		assert.NotNil(t, a.CreatedBy)
	}
	assert.Equal(t, int64(-2), attachments[2].CreatedByID)
	assert.Equal(t, int64(-2), attachments[2].CreatedBy.ID)
}

func TestTaskAttachment_Delete(t *testing.T) {
	u := &user.User{ID: 1}

	t.Run("Normal", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 1}
		err := ta.Delete(s, u)
		require.NoError(t, err)
		// Check if the file itself was deleted
		_, err = files.FileStat(ta.File) // The new file has the id 2 since it's the second attachment
		require.Error(t, err)
		assert.True(t, os.IsNotExist(err))
	})
	t.Run("Nonexisting", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 9999}
		err := ta.Delete(s, u)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentDoesNotExist(err))
	})
	t.Run("Existing attachment, nonexisting file", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		ta := &TaskAttachment{ID: 2}
		err := ta.Delete(s, u)
		require.NoError(t, err)
	})
}

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

	// Insert right after the SOI marker.
	out := append([]byte{}, encoded[:2]...)
	for range segments {
		out = append(out, segment...)
	}
	return append(out, encoded[2:]...)
}

// Only signature + IHDR: enough for image.DecodeConfig without allocating the pixels.
func pngHeader(width, height uint32) []byte {
	ihdr := []byte("IHDR")
	ihdr = binary.BigEndian.AppendUint32(ihdr, width)
	ihdr = binary.BigEndian.AppendUint32(ihdr, height)
	ihdr = append(ihdr, 8, 6, 0, 0, 0)

	header := []byte("\x89PNG\r\n\x1a\n")
	header = binary.BigEndian.AppendUint32(header, 13)
	header = append(header, ihdr...)
	return binary.BigEndian.AppendUint32(header, crc32.ChecksumIEEE(ihdr))
}

func imageAttachment(t *testing.T, id int64, content []byte) *TaskAttachment {
	t.Helper()
	file := &files.File{
		ID:   id,
		Mime: "image/png",
	}
	require.NoError(t, file.Save(bytes.NewReader(content)))
	return &TaskAttachment{
		ID:     id,
		FileID: id,
		File:   file,
	}
}

func TestTaskAttachment_Preview(t *testing.T) {
	t.Run("image", func(t *testing.T) {
		ta := imageAttachment(t, 910001, testPNG(t, 400, 300))

		preview, err := ta.loadPreviewOrFile(t.Context(), PreviewMedium)
		require.NoError(t, err)

		cfg, format, err := image.DecodeConfig(bytes.NewReader(preview))
		require.NoError(t, err)
		assert.Equal(t, "png", format)
		assert.Equal(t, PreviewMedium.GetSize(), cfg.Width)
		assert.Equal(t, 150, cfg.Height)
		assert.Nil(t, ta.File.File, "the original must not be opened when a preview is served")
	})

	t.Run("missing file is an error", func(t *testing.T) {
		ta := &TaskAttachment{
			ID: 910005,
			File: &files.File{
				ID:   910005,
				Mime: "image/png",
			},
		}

		_, err := ta.loadPreviewOrFile(t.Context(), PreviewMedium)
		require.Error(t, err)
		assert.True(t, files.IsErrFileDoesNotExist(err))
	})

	t.Run("busy server returns an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910006, testPNG(t, 400, 300))
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)

		preview, err := ta.loadPreviewOrFile(t.Context(), PreviewMedium)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentPreviewBusy(err), "err: %v", err)
		assert.Nil(t, preview)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("cancelled caller gets an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910007, testPNG(t, 400, 300))
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		preview, err := ta.loadPreviewOrFile(ctx, PreviewMedium)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, preview)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("cached preview needs no capacity", func(t *testing.T) {
		ta := imageAttachment(t, 910008, testPNG(t, 400, 300))
		generated, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)

		cached, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		assert.Equal(t, generated, cached)
	})

	fallbacks := map[string]struct {
		id      int64
		content []byte
		size    uint64
	}{
		"undecodable image falls back to the original": {
			id:      910002,
			content: append(pngHeader(400, 300), []byte("not image data")...),
		},
		"image over the pixel limit falls back to the original": {
			id:      910003,
			content: pngHeader(10000, 10000),
		},
		"file over the size limit falls back to the original": {
			id:      910004,
			content: testPNG(t, 400, 300),
			size:    maxPreviewSourceBytes + 1,
		},
	}
	for name, tc := range fallbacks {
		t.Run(name, func(t *testing.T) {
			ta := imageAttachment(t, tc.id, tc.content)
			ta.File.Size = tc.size

			preview, err := ta.loadPreviewOrFile(t.Context(), PreviewMedium)
			require.NoError(t, err)
			assert.Nil(t, preview)

			original, err := io.ReadAll(ta.File.File)
			require.NoError(t, err)
			require.NoError(t, ta.File.File.Close())
			assert.Equal(t, tc.content, original)
		})
	}
}

func exhaustAttachmentPreviewBudget(t *testing.T, queueTimeout time.Duration) {
	t.Helper()
	original := attachmentPreviews
	t.Cleanup(func() { attachmentPreviews = original })

	attachmentPreviews = newPreviewGenerator(1, queueTimeout)
	require.True(t, attachmentPreviews.slots.TryAcquire(1))
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

func TestPreviewGenerator_ClosesEverySourceHandle(t *testing.T) {
	tests := map[string]struct {
		content        []byte
		opened         int32
		notPreviewable bool
		tooLarge       bool
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
			content:        append(pngHeader(400, 300), []byte("not image data")...),
			opened:         2,
			notPreviewable: true,
		},
		"image over the pixel limit is never opened for decoding": {
			content:        pngHeader(10000, 10000),
			opened:         1,
			notPreviewable: true,
			tooLarge:       true,
		},
		"image whose header is past the read limit is never opened for decoding": {
			content:        testTIFF(t, 1024, maxPreviewHeaderBytes/(4*1024)+1),
			opened:         1,
			notPreviewable: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			source := &trackedSource{content: tc.content}

			_, err := newPreviewGenerator(previewPixelBudget, time.Minute).generate(t.Context(), source.open, PreviewMedium)

			assert.Equal(t, tc.notPreviewable, IsErrTaskAttachmentNotPreviewable(err), "err: %v", err)
			assert.Equal(t, tc.tooLarge, imageutils.IsErrImageTooLarge(err), "err: %v", err)
			assert.Equal(t, tc.opened, source.opened.Load())
			assert.Equal(t, source.opened.Load(), source.closed.Load())
		})
	}

	t.Run("source that cannot be opened is an error", func(t *testing.T) {
		openErr := errors.New("storage unavailable")

		open := func() (io.ReadCloser, error) { return nil, openErr }
		_, err := newPreviewGenerator(previewPixelBudget, time.Minute).generate(t.Context(), open, PreviewMedium)

		require.ErrorIs(t, err, openErr)
		assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
	})
}

// Measures decoding pixels independently of the generator's own semaphore.
type decodeTracker struct {
	gate chan struct{}

	mu       sync.Mutex
	inFlight int64
	peak     int64
	decodes  int
}

func newDecodeTracker() *decodeTracker {
	return &decodeTracker{gate: make(chan struct{})}
}

func (d *decodeTracker) add(pixels int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.inFlight += pixels
	d.peak = max(d.peak, d.inFlight)
	if pixels > 0 {
		d.decodes++
	}
}

func (d *decodeTracker) startedDecodes() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.decodes
}

func (d *decodeTracker) pixelsInFlight() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.inFlight
}

func (d *decodeTracker) peakPixels() int64 {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.peak
}

func (d *decodeTracker) source(t *testing.T, content []byte) sourceOpener {
	t.Helper()
	cfg, _, err := image.DecodeConfig(bytes.NewReader(content))
	require.NoError(t, err)
	pixels := int64(cfg.Width) * int64(cfg.Height)

	var opened atomic.Int32
	return func() (io.ReadCloser, error) {
		// The first open only reads the header, the second one decodes.
		if opened.Add(1) == 1 {
			return io.NopCloser(bytes.NewReader(content)), nil
		}
		d.add(pixels)
		return &gatedDecode{
			Reader:  bytes.NewReader(content),
			tracker: d,
			pixels:  pixels,
		}, nil
	}
}

type gatedDecode struct {
	io.Reader
	tracker *decodeTracker
	pixels  int64
}

func (r *gatedDecode) Read(p []byte) (int, error) {
	<-r.tracker.gate
	return r.Reader.Read(p)
}

func (r *gatedDecode) Close() error {
	r.tracker.add(-r.pixels)
	return nil
}

func TestPreviewGenerator_BoundsDecodedPixels(t *testing.T) {
	const (
		small  = 100 * 100
		budget = 2 * small
	)

	t.Run("distinct images in parallel stay within the budget", func(t *testing.T) {
		g := newPreviewGenerator(budget, time.Minute)
		tracker := newDecodeTracker()
		content := testPNG(t, 100, 100)

		const callers = 6
		results := make(chan error, callers)
		for range callers {
			source := tracker.source(t, content)
			go func() {
				_, err := g.generate(t.Context(), source, PreviewSmall)
				results <- err
			}()
		}

		require.Eventually(t, func() bool { return tracker.pixelsInFlight() >= budget }, 5*time.Second, time.Millisecond)
		assert.Never(t, func() bool { return tracker.peakPixels() > budget }, 200*time.Millisecond, 5*time.Millisecond)

		close(tracker.gate)
		for range callers {
			require.NoError(t, <-results)
		}
		assert.Equal(t, int64(budget), tracker.peakPixels())
	})

	t.Run("a larger image takes more of the budget", func(t *testing.T) {
		g := newPreviewGenerator(budget, time.Minute)
		tracker := newDecodeTracker()

		results := make(chan error, 2)
		generate := func(source sourceOpener) {
			_, err := g.generate(t.Context(), source, PreviewSmall)
			results <- err
		}

		large := tracker.source(t, testPNG(t, 200, 100))
		go generate(large)
		require.Eventually(t, func() bool { return tracker.pixelsInFlight() == budget }, 5*time.Second, time.Millisecond)

		smallSource := tracker.source(t, testPNG(t, 100, 100))
		go generate(smallSource)
		assert.Never(t, func() bool { return tracker.peakPixels() > budget }, 200*time.Millisecond, 5*time.Millisecond)

		close(tracker.gate)
		require.NoError(t, <-results)
		require.NoError(t, <-results)
		assert.Equal(t, int64(budget), tracker.peakPixels())
	})
}

func TestPreviewGenerator_Queue(t *testing.T) {
	const pixels = 100 * 100

	t.Run("cancelled caller leaves the queue with an error", func(t *testing.T) {
		g := newPreviewGenerator(pixels, time.Minute)
		require.True(t, g.slots.TryAcquire(pixels))
		source := &trackedSource{content: testPNG(t, 100, 100)}
		ctx, cancel := context.WithCancel(t.Context())

		result := make(chan error, 1)
		go func() {
			_, err := g.generate(ctx, source.open, PreviewSmall)
			result <- err
		}()
		require.Eventually(t, func() bool { return source.opened.Load() == 1 }, 5*time.Second, time.Millisecond)
		cancel()

		select {
		case err := <-result:
			require.ErrorIs(t, err, context.Canceled)
			assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
		case <-time.After(5 * time.Second):
			require.FailNow(t, "cancelled caller is still queued")
		}
		assert.Equal(t, int32(1), source.opened.Load(), "the source must not be opened for decoding")

		g.slots.Release(pixels)
		assert.True(t, g.slots.TryAcquire(pixels), "the cancelled caller must not hold any budget")
	})

	t.Run("gives up with a 503 when no capacity frees up", func(t *testing.T) {
		g := newPreviewGenerator(pixels, 10*time.Millisecond)
		require.True(t, g.slots.TryAcquire(pixels))
		source := &trackedSource{content: testPNG(t, 100, 100)}

		_, err := g.generate(t.Context(), source.open, PreviewSmall)

		var busy ErrTaskAttachmentPreviewBusy
		require.ErrorAs(t, err, &busy)
		assert.Equal(t, http.StatusServiceUnavailable, busy.HTTPError().HTTPCode)
		assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
		assert.Equal(t, int32(1), source.opened.Load(), "the source must not be opened for decoding")
	})
}

// The preview cache outlives a test, so a fixed key would be a cache hit when the test runs again.
func uniquePreviewCacheKey(t *testing.T) string {
	t.Helper()
	return t.Name() + "_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}

func TestPreviewGenerator_SharesGeneration(t *testing.T) {
	const pixels = 100 * 100

	t.Run("parallel calls for the same preview generate once", func(t *testing.T) {
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		tracker := newDecodeTracker()
		source := tracker.source(t, testPNG(t, 100, 100))
		cacheKey := uniquePreviewCacheKey(t)

		const callers = 6
		type result struct {
			preview []byte
			err     error
		}
		results := make(chan result, callers)
		for range callers {
			go func() {
				preview, err := g.get(t.Context(), cacheKey, source, PreviewSmall)
				results <- result{
					preview: preview,
					err:     err,
				}
			}()
		}

		require.Eventually(t, func() bool { return tracker.startedDecodes() > 0 }, 5*time.Second, time.Millisecond)
		assert.Never(t, func() bool { return tracker.startedDecodes() > 1 }, 200*time.Millisecond, 5*time.Millisecond)

		close(tracker.gate)
		for range callers {
			r := <-results
			require.NoError(t, r.err)
			_, err := png.DecodeConfig(bytes.NewReader(r.preview))
			require.NoError(t, err)
		}
		assert.Equal(t, 1, tracker.startedDecodes())
	})

	t.Run("a caller cancelling does not fail another caller of the same flight", func(t *testing.T) {
		g := newPreviewGenerator(pixels, time.Minute)
		require.True(t, g.slots.TryAcquire(pixels))
		content := testPNG(t, 100, 100)
		leaderSource := &trackedSource{content: content}
		followerSource := &trackedSource{content: content}
		cacheKey := uniquePreviewCacheKey(t)

		leaderCtx, cancelLeader := context.WithCancel(t.Context())
		leader := make(chan error, 1)
		go func() {
			_, err := g.get(leaderCtx, cacheKey, leaderSource.open, PreviewSmall)
			leader <- err
		}()
		require.Eventually(t, func() bool { return leaderSource.opened.Load() == 1 }, 5*time.Second, time.Millisecond)

		type result struct {
			preview []byte
			err     error
		}
		follower := make(chan result, 1)
		go func() {
			preview, err := g.get(t.Context(), cacheKey, followerSource.open, PreviewSmall)
			follower <- result{
				preview: preview,
				err:     err,
			}
		}()

		cancelLeader()
		select {
		case err := <-leader:
			require.ErrorIs(t, err, context.Canceled)
		case <-time.After(5 * time.Second):
			require.FailNow(t, "cancelled caller is still waiting")
		}

		g.slots.Release(pixels)
		r := <-follower
		require.NoError(t, r.err)
		_, err := png.DecodeConfig(bytes.NewReader(r.preview))
		require.NoError(t, err)
		assert.Zero(t, followerSource.opened.Load(), "the follower must share the flight instead of generating itself")
	})

	t.Run("a panicking decode is an error", func(t *testing.T) {
		g := newPreviewGenerator(pixels, time.Minute)

		_, err := g.get(t.Context(), uniquePreviewCacheKey(t), func() (io.ReadCloser, error) { panic("broken decoder") }, PreviewSmall)

		require.ErrorContains(t, err, "broken decoder")
		assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
	})
}

func TestRenderPreview_NeverUpscales(t *testing.T) {
	tests := map[string]struct {
		width  int
		height int
	}{
		"small image": {
			width:  50,
			height: 40,
		},
		// Upscaled to xl this would be 800x16,000,000 pixels.
		"tall one pixel wide image": {
			width:  1,
			height: 20000,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			preview, err := renderPreview(bytes.NewReader(testPNG(t, tc.width, tc.height)), PreviewExtraLarge)
			require.NoError(t, err)

			cfg, err := png.DecodeConfig(bytes.NewReader(preview))
			require.NoError(t, err)
			assert.Equal(t, tc.width, cfg.Width)
			assert.Equal(t, tc.height, cfg.Height)
		})
	}
}

func TestTaskAttachment_Permissions(t *testing.T) {
	u := &user.User{ID: 1}
	t.Run("Can Read", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, _, err := ta.CanRead(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, _, err := ta.CanRead(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Delete", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanDelete(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
	t.Run("Can Create", func(t *testing.T) {
		t.Run("Allowed", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 1}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.True(t, can)
		})
		t.Run("Forbidden, no access", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 14}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
		t.Run("Forbidden, shared read only", func(t *testing.T) {
			db.LoadAndAssertFixtures(t)
			s := db.NewSession()
			defer s.Close()

			ta := &TaskAttachment{TaskID: 15}
			can, err := ta.CanCreate(s, u)
			require.NoError(t, err)
			assert.False(t, can)
		})
	})
}
