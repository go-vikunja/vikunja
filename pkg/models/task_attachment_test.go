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
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/modules/imageutils"
	"code.vikunja.io/api/pkg/modules/keyvalue"
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
	t.Run("removes the previews", func(t *testing.T) {
		files.InitTestFileFixtures(t)
		db.LoadAndAssertFixtures(t)
		s := db.NewSession()
		defer s.Close()

		content := testPNG(t, 400, 300)
		ta := &TaskAttachment{TaskID: 1}
		require.NoError(t, ta.NewAttachment(s, bytes.NewReader(content), "image.png", uint64(len(content)), u))
		preview, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		require.NoError(t, preview.Close())
		storedPreview(t, ta.FileID, PreviewMedium)

		require.NoError(t, ta.Delete(s, u))

		_, _, err = ta.File.OpenPreview(PreviewMedium)
		require.ErrorIs(t, err, fs.ErrNotExist)
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

func opaqueImage(width, height int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.NewUniform(color.NRGBA{
		R: 200,
		G: 100,
		B: 50,
		A: 255,
	}), image.Point{}, draw.Src)
	return img
}

func encodedImage(t *testing.T, encode func(io.Writer, image.Image) error, img image.Image) []byte {
	t.Helper()
	var buf bytes.Buffer
	require.NoError(t, encode(&buf, img))
	return buf.Bytes()
}

func TestRemoveLegacyTaskAttachmentPreviews(t *testing.T) {
	require.NoError(t, keyvalue.Put("task_attachment_preview_1_size_md", []byte("preview")))
	require.NoError(t, keyvalue.Put("task_attachment_other", "kept"))
	t.Cleanup(func() { _ = keyvalue.Del("task_attachment_other") })

	require.NoError(t, RemoveLegacyTaskAttachmentPreviews())

	_, exists, err := keyvalue.Get("task_attachment_preview_1_size_md")
	require.NoError(t, err)
	assert.False(t, exists)
	_, exists, err = keyvalue.Get("task_attachment_other")
	require.NoError(t, err)
	assert.True(t, exists)
}

func TestPreviewFormat(t *testing.T) {
	transparentCorner := opaqueImage(400, 300)
	transparentCorner.SetNRGBA(0, 0, color.NRGBA{})
	encodeJPEG := func(w io.Writer, img image.Image) error { return jpeg.Encode(w, img, nil) }

	tests := map[string]struct {
		source []byte
		format string
		mime   string
	}{
		"opaque png becomes a jpeg": {
			source: encodedImage(t, png.Encode, opaqueImage(400, 300)),
			format: "jpeg",
			mime:   "image/jpeg",
		},
		"jpeg stays a jpeg": {
			source: encodedImage(t, encodeJPEG, opaqueImage(400, 300)),
			format: "jpeg",
			mime:   "image/jpeg",
		},
		"fully transparent png stays a png": {
			source: testPNG(t, 400, 300),
			format: "png",
			mime:   "image/png",
		},
		"png with a single transparent pixel stays a png": {
			source: encodedImage(t, png.Encode, transparentCorner),
			format: "png",
			mime:   "image/png",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fileID := uniquePreviewFileID()
			source := &trackedSource{content: tc.source}
			g := newPreviewGenerator(previewPixelBudget, time.Minute)

			// The first preview is the generated one, the second the stored one.
			for range 2 {
				preview, err := g.get(t.Context(), fileID, source.open, PreviewMedium)
				require.NoError(t, err)
				assert.Equal(t, tc.mime, preview.Mime)

				cfg, format, err := image.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
				require.NoError(t, err)
				assert.Equal(t, tc.format, format)
				assert.Equal(t, PreviewMedium.GetSize(), cfg.Width)
				assert.Equal(t, 150, cfg.Height)
			}
			assert.Equal(t, int32(2), source.opened.Load())
		})
	}

	t.Run("transparency survives", func(t *testing.T) {
		preview, err := renderPreview(bytes.NewReader(encodedImage(t, png.Encode, transparentCorner)), PreviewMedium)
		require.NoError(t, err)

		img, err := png.Decode(bytes.NewReader(preview))
		require.NoError(t, err)
		_, _, _, alpha := img.At(0, 0).RGBA()
		assert.Less(t, alpha, uint32(0xffff))
	})
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

func readPreview(t *testing.T, preview *TaskAttachmentPreview) []byte {
	t.Helper()
	content, err := io.ReadAll(preview)
	require.NoError(t, err)
	require.NoError(t, preview.Close())
	assert.Equal(t, int64(len(content)), preview.Size)
	return content
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

		download, err := ta.openPreviewOrFile(t.Context(), PreviewMedium)
		require.NoError(t, err)
		preview, isPreview := download.(*TaskAttachmentPreview)
		require.True(t, isPreview, "got %T", download)
		assert.Equal(t, "image/png", preview.Mime)

		cfg, format, err := image.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
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

		_, err := ta.openPreviewOrFile(t.Context(), PreviewMedium)
		require.Error(t, err)
		assert.True(t, files.IsErrFileDoesNotExist(err))
	})

	t.Run("busy server returns an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910006, testPNG(t, 400, 300))
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)

		download, err := ta.openPreviewOrFile(t.Context(), PreviewMedium)
		require.Error(t, err)
		assert.True(t, IsErrTaskAttachmentPreviewBusy(err), "err: %v", err)
		assert.Nil(t, download)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("cancelled caller gets an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910007, testPNG(t, 400, 300))
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		download, err := ta.openPreviewOrFile(ctx, PreviewMedium)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, download)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("stored preview needs no capacity", func(t *testing.T) {
		ta := imageAttachment(t, 910008, testPNG(t, 400, 300))
		generated, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		exhaustAttachmentPreviewBudget(t, 10*time.Millisecond)

		stored, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		assert.Equal(t, readPreview(t, generated), readPreview(t, stored))
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

			download, err := ta.openPreviewOrFile(t.Context(), PreviewMedium)
			require.NoError(t, err)
			original, isOriginal := download.(*TaskAttachmentOriginal)
			require.True(t, isOriginal, "got %T", download)

			content, err := io.ReadAll(original.File.File)
			require.NoError(t, err)
			require.NoError(t, original.Close())
			assert.Equal(t, tc.content, content)
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

var lastPreviewTestFileID atomic.Int64

// Stored previews outlive a test, so a fixed file would already have one when the test runs again.
func uniquePreviewFileID() int64 {
	return 920000 + lastPreviewTestFileID.Add(1)
}

func storedPreview(t *testing.T, fileID int64, size PreviewSize) []byte {
	t.Helper()
	content, length, err := (&files.File{ID: fileID}).OpenPreview(size)
	require.NoError(t, err)
	preview, err := io.ReadAll(content)
	require.NoError(t, err)
	require.NoError(t, content.Close())
	assert.Equal(t, int64(len(preview)), length)
	return preview
}

type brokenPreviewStore struct {
	previewStore
	openErr error
	saveErr error
}

func (s brokenPreviewStore) open(fileID int64, size PreviewSize) (io.ReadCloser, int64, error) {
	if s.openErr != nil {
		return nil, 0, s.openErr
	}
	return s.previewStore.open(fileID, size)
}

func (s brokenPreviewStore) save(fileID int64, size PreviewSize, preview []byte) error {
	if s.saveErr != nil {
		return s.saveErr
	}
	return s.previewStore.save(fileID, size, preview)
}

type trackedPreviewStore struct {
	previewStore
	opened atomic.Int32
	closed atomic.Int32
}

func (s *trackedPreviewStore) open(fileID int64, size PreviewSize) (io.ReadCloser, int64, error) {
	preview, length, err := s.previewStore.open(fileID, size)
	if err != nil {
		return nil, 0, err
	}
	s.opened.Add(1)
	return &trackedStoredPreview{
		ReadCloser: preview,
		store:      s,
	}, length, nil
}

type trackedStoredPreview struct {
	io.ReadCloser
	store *trackedPreviewStore
}

func (p *trackedStoredPreview) Close() error {
	p.store.closed.Add(1)
	return p.ReadCloser.Close()
}

func failingSource(t *testing.T) sourceOpener {
	t.Helper()
	return func() (io.ReadCloser, error) {
		assert.Fail(t, "the source must not be opened")
		return nil, errors.New("source opened")
	}
}

func TestPreviewGenerator_StoresPreviews(t *testing.T) {
	t.Run("a stored preview is served without generating", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		stored := testPNG(t, 100, 100)
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, stored))
		source := &trackedSource{content: testPNG(t, 400, 300)}

		preview, err := newPreviewGenerator(previewPixelBudget, time.Minute).get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		assert.Equal(t, stored, readPreview(t, preview))
		assert.Zero(t, source.opened.Load())
	})

	t.Run("a stored preview is streamed and closed by the caller", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		var stored bytes.Buffer
		require.NoError(t, jpeg.Encode(&stored, image.NewGray(image.Rect(0, 0, 100, 100)), nil))
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, stored.Bytes()))
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		store := &trackedPreviewStore{previewStore: g.store}
		g.store = store

		preview, err := g.get(t.Context(), fileID, failingSource(t), PreviewSmall)

		require.NoError(t, err)
		assert.Equal(t, int32(1), store.opened.Load())
		assert.Zero(t, store.closed.Load(), "the preview must not be read into memory")
		assert.Equal(t, "image/jpeg", preview.Mime, "the mime type must come from the stored bytes")
		assert.Equal(t, stored.Bytes(), readPreview(t, preview))
		assert.Equal(t, int32(1), store.closed.Load())
	})

	t.Run("a stored preview that is not an image is generated again", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, nil))
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		store := &trackedPreviewStore{previewStore: g.store}
		g.store = store

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		generated := readPreview(t, preview)
		_, err = png.DecodeConfig(bytes.NewReader(generated))
		require.NoError(t, err)
		assert.Equal(t, generated, storedPreview(t, fileID, PreviewSmall))
		assert.Equal(t, store.opened.Load(), store.closed.Load())
	})

	t.Run("a generated preview is not generated again, even after a restart", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := newPreviewGenerator(previewPixelBudget, time.Minute)

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)
		require.NoError(t, err)
		generated := readPreview(t, preview)
		require.Equal(t, int32(2), source.opened.Load())
		assert.Equal(t, generated, storedPreview(t, fileID, PreviewSmall))

		// A restart only keeps what is in the storage.
		generators := map[string]*previewGenerator{
			"same generator":          g,
			"generator after restart": newPreviewGenerator(previewPixelBudget, time.Minute),
		}
		for name, generator := range generators {
			preview, err := generator.get(t.Context(), fileID, source.open, PreviewSmall)
			require.NoError(t, err, name)
			assert.Equal(t, generated, readPreview(t, preview), name)
			assert.Equal(t, int32(2), source.opened.Load(), name)
		}
	})

	t.Run("previews are stored per size", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := newPreviewGenerator(previewPixelBudget, time.Minute)

		for _, size := range []PreviewSize{
			PreviewSmall,
			PreviewMedium,
		} {
			preview, err := g.get(t.Context(), fileID, source.open, size)
			require.NoError(t, err)
			cfg, err := png.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
			require.NoError(t, err)
			assert.Equal(t, size.GetSize(), cfg.Width)
		}
	})

	t.Run("a preview that cannot be stored is still served", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		g.store = brokenPreviewStore{
			previewStore: g.store,
			saveErr:      errors.New("storage is full"),
		}

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		cfg, err := png.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
		require.NoError(t, err)
		assert.Equal(t, PreviewSmall.GetSize(), cfg.Width)
		_, _, err = (&files.File{ID: fileID}).OpenPreview(PreviewSmall)
		require.ErrorIs(t, err, fs.ErrNotExist)
	})

	t.Run("a stored preview that cannot be read is generated instead", func(t *testing.T) {
		fileID := uniquePreviewFileID()
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, testPNG(t, 100, 100)))
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		g.store = brokenPreviewStore{
			previewStore: g.store,
			openErr:      errors.New("storage unavailable"),
		}

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		cfg, err := png.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
		require.NoError(t, err)
		assert.Equal(t, 75, cfg.Height, "the preview must come from the source, not the storage")
		assert.Equal(t, int32(2), source.opened.Load())
	})
}

