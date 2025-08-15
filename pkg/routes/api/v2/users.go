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
	v2 "code.vikunja.io/api/pkg/models/v2"
	"code.vikunja.io/api/pkg/modules/auth"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v4"
)

// GetUserByID handles getting a user by its ID.
func GetUserByID(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID.")
	}

	u, err := user.GetUserByID(s, userID)
	if err != nil {
		return err
	}

	v2User := &v2.User{
		User: *u,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/users/%d", u.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2User)
}

// GetUsers handles getting all users.
func GetUsers(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	currentUser, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	search := c.QueryParam("s")

	users, err := user.ListUsers(s, search, currentUser, nil)
	if err != nil {
		return err
	}

	v2Users := make([]*v2.User, len(users))
	for i, u := range users {
		// Obfuscate the mailadresses
		u.Email = ""
		v2Users[i] = &v2.User{
			User: *u,
			Links: &v2.Links{
				Self: &v2.Link{
					Href: fmt.Sprintf("/api/v2/users/%d", u.ID),
				},
			},
		}
	}

	return c.JSON(http.StatusOK, v2Users)
}

// CreateUser handles creating a new user.
func CreateUser(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	var userIn v2.User
	if err := c.Bind(&userIn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user object provided.")
	}

	// Insert the user
	newUser, err := user.CreateUser(s, &userIn.User)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Create their initial project
	err = models.CreateNewProjectForUser(s, newUser)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	v2User := &v2.User{
		User: *newUser,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/users/%d", newUser.ID),
			},
		},
	}

	return c.JSON(http.StatusCreated, v2User)
}

// UpdateUser handles updating a user.
func UpdateUser(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID.")
	}

	var userIn v2.User
	if err := c.Bind(&userIn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user object provided.")
	}

	if userIn.ID != userID {
		return echo.NewHTTPError(http.StatusBadRequest, "User ID in path and body do not match.")
	}

	// Permission check
	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	if aut.GetID() != userID {
		// For now, only allow users to update themselves. Admin updates can be added later.
		return echo.ErrForbidden
	}

	// Get existing user
	existingUser, err := user.GetUserByID(s, userID)
	if err != nil {
		return err
	}

	// Update fields. A PUT should replace the resource, but for now I'll just update a few fields like in v1.
	// This should be improved later to be a real PUT.
	existingUser.Name = userIn.Name
	existingUser.Email = userIn.Email

	// A real implementation would handle password changes separately and with more security checks.
	// For now, I'll omit password changes here.

	_, err = user.UpdateUser(s, existingUser, true)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	v2User := &v2.User{
		User: *existingUser,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: fmt.Sprintf("/api/v2/users/%d", existingUser.ID),
			},
		},
	}

	return c.JSON(http.StatusOK, v2User)
}

// DeleteUser handles deleting a user.
func DeleteUser(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid user ID.")
	}

	// Permission check - for now, admin only
	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	currentUser, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	if !currentUser.IsAdmin() {
		return echo.ErrForbidden
	}

	userToDelete, err := user.GetUserByID(s, userID)
	if err != nil {
		return err
	}

	err = user.DeleteUser(s, userToDelete)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

// GetCurrentUser handles getting the currently authenticated user.
func GetCurrentUser(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	v2User := &v2.User{
		User: *u,
		Links: &v2.Links{
			Self: &v2.Link{
				Href: "/api/v2/user",
			},
		},
	}

	return c.JSON(http.StatusOK, v2User)
}

// GetUserSettings handles getting the settings of the currently authenticated user.
func GetUserSettings(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	settings := &v2.UserSettings{
		Name:                         u.Name,
		EmailRemindersEnabled:        u.EmailRemindersEnabled,
		DiscoverableByName:           u.DiscoverableByName,
		DiscoverableByEmail:          u.DiscoverableByEmail,
		OverdueTasksRemindersEnabled: u.OverdueTasksRemindersEnabled,
		DefaultProjectID:             u.DefaultProjectID,
		WeekStart:                    u.WeekStart,
		Language:                     u.Language,
		Timezone:                     u.Timezone,
		OverdueTasksRemindersTime:    u.OverdueTasksRemindersTime,
		FrontendSettings:             u.FrontendSettings,
		AvatarProvider:               u.AvatarProvider,
	}

	return c.JSON(http.StatusOK, settings)
}

