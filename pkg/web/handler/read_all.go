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

package handler

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	currentAuth, err := config.AuthProvider.AuthObject(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.")
	}

	// Get the object & bind params to struct
	if err := ctx.Bind(currentStruct); err != nil {
		config.LoggingProvider.Debugf("Invalid model error. Internal error was: %s", err.Error())
		if he, is := err.(*echo.HTTPError); is {
			return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided. Error was: %s", he.Message))
		}
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid model provided."))
	}

	// Pagination
	page := ctx.QueryParam("page")
	if page == "" {
		page = "1"
	}
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		config.LoggingProvider.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Bad page requested.")
	}
	if pageNumber < 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Page number cannot be negative.")
	}

	// Items per page
	var perPageNumber int
	perPage := ctx.QueryParam("per_page")
	// If we dont have an "items per page" parameter, we want to use the default.
	// To prevent Atoi from failing, we check this here.
	if perPage != "" {
		perPageNumber, err = strconv.Atoi(perPage)
		if err != nil {
			config.LoggingProvider.Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, "Bad per page amount requested.")
		}
	}
	// Set default page count
	if perPageNumber == 0 {
		perPageNumber = config.MaxItemsPerPage
	}
	if perPageNumber < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Per page amount cannot be negative.")
	}
	if perPageNumber > config.MaxItemsPerPage {
		perPageNumber = config.MaxItemsPerPage
	}

	// Create the db session
	s := config.SessionFactory()
	defer func() {
		err = s.Close()
		if err != nil {
			config.LoggingProvider.Errorf("Could not close session: %s", err)
		}
	}()

	// Search
	search := ctx.QueryParam("s")

	result, resultCount, numberOfItems, err := currentStruct.ReadAll(s, currentAuth, search, pageNumber, perPageNumber)
	if err != nil {
		_ = s.Rollback()
		return HandleHTTPError(err, ctx)
	}

	// Calculate the number of pages from the number of items
	// We always round up, because if we don't have a number of items which is exactly dividable by the number of items per page,
	// we would get a result that is one page off.
	var numberOfPages = math.Ceil(float64(numberOfItems) / float64(perPageNumber))
	// If we return all results, we only have one page
	if pageNumber < 0 {
		numberOfPages = 1
	}
	// If we don't have results, we don't have a page
	if resultCount == 0 {
		numberOfPages = 0
	}

	ctx.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	ctx.Response().Header().Set("x-pagination-result-count", strconv.FormatInt(int64(resultCount), 10))
	ctx.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	err = s.Commit()
	if err != nil {
		return HandleHTTPError(err, ctx)
	}

	err = ctx.JSON(http.StatusOK, result)
	if err != nil {
		return HandleHTTPError(err, ctx)
	}
	return err
}
