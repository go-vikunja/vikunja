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

package files

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"path"
	"slices"

	"code.vikunja.io/api/pkg/log"
)

type PreviewSize string

const (
	PreviewSizeUnknown PreviewSize = "unknown"
	PreviewSmall       PreviewSize = "sm"
	PreviewMedium      PreviewSize = "md"
	PreviewLarge       PreviewSize = "lg"
	PreviewExtraLarge  PreviewSize = "xl"
)

func (previewSize PreviewSize) GetSize() int {
	switch previewSize {
	case PreviewSmall:
		return 100
	case PreviewMedium:
		return 200
	case PreviewLarge:
		return 400
	case PreviewExtraLarge:
		return 800
	case PreviewSizeUnknown:
		return 0
	default:
		return 200
	}
}

const previewDir = "previews"

// Listed so all previews of a file can be removed without listing the storage.
var StoredPreviewSizes = []PreviewSize{
	PreviewSmall,
	PreviewMedium,
	PreviewLarge,
	PreviewExtraLarge,
}

func GetPreviewSizeFromString(size string) PreviewSize {
	if slices.Contains(StoredPreviewSizes, PreviewSize(size)) {
		return PreviewSize(size)
	}
	return PreviewSizeUnknown
}

func (f *File) previewPath(size PreviewSize) string {
	return path.Join(previewDir, f.fileID()+"_"+string(size))
}

func checkStoredPreviewSize(size PreviewSize) error {
	if !slices.Contains(StoredPreviewSizes, size) {
		return fmt.Errorf("previews of size %q are not stored", size)
	}
	return nil
}

// OpenPreview returns an error wrapping fs.ErrNotExist when no preview of that size is stored.
func (f *File) OpenPreview(size PreviewSize) (preview io.ReadCloser, length int64, err error) {
	if err := checkStoredPreviewSize(size); err != nil {
		return nil, 0, err
	}

	preview, length, err = storage.OpenWithSize(f.previewPath(size))
	if err != nil {
		return nil, 0, fmt.Errorf("failed to open the %s preview of file %d: %w", size, f.ID, err)
	}
	return preview, length, nil
}

func (f *File) SavePreview(size PreviewSize, preview []byte) error {
	if err := checkStoredPreviewSize(size); err != nil {
		return err
	}

	if err := storage.MkdirAll(previewDir, 0700); err != nil {
		return fmt.Errorf("failed to create the preview directory: %w", err)
	}
	if err := storage.Write(f.previewPath(size), bytes.NewReader(preview), uint64(len(preview))); err != nil {
		return fmt.Errorf("failed to save the %s preview of file %d: %w", size, f.ID, err)
	}

	// A Delete racing this write has already removed the previews.
	if _, err := storage.Stat(f.fileID()); errors.Is(err, fs.ErrNotExist) {
		f.removePreview(size)
		return ErrFileDoesNotExist{FileID: f.ID}
	}
	return nil
}

// removePreviews only logs failures: previews are derived data and must not fail what happens to their file.
func (f *File) removePreviews() {
	for _, size := range StoredPreviewSizes {
		f.removePreview(size)
	}
}

func (f *File) removePreview(size PreviewSize) {
	err := storage.Remove(f.previewPath(size))
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		log.Errorf("Could not remove the %s preview of file %d: %v", size, f.ID, err)
	}
}
