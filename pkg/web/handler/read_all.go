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

package handler

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"reflect"
	"strconv"

	vconfig "code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"

	"github.com/labstack/echo/v5"
)

// Shared with the MCP dispatcher, which renders them as plain tool-result text.
// The wording is the v1 400 response body, verbatim.
var (
	//nolint:revive,staticcheck // verbatim v1 API response text
	ErrNegativePage = errors.New("Page number cannot be negative.")
	//nolint:revive,staticcheck // verbatim v1 API response text
	ErrNegativePerPage = errors.New("Per page amount cannot be negative.")
)

// NormalizePagination clamps a (page, per_page) pair to what the models expect:
// page < 1 makes them drop the LIMIT clause, so an unset page becomes 1 and an
// unset or oversized per_page the configured maximum.
func NormalizePagination(page, perPage int) (int, int, error) {
	if page < 0 {
		return 0, 0, ErrNegativePage
	}
	if page == 0 {
		page = 1
	}
	if perPage < 0 {
		return 0, 0, ErrNegativePerPage
	}
	maxPerPage := vconfig.ServiceMaxItemsPerPage.GetInt()
	if perPage == 0 || perPage > maxPerPage {
		perPage = maxPerPage
	}
	return page, perPage, nil
}

// ReadAllWeb is the webhandler to get all objects of a type
func (c *WebHandler) ReadAllWeb(ctx *echo.Context) error {
	// Get our model
	currentStruct := c.EmptyStruct()

	currentAuth, err := auth.GetAuthFromClaims(ctx)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not determine the current user.").Wrap(err)
	}

	// Get the object & bind params to struct
	if err := ctx.Bind(currentStruct); err != nil {
		log.Debugf("Invalid model error. Internal error was: %s", err.Error())
		var he *echo.HTTPError
		if errors.As(err, &he) {
			return models.ErrInvalidModel{Message: fmt.Sprintf("%v", he.Message), Err: err}
		}
		return models.ErrInvalidModel{Err: err}
	}

	// Pagination
	page := ctx.QueryParam("page")
	if page == "" {
		page = "1"
	}
	pageNumber, err := strconv.Atoi(page)
	if err != nil {
		log.Error(err.Error())
		return echo.NewHTTPError(http.StatusBadRequest, "Bad page requested.").Wrap(err)
	}

	// Items per page
	var perPageNumber int
	perPage := ctx.QueryParam("per_page")
	// If we dont have an "items per page" parameter, we want to use the default.
	// To prevent Atoi from failing, we check this here.
	if perPage != "" {
		perPageNumber, err = strconv.Atoi(perPage)
		if err != nil {
			log.Error(err.Error())
			return echo.NewHTTPError(http.StatusBadRequest, "Bad per page amount requested.").Wrap(err)
		}
	}

	pageNumber, perPageNumber, err = NormalizePagination(pageNumber, perPageNumber)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).Wrap(err)
	}

	// Search
	search := ctx.QueryParam("s")

	result, resultCount, numberOfItems, err := DoReadAll(ctx.Request().Context(), currentStruct, currentAuth, search, pageNumber, perPageNumber)
	if err != nil {
		return err
	}

	// Calculate the number of pages from the number of items
	// We always round up, because if we don't have a number of items which is exactly dividable by the number of items per page,
	// we would get a result that is one page off.
	var numberOfPages = math.Ceil(float64(numberOfItems) / float64(perPageNumber))
	// If we don't have results, we don't have a page
	if resultCount == 0 {
		numberOfPages = 0
	}

	ctx.Response().Header().Set("x-pagination-total-pages", strconv.FormatFloat(numberOfPages, 'f', 0, 64))
	ctx.Response().Header().Set("x-pagination-result-count", strconv.FormatInt(int64(resultCount), 10))
	ctx.Response().Header().Set("Access-Control-Expose-Headers", "x-pagination-total-pages, x-pagination-result-count")

	// Ensure we return an empty array instead of null when there are no results.
	// We need to use reflection here because a nil slice wrapped in an interface{}
	// is not equal to nil (the interface contains a nil value but is not nil itself).
	if result == nil || (reflect.ValueOf(result).Kind() == reflect.Slice && reflect.ValueOf(result).IsNil()) {
		result = []interface{}{}
	}

	return ctx.JSON(http.StatusOK, result)
}
