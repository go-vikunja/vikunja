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
	// Token scheme prefix, matching what caldav-go generates.
	syncTokenPrefix = "data:,"
)

// parseSyncToken extracts the project ID and timestamp from a sync token
// of the form data:,"PROJECT_ID-UNIX_TIMESTAMP".
func parseSyncToken(token string) (projectID int64, ts time.Time, ok bool) {
	// Quotes may sit outside or inside the data:, prefix.
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)
	token = strings.TrimPrefix(token, syncTokenPrefix)
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)

	if token == "" {
		return 0, time.Time{}, false
	}

	// Split on the last '-' so a negative project ID can't break parsing.
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

	for _, open := range []string{openTag, openTagNS} {
		startIdx := strings.Index(lower, strings.ToLower(open))
		if startIdx < 0 {
			continue
		}
		valueStart := startIdx + len(open)
		endIdx := strings.Index(body[valueStart:], "<")
		if endIdx < 0 {
			continue
		}
		return strings.TrimSpace(body[valueStart : valueStart+endIdx])
	}
	return ""
}

func requestsCalendarData(body string) bool {
	return strings.Contains(body, "calendar-data")
}

// writeForbiddenValidSyncToken writes the RFC 6578 §3.6 error response,
// telling the client to full-resync with an empty token.
func writeForbiddenValidSyncToken(c *echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/xml; charset=utf-8")
	c.Response().WriteHeader(http.StatusForbidden)
	_, err := fmt.Fprint(c.Response(), `<?xml version="1.0" encoding="utf-8"?>`+"\n"+
		`<D:error xmlns:D="DAV:"><D:valid-sync-token/></D:error>`)
	return err
}

// handleSyncCollectionReport handles a CalDAV REPORT sync-collection
// request (RFC 6578):
//   - Empty sync-token  → return all tasks (initial sync or forced full resync)
//   - Valid sync-token  → return only tasks changed/created since the token timestamp
//     plus 404 entries for tasks deleted since then
//   - Invalid/unknown token → 403 + <D:valid-sync-token/> so client resets
func handleSyncCollectionReport(c *echo.Context, body string, storage *VikunjaCaldavProjectStorage) error {
	rawToken := extractSyncTokenFromBody(body)
	includeCalendarData := requestsCalendarData(body)

	projectID := storage.project.ID

	log.Debugf("[CALDAV sync-collection] project=%d token=%q calendar-data=%v", projectID, rawToken, includeCalendarData)

	rr, err := storage.getProjectRessource(true)
	if err != nil {
		log.Errorf("[CALDAV sync-collection] Failed to load project resource: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError, "internal server error").Wrap(err)
	}

	// The etag is "PROJECT_ID-UNIX_TS"; the new sync token reuses it.
	newEtag := rr.CalculateEtag()
	newEtagClean := strings.Trim(newEtag, `"`)
	newToken := syncTokenPrefix + `"` + newEtagClean + `"`

	if rawToken == "" {
		log.Debugf("[CALDAV sync-collection] project=%d empty token → full sync (%d tasks)", projectID, len(rr.projectTasks))
		return writeSyncResponse(c, rr, rr.projectTasks, nil, newToken, includeCalendarData)
	}

	tokenProjectID, tokenTS, ok := parseSyncToken(rawToken)
	if !ok {
		log.Warningf("[CALDAV sync-collection] project=%d could not parse token %q → 403 valid-sync-token", projectID, rawToken)
		return writeForbiddenValidSyncToken(c)
	}

	if tokenProjectID != projectID {
		log.Warningf("[CALDAV sync-collection] token project %d != request project %d → 403 valid-sync-token", tokenProjectID, projectID)
		return writeForbiddenValidSyncToken(c)
	}

	// Tokens older than the soft-delete retention window cannot see deletions
	// whose tasks were already purged — force a full resync instead of
	// silently serving an incomplete delta.
	if time.Since(tokenTS) > models.TaskDeleteRetention {
		log.Debugf("[CALDAV sync-collection] project=%d token %q older than retention → 403 valid-sync-token", projectID, rawToken)
		return writeForbiddenValidSyncToken(c)
	}

	// Inclusive comparison: tokens have second granularity, so strict After
	// would miss a change in the same second the token was minted.
	// Re-reporting is cheap — the etag is unchanged, so clients skip the download.
	var changedTasks []*models.TaskWithComments
	for _, t := range rr.projectTasks {
		if !t.Updated.Before(tokenTS) || !t.Created.Before(tokenTS) {
			changedTasks = append(changedTasks, t)
		}
	}

	s := db.NewSession()
	defer s.Close()
	deletions, err := models.GetDeletedTasksSince(s, projectID, tokenTS)
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

// writeSyncResponse writes the RFC 6578 §3.8/3.9 207 Multistatus response:
// changed tasks as 200 entries, deleted tasks as 404 entries.
func writeSyncResponse(
	c *echo.Context,
	rr VikunjaProjectResourceAdapter,
	tasks []*models.TaskWithComments,
	deletions []*models.Task,
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

	for _, d := range deletions {
		href := getTaskURL(d)
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

func xmlEscape(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&#39;")
	return s
}
