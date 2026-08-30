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

package vikunjafile

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"crypto/rand"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io"
	"os"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/user"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildExportWithFiles(t *testing.T, data string, entries []struct {
	name    string
	content []byte
}) []byte {
	t.Helper()
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)
	for name, content := range map[string][]byte{"VERSION": []byte("dev"), "data.json": []byte(data)} {
		w, err := zw.Create(name)
		require.NoError(t, err)
		_, err = w.Write(content)
		require.NoError(t, err)
	}
	for _, entry := range entries {
		w, err := zw.Create(entry.name)
		require.NoError(t, err)
		_, err = w.Write(entry.content)
		require.NoError(t, err)
	}
	require.NoError(t, zw.Close())
	return zipBuf.Bytes()
}

func existingStorage(t *testing.T) int64 {
	t.Helper()
	s := db.NewSession()
	defer s.Close()
	var size int64
	_, err := s.Table("files").Select("COALESCE(SUM(size), 0)").Where("created_by_id = ?", 1).Get(&size)
	require.NoError(t, err)
	return size
}

const testLimitsDataJSON = `[{
	"id": 1,
	"title": "import project",
	"tasks": [{
		"id": 1,
		"title": "import task",
		"attachments": [{
			"id": 1,
			"file": {"id": 1, "name": "blob.bin", "size": 1}
		}],
		"comments": []
	}],
	"views": []
}]`

func TestErrVikunjaFileImportTooLargeHTTPErrorCode(t *testing.T) {
	httpErr := (&ErrVikunjaFileImportTooLarge{Reason: "test"}).HTTPError()
	assert.Equal(t, 14007, httpErr.Code)
}

func buildExportZip(t *testing.T, attachmentContent []byte) []byte {
	return buildExportWithFiles(t, testLimitsDataJSON, []struct {
		name    string
		content []byte
	}{{name: "files/1", content: attachmentContent}})
}

// buildLyingZipEntryZip forges metadata to exercise runtime byte accounting.
func buildLyingZipEntryZip(t *testing.T, declaredUncompressed uint64, actualUncompressed int) []byte {
	t.Helper()
	var zipBuf bytes.Buffer
	zw := zip.NewWriter(&zipBuf)

	vf, err := zw.Create("VERSION")
	require.NoError(t, err)
	_, err = vf.Write([]byte("dev"))
	require.NoError(t, err)

	df, err := zw.Create("data.json")
	require.NoError(t, err)
	_, err = df.Write([]byte(testLimitsDataJSON))
	require.NoError(t, err)

	// CreateRaw preserves the forged uncompressed size.
	var compressed bytes.Buffer
	fw, err := flate.NewWriter(&compressed, flate.DefaultCompression)
	require.NoError(t, err)
	_, err = fw.Write(bytes.Repeat([]byte{0x42}, actualUncompressed))
	require.NoError(t, err)
	require.NoError(t, fw.Close())

	fh := &zip.FileHeader{Name: "files/1", Method: zip.Deflate}
	fh.CompressedSize64 = uint64(compressed.Len())
	fh.UncompressedSize64 = declaredUncompressed
	raw, err := zw.CreateRaw(fh)
	require.NoError(t, err)
	_, err = raw.Write(compressed.Bytes())
	require.NoError(t, err)

	require.NoError(t, zw.Close())
	return zipBuf.Bytes()
}

func setLimits(t *testing.T, maxSize, maxUserStorage string, maxFiles int64) {
	t.Helper()
	config.MigrationVikunjaFileMaxSize.Set(maxSize)
	config.MigrationVikunjaFileMaxUserStorage.Set(maxUserStorage)
	config.MigrationVikunjaFileMaxFiles.Set(maxFiles)
	t.Cleanup(func() {
		config.MigrationVikunjaFileMaxSize.Set("256MB")
		config.MigrationVikunjaFileMaxUserStorage.Set("1GB")
		config.MigrationVikunjaFileMaxFiles.Set(10000)
	})
}

func runMigrate(t *testing.T, export []byte) error {
	t.Helper()
	db.LoadAndAssertFixtures(t)
	m := &FileMigrator{}
	u := &user.User{ID: 1}
	reader := bytes.NewReader(export)
	return m.Migrate(u, reader, int64(reader.Len()))
}

func assertTooLarge(t *testing.T, err error) {
	t.Helper()
	require.Error(t, err)
	var tooLarge *ErrVikunjaFileImportTooLarge
	require.ErrorAs(t, err, &tooLarge, "expected ErrVikunjaFileImportTooLarge, got %v", err)
}

