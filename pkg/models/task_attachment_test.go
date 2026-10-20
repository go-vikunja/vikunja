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
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"io/fs"
	"net/http"
	"os"
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
			fileID := uniquePreviewFileID(t)
			source := &trackedSource{content: tc.source}
			g := &previewGenerator{}

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
		previews, err := generatePreviews(t.Context(), imageutils.BytesOpener(encodedImage(t, png.Encode, transparentCorner)))
		require.NoError(t, err)

		img, err := png.Decode(bytes.NewReader(previews[PreviewMedium]))
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

		preview, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
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

		_, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.Error(t, err)
		assert.True(t, files.IsErrFileDoesNotExist(err))
	})

	t.Run("busy server returns an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910006, testPNG(t, 400, 300))
		t.Cleanup(imageutils.ExhaustBudgetForTests(10 * time.Millisecond))

		preview, err := ta.GetPreview(t.Context(), PreviewMedium)
		var busy imageutils.ErrDecodeBusy
		require.ErrorAs(t, err, &busy)
		assert.Equal(t, http.StatusServiceUnavailable, busy.HTTPError().HTTPCode)
		assert.Equal(t, imageutils.ErrCodeDecodeBusy, busy.HTTPError().Code)
		assert.Nil(t, preview)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("cancelled caller gets an error instead of the original", func(t *testing.T) {
		ta := imageAttachment(t, 910007, testPNG(t, 400, 300))
		t.Cleanup(imageutils.ExhaustBudgetForTests(10 * time.Millisecond))
		ctx, cancel := context.WithCancel(t.Context())
		cancel()

		preview, err := ta.GetPreview(ctx, PreviewMedium)
		require.ErrorIs(t, err, context.Canceled)
		assert.Nil(t, preview)
		assert.Nil(t, ta.File.File, "the original must not be opened")
	})

	t.Run("stored preview needs no capacity", func(t *testing.T) {
		ta := imageAttachment(t, 910008, testPNG(t, 400, 300))
		generated, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		t.Cleanup(imageutils.ExhaustBudgetForTests(10 * time.Millisecond))

		stored, err := ta.GetPreview(t.Context(), PreviewMedium)
		require.NoError(t, err)
		assert.Equal(t, readPreview(t, generated), readPreview(t, stored))
	})

	notPreviewable := map[string]struct {
		id      int64
		mime    string
		content []byte
		size    uint64
		cause   func(error) bool
	}{
		"non-image": {
			id:      910010,
			mime:    "text/plain",
			content: []byte("not an image"),
		},
		"undecodable image": {
			id:      910002,
			content: append(pngHeader(400, 300), []byte("not image data")...),
			cause:   func(err error) bool { return errors.Is(err, imageutils.ErrUndecodable) },
		},
		"image over the pixel limit": {
			id:      910003,
			content: pngHeader(10000, 10000),
			cause:   imageutils.IsErrImageTooLarge,
		},
		"file over the size limit": {
			id:      910004,
			content: testPNG(t, 400, 300),
			size:    maxPreviewSourceBytes + 1,
		},
	}
	for name, tc := range notPreviewable {
		t.Run(name+" is not previewable instead of returning the original", func(t *testing.T) {
			ta := imageAttachment(t, tc.id, tc.content)
			ta.File.Size = tc.size
			if tc.mime != "" {
				ta.File.Mime = tc.mime
			}

			preview, err := ta.GetPreview(t.Context(), PreviewMedium)
			require.Error(t, err)
			assert.True(t, IsErrTaskAttachmentNotPreviewable(err), "err: %v", err)
			if tc.cause != nil {
				assert.True(t, tc.cause(err), "err: %v", err)
			}
			assert.Nil(t, preview)
			assert.Nil(t, ta.File.File, "the original must not be opened")
		})
	}
}

type trackedSource struct {
	content []byte
	opened  atomic.Int32
}

func (s *trackedSource) open() (io.ReadCloser, error) {
	s.opened.Add(1)
	return io.NopCloser(bytes.NewReader(s.content)), nil
}

