// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package files

import "bytes"

// Dump dumps all saved files
// This only includes the raw files, no db entries.
func Dump() (allFiles map[int64][]byte, err error) {
	files := []*File{}
	err = x.Find(&files)
	if err != nil {
		return
	}

	allFiles = make(map[int64][]byte, len(files))
	for _, file := range files {
		if err := file.LoadFileByID(); err != nil {
			return nil, err
		}
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(file.File); err != nil {
			return nil, err
		}
		allFiles[file.ID] = buf.Bytes()
	}

	return
}
