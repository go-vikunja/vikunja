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

package models

import (
	"errors"
	"fmt"

	"code.vikunja.io/api/pkg/web"
	"xorm.io/xorm"
)

type ProjectUrgencyWeights struct {
	ProjectID      int64                  `json:"-" param:"project"`
	UrgencyWeights []ProjectUrgencyWeight `json:"urgency_weights"`
}

type ProjectUrgencyWeight struct {
	Property UrgencyProperty `json:"property"`
	Weight   float64         `json:"weight"`
	Filter   *BasicFilter    `json:"filter,omitempty"`
}

func (u *ProjectUrgencyWeights) CanRead(s *xorm.Session, auth web.Auth) (bool, int, error) {
	if isInstanceAdmin(s, auth) {
		return true, int(PermissionAdmin), nil
	}
	filterID := GetSavedFilterIDFromProjectID(u.ProjectID)
	if filterID > 0 {
		savedFilter := &SavedFilter{ID: filterID}
		return savedFilter.CanRead(s, auth)
	}
	project := &Project{ID: u.ProjectID}
	return project.CanRead(s, auth)
}

func (u *ProjectUrgencyWeights) CanUpdate(s *xorm.Session, auth web.Auth) (bool, error) {
	if isInstanceAdmin(s, auth) {
		return true, nil
	}
	filterID := GetSavedFilterIDFromProjectID(u.ProjectID)
	if filterID > 0 {
		savedFilter := &SavedFilter{ID: filterID}
		return savedFilter.CanUpdate(s, auth)
	}
	project := &Project{ID: u.ProjectID}
	return project.IsAdmin(s, auth)
}

func (u *ProjectUrgencyWeights) CanCreate(*xorm.Session, web.Auth) (bool, error) {
	return false, errors.New("not implemented")
}

func (u *ProjectUrgencyWeights) CanDelete(*xorm.Session, web.Auth) (bool, error) {
	return false, errors.New("not implemented")
}

// ReadAll returns the currently set project urgency weights
// @Summary Return project urgency weights
// @Description Returns the project's urgency weights.
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} ProjectUrgencyWeights
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/avatar [get]
func (u *ProjectUrgencyWeights) ReadAll(s *xorm.Session, _ web.Auth, _ string, _ int, _ int) (result any, resultCount int, numberOfTotalItems int64, err error) {
	// TODO should we implement pagination?
	// TODO check different user can't set or get weights from project they don't have access to
	weights, err := GetUrgencyWeights(s, u.ProjectID)
	if err != nil {
		return nil, 0, 0, err
	}
	u.UrgencyWeights = make([]ProjectUrgencyWeight, 0, len(weights))
	for _, weight := range weights {
		var filter *BasicFilter
		if weight.Filter != nil {
			filter = &BasicFilter{
				Query:        weight.Filter.Filter,
				IncludeNulls: weight.Filter.FilterIncludeNulls,
			}
		}
		var property UrgencyProperty
		if err := property.UnmarshalText([]byte(weight.Property)); err != nil {
			return nil, 0, 0, err
		}
		u.UrgencyWeights = append(u.UrgencyWeights, ProjectUrgencyWeight{
			Property: property,
			Weight:   weight.Weight,
			Filter:   filter,
		})
	}
	return *u, 1, 1, nil
}

// Update is the handler to change a project's urgency weights
// @Summary Change a project's urgency weights
// @tags project
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param urgency_weights body UrgencyWeights true "The updated project urgency weights"
// @Success 200 {object} models.Message
// @Failure 400 {object} web.HTTPError "Something's invalid."
// @Failure 500 {object} models.Message "Internal server error."
// @Router /user/settings/urgency_weights [post]
func (u *ProjectUrgencyWeights) Update(s *xorm.Session, _ web.Auth) error {
	// TODO check different user can't set or get weights from project they don't have access to
	var weights []UrgencyWeight
	for _, weight := range u.UrgencyWeights {
		var filter *TaskCollection
		if weight.Filter != nil {
			filter = &TaskCollection{
				Filter:             weight.Filter.Query,
				FilterIncludeNulls: weight.Filter.IncludeNulls,
			}
			if err := filter.ValidateFilterString(); err != nil {
				return ErrInvalidModel{Err: err}
			}
		}
		if weight.Weight < 1 {
			return ErrInvalidModel{Err: fmt.Errorf("property %q weight was %.2f, must be at least 1", weight.Property, weight.Weight)}
		}
		propertyName, err := weight.Property.MarshalText()
		if err != nil {
			return err
		}
		weights = append(weights, UrgencyWeight{
			Property: string(propertyName),
			Weight:   weight.Weight,
			Filter:   filter,
		})
	}
	return SetUrgencyWeights(s, u.ProjectID, weights)
}

func (*ProjectUrgencyWeights) Create(*xorm.Session, web.Auth) error {
	return errors.New("not implemented")
}

func (*ProjectUrgencyWeights) Delete(*xorm.Session, web.Auth) error {
	return errors.New("not implemented")
}

func (*ProjectUrgencyWeights) ReadOne(*xorm.Session, web.Auth) error {
	return errors.New("not implemented")
}
