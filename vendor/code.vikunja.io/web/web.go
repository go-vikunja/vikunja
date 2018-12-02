//  Copyright (c) 2018 Vikunja and contributors.
//
//  This program is free software: you can redistribute it and/or modify
//  it under the terms of the GNU Lesser General Public License as published by
//  the Free Software Foundation, either version 3 of the License, or
//  (at your option) any later version.
//
//  This program is distributed in the hope that it will be useful,
//  but WITHOUT ANY WARRANTY; without even the implied warranty of
//  MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//  GNU Lesser General Public License for more details.
//
//  You should have received a copy of the GNU Lesser General Public License
//  along with this program.  If not, see <http://www.gnu.org/licenses/>.

package web

import "github.com/labstack/echo"

// Rights defines rights methods
type Rights interface {
	IsAdmin(Auth) bool
	CanWrite(Auth) bool
	CanRead(Auth) bool
	CanDelete(Auth) bool
	CanUpdate(Auth) bool
	CanCreate(Auth) bool
}

// CRUDable defines the crud methods
type CRUDable interface {
	Create(Auth) error
	ReadOne() error
	ReadAll(string, Auth, int) (interface{}, error)
	Update() error
	Delete() error
}

// HTTPErrorProcessor is executed when the defined error is thrown, it will make sure the user sees an appropriate error message and http status code
type HTTPErrorProcessor interface {
	HTTPError() HTTPError
}

// HTTPError holds informations about an http error
type HTTPError struct {
	HTTPCode int    `json:"-"`
	Code     int    `json:"code"`
	Message  string `json:"message"`
}

// Auth defines the authentication interface used to get some auth thing
type Auth interface {
	AuthDummy()
}

// Authprovider is a holder for the implementation of an authprovider by the application
type Authprovider interface {
	GetAuthObject(echo.Context) (Auth, error)
}

// Auths holds the authobject
type Auths struct {
	AuthObject func(echo.Context) (Auth, error)
}
