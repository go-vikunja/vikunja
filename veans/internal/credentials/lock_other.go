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

//go:build !unix

package credentials

import "os"

// Non-Unix platforms (Windows) don't get advisory file locking. Two
// concurrent `veans login` runs on the same Windows machine can race and
// lose a token; in practice veans runs from agent-driven shells on Linux
// or macOS, so this trade-off is acceptable.
func flockExclusive(*os.File) error { return nil }
func flockUnlock(*os.File) error    { return nil }
