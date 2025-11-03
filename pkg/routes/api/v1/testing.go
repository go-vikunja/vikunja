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
	"code.vikunja.io/api/pkg/log"

	"github.com/labstack/echo/v4"
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
func HandleTesting(c echo.Context) error {
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

	truncate := c.QueryParam("truncate")
	if truncate == "true" || truncate == "" {
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

	return c.JSON(http.StatusCreated, nil)
}
