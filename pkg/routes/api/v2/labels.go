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

package v2

import (
	"fmt"
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v4"
)

// DeleteLabel handles deleting a label.
func DeleteLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID.")
	}

	l := &models.Label{ID: labelID}
	if err := l.Delete(s, u); err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateLabel handles updating a label.
func UpdateLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID.")
	}

	var l models.Label
	if err := c.Bind(&l); err != nil {
		return err
	}
	l.ID = labelID

	if err := l.Update(s, u); err != nil {
		return err
	}

	v2Label := &v2.Label{
		Label: l,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/labels/%d", l.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Label)
}

// GetLabel handles getting a label by its ID.
func GetLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID.")
	}

	l := &models.Label{ID: labelID}
	if err := l.ReadOne(s, u); err != nil {
		return err
	}

	v2Label := &v2.Label{
		Label: *l,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/labels/%d", l.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2Label)
}

// CreateLabel handles creating a new label.
func CreateLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	u, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	var l models.Label
	if err := c.Bind(&l); err != nil {
		return err
	}

	if err := l.Create(s, u); err != nil {
		return err
	}

	v2Label := &v2.Label{
		Label: l,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/labels/%d", l.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2Label)
}

// GetLabels handles getting all labels for the current user.
func GetLabels(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	page, perPage := v2.GetPageAndPerPage(c)
	search := c.QueryParam("s")

	labelsInterface, _, _, err := models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		Search:              []string{search},
		User:                aut,
		Page:                page,
		PerPage:             perPage,
		GetUnusedLabels:     true,
		GroupByLabelIDsOnly: true,
		GetForUser:          true,
	})
	if err != nil {
		return err
	}
	labels, ok := labelsInterface.([]*models.Label)
	if !ok {
		return fmt.Errorf("could not convert labels to []*models.Label")
	}

	v2Labels := make([]*v2.Label, len(labels))
	for i, l := range labels {
		v2Labels[i] = &v2.Label{
			Label: *l,
			Links: &v2.Links{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/labels/%d", l.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Labels)
}