// TestVikunjaFileLimits covers the import budgets from GHSA-w7jp-mf2v-8342.
func TestVikunjaFileLimits(t *testing.T) {
	t.Run("compressed expansion is rejected by the preflight", func(t *testing.T) {
		setLimits(t, "1MB", "1GB", 10000)

		content := make([]byte, 2*1024*1024)
		_, err := io.ReadFull(rand.Reader, content)
		require.NoError(t, err)

		err = runMigrate(t, buildExportZip(t, content))
		assertTooLarge(t, err)
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "import project"})
	})

	t.Run("lying metadata is rejected by the preflight, real reads stay budgeted", func(t *testing.T) {
		setLimits(t, "1MB", "1GB", 10000)

		export := buildLyingZipEntryZip(t, 2*1024*1024, 1024)
		err := runMigrate(t, export)
		assertTooLarge(t, err)
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "import project"})
		db.AssertMissing(t, "files", map[string]interface{}{"name": "blob.bin"})
	})

	t.Run("actual decompressed bytes are counted while reading", func(t *testing.T) {
		setLimits(t, "2KB", "1GB", 10000)

		b := &importBudget{remaining: 2 * 1024}
		require.NoError(t, b.count(1024))
		require.Error(t, b.count(1025), "the actual byte count must trip the budget")

		err := runMigrate(t, buildExportZip(t, []byte("hello")))
		require.NoError(t, err)
	})

	t.Run("the file count is bounded", func(t *testing.T) {
		setLimits(t, "256MB", "1GB", 2)

		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)
		for _, name := range []string{"VERSION", "data.json"} {
			w, err := zw.Create(name)
			require.NoError(t, err)
			_, err = w.Write([]byte("dev"))
			require.NoError(t, err)
		}
		for _, id := range []string{"1", "2", "3"} {
			w, err := zw.Create("files/" + id)
			require.NoError(t, err)
			_, err = w.Write([]byte("blob"))
			require.NoError(t, err)
		}
		require.NoError(t, zw.Close())

		err := runMigrate(t, zipBuf.Bytes())
		assertTooLarge(t, err)
	})

	t.Run("duplicate file ids are rejected", func(t *testing.T) {
		setLimits(t, "256MB", "1GB", 10000)
		export := buildExportWithFiles(t, testLimitsDataJSON, []struct {
			name    string
			content []byte
		}{
			{name: "files/1", content: []byte("first")},
			{name: "files/1", content: []byte("second")},
		})

		err := runMigrate(t, export)
		require.ErrorContains(t, err, "duplicate file id")
	})

	t.Run("duplicate entries cannot bypass the file count", func(t *testing.T) {
		setLimits(t, "256MB", "1GB", 1)
		export := buildExportWithFiles(t, testLimitsDataJSON, []struct {
			name    string
			content []byte
		}{
			{name: "files/1", content: []byte("first")},
			{name: "files/1", content: []byte("second")},
		})

		assertTooLarge(t, runMigrate(t, export))
	})

	t.Run("metadata does not consume the storage quota", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		quota := existingStorage(t) + int64(len("hello"))
		setLimits(t, "256MB", fmt.Sprintf("%dB", quota), 10000)

		err := runMigrate(t, buildExportZip(t, []byte("hello")))
		require.NoError(t, err)
		db.AssertExists(t, "files", map[string]interface{}{"name": "blob.bin", "size": 5}, false)
	})

	t.Run("actual file bytes enforce the storage quota", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setLimits(t, "256MB", fmt.Sprintf("%dB", existingStorage(t)+1024), 10000)

		err := runMigrate(t, buildExportZip(t, bytes.Repeat([]byte("x"), 8*1024)))
		assertTooLarge(t, err)
		db.AssertMissing(t, "files", map[string]interface{}{"name": "blob.bin"})
	})

	t.Run("stored background bytes enforce the storage quota", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)

		img := image.NewRGBA(image.Rect(0, 0, 512, 512))
		draw.Draw(img, img.Bounds(), &image.Uniform{C: color.RGBA{R: 12, G: 34, B: 56, A: 255}}, image.Point{}, draw.Src)
		var background bytes.Buffer
		require.NoError(t, png.Encode(&background, img))

		setLimits(t, "256MB", fmt.Sprintf("%dB", existingStorage(t)+int64(background.Len())), 10000)
		data := `[{"id":1,"title":"background quota project","background_information":{"id":1},"tasks":[],"views":[]}]`
		export := buildExportWithFiles(t, data, []struct {
			name    string
			content []byte
		}{{name: "files/1", content: background.Bytes()}})

		assertTooLarge(t, runMigrate(t, export))
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "background quota project"})
	})

	t.Run("reused entries debit storage on every open", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		setLimits(t, "256MB", fmt.Sprintf("%dB", existingStorage(t)+1000), 10000)
		data := `[{
			"id": 1,
			"title": "reused file project",
			"tasks": [{"id": 1, "title": "one", "attachments": [{"id": 1, "file": {"id": 1, "name": "one.bin"}}]},
			          {"id": 2, "title": "two", "attachments": [{"id": 2, "file": {"id": 1, "name": "two.bin"}}]}],
			"views": []
		}]`
		export := buildExportWithFiles(t, data, []struct {
			name    string
			content []byte
		}{{name: "files/1", content: bytes.Repeat([]byte("x"), 600)}})

		assertTooLarge(t, runMigrate(t, export))
		db.AssertMissing(t, "projects", map[string]interface{}{"title": "reused file project"})
	})

	t.Run("a failed import cleans up blobs it already created", func(t *testing.T) {
		db.LoadAndAssertFixtures(t)
		config.MigrationVikunjaFileMaxSize.Set("256KB")
		config.MigrationVikunjaFileMaxFiles.Set(10000)
		config.MigrationVikunjaFileMaxUserStorage.Set("1GB")
		t.Cleanup(func() {
			config.MigrationVikunjaFileMaxSize.Set("256MB")
		})

		dataJSON := `[{
			"id": 1,
			"title": "cleanup import project",
			"tasks": [{
				"id": 1,
				"title": "first task with a small attachment",
				"attachments": [{"id": 1, "file": {"id": 1, "name": "small.bin", "size": 1}}],
				"comments": []
			},{
				"id": 2,
				"title": "second task with the budget-breaking attachment",
				"attachments": [{"id": 2, "file": {"id": 2, "name": "big.bin", "size": 1}}],
				"comments": []
			}],
			"views": []
		}]`

		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)
		vf, err := zw.Create("VERSION")
		require.NoError(t, err)
		_, err = vf.Write([]byte("dev"))
		require.NoError(t, err)
		df, err := zw.Create("data.json")
		require.NoError(t, err)
		_, err = df.Write([]byte(dataJSON))
		require.NoError(t, err)
		sf, err := zw.Create("files/1")
		require.NoError(t, err)
		_, err = sf.Write(bytes.Repeat([]byte{0x61}, 32*1024))
		require.NoError(t, err)

		// Fail on the second entry after the first blob is written.
		var compressed bytes.Buffer
		fw, err := flate.NewWriter(&compressed, flate.DefaultCompression)
		require.NoError(t, err)
		_, err = fw.Write(bytes.Repeat([]byte{0x42}, 2*1024*1024))
		require.NoError(t, err)
		require.NoError(t, fw.Close())

		fh := &zip.FileHeader{Name: "files/2", Method: zip.Deflate}
		fh.CompressedSize64 = uint64(compressed.Len())
		fh.UncompressedSize64 = 100
		raw, err := zw.CreateRaw(fh)
		require.NoError(t, err)
		_, err = raw.Write(compressed.Bytes())
		require.NoError(t, err)
		require.NoError(t, zw.Close())

		// The next database ID identifies the blob written before rollback.
		var maxFileID int64
		qs := db.NewSession()
		_, err = qs.Table("files").Select("COALESCE(MAX(id), 0)").Get(&maxFileID)
		require.NoError(t, err)
		_ = qs.Close()

		// A leftover blob from an earlier import may already occupy
		// maxFileID+1; the new attachment rewrites it, and the cleanup must
		// remove it again. Find the first free slot to also prove the blob
		// really belongs to this import.
		probeID := maxFileID + 1
		for {
			_, statErr := files.FileStat(&files.File{ID: probeID})
			if statErr != nil {
				break
			}
			probeID++
		}

		err = runMigrate(t, zipBuf.Bytes())
		require.Error(t, err, "the lying zip entry must abort the import")

		db.AssertMissing(t, "projects", map[string]interface{}{"title": "cleanup import project"})
		db.AssertMissing(t, "files", map[string]interface{}{"name": "small.bin"})

		_, err = files.FileStat(&files.File{ID: probeID})
		require.Error(t, err, "the blob of the rolled-back attachment must be removed")
		assert.True(t, os.IsNotExist(err), "expected a missing blob, got %v", err)
	})

	t.Run("a small import with an attachment still works", func(t *testing.T) {
		setLimits(t, "256MB", "1GB", 10000)

		err := runMigrate(t, buildExportZip(t, []byte("hello attachment")))
		require.NoError(t, err)

		db.AssertExists(t, "projects", map[string]interface{}{
			"title":    "import project",
			"owner_id": 1,
		}, false)
		db.AssertExists(t, "files", map[string]interface{}{
			"name":          "blob.bin",
			"created_by_id": 1,
		}, false)
	})
}
