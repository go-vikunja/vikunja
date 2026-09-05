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

package web

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var errCodeConstRegex = regexp.MustCompile(`\bErrCode(\w+)\s*=\s*(\d+)\b`)

// TestErrorCodesAreUnique guards the world-error code space, which is flat across all packages:
// clients (and the frontend translations) map a code to a message without knowing which package
// raised it, so two packages picking the same number silently mislabels one of them.
func TestErrorCodesAreUnique(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	pkgDir := filepath.Dir(filepath.Dir(thisFile))

	pkgFS := os.DirFS(pkgDir)
	codes := make(map[string][]string)
	err := fs.WalkDir(pkgFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		content, err := fs.ReadFile(pkgFS, path)
		if err != nil {
			return err
		}

		for _, match := range errCodeConstRegex.FindAllStringSubmatch(string(content), -1) {
			codes[match[2]] = append(codes[match[2]], "ErrCode"+match[1])
		}
		return nil
	})
	require.NoError(t, err)
	require.NotEmpty(t, codes, "no error code constants found, the test is not looking at the right place")

	duplicates := []string{}
	for code, names := range codes {
		unique := map[string]bool{}
		for _, name := range names {
			unique[name] = true
		}
		if len(unique) < 2 {
			continue
		}
		sorted := make([]string, 0, len(unique))
		for name := range unique {
			sorted = append(sorted, name)
		}
		sort.Strings(sorted)
		duplicates = append(duplicates, code+": "+strings.Join(sorted, ", "))
	}
	sort.Strings(duplicates)

	assert.Emptyf(t, duplicates, "error codes must be unique across all packages, got duplicates:\n%s", strings.Join(duplicates, "\n"))
}
