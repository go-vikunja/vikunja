// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package integrations

import (
	"os"
	"testing"

	"github.com/labstack/echo/v4"
)

var testRouter *echo.Echo

// TestMain is the main test function used to bootstrap the test env
func TestMain(m *testing.M) {
	var err error
	testRouter, err = setupTestEnv()
	if err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}
