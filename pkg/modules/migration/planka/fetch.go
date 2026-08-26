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

package planka

import (
	"fmt"
	"net/url"
	"slices"
	"sort"
	"time"

	"code.vikunja.io/api/pkg/log"
)

// maxPages stops the paged loops from spinning forever on a server that keeps returning cards.
const maxPages = 10000

func fetchAll(c *client) (*plankaData, error) {
	log.Debugf("[Planka Migration] Fetching projects")

	projects := &projectsResponse{}
	if err := c.get("/api/projects", nil, projects); err != nil {
		return nil, err
	}

	data := &plankaData{
		CurrentUserID:         c.currentUserID,
		Users:                 map[string]plankaUser{},
		BaseCustomFieldGroups: map[string]plankaBaseCustomFieldGroup{},
		BaseCustomFields:      map[string]plankaCustomField{},
	}
	for _, g := range projects.Included.BaseCustomFieldGroups {
		data.BaseCustomFieldGroups[g.ID] = g
	}
	// fields of base groups only come with the projects, board payloads don't repeat them
	for _, f := range projects.Included.CustomFields {
		data.BaseCustomFields[f.ID] = f
	}

	boardsByProject := map[string][]plankaBoard{}
	for _, b := range projects.Included.Boards {
		boardsByProject[b.ProjectID] = append(boardsByProject[b.ProjectID], b)
	}

	log.Debugf("[Planka Migration] Got %d projects with %d boards", len(projects.Items), len(projects.Included.Boards))

	for _, p := range projects.Items {
		pd := &plankaProjectData{plankaProject: p}
		boards := boardsByProject[p.ID]
		sort.SliceStable(boards, func(i, j int) bool { return boards[i].Position < boards[j].Position })

		for _, b := range boards {
			bd, err := fetchBoard(c, b.ID)
			if err != nil {
				return nil, fmt.Errorf("fetching board %s: %w", b.ID, err)
			}
			for _, u := range bd.Users {
				data.Users[u.ID] = plankaUser{ID: u.ID, Name: u.Name, Username: u.Username}
			}
			pd.Boards = append(pd.Boards, bd)
		}
		data.Projects = append(data.Projects, pd)
	}

	return data, nil
}

func fetchBoard(c *client, boardID string) (*plankaBoardData, error) {
	log.Debugf("[Planka Migration] Fetching board %s", boardID)

	resp := &boardResponse{}
	if err := c.get("/api/boards/"+url.PathEscape(boardID), nil, resp); err != nil {
		return nil, err
	}

	// Planka v1 lists have no type; v2 always sets one.
	for _, l := range resp.Included.Lists {
		if l.Type == "" {
			return nil, &ErrUnsupportedVersion{}
		}
	}

	bd := &plankaBoardData{
		plankaBoard:   resp.Item,
		boardIncluded: resp.Included,
		Comments:      map[string][]plankaComment{},
	}

	// Cards of archive lists are not in the board payload.
	for _, l := range resp.Included.Lists {
		if l.Type != listTypeArchive {
			continue
		}
		if err := fetchArchivedCards(c, l.ID, bd); err != nil {
			log.Errorf("[Planka Migration] Could not fetch all cards of archive list %s on board %s, keeping the ones already fetched: %s", l.ID, boardID, err)
		}
	}

	for _, card := range bd.Cards {
		if card.CommentsTotal == 0 {
			continue
		}
		comments, users, err := fetchComments(c, card.ID)
		if err != nil {
			log.Errorf("[Planka Migration] Could not fetch all comments of card %s, keeping the %d already fetched: %s", card.ID, len(comments), err)
		}
		bd.Comments[card.ID] = comments
		bd.Users = append(bd.Users, users...)
	}

	log.Debugf("[Planka Migration] Fetched board %s with %d lists and %d cards", boardID, len(bd.Lists), len(bd.Cards))

	return bd, nil
}

// fetchArchivedCards pages GET /api/lists/:id/cards (sorted listChangedAt DESC, id DESC) and merges
// the cards and their included data into bd.
func fetchArchivedCards(c *client, listID string, bd *plankaBoardData) error {
	var (
		beforeAt time.Time
		beforeID string
	)
	for page := 0; ; page++ {
		if page >= maxPages {
			return fmt.Errorf("cards of list %s did not end after %d pages", listID, maxPages)
		}

		query := url.Values{}
		if beforeID != "" {
			query.Set("before[listChangedAt]", beforeAt.UTC().Format(time.RFC3339Nano))
			query.Set("before[id]", beforeID)
		}

		resp := &listCardsResponse{}
		if err := c.get("/api/lists/"+url.PathEscape(listID)+"/cards", query, resp); err != nil {
			return err
		}
		if len(resp.Items) == 0 {
			return nil
		}

		bd.Cards = append(bd.Cards, resp.Items...)
		bd.Users = append(bd.Users, resp.Included.Users...)
		bd.CardLabels = append(bd.CardLabels, resp.Included.CardLabels...)
		bd.TaskLists = append(bd.TaskLists, resp.Included.TaskLists...)
		bd.Tasks = append(bd.Tasks, resp.Included.Tasks...)
		bd.Attachments = append(bd.Attachments, resp.Included.Attachments...)
		bd.CustomFieldGroups = append(bd.CustomFieldGroups, resp.Included.CustomFieldGroups...)
		bd.CustomFields = append(bd.CustomFields, resp.Included.CustomFields...)
		bd.CustomFieldValues = append(bd.CustomFieldValues, resp.Included.CustomFieldValues...)

		last := resp.Items[len(resp.Items)-1]
		if last.ID == beforeID {
			return nil
		}
		// planka rejects a cursor without before[listChangedAt]; cards with a null one sort first, so
		// this can only happen on a first page which is already complete in the common case.
		if last.ListChangedAt == nil {
			log.Warningf("[Planka Migration] Card %s of list %s has no listChangedAt, cannot page further, the remaining archived cards of that list are not imported", last.ID, listID)
			return nil
		}
		beforeAt, beforeID = *last.ListChangedAt, last.ID
	}
}

// fetchComments pages GET /api/cards/:id/comments (sorted id DESC) and returns them oldest first.
func fetchComments(c *client, cardID string) (all []plankaComment, users []plankaUser, err error) {
	// planka sorts newest first; partial results returned with an error must be ordered as well
	defer func() { slices.Reverse(all) }()

	var beforeID string
	for page := 0; ; page++ {
		if page >= maxPages {
			return all, users, fmt.Errorf("comments of card %s did not end after %d pages", cardID, maxPages)
		}

		query := url.Values{}
		if beforeID != "" {
			query.Set("beforeId", beforeID)
		}

		resp := &commentsResponse{}
		if err := c.get("/api/cards/"+url.PathEscape(cardID)+"/comments", query, resp); err != nil {
			return all, users, err
		}
		if len(resp.Items) == 0 {
			break
		}

		all = append(all, resp.Items...)
		users = append(users, resp.Included.Users...)

		last := resp.Items[len(resp.Items)-1]
		if last.ID == beforeID {
			break
		}
		beforeID = last.ID
	}

	return all, users, nil
}