func TestGeneratePreviews(t *testing.T) {
	t.Run("source that cannot be opened is an error", func(t *testing.T) {
		openErr := errors.New("storage unavailable")

		_, err := generatePreviews(t.Context(), func() (io.ReadCloser, error) { return nil, openErr })

		require.ErrorIs(t, err, openErr)
		assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
	})

	neverUpscaled := map[string]struct {
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
	for name, tc := range neverUpscaled {
		t.Run(name+" is never upscaled", func(t *testing.T) {
			previews, err := generatePreviews(t.Context(), imageutils.BytesOpener(testPNG(t, tc.width, tc.height)))
			require.NoError(t, err)

			cfg, _, err := image.DecodeConfig(bytes.NewReader(previews[PreviewExtraLarge]))
			require.NoError(t, err)
			assert.Equal(t, tc.width, cfg.Width)
			assert.Equal(t, tc.height, cfg.Height)
		})
	}
}

// Blocks the decode of a source until the gate is closed.
type gatedSource struct {
	content []byte
	gate    chan struct{}
	opened  atomic.Int32
}

func newGatedSource(content []byte) *gatedSource {
	return &gatedSource{
		content: content,
		gate:    make(chan struct{}),
	}
}

func (s *gatedSource) open() (io.ReadCloser, error) {
	// The first open only reads the header, the second one decodes.
	if s.opened.Add(1) == 1 {
		return io.NopCloser(bytes.NewReader(s.content)), nil
	}
	return io.NopCloser(&gatedReader{
		Reader: bytes.NewReader(s.content),
		gate:   s.gate,
	}), nil
}

func (s *gatedSource) decodes() int32 {
	return max(s.opened.Load()-1, 0)
}

type gatedReader struct {
	io.Reader
	gate chan struct{}
}

func (r *gatedReader) Read(p []byte) (int, error) {
	<-r.gate
	return r.Reader.Read(p)
}

var lastPreviewTestFileID atomic.Int64

// Stored previews outlive a test, so a fixed file would already have one when the test runs again.
func uniquePreviewFileID(t *testing.T) int64 {
	t.Helper()
	fileID := 920000 + lastPreviewTestFileID.Add(1)
	require.NoError(t, (&files.File{ID: fileID}).Save(bytes.NewReader([]byte("original"))))
	return fileID
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

func failingSource(t *testing.T) imageutils.Opener {
	t.Helper()
	return func() (io.ReadCloser, error) {
		assert.Fail(t, "the source must not be opened")
		return nil, errors.New("source opened")
	}
}

func TestPreviewGenerator_StoresPreviews(t *testing.T) {
	t.Run("a stored preview is served without generating", func(t *testing.T) {
		fileID := uniquePreviewFileID(t)
		stored := testPNG(t, 100, 100)
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, stored))
		source := &trackedSource{content: testPNG(t, 400, 300)}

		preview, err := (&previewGenerator{}).get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		assert.Equal(t, stored, readPreview(t, preview))
		assert.Zero(t, source.opened.Load())
	})

	t.Run("a stored preview is streamed and closed by the caller", func(t *testing.T) {
		fileID := uniquePreviewFileID(t)
		var stored bytes.Buffer
		require.NoError(t, jpeg.Encode(&stored, image.NewGray(image.Rect(0, 0, 100, 100)), nil))
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, stored.Bytes()))

		preview, err := (&previewGenerator{}).get(t.Context(), fileID, failingSource(t), PreviewSmall)

		require.NoError(t, err)
		assert.Equal(t, "image/jpeg", preview.Mime, "the mime type must come from the stored bytes")
		assert.Equal(t, stored.Bytes(), readPreview(t, preview))
	})

	t.Run("a stored preview that is not an image is generated again", func(t *testing.T) {
		fileID := uniquePreviewFileID(t)
		require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, nil))
		source := &trackedSource{content: testPNG(t, 400, 300)}

		preview, err := (&previewGenerator{}).get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		generated := readPreview(t, preview)
		_, err = png.DecodeConfig(bytes.NewReader(generated))
		require.NoError(t, err)
		assert.Equal(t, generated, storedPreview(t, fileID, PreviewSmall))
	})

	t.Run("a stored preview that is neither jpeg nor png is generated again", func(t *testing.T) {
		svg := []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>alert(1)</script></svg>`)
		var gifPreview bytes.Buffer
		require.NoError(t, gif.Encode(&gifPreview, image.NewGray(image.Rect(0, 0, 100, 100)), nil))
		stored := map[string][]byte{
			"svg": svg,
			"gif": gifPreview.Bytes(),
		}
		for name, content := range stored {
			t.Run(name, func(t *testing.T) {
				fileID := uniquePreviewFileID(t)
				require.NoError(t, (&files.File{ID: fileID}).SavePreview(PreviewSmall, content))
				source := &trackedSource{content: testPNG(t, 400, 300)}

				preview, err := (&previewGenerator{}).get(t.Context(), fileID, source.open, PreviewSmall)

				require.NoError(t, err)
				assert.Equal(t, "image/png", preview.Mime)
				assert.Equal(t, readPreview(t, preview), storedPreview(t, fileID, PreviewSmall))
			})
		}
	})

	t.Run("a generated preview is not generated again, even after a restart", func(t *testing.T) {
		fileID := uniquePreviewFileID(t)
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := &previewGenerator{}

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)
		require.NoError(t, err)
		generated := readPreview(t, preview)
		require.Equal(t, int32(2), source.opened.Load())
		assert.Equal(t, generated, storedPreview(t, fileID, PreviewSmall))

		// A restart only keeps what is in the storage.
		generators := map[string]*previewGenerator{
			"same generator":          g,
			"generator after restart": {},
		}
		for name, generator := range generators {
			preview, err := generator.get(t.Context(), fileID, source.open, PreviewSmall)
			require.NoError(t, err, name)
			assert.Equal(t, generated, readPreview(t, preview), name)
			assert.Equal(t, int32(2), source.opened.Load(), name)
		}
	})

	t.Run("one request stores every size", func(t *testing.T) {
		fileID := uniquePreviewFileID(t)
		source := &trackedSource{content: testPNG(t, 400, 300)}
		g := &previewGenerator{}

		preview, err := g.get(t.Context(), fileID, source.open, PreviewSmall)
		require.NoError(t, err)
		assert.Equal(t, storedPreview(t, fileID, PreviewSmall), readPreview(t, preview))
		require.Equal(t, int32(2), source.opened.Load())

		dimensions := map[PreviewSize]image.Point{
			PreviewSmall:      {X: 100, Y: 75},
			PreviewMedium:     {X: 200, Y: 150},
			PreviewLarge:      {X: 400, Y: 300},
			PreviewExtraLarge: {X: 400, Y: 300},
		}
		require.Len(t, dimensions, len(files.StoredPreviewSizes))
		for size, want := range dimensions {
			stored := storedPreview(t, fileID, size)
			cfg, err := png.DecodeConfig(bytes.NewReader(stored))
			require.NoError(t, err, size)
			assert.Equal(t, want, image.Point{X: cfg.Width, Y: cfg.Height}, size)

			preview, err := g.get(t.Context(), fileID, failingSource(t), size)
			require.NoError(t, err, size)
			assert.Equal(t, stored, readPreview(t, preview), size)
		}
	})

	t.Run("a preview of a deleted file is served but not stored", func(t *testing.T) {
		fileID := 920000 + lastPreviewTestFileID.Add(1)
		source := &trackedSource{content: testPNG(t, 400, 300)}

		preview, err := (&previewGenerator{}).get(t.Context(), fileID, source.open, PreviewSmall)

		require.NoError(t, err)
		cfg, err := png.DecodeConfig(bytes.NewReader(readPreview(t, preview)))
		require.NoError(t, err)
		assert.Equal(t, PreviewSmall.GetSize(), cfg.Width)
		for _, size := range files.StoredPreviewSizes {
			_, _, err = (&files.File{ID: fileID}).OpenPreview(size)
			require.ErrorIs(t, err, fs.ErrNotExist, size)
		}
	})
}

func TestPreviewGenerator_SharesGeneration(t *testing.T) {
	t.Run("parallel calls for any size of the same file generate once", func(t *testing.T) {
		g := &previewGenerator{}
		source := newGatedSource(testPNG(t, 400, 300))
		fileID := uniquePreviewFileID(t)

		const callers = 6
		type result struct {
			size    PreviewSize
			preview *TaskAttachmentPreview
			err     error
		}
		results := make(chan result, callers)
		for i := range callers {
			size := []PreviewSize{
				PreviewSmall,
				PreviewLarge,
			}[i%2]
			go func() {
				preview, err := g.get(t.Context(), fileID, source.open, size)
				results <- result{
					size:    size,
					preview: preview,
					err:     err,
				}
			}()
		}

		require.Eventually(t, func() bool { return source.decodes() > 0 }, 5*time.Second, time.Millisecond)
		assert.Never(t, func() bool { return source.opened.Load() > 2 }, 200*time.Millisecond, 5*time.Millisecond)

		close(source.gate)
		for range callers {
			r := <-results
			require.NoError(t, r.err)
			// Every caller reads the whole preview although they share one flight.
			img, err := png.Decode(bytes.NewReader(readPreview(t, r.preview)))
			require.NoError(t, err)
			assert.Equal(t, r.size.GetSize(), img.Bounds().Dx())
		}
		assert.Equal(t, int32(1), source.decodes())
	})

	t.Run("a caller cancelling does not fail another caller of the same flight", func(t *testing.T) {
		g := &previewGenerator{}
		release := imageutils.ExhaustBudgetForTests(time.Minute)
		t.Cleanup(release)
		content := testPNG(t, 100, 100)
		leaderSource := &trackedSource{content: content}
		followerSource := &trackedSource{content: content}
		fileID := uniquePreviewFileID(t)

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

		release()
		r := <-follower
		require.NoError(t, r.err)
		_, err := png.DecodeConfig(bytes.NewReader(readPreview(t, r.preview)))
		require.NoError(t, err)
		assert.Zero(t, followerSource.opened.Load(), "the follower must share the flight instead of generating itself")
	})

	t.Run("a panicking decode is an error", func(t *testing.T) {
		g := &previewGenerator{}

		_, err := g.get(t.Context(), uniquePreviewFileID(t), func() (io.ReadCloser, error) { panic("broken decoder") }, PreviewSmall)

		require.ErrorContains(t, err, "broken decoder")
		assert.False(t, IsErrTaskAttachmentNotPreviewable(err))
	})
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
