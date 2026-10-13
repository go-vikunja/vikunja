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
