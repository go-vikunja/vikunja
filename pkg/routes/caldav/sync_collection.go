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

package caldav

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	caldavpkg "code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"github.com/labstack/echo/v5"
)

const (
	// syncTokenPrefix is the scheme prefix for sync tokens, matching what caldav-go generates.
	syncTokenPrefix = "data:,"
)

// parseSyncToken extracts the projectID and timestamp from a sync token.
// The token format is: data:,"PROJECT_ID-UNIX_TIMESTAMP" (with optional surrounding quotes).
// Returns (projectID, ts, true) on success, or (0, zero, false) on failure.
func parseSyncToken(token string) (projectID int64, ts time.Time, ok bool) {
	// Strip optional outer quotes
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)
	// Strip the data:, prefix
	token = strings.TrimPrefix(token, syncTokenPrefix)
	// Remove any residual quotes after prefix stripping
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)

	if token == "" {
		return 0, time.Time{}, false
	}

	// Format is "PROJECT_ID-UNIX_TIMESTAMP"
	// Project IDs may be negative (unlikely) but we split on the last '-' to be safe.
	idx := strings.LastIndex(token, "-")
	if idx < 0 {
		return 0, time.Time{}, false
	}

	projectIDStr := token[:idx]
	tsStr := token[idx+1:]

	pid, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return 0, time.Time{}, false
	}

	unixSec, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return 0, time.Time{}, false
	}

	return pid, time.Unix(unixSec, 0).UTC(), true
}

// extractSyncTokenFromBody pulls the <D:sync-token> text content from the
// raw REPORT body. Returns "" if not present.
func extractSyncTokenFromBody(body string) string {
	const openTag = "<sync-token>"
	const openTagNS = ":sync-token>"
	lower := strings.ToLower(body)

	// Handle both namespaced and un-namespaced variants
	for _, open := range []string{openTag, openTagNS} {
		startIdx := strings.Index(lower, strings.ToLower(open))
		if startIdx < 0 {
			continue
		}
		// Value starts right after the opening tag
		valueStart := startIdx + len(open)
		// Find the matching closing tag (look for '<' after valueStart)
		endIdx := strings.Index(body[valueStart:], "<")
		if endIdx < 0 {
			continue
		}
		return strings.TrimSpace(body[valueStart : valueStart+endIdx])
	}
	return ""
}

// requestsCalendarData returns true if the REPORT body includes a <calendar-data> prop request.
func requestsCalendarData(body string) bool {
	return strings.Contains(body, "calendar-data")
}

// writeForbiddenValidSyncToken writes the RFC 6578 §3.6 error response
// telling the client that its sync token is no longer valid. The client
// must perform a full resync with an empty token.
func writeForbiddenValidSyncToken(c *echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusForbidden)
	_, err := fmt.Fprint(c.Response(), `<?xml version="1.0" encoding="utf-8"?>`+"\n"+
		`<D:error xmlns:D="DAV:"><D:valid-sync-token/></D:error>`)
	return err
}

