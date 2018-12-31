//  Vikunja is a todo-list application to facilitate your life.
//  Copyright 2018 Vikunja and contributors. All rights reserved.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU General Public License for more details.
//
//  You should have received a copy of the GNU General Public License
//  along with this program.  If not, see <https://www.gnu.org/licenses/>.

package models

import "code.vikunja.io/web"

// Create creates a new label
// @Summary Create a label
// @Description Creates a new label.
// @tags labels
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param label body models.Label true "The label object"
// @Success 200 {object} models.Label "The created label object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels [put]
func (l *Label) Create(a web.Auth) (err error) {
	u, err := getUserWithError(a)
	if err != nil {
		return
	}

	l.CreatedBy = u
	l.CreatedByID = u.ID

	_, err = x.Insert(l)
	return
}

// Update updates a label
// @Summary Update a label
// @Description Update an existing label. The user needs to be the creator of the label to be able to do this.
// @tags labels
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Label ID"
// @Param label body models.Label true "The label object"
// @Success 200 {object} models.Label "The created label object."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid label object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to update the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [put]
func (l *Label) Update() (err error) {
	_, err = x.ID(l.ID).Update(l)
	if err != nil {
		return
	}

	err = l.ReadOne()
	return
}

// Delete deletes a label
// @Summary Delete a label
// @Description Delete an existing label. The user needs to be the creator of the label to be able to do this.
// @tags labels
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Label ID"
// @Success 200 {object} models.Label "The label was successfully deleted."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "Not allowed to delete the label."
// @Failure 404 {object} code.vikunja.io/web.HTTPError "Label not found."
// @Failure 500 {object} models.Message "Internal error"
// @Router /labels/{id} [delete]
func (l *Label) Delete() (err error) {
	_, err = x.ID(l.ID).Delete(&Label{})
	return err
}