func TestPreviewGenerator_SharesGeneration(t *testing.T) {
	const pixels = 100 * 100

	t.Run("parallel calls for the same preview generate once", func(t *testing.T) {
		g := newPreviewGenerator(previewPixelBudget, time.Minute)
		tracker := newDecodeTracker()
		source := tracker.source(t, testPNG(t, 100, 100))
		fileID := uniquePreviewFileID()

		const callers = 6
		type result struct {
			preview *TaskAttachmentPreview
			err     error
		}
		results := make(chan result, callers)
		for range callers {
			go func() {
				preview, err := g.get(t.Context(), fileID, source, PreviewSmall)
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
			// Every caller reads the whole preview although they share one flight.
			_, err := png.Decode(bytes.NewReader(readPreview(t, r.preview)))
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
		fileID := uniquePreviewFileID()

		leaderCtx, cancelLeader := context.WithCancel(t.Context())
		leader := make(chan error, 1)
		go func() {
			_, err := g.get(leaderCtx, fileID, leaderSource.open, PreviewSmall)
			leader <- err
		}()
		require.Eventually(t, func() bool { return leaderSource.opened.Load() == 1 }, 5*time.Second, time.Millisecond)

		type result struct {
			preview *TaskAttachmentPreview
			err     error
		}
		follower := make(chan result, 1)
		go func() {
			preview, err := g.get(t.Context(), fileID, followerSource.open, PreviewSmall)
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
		_, err := png.DecodeConfig(bytes.NewReader(readPreview(t, r.preview)))
		require.NoError(t, err)
		assert.Zero(t, followerSource.opened.Load(), "the follower must share the flight instead of generating itself")
	})

	t.Run("a panicking decode is an error", func(t *testing.T) {
		g := newPreviewGenerator(pixels, time.Minute)

		_, err := g.get(t.Context(), uniquePreviewFileID(), func() (io.ReadCloser, error) { panic("broken decoder") }, PreviewSmall)

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

			cfg, _, err := image.DecodeConfig(bytes.NewReader(preview))
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
