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
	"code.vikunja.io/api/pkg/modules/humabridge"
	"code.vikunja.io/api/pkg/routes/feeds"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humaecho"
	"github.com/labstack/echo/v5"
)

// CalendarFeedTokenParam is the query parameter carrying the feeds-scoped API
// token on the calendar feeds.
const CalendarFeedTokenParam = "token"

// RegisterCalendarFeedRoutes wires the iCalendar task feeds onto the Huma API.
// Calendar apps subscribe to a bare URL and send no auth header, so these ops
// take a feeds-scoped API token in the query string.
func RegisterCalendarFeedRoutes(api huma.API) {
	tags := []string{"service"}
	security := []map[string][]string{{"FeedTokenQuery": {}}}
	responses := map[string]*huma.Response{
		"200": {
			Description: "The tasks as an iCalendar feed.",
			Content: map[string]*huma.MediaType{
				"text/calendar": {
					Schema: &huma.Schema{Type: huma.TypeString, Format: "binary"},
				},
			},
		},
	}

	Register(api, huma.Operation{
		OperationID: "projects-calendar-feed",
		Summary:     "Project iCalendar feed",
		Description: "Returns the project's tasks as a read-only iCalendar feed (one VEVENT per task with a due, start or end date) for calendar apps to subscribe to. Authenticated with a feeds-scoped API token in the `token` query parameter; the token owner must be able to read the project. Done tasks are left out unless include_done is set.",
		Method:      http.MethodGet,
		Path:        "/projects/{project}/calendar.ics",
		Tags:        tags,
		Security:    security,
		Responses:   responses,
	}, projectCalendarFeed)
	Register(api, huma.Operation{
		OperationID: "user-calendar-feed",
		Summary:     "User iCalendar feed",
		Description: "Returns the tasks of every project the token owner can read as a read-only iCalendar feed. Same authentication and date mapping as the project feed.",
		Method:      http.MethodGet,
		Path:        "/user/calendar.ics",
		Tags:        tags,
		Security:    security,
		Responses:   responses,
	}, userCalendarFeed)
}

func init() { AddRouteRegistrar(RegisterCalendarFeedRoutes) }

type projectCalendarFeedInput struct {
	ProjectID   int64 `path:"project" doc:"The numeric id of the project."`
	IncludeDone bool  `query:"include_done" doc:"If true, done tasks are included in the feed."`
}

type userCalendarFeedInput struct {
	IncludeDone bool `query:"include_done" doc:"If true, done tasks are included in the feed."`
}

func projectCalendarFeed(ctx context.Context, in *projectCalendarFeedInput) (*huma.StreamResponse, error) {
	return calendarFeed(ctx, in.ProjectID, in.IncludeDone)
}

func userCalendarFeed(ctx context.Context, in *userCalendarFeedInput) (*huma.StreamResponse, error) {
	return calendarFeed(ctx, 0, in.IncludeDone)
}

// calendarFeed authenticates the query token itself: the path is in
// unauthenticatedAPIPaths, so the JWT middleware lets the request through and
// a query token is never accepted anywhere else.
func calendarFeed(ctx context.Context, projectID int64, includeDone bool) (*huma.StreamResponse, error) {
	c := humabridge.EchoContextFrom(ctx)
	if c == nil {
		return nil, huma.Error500InternalServerError("could not resolve request context")
	}

	s := db.NewSession()
	defer s.Close()

	u, err := feeds.AuthenticateFeedURLToken(s, (*c).QueryParam(CalendarFeedTokenParam))
	if err != nil {
		return nil, translateDomainError(err)
	}
	if u == nil {
		return nil, huma.Error401Unauthorized(http.StatusText(http.StatusUnauthorized))
	}

	ical, err := feeds.BuildTasksCalendar(s, u, projectID, includeDone)
	if err != nil {
		return nil, translateDomainError(err)
	}

	return &huma.StreamResponse{Body: func(hctx huma.Context) {
		ec := humaecho.Unwrap(hctx)
		(*ec).Response().Header().Set(echo.HeaderContentType, feeds.CalendarContentType)
		_, _ = (*ec).Response().Write([]byte(ical))
	}}, nil
}
