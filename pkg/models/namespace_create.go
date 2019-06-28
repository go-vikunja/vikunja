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

import (
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/web"
)

// Create implements the creation method via the interface
// @Summary Creates a new namespace
// @Description Creates a new namespace.
// @tags namespace
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param namespace body models.Namespace true "The namespace you want to create."
// @Success 200 {object} models.Namespace "The created namespace."
// @Failure 400 {object} code.vikunja.io/web.HTTPError "Invalid namespace object provided."
// @Failure 403 {object} code.vikunja.io/web.HTTPError "The user does not have access to the namespace"
// @Failure 500 {object} models.Message "Internal error"
// @Router /namespaces [put]
func (n *Namespace) Create(a web.Auth) (err error) {
	// Check if we have at least a name
	if n.Name == "" {
		return ErrNamespaceNameCannotBeEmpty{NamespaceID: 0, UserID: a.GetID()}
	}
	n.ID = 0 // This would otherwise prevent the creation of new lists after one was created

	// Check if the User exists
	n.Owner, err = GetUserByID(a.GetID())
	if err != nil {
		return
	}
	n.OwnerID = n.Owner.ID

	// Insert
	if _, err = x.Insert(n); err != nil {
		return err
	}

	metrics.UpdateCount(1, metrics.NamespaceCountKey)
	return
}
