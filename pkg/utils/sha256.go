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
	"crypto/sha256"
	"encoding/hex"
)

// Sha256Hex returns the full hex-encoded SHA-256 hash of a string.
func Sha256Hex(cleartext string) string {
	h := sha256.Sum256([]byte(cleartext))
	return hex.EncodeToString(h[:])
}

// Sha256 calculates a sha256 hash from a string, truncated to 45 characters.
func Sha256(cleartext string) string {
	return Sha256Hex(cleartext)[:45]
}
