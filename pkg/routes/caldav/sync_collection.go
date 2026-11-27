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
	"net/http"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
	"github.com/labstack/echo/v5"
	"xorm.io/xorm"
)

const (
	// Token scheme prefix, matching what caldav-go generates.
	syncTokenPrefix = "data:,"
)

func parseSyncToken(token string) (projectID int64, ts time.Time, purgeGeneration int64, ok bool) {
	// Quotes may sit outside or inside the data:, prefix.
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)
	token = strings.TrimPrefix(token, syncTokenPrefix)
	token = strings.TrimPrefix(token, `"`)
	token = strings.TrimSuffix(token, `"`)

	if timestamp, generation, found := strings.Cut(token, ";purge="); found {
		var err error
		purgeGeneration, err = strconv.ParseInt(generation, 10, 64)
		if err != nil || purgeGeneration <= 0 {
			return 0, time.Time{}, 0, false
		}
		token = timestamp
	}

	if token == "" {
		return 0, time.Time{}, 0, false
	}

	// Split on the last '-' so a negative project ID can't break parsing.
	idx := strings.LastIndex(token, "-")
	if idx < 0 {
		return 0, time.Time{}, 0, false
	}

	projectIDStr := token[:idx]
	tsStr := token[idx+1:]

	pid, err := strconv.ParseInt(projectIDStr, 10, 64)
	if err != nil {
		return 0, time.Time{}, 0, false
	}

	unixSec, err := strconv.ParseInt(tsStr, 10, 64)
	if err != nil {
		return 0, time.Time{}, 0, false
	}

	return pid, time.Unix(unixSec, 0).UTC(), purgeGeneration, true
}

// RFC 6578 §3.6: tells the client to full-resync with an empty token.
func writeForbiddenValidSyncToken(c *echo.Context) error {
	return writeDAVError(c, http.StatusForbidden, davName("valid-sync-token"))
}

// handleSyncCollectionReport answers a sync-collection REPORT (RFC 6578) on a collection
// the user can read:
//   - Empty sync-token  → return all tasks (initial sync or forced full resync)
//   - Valid sync-token  → return only tasks changed/created since the token timestamp
//     plus 404 entries for tasks deleted since then
//   - Invalid/unknown token → 403 + <D:valid-sync-token/> so client resets
func handleSyncCollectionReport(c *echo.Context, s *xorm.Session, u *user.User, collection readableCollection, req syncCollectionRequest, ctx propContext) error {
	rawToken := req.token
	projectID := collection.project.ID

	log.Debugf("[CALDAV sync-collection] project=%d token=%q", projectID, rawToken)

	// Deletions are looked up by project_id, which no task row ever has for a
	// pseudo project, so a delta here would never contain tombstones. Checked
	// before loading the collection because clients poll this path forever.
	if rawToken != "" && collection.isPseudo() {
		log.Debugf("[CALDAV sync-collection] project=%d is a pseudo project → 403 valid-sync-token to force a full resync", projectID)
		return writeForbiddenValidSyncToken(c)
	}

	tasks, err := readCollectionTasks(s, u, collection)
	if err != nil {
		return writeError(c, err)
	}
	newToken := syncTokenPrefix + collectionETag(collection.project, tasks)

	// A pseudo project gets a token here that the branch above will always
	// reject; RFC 6578 mandates a sync-token in the multistatus regardless.
	if rawToken == "" {
		log.Debugf("[CALDAV sync-collection] project=%d empty token → full sync (%d tasks)", projectID, len(tasks))
		return writeSyncResponse(c, s, u, collection, req, ctx, tasks, nil, newToken)
	}

	tokenProjectID, tokenTS, tokenPurgeGeneration, ok := parseSyncToken(rawToken)
	if !ok {
		log.Warningf("[CALDAV sync-collection] project=%d could not parse token %q → 403 valid-sync-token", projectID, rawToken)
		return writeForbiddenValidSyncToken(c)
	}

	if tokenProjectID != projectID {
		log.Warningf("[CALDAV sync-collection] token project %d != request project %d → 403 valid-sync-token", tokenProjectID, projectID)
		return writeForbiddenValidSyncToken(c)
	}

	purgeGeneration := int64(0)
	if !collection.project.LastPurgedAt.IsZero() {
		purgeGeneration = collection.project.LastPurgedAt.Unix()
	}
	if tokenPurgeGeneration != purgeGeneration {
		return writeForbiddenValidSyncToken(c)
	}

	// Tokens older than the soft-delete retention window cannot see deletions
	// whose tasks were already purged — force a full resync instead of
	// silently serving an incomplete delta.
	if time.Since(tokenTS) > models.TrashRetention {
		log.Debugf("[CALDAV sync-collection] project=%d token %q older than retention → 403 valid-sync-token", projectID, rawToken)
		return writeForbiddenValidSyncToken(c)
	}

	// Inclusive comparison: tokens have second granularity, so strict After
	// would miss a change in the same second the token was minted.
	// Re-reporting is cheap — the etag is unchanged, so clients skip the download.
	var changedTasks []*models.TaskWithComments
	for _, t := range tasks {
		if !t.Updated.Before(tokenTS) || !t.Created.Before(tokenTS) {
			changedTasks = append(changedTasks, t)
		}
	}

	deletions, err := models.GetDeletedTasksSince(s, projectID, tokenTS)
	if err != nil {
		return writeError(c, err)
	}

	log.Debugf("[CALDAV sync-collection] project=%d delta: %d changed tasks, %d deletions since %v → newToken=%q",
		projectID, len(changedTasks), len(deletions), tokenTS.UTC(), newToken)
	return writeSyncResponse(c, s, u, collection, req, ctx, changedTasks, deletions, newToken)
}

// writeSyncResponse writes the RFC 6578 §3.8/3.9 207 Multistatus response:
// changed tasks with the requested properties, deleted tasks as 404 entries.
func writeSyncResponse(
	c *echo.Context,
	s *xorm.Session,
	u *user.User,
	collection readableCollection,
	req syncCollectionRequest,
	ctx propContext,
	tasks []*models.TaskWithComments,
	deletions []*models.Task,
	newToken string,
) error {
	if err := s.Commit(); err != nil {
		return writeError(c, err)
	}

	ms := multistatus{
		syncToken: newToken,
		responses: make([]msResponse, 0, len(tasks)+len(deletions)),
		minimal:   prefersMinimal(c),
	}

	// readCollectionTasks loaded the owner.
	owner := ownerName(u, &collection.project.Project)
	props := reportProps(req.props)
	for _, t := range tasks {
		resource := &taskResource{href: taskURL(collection.project.ID, &t.Task), task: &t.Task, owner: owner}
		ms.responses = append(ms.responses, resource.answer(ctx, props))
	}

	for _, d := range deletions {
		ms.responses = append(ms.responses, statusResponse{
			href:   taskURL(collection.project.ID, d),
			status: http.StatusNotFound,
		})
	}

	return writeMultistatus(c, ms)
}
