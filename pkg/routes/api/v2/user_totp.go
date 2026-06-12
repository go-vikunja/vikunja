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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/danielgtaylor/huma/v2"
	"xorm.io/xorm"
)

type totpStatusBody struct {
	Body *user.TOTP
}

type totpEnableBody struct {
	Body struct {
		Passcode string `json:"passcode" doc:"The current totp passcode, used to confirm the authenticator is set up correctly."`
	}
}

type totpDisableBody struct {
	Body struct {
		Password string `json:"password" doc:"The current user's password, required to disable totp."`
	}
}

type totpMessageBody struct {
	Body models.Message
}

// totpQrCodeResponse carries the qr code jpeg bytes plus a fixed Content-Type.
// Huma writes the []byte Body straight to the wire; the header field overrides
// content negotiation so image/jpeg reaches the client (matching v1).
type totpQrCodeResponse struct {
	ContentType string `header:"Content-Type"`
	Body        []byte
}

// RegisterTOTPRoutes wires the current-user totp (2FA) operations onto the Huma
// API. Totp is a local-account feature; every handler rejects OIDC/LDAP users.
func RegisterTOTPRoutes(api huma.API) {
	if !config.ServiceEnableTotp.GetBool() {
		return
	}

	tags := []string{"user"}

	Register(api, huma.Operation{
		OperationID: "totp-get",
		Summary:     "Get totp status",
		Description: "Returns the authenticated user's current totp setting. Fails with 412 if totp was never enrolled. Local accounts only.",
		Method:      http.MethodGet,
		Path:        "/user/settings/totp",
		Tags:        tags,
	}, totpGet)

	Register(api, huma.Operation{
		OperationID: "totp-enroll",
		Summary:     "Enroll into totp",
		Description: "Creates the totp secret for the authenticated user. The setup must still be confirmed via the enable endpoint before it takes effect. Local accounts only.",
		Method:      http.MethodPost,
		Path:        "/user/settings/totp/enroll",
		// v1 returns 200 here, not 201: enrollment is an inactive draft, not a usable resource yet.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, totpEnroll)

	Register(api, huma.Operation{
		OperationID: "totp-enable",
		Summary:     "Enable totp",
		Description: "Activates a previously enrolled totp setting by confirming a passcode. All existing sessions are invalidated. Local accounts only.",
		Method:      http.MethodPost,
		Path:        "/user/settings/totp/enable",
		// Confirms an existing enrollment; creates no new resource.
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, totpEnable)

	Register(api, huma.Operation{
		OperationID:   "totp-disable",
		Summary:       "Disable totp",
		Description:   "Removes all totp settings for the authenticated user. Requires the current password for confirmation. Local accounts only.",
		Method:        http.MethodPost,
		Path:          "/user/settings/totp/disable",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, totpDisable)

	Register(api, huma.Operation{
		OperationID: "totp-qrcode",
		Summary:     "Get the totp enrollment qr code",
		Description: "Returns the qr code for the authenticated user's enrolled totp setting as a jpeg image, for scanning into an authenticator app. Requires a prior enrollment. Local accounts only.",
		Method:      http.MethodGet,
		Path:        "/user/settings/totp/qrcode",
		Tags:        tags,
		// Spell out the binary response; a bare []byte Body would otherwise be
		// modeled as a base64 JSON string instead of binary image data.
		Responses: map[string]*huma.Response{
			"200": {
				Description: "The qr code as a jpeg image.",
				Content: map[string]*huma.MediaType{
					"image/jpeg": {
						Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
					},
				},
			},
		},
	}, totpQrCode)
}

func init() { AddRouteRegistrar(RegisterTOTPRoutes) }

// localUserFromCtx resolves the authenticated user and refuses anything that is
// not a local account, mirroring v1's getLocalUserFromContext. The caller owns
// the returned session. CheckUserPassword and IsLocalUser need the full DB
// record (password hash, issuer), so this loads it rather than trusting the
// token claims.
func localUserFromCtx(ctx context.Context) (*user.User, *xorm.Session, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, nil, err
	}

	s := db.NewSession()
	u, err := models.GetUserOrLinkShareUser(s, a)
	if err != nil {
		s.Close()
		return nil, nil, translateDomainError(err)
	}
	// A link share resolves to a synthetic, non-local user; any other auth type
	// yields nil. Both must be refused — totp is a real-account-only feature.
	if u == nil || !u.IsLocalUser() {
		s.Close()
		return nil, nil, translateDomainError(&user.ErrAccountIsNotLocal{})
	}

	return u, s, nil
}

func totpGet(ctx context.Context, _ *struct{}) (*totpStatusBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	t, err := user.GetTOTPForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &totpStatusBody{Body: t}, nil
}

func totpEnroll(ctx context.Context, _ *struct{}) (*totpStatusBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	t, err := user.EnrollTOTP(s, u)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &totpStatusBody{Body: t}, nil
}

func totpEnable(ctx context.Context, in *totpEnableBody) (*totpMessageBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	if err := user.EnableTOTP(s, &user.TOTPPasscode{User: u, Passcode: in.Body.Passcode}); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := models.DeleteAllUserSessions(s, u.ID); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &totpMessageBody{Body: models.Message{Message: "TOTP was enabled successfully."}}, nil
}

func totpDisable(ctx context.Context, in *totpDisableBody) (*totpMessageBody, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	if err := user.CheckUserPassword(u, in.Body.Password); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := user.DisableTOTP(s, u); err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &totpMessageBody{Body: models.Message{Message: "TOTP was disabled successfully."}}, nil
}

func totpQrCode(ctx context.Context, _ *struct{}) (*totpQrCodeResponse, error) {
	u, s, err := localUserFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	defer s.Close()

	qrcode, err := user.GetTOTPQrCodeAsJpegForUser(s, u)
	if err != nil {
		_ = s.Rollback()
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		return nil, translateDomainError(err)
	}
	return &totpQrCodeResponse{ContentType: "image/jpeg", Body: qrcode}, nil
}
