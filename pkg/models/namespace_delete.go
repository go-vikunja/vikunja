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

import _ "code.vikunja.io/web" // For swaggerdocs generation

// Delete deletes a namespace
// @Summary Deletes a namespace
// @Description Delets a namespace
// @tags namespace
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "Namespace ID"
// @Success 200 {object} models.Message "The namespace was successfully deleted."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces/{id} [delete]
func (n *Namespace) Delete() (err error) {

	// Check if the namespace exists
	_, err = GetNamespaceByID(n.ID)
	if err != nil {
		return
	}

	// Delete the namespace
	_, err = x.ID(n.ID).Delete(&Namespace{})
	if err != nil {
		return
	}

	// Delete all lists with their tasks
	lists, err := GetListsByNamespaceID(n.ID)
	var listIDs []int64
	// We need to do that for here because we need the list ids to delete two times:
	// 1) to delete the lists itself
	// 2) to delete the list tasks
	for _, l := range lists {
		listIDs = append(listIDs, l.ID)
	}

	// Delete tasks
	_, err = x.In("list_id", listIDs).Delete(&ListTask{})
	if err != nil {
		return
	}

	// Delete the lists
	_, err = x.In("id", listIDs).Delete(&List{})
	if err != nil {
		return
	}

	return
}
