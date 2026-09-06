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

package entitlement

import (
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/web"
)

// Error codes 20xxx belong to this package.

// ErrLimitReached: the owner is at the limit, so nothing new may be created.
// Existing data stays usable.
type ErrLimitReached struct {
	Feature Feature
	Limit   int64
	Current int64
}

func IsErrLimitReached(err error) bool {
	_, ok := err.(ErrLimitReached)
	return ok
}

func (err ErrLimitReached) Error() string {
	return fmt.Sprintf("Limit for %s reached (limit: %d, current: %d)", err.Feature, err.Limit, err.Current)
}

const ErrCodeLimitReached = 20001

func (err ErrLimitReached) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeLimitReached,
		Message:  fmt.Sprintf("The limit for %s has been reached.", err.Feature),
	}
}

// ErrFeatureDisabledForUser: the instance offers the feature, but this user's plan does not include it.
type ErrFeatureDisabledForUser struct {
	Feature Feature
}

func IsErrFeatureDisabledForUser(err error) bool {
	_, ok := err.(ErrFeatureDisabledForUser)
	return ok
}

func (err ErrFeatureDisabledForUser) Error() string {
	return fmt.Sprintf("Feature %s is not enabled for this user", err.Feature)
}

const ErrCodeFeatureDisabledForUser = 20002

func (err ErrFeatureDisabledForUser) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusForbidden,
		Code:     ErrCodeFeatureDisabledForUser,
		Message:  fmt.Sprintf("The feature %s is not enabled for this user.", err.Feature),
	}
}

// ErrFeatureNotLicensed keeps the 404 the license gates always served, so
// unlicensed routes stay indistinguishable from unregistered ones.
type ErrFeatureNotLicensed struct {
	Feature Feature
}

func IsErrFeatureNotLicensed(err error) bool {
	_, ok := err.(ErrFeatureNotLicensed)
	return ok
}

func (err ErrFeatureNotLicensed) Error() string {
	return fmt.Sprintf("Feature %s is not licensed on this instance", err.Feature)
}

const ErrCodeFeatureNotLicensed = 20003

func (err ErrFeatureNotLicensed) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusNotFound,
		Code:     ErrCodeFeatureNotLicensed,
		Message:  "Not Found",
	}
}

type ErrUnknownFeature struct {
	Feature Feature
}

func IsErrUnknownFeature(err error) bool {
	_, ok := err.(ErrUnknownFeature)
	return ok
}

func (err ErrUnknownFeature) Error() string {
	return fmt.Sprintf("Unknown entitlement feature %q", err.Feature)
}

const ErrCodeUnknownFeature = 20004

func (err ErrUnknownFeature) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: http.StatusBadRequest,
		Code:     ErrCodeUnknownFeature,
		Message:  fmt.Sprintf("Unknown entitlement feature %q.", err.Feature),
	}
}
