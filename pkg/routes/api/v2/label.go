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
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/labstack/echo/v4"
)

// RegisterLabels registers all label routes
func RegisterLabels(a *echo.Group) {
	labels := a.Group("/labels")
	labels.GET("", GetAllLabels)
	labels.POST("", CreateLabel)
	labels.GET("/:id", GetLabel)
	labels.PUT("/:id", UpdateLabel)
	labels.DELETE("/:id", DeleteLabel)
}

type LabelResponse struct {
	*models.Label
	Links *LabelLinks `json:"_links"`
}

type LabelLinks struct {
	Self string `json:"self"`
}

func labelResponse(label *models.Label) *LabelResponse {
	return &LabelResponse{
		Label: label,
		Links: &LabelLinks{
			Self: fmt.Sprintf("/api/v2/labels/%d", label.ID),
		},
	}
}

func labelsResponse(labels []*models.Label) []*LabelResponse {
	labelResponses := make([]*LabelResponse, len(labels))
	for i, label := range labels {
		labelResponses[i] = labelResponse(label)
	}
	return labelResponses
}

// GetAllLabels handles retrieving all labels for a user
func GetAllLabels(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelService := services.NewLabelService(s.Engine())
	labels, err := labelService.GetAll(s, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, labelsResponse(labels))
}

// CreateLabel creates a new label
func CreateLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	l := new(models.Label)
	if err := c.Bind(l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object provided.").SetInternal(err)
	}

	if err := c.Validate(l); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelService := services.NewLabelService(s.Engine())
	if err := labelService.Create(s, l, u); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusCreated, labelResponse(l))
}

// GetLabel retrieves a single label by its ID
func GetLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelService := services.NewLabelService(s.Engine())
	label, err := labelService.Get(s, labelID, u)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, labelResponse(label))
}

// UpdateLabel handles updating a label
func UpdateLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	updatePayload := new(models.Label)
	if err := c.Bind(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label object provided.").SetInternal(err)
	}
	updatePayload.ID = labelID

	if err := c.Validate(updatePayload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error()).SetInternal(err)
	}

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelService := services.NewLabelService(s.Engine())
	if err := labelService.Update(s, updatePayload, u); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.JSON(http.StatusOK, labelResponse(updatePayload))
}

// DeleteLabel handles deleting a label
func DeleteLabel(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	auth, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid label ID").SetInternal(err)
	}

	u, err := user.GetFromAuth(auth)
	if err != nil {
		return handler.HandleHTTPError(err)
	}

	labelService := services.NewLabelService(s.Engine())
	label := &models.Label{ID: labelID, CreatedByID: u.ID}
	if err := labelService.Delete(s, label, u); err != nil {
		return handler.HandleHTTPError(err)
	}

	if err := s.Commit(); err != nil {
		return handler.HandleHTTPError(err)
	}

	return c.NoContent(http.StatusNoContent)
}
