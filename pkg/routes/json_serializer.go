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

package routes

import (
	"encoding/json"
	"io"
	"net/http"

	"code.vikunja.io/api/pkg/utils"

	"github.com/labstack/echo/v5"
)

type jsonSerializer struct {
	echo.DefaultJSONSerializer
}

func (jsonSerializer) Deserialize(c *echo.Context, target any) error {
	body, err := io.ReadAll(c.Request().Body)
	if err != nil {
		if echo.StatusCode(err) != 0 {
			return err
		}
		return echo.ErrBadRequest.Wrap(err)
	}
	if err := utils.CheckJSONForNUL(body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).Wrap(err)
	}
	if err := json.Unmarshal(body, target); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}
	return nil
}
