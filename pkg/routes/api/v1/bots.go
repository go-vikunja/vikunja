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

package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

func fetchOwnedBot(s *xorm.Session, c *echo.Context, caller *user.User) (*user.User, error) {
	botID, err := strconv.ParseInt(c.Param("bot"), 10, 64)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusBadRequest, "invalid bot id")
	}
	bot, err := user.GetUserByID(s, botID)
	if err != nil {
		return nil, err
	}
	if bot.BotOwnerID != caller.ID {
		return nil, &user.ErrBotNotOwned{UserID: botID}
	}
	return bot, nil
}

// CreateBotToken creates a new api token owned by a bot user.
// @Summary Create a new api token for a bot user
// @Description Creates a new api token owned by the specified bot. The bot must be owned by the calling user.
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param bot path int true "Bot user ID"
// @Param token body models.APIToken true "The token object with required fields"
// @Success 201 {object} models.APIToken "The created token."
// @Failure 400 {object} web.HTTPError "Invalid input."
// @Failure 403 {object} web.HTTPError "You do not own this bot."
// @Router /user/bots/{bot}/tokens [put]
func CreateBotToken(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	caller, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	bot, err := fetchOwnedBot(s, c, caller)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	token := &models.APIToken{}
	if err := c.Bind(token); err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	token.OwnerID = bot.ID
	if err := token.Create(s, caller); err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusCreated, token)
}

// ListBotTokens returns all api tokens owned by a bot user.
// @Summary List api tokens for a bot user
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param bot path int true "Bot user ID"
// @Success 200 {array} models.APIToken
// @Failure 403 {object} web.HTTPError "You do not own this bot."
// @Router /user/bots/{bot}/tokens [get]
func ListBotTokens(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	caller, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	bot, err := fetchOwnedBot(s, c, caller)
	if err != nil {
		_ = s.Rollback()
		return err
	}

	tokens := []*models.APIToken{}
	if err := s.Where("owner_id = ?", bot.ID).Find(&tokens); err != nil {
		_ = s.Rollback()
		return err
	}
	if err := s.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, tokens)
}

// DeleteBotToken deletes an api token owned by a bot user.
// @Summary Delete an api token for a bot user
// @tags api
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param bot path int true "Bot user ID"
// @Param token path int true "Token ID"
// @Success 200 {object} models.Message
// @Failure 403 {object} web.HTTPError "You do not own this bot."
// @Failure 404 {object} web.HTTPError "The token does not exist."
// @Router /user/bots/{bot}/tokens/{token} [delete]
func DeleteBotToken(c *echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	caller, err := user.GetCurrentUserFromDB(s, c)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	bot, err := fetchOwnedBot(s, c, caller)
	if err != nil {
		_ = s.Rollback()
		return err
	}
	tokenID, err := strconv.ParseInt(c.Param("token"), 10, 64)
	if err != nil {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusBadRequest, "invalid token id")
	}

	deleted, err := s.Where("id = ? AND owner_id = ?", tokenID, bot.ID).Delete(&models.APIToken{})
	if err != nil {
		_ = s.Rollback()
		return err
	}
	if deleted == 0 {
		_ = s.Rollback()
		return echo.NewHTTPError(http.StatusNotFound, "token not found")
	}
	if err := s.Commit(); err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &models.Message{Message: "The token was deleted successfully."})
}
