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

package feeds

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/i18n"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	"github.com/gorilla/feeds"
	"github.com/labstack/echo/v5"
)

const feedItemLimit = 50

// NotificationsAtomFeed serves the authenticated user's notifications as an
// Atom feed. Notifications are not marked as read by being fetched here.
func NotificationsAtomFeed(c *echo.Context) error {
	u, ok := c.Get("userBasicAuth").(*user.User)
	if !ok {
		return echo.NewHTTPError(http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
	}

	s := db.NewSession()
	defer s.Close()

	rows, _, _, err := notifications.GetNotificationsForUser(s, u.ID, feedItemLimit, 0)
	if err != nil {
		return err
	}

	publicURL := config.ServicePublicURL.GetString()
	feed := &feeds.Feed{
		Title:   i18n.T(u.Language, "feeds.notifications.title", u.GetName()),
		Link:    &feeds.Link{Href: publicURL + "feeds/notifications.atom"},
		Author:  &feeds.Author{Name: u.GetName()},
		Updated: time.Now(),
	}

	for _, row := range rows {
		typed, ok := notifications.Lookup(row.Name)
		if !ok {
			continue
		}

		raw, err := json.Marshal(row.Notification)
		if err != nil {
			continue
		}
		if err := json.Unmarshal(raw, typed); err != nil {
			continue
		}

		titler, ok := typed.(notifications.Titler)
		if !ok {
			continue
		}

		feed.Items = append(feed.Items, &feeds.Item{
			Id:      "vikunja-notification-" + strconv.FormatInt(row.ID, 10),
			Title:   titler.ToTitle(u.Language),
			Created: row.Created,
			Link:    &feeds.Link{Href: publicURL},
		})
	}

	atom, err := feed.ToAtom()
	if err != nil {
		return err
	}

	c.Response().Header().Set(echo.HeaderContentType, "application/atom+xml; charset=utf-8")
	return c.String(http.StatusOK, atom)
}