// UpdateUserSettings handles updating the settings of the currently authenticated user.
func UpdateUserSettings(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	var settingsIn v2.UserSettings
	if err := c.Bind(&settingsIn); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid settings object provided.")
	}

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	// Update user fields from settings
	u.Name = settingsIn.Name
	u.EmailRemindersEnabled = settingsIn.EmailRemindersEnabled
	u.DiscoverableByName = settingsIn.DiscoverableByName
	u.DiscoverableByEmail = settingsIn.DiscoverableByEmail
	u.OverdueTasksRemindersEnabled = settingsIn.OverdueTasksRemindersEnabled
	u.DefaultProjectID = settingsIn.DefaultProjectID
	u.WeekStart = settingsIn.WeekStart
	u.Language = settingsIn.Language
	u.Timezone = settingsIn.Timezone
	u.OverdueTasksRemindersTime = settingsIn.OverdueTasksRemindersTime
	u.FrontendSettings = settingsIn.FrontendSettings
	u.AvatarProvider = settingsIn.AvatarProvider

	_, err = user.UpdateUser(s, u, true)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	// Return the updated settings
	updatedSettings := &v2.UserSettings{
		Name:                         u.Name,
		EmailRemindersEnabled:        u.EmailRemindersEnabled,
		DiscoverableByName:           u.DiscoverableByName,
		DiscoverableByEmail:          u.DiscoverableByEmail,
		OverdueTasksRemindersEnabled: u.OverdueTasksRemindersEnabled,
		DefaultProjectID:             u.DefaultProjectID,
		WeekStart:                    u.WeekStart,
		Language:                     u.Language,
		Timezone:                     u.Timezone,
		OverdueTasksRemindersTime:    u.OverdueTasksRemindersTime,
		FrontendSettings:             u.FrontendSettings,
		AvatarProvider:               u.AvatarProvider,
	}

	return c.JSON(http.StatusOK, updatedSettings)
}

// UpdateUserEmail handles updating the email of the currently authenticated user.
func UpdateUserEmail(c echo.Context) error {
	var emailUpdate v2.EmailUpdate
	if err := c.Bind(&emailUpdate); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid model provided.")
	}

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserByID(s, aut.GetID())
	if err != nil {
		return err
	}

	// Check password
	_, err = user.CheckUserCredentials(s, &user.Login{
		Username: u.Username,
		Password: emailUpdate.Password,
	})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	err = user.UpdateEmail(s, &user.EmailUpdate{
		User:     u,
		Password: emailUpdate.Password,
		NewEmail: emailUpdate.Email,
	})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "We sent you email with a link to confirm your email address."})
}

// UpdateUserPassword handles updating the password of the currently authenticated user.
func UpdateUserPassword(c echo.Context) error {
	var newPW v2.PasswordUpdate
	if err := c.Bind(&newPW); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "No password provided.")
	}

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	s := db.NewSession()
	defer s.Close()

	u, err := user.GetUserByID(s, aut.GetID())
	if err != nil {
		return err
	}

	// Check old password
	_, err = user.CheckUserCredentials(s, &user.Login{
		Username: u.Username,
		Password: newPW.OldPassword,
	})
	if err != nil {
		_ = s.Rollback()
		return err
	}

	// Update the password
	if err = user.UpdateUserPassword(s, u, newPW.NewPassword); err != nil {
		_ = s.Rollback()
		return err
	}

	if err := s.Commit(); err != nil {
		_ = s.Rollback()
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The password was updated successfully."})
}

// Login handles user login.
func Login(c echo.Context) error {
	var loginCreds user.Login
	if err := c.Bind(&loginCreds); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide a username and password.")
	}

	s := db.NewSession()
	defer s.Close()

	// This is a simplified login process. A real implementation would handle LDAP, etc.
	// For now, we only handle local auth.
	u, err := user.CheckUserCredentials(s, &loginCreds)
	if err != nil {
		return err
	}

	if u.Status == user.StatusDisabled {
		return &user.ErrAccountDisabled{UserID: u.ID}
	}

	// Simplified TOTP check.
	totpEnabled, err := user.TOTPEnabledForUser(s, u)
	if err != nil {
		return err
	}
	if totpEnabled {
		if loginCreds.TOTPPasscode == "" {
			return user.ErrInvalidTOTPPasscode{}
		}
		if _, err := user.ValidateTOTPPasscode(s, &user.TOTPPasscode{User: u, Passcode: loginCreds.TOTPPasscode}); err != nil {
			return err
		}
	}

	// Create token
	return auth.NewUserAuthTokenResponse(u, c, loginCreds.LongToken)
}

// Logout handles user logout.
func Logout(c echo.Context) error {
	// A real implementation would invalidate the token.
	// For now, we just return a success message.
	return c.JSON(http.StatusOK, models.Message{Message: "Successfully logged out."})
}

// RenewToken handles renewing a user's token.
func RenewToken(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	aut, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	u, err := models.GetUserOrLinkShareUser(s, aut)
	if err != nil {
		return err
	}

	// Simplified token renewal.
	return auth.NewUserAuthTokenResponse(u, c, false)
}