// handleSyncCollectionReport handles a CalDAV REPORT sync-collection request
// (RFC 6578). It replaces the default caldav-go behaviour which returns 412.
//
// Protocol overview:
//   - Empty sync-token  → return all tasks (initial sync or forced full resync)
//   - Valid sync-token  → return only tasks changed/created since the token timestamp
//     plus 404 entries for tasks deleted since then
//   - Invalid/unknown token → 403 + <D:valid-sync-token/> so client resets
func handleSyncCollectionReport(c *echo.Context, body string, storage *VikunjaCaldavProjectStorage) error {
	rawToken := extractSyncTokenFromBody(body)
	includeCalendarData := requestsCalendarData(body)

	projectID := storage.project.ID

	log.Debugf("[CALDAV sync-collection] project=%d token=%q calendar-data=%v", projectID, rawToken, includeCalendarData)

	// Load project and all its tasks
	rr, err := storage.getProjectRessource(true)
	if err != nil {
		log.Errorf("[CALDAV sync-collection] Failed to load project resource: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}

	// Determine the new sync token (based on latest task/project modification time)
	newEtag := rr.CalculateEtag()
	// newEtag is "PROJECT_ID-UNIX_TS" (with surrounding quotes stripped below)
	newEtagClean := strings.Trim(newEtag, `"`)
	newToken := syncTokenPrefix + `"` + newEtagClean + `"`

	// Case 1: empty token → full sync (all current tasks as "changed")
	if rawToken == "" {
		log.Debugf("[CALDAV sync-collection] project=%d empty token → full sync (%d tasks)", projectID, len(rr.projectTasks))
		return writeSyncResponse(c, rr, rr.projectTasks, nil, newToken, includeCalendarData)
	}

	// Case 2: non-empty token — try to parse it
	tokenProjectID, tokenTS, ok := parseSyncToken(rawToken)
	if !ok {
		// Unparseable token — tell client to start over
		log.Warningf("[CALDAV sync-collection] project=%d could not parse token %q → 403 valid-sync-token", projectID, rawToken)
		return writeForbiddenValidSyncToken(c)
	}

	// Token belongs to a different project — reject it
	if tokenProjectID != projectID {
		log.Warningf("[CALDAV sync-collection] token project %d != request project %d → 403 valid-sync-token", tokenProjectID, projectID)
		return writeForbiddenValidSyncToken(c)
	}

	// Case 3: valid token — compute delta
	var changedTasks []*models.TaskWithComments
	for _, t := range rr.projectTasks {
		if t.Updated.After(tokenTS) || t.Created.After(tokenTS) {
			changedTasks = append(changedTasks, t)
		}
	}

	// Fetch deletions since the token timestamp
	s := db.NewSession()
	defer s.Close()
	deletions, err := models.GetCaldavDeletionsSince(s, projectID, tokenTS)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("[CALDAV sync-collection] Failed to fetch deletions for project %d: %v", projectID, err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}
	if err := s.Commit(); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}

	log.Debugf("[CALDAV sync-collection] project=%d delta: %d changed tasks, %d deletions since %v → newToken=%q",
		projectID, len(changedTasks), len(deletions), tokenTS.UTC(), newToken)
	return writeSyncResponse(c, rr, changedTasks, deletions, newToken, includeCalendarData)
}

// writeSyncResponse writes the RFC 6578 §3.8/3.9 207 Multistatus response.
// tasks is the list of tasks to report as changed/present (may be empty for no-change delta).
// deletions is the list of tasks deleted since the sync token (may be nil).
func writeSyncResponse(
	c *echo.Context,
	rr VikunjaProjectResourceAdapter,
	tasks []*models.TaskWithComments,
	deletions []*models.TaskCaldavDeletion,
	newToken string,
	includeCalendarData bool,
) error {
	var sb strings.Builder
	sb.WriteString(`<?xml version="1.0" encoding="utf-8"?>` + "\n")
	sb.WriteString(`<D:multistatus xmlns:D="DAV:" xmlns:C="urn:ietf:params:xml:ns:caldav">` + "\n")

	for _, t := range tasks {
		href := getTaskURL(&t.Task)
		etag := `"` + strconv.FormatInt(t.ID, 10) + `-` + strconv.FormatInt(t.Updated.Unix(), 10) + `"`

		sb.WriteString("  <D:response>\n")
		sb.WriteString("    <D:href>" + xmlEscape(href) + "</D:href>\n")
		sb.WriteString("    <D:propstat>\n")
		sb.WriteString("      <D:prop>\n")
		sb.WriteString("        <D:getetag>" + xmlEscape(etag) + "</D:getetag>\n")
		if includeCalendarData {
			project := rr.project
			calData := caldavpkg.GetCaldavTodosForTasks(project, []*models.TaskWithComments{t})
			sb.WriteString("        <C:calendar-data>" + xmlEscape(calData) + "</C:calendar-data>\n")
		}
		sb.WriteString("      </D:prop>\n")
		sb.WriteString("      <D:status>HTTP/1.1 200 OK</D:status>\n")
		sb.WriteString("    </D:propstat>\n")
		sb.WriteString("  </D:response>\n")
	}

	// 404 entries for deleted tasks
	for _, d := range deletions {
		href := ProjectBasePath + "/" + strconv.FormatInt(d.ProjectID, 10) + "/" + d.UID + ".ics"
		sb.WriteString("  <D:response>\n")
		sb.WriteString("    <D:href>" + xmlEscape(href) + "</D:href>\n")
		sb.WriteString("    <D:status>HTTP/1.1 404 Not Found</D:status>\n")
		sb.WriteString("  </D:response>\n")
	}

	sb.WriteString("  <D:sync-token>" + xmlEscape(newToken) + "</D:sync-token>\n")
	sb.WriteString("</D:multistatus>\n")

	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusMultiStatus)
	_, err := fmt.Fprint(c.Response(), sb.String())
	return err
}

// xmlEscape escapes the five XML special characters.
func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
