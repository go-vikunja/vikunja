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

package apiv2

import (
	"io"
	"os"
	"strings"
	"testing"

	"code.vikunja.io/api/pkg/config"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestRegisterAllCreatesAllSchemaLinks guards against response body types Huma
// cannot build its $schema wrapper for. Huma prepends a $schema field via
// reflect.StructOf, which panics on embedded fields with a non-empty method
// set anywhere but position 0 (go#15924); Huma recovers, prints a warning to
// stderr and silently drops $schema + the Link describedBy header for that
// type. Typical trigger: a body embedding a model pointer, or a model whose
// value method set is non-empty because it doesn't shadow every method of its
// embedded web.CRUDable/web.Permissions interfaces (issue #3272).
func TestRegisterAllCreatesAllSchemaLinks(t *testing.T) {
	config.InitDefaultConfig()

	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stderr
	os.Stderr = w

	func() {
		defer func() { os.Stderr = orig; w.Close() }()
		e := echo.New()
		RegisterAll(NewAPI(e, e.Group(GroupPrefix)))
	}()

	out, err := io.ReadAll(r)
	require.NoError(t, err)
	for line := range strings.SplitSeq(string(out), "\n") {
		if strings.Contains(line, "unable to create schema link") {
			assert.Fail(t, "Huma could not create a schema link", "%s — the response body type embeds a type with a non-empty method set; see this test's doc comment", line)
		}
	}
}
