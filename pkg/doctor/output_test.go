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

package doctor

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintGroup(t *testing.T) {
	var buf bytes.Buffer

	PrintHeader(&buf)
	PrintGroup(&buf, CheckGroup{
		Name: "Files (s3)",
		Results: []CheckResult{
			{Name: "Endpoint", Passed: true, Value: "http://localhost:9000"},
			{Name: "Initialization", Passed: false, Error: "S3 endpoint http://localhost:9000 did not respond within 12s"},
			{Name: "CORS", Passed: true, Value: "2 origins", Lines: []string{"a", "b"}},
		},
	})

	assert.Equal(t, `Vikunja Doctor
==============

Files (s3)
  ✓ Endpoint: http://localhost:9000
  ✗ Initialization: S3 endpoint http://localhost:9000 did not respond within 12s
  ✓ CORS: 2 origins
      a
      b

`, buf.String())
}
