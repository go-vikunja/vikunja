// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package v1

import (
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/web/handler"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

// GetListsByNamespaceID is the web handler to delete a namespace
// TODO: depricate this in favour of namespace.ReadOne() <-- should also return the lists
// @Summary Get all lists in a namespace
// @Description Returns all lists inside of a namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Param id path int true "Namespace ID"
// @Security JWTKeyAuth
// @Success 200 {array} models.List "The lists."
// @Failure 403 {object} models.Message "No access to that namespace."
// @Failure 404 {object} models.Message "The namespace does not exist."
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id}/lists [get]
func GetListsByNamespaceID(c echo.Context) error {
	// Get our namespace
	namespace, err := getNamespace(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	// Get the lists
	doer, err := user.GetCurrentUser(c)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}

	lists, err := models.GetListsByNamespaceID(namespace.ID, doer)
	if err != nil {
		return handler.HandleHTTPError(err, c)
	}
	return c.JSON(http.StatusOK, lists)
}

func getNamespace(c echo.Context) (namespace *models.Namespace, err error) {
	// Check if we have our ID
	id := c.Param("namespace")
	// Make int
	namespaceID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return
	}

	if namespaceID == -1 {
		namespace = &models.PseudoNamespace
		return
	}

	// Check if the user has acces to that namespace
	user, err := user.GetCurrentUser(c)
	if err != nil {
		return
	}
	namespace = &models.Namespace{ID: namespaceID}
	canRead, _, err := namespace.CanRead(user)
	if err != nil {
		return namespace, err
	}
	if !canRead {
		return nil, echo.ErrForbidden
	}
	return
}
