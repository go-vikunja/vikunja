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

package apiv2

import (
	"context"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type userDeletionPasswordBody struct {
	Body struct {
		Password string `json:"password" doc:"The authenticated user's password. Required for local users; ignored for users authenticated via an external provider."`
	}
}

type userDeletionConfirmBody struct {
	Body struct {
		Token string `json:"token" required:"true" minLength:"1" doc:"The deletion confirmation token from the email sent by the request-deletion endpoint."`
	}
}

func RegisterUserDeletionRoutes(api huma.API) {
	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID:   "user-deletion-request",
		Summary:       "Request account deletion",
		Description:   "Starts deletion of the authenticated user's account. Local users must provide their password; a confirmation email is then sent and deletion only proceeds once confirmed.",
		Method:        http.MethodPost,
		Path:          "/user/deletion/request",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, userDeletionRequest)

	Register(api, huma.Operation{
		OperationID:   "user-deletion-confirm",
		Summary:       "Confirm account deletion",
		Description:   "Confirms a requested account deletion using the token from the confirmation email and schedules the account for deletion.",
		Method:        http.MethodPost,
		Path:          "/user/deletion/confirm",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, userDeletionConfirm)

	Register(api, huma.Operation{
		OperationID:   "user-deletion-cancel",
		Summary:       "Cancel account deletion",
		Description:   "Cancels a scheduled account deletion. Local users must provide their password.",
		Method:        http.MethodPost,
		Path:          "/user/deletion/cancel",
		Tags:          tags,
		DefaultStatus: http.StatusNoContent,
	}, userDeletionCancel)
}

func init() { AddRouteRegistrar(RegisterUserDeletionRoutes) }

// authUserFromCtx resolves the full DB user for the authenticated caller, refusing
// link shares (which have no account to delete) with a 403.
func authUserFromCtx(ctx context.Context, s *xorm.Session) (*user.User, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	authUser, is := a.(*user.User)
	if !is {
		return nil, huma.Error403Forbidden("only users can manage account deletion")
	}
	// The auth user from the JWT claims is partial; re-fetch for the password hash.
	u, err := user.GetUserByID(s, authUser.ID)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return u, nil
}

func userDeletionRequest(ctx context.Context, in *userDeletionPasswordBody) (*emptyBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := authUserFromCtx(ctx, s)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if u.IsLocalUser() {
		if err := user.CheckUserPassword(u, in.Body.Password); err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	if err := user.RequestDeletion(s, u); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func userDeletionConfirm(ctx context.Context, in *userDeletionConfirmBody) (*emptyBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := authUserFromCtx(ctx, s)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if err := user.ConfirmDeletion(s, u, in.Body.Token); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

func userDeletionCancel(ctx context.Context, in *userDeletionPasswordBody) (*emptyBody, error) {
	s := db.NewSession()
	defer s.Close()

	u, err := authUserFromCtx(ctx, s)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}

	if u.IsLocalUser() {
		if err := user.CheckUserPassword(u, in.Body.Password); err != nil {
			_ = s.Rollback()
			return nil, translateDomainError(err)
		}
	}

	if err := user.CancelDeletion(s, u); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}

	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}
