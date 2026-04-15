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

package v1

import (
	"bytes"
	"encoding/json"
	"net/http"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v5"
)

// HandleTesting is the web handler to reset the db
// @Summary Reset the db to a defined state
// @Description Fills the specified table with the content provided in the payload. You need to enable the testing endpoint before doing this and provide the `Authorization: <token>` secret when making requests to this endpoint. See docs for more details.
// @tags testing
// @Accept json
// @Produce json
// @Param table path string true "The table to reset"
// @Success 201 {array} user.User "Everything has been imported successfully."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /test/{table} [patch]
func HandleTesting(c *echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token != config.ServiceTestingtoken.GetString() {
		return echo.ErrForbidden
	}

	table := c.Param("table")

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(c.Request().Body); err != nil {
		return err
	}

	content := []map[string]interface{}{}
	err := json.Unmarshal(buf.Bytes(), &content)
	if err != nil {
		log.Errorf("Error replacing table data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Wait for all async event handlers from the previous test to complete
	// before modifying the database. Without this, handlers hold SQLite
	// connections and starve this request's truncate/insert operations.
	events.WaitForPendingHandlers()

	truncate := c.QueryParam("truncate")
	if truncate == "true" || truncate == "" {
		// When truncating certain tables, also truncate dependent tables
		// whose rows reference the truncated table by user/entity ID.
		// Without foreign key cascades, stale rows would persist and
		// pollute subsequent tests that reuse the same auto-increment IDs.
		dependentTables := map[string][]string{
			"users": {"notifications"},
		}
		if deps, ok := dependentTables[table]; ok {
			for _, dep := range deps {
				if err = db.RestoreAndTruncate(dep, nil); err != nil {
					log.Errorf("Error truncating dependent table %s: %v", dep, err)
					return c.JSON(http.StatusInternalServerError, map[string]interface{}{
						"error":   true,
						"message": err.Error(),
					})
				}
			}
		}
		err = db.RestoreAndTruncate(table, content)
	} else {
		err = db.Restore(table, content)
	}

	if err != nil {
		log.Errorf("Error replacing table data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		})
	}

	// Seeding the license_status table only updates the DB row — the in-memory
	// license state is populated once at startup. Re-apply from cache so tests
	// that seed a valid Response get the licensed features without restarting.
	if table == "license_status" {
		if err := license.ReloadFromCache(); err != nil {
			log.Errorf("Error reloading license from seeded cache: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]interface{}{
				"error":   true,
				"message": err.Error(),
			})
		}
	}

	s := db.NewSession()
	defer s.Close()
	data := []map[string]interface{}{}
	err = s.Table(table).Find(&data)
	if err != nil {
		log.Errorf("Error fetching table data: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, data)
}

// HandleTestingTruncateAll truncates all tables in the database
// @Summary Truncate all tables
// @Description Removes all data from every Vikunja table. Used by e2e tests to ensure clean state before each test. Requires the testing token.
// @tags testing
// @Produce json
// @Success 200 {object} map[string]string "All tables truncated."
// @Failure 403 {object} web.HTTPError "Forbidden"
// @Failure 500 {object} models.Message "Internal server error."
// @Router /test/all [delete]
func HandleTestingTruncateAll(c *echo.Context) error {
	token := c.Request().Header.Get("Authorization")
	if token != config.ServiceTestingtoken.GetString() {
		return echo.ErrForbidden
	}

	events.WaitForPendingHandlers()

	if err := db.TruncateAllTables(); err != nil {
		log.Errorf("Error truncating all tables: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error":   true,
			"message": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "ok",
	})
}
