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
	"archive/zip"
	"fmt"
	"io"
	"strconv"
)

// Change to deflate to gain better compression
// see https://pkg.go.dev/archive/zip#pkg-constants
const CompressionUsed = zip.Deflate

func WriteBytesToZip(filename string, data []byte, writer *zip.Writer) (err error) {
	header := &zip.FileHeader{
		Name:   filename,
		Method: CompressionUsed,
	}
	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return
}

// WriteFilesToZip writes a bunch of files from the db to a zip file. It expects a map with the file id
// as key and its content as io.ReadCloser.
func WriteFilesToZip(files map[int64]io.ReadCloser, wr *zip.Writer) (err error) {
	for fid, file := range files {
		header := &zip.FileHeader{
			Name:   "files/" + strconv.FormatInt(fid, 10),
			Method: CompressionUsed,
		}
		w, err := wr.CreateHeader(header)
		if err != nil {
			return err
		}
		_, err = io.Copy(w, file)
		if err != nil {
			return fmt.Errorf("error writing file %d: %w", fid, err)
		}
		_ = file.Close()
	}

	return nil
}
