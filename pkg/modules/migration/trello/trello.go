// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package trello

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"github.com/adlio/trello"
)

// Migration represents the trello migration struct
type Migration struct {
	Token string `json:"code"`
}

var trelloColorMap map[string]string

func init() {
	trelloColorMap = make(map[string]string, 10)
	trelloColorMap = map[string]string{
		"green":       "61bd4f",
		"yellow":      "f2d600",
		"orange":      "ff9f1a",
		"red":         "eb5a46",
		"sky":         "00c2e0",
		"lime":        "51e898",
		"purple":      "c377e0",
		"blue":        "0079bf",
		"pink":        "ff78cb",
		"black":       "344563",
		"transparent": "", // Empty
	}
}

// Name is used to get the name of the trello migration - we're using the docs here to annotate the status route.
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not. This is useful to show a confirmation message in the frontend if the user is trying to do the same migration again.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/trello/status [get]
func (m *Migration) Name() string {
	return "trello"
}

// AuthURL returns the url users need to authenticate against
// @Summary Get the auth url from trello
// @Description Returns the auth url where the user needs to get its auth code. This code can then be used to migrate everything from trello to Vikunja.
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} handler.AuthURL "The auth url."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/trello/auth [get]
func (m *Migration) AuthURL() string {
	return "https://trello.com/1/authorize" +
		"?expiration=1hour" +
		"&scope=read" +
		"&callback_method=fragment" +
		"&response_type=token" +
		"&name=Vikunja%20Migration" +
		"&key=" + config.MigrationTrelloKey.GetString() +
		"&return_url=" + config.MigrationTrelloRedirectURL.GetString()
}

func getTrelloData(token string) (trelloData []*trello.Board, err error) {
	allArg := trello.Arguments{"fields": "all"}

	client := trello.NewClient(config.MigrationTrelloKey.GetString(), token)
	client.Logger = log.GetLogger()

	log.Debugf("[Trello Migration] Getting boards...")

	trelloData, err = client.GetMyBoards(trello.Defaults())
	if err != nil {
		return
	}

	log.Debugf("[Trello Migration] Got %d trello boards", len(trelloData))

	for _, board := range trelloData {
		log.Debugf("[Trello Migration] Getting lists for board %s", board.ID)

		board.Lists, err = board.GetLists(trello.Defaults())
		if err != nil {
			return
		}

		log.Debugf("[Trello Migration] Got %d lists for board %s", len(board.Lists), board.ID)

		listMap := make(map[string]*trello.List, len(board.Lists))
		for _, list := range board.Lists {
			listMap[list.ID] = list
		}

		log.Debugf("[Trello Migration] Getting cards for board %s", board.ID)

		cards, err := board.GetCards(allArg)
		if err != nil {
			return nil, err
		}

		log.Debugf("[Trello Migration] Got %d cards for board %s", len(cards), board.ID)

		for _, card := range cards {
			list, exists := listMap[card.IDList]
			if !exists {
				continue
			}

			card.Attachments, err = card.GetAttachments(allArg)
			if err != nil {
				return nil, err
			}

			list.Cards = append(list.Cards, card)
		}

		log.Debugf("[Trello Migration] Looked for attachements on all cards of board %s", board.ID)
	}

	return
}

// Converts all previously obtained data from trello into the vikunja format.
// `trelloData` should contain all boards with their lists and cards respectively.
func convertTrelloDataToVikunja(trelloData []*trello.Board, token string) (fullVikunjaHierachie []*models.NamespaceWithListsAndTasks, err error) {

	log.Debugf("[Trello Migration] ")

	fullVikunjaHierachie = []*models.NamespaceWithListsAndTasks{
		{
			Namespace: models.Namespace{
				Title: "Imported from Trello",
			},
			Lists: []*models.ListWithTasksAndBuckets{},
		},
	}

	var bucketID int64 = 1

	log.Debugf("[Trello Migration] Converting %d boards to vikunja lists", len(trelloData))

	for _, board := range trelloData {
		list := &models.ListWithTasksAndBuckets{
			List: models.List{
				Title:       board.Name,
				Description: board.Desc,
				IsArchived:  board.Closed,
			},
		}

		// Background
		// We're pretty much abusing the backgroundinformation field here - not sure if this is really better than adding a new property to the list
		if board.Prefs.BackgroundImage != "" {
			log.Debugf("[Trello Migration] Downloading background %s for board %s", board.Prefs.BackgroundImage, board.ID)
			buf, err := migration.DownloadFile(board.Prefs.BackgroundImage)
			if err != nil {
				return nil, err
			}
			log.Debugf("[Trello Migration] Downloaded background %s for board %s", board.Prefs.BackgroundImage, board.ID)
			list.BackgroundInformation = buf
		} else {
			log.Debugf("[Trello Migration] Board %s does not have a background image, not copying...", board.ID)
		}

		for _, l := range board.Lists {
			bucket := &models.Bucket{
				ID:    bucketID,
				Title: l.Name,
			}

			log.Debugf("[Trello Migration] Converting %d cards to tasks from board %s", len(l.Cards), board.ID)

			for _, card := range l.Cards {

				log.Debugf("[Trello Migration] Converting card %s", card.ID)

				// The usual stuff: Title, description, position, bucket id
				task := &models.Task{
					Title:          card.Name,
					Description:    card.Desc,
					KanbanPosition: card.Pos,
					BucketID:       bucketID,
				}

				if card.Due != nil {
					task.DueDate = *card.Due
				}

				// Checklists (as markdown in description)
				for _, checklist := range card.Checklists {
					task.Description += "\n\n## " + checklist.Name + "\n"

					for _, item := range checklist.CheckItems {
						task.Description += "\n* "
						if item.State == "completed" {
							task.Description += "[x]"
						} else {
							task.Description += "[ ]"
						}
						task.Description += " " + item.Name
					}
				}
				if len(card.Checklists) > 0 {
					log.Debugf("[Trello Migration] Converted %d checklists from card %s", len(card.Checklists), card.ID)
				}

				// Labels
				for _, label := range card.Labels {
					color, exists := trelloColorMap[label.Color]
					if !exists {
						log.Debugf("[Trello Migration] Color %s not mapped for trello card %s, not adding label", label.Color, card.ID)
						continue
					}

					task.Labels = append(task.Labels, &models.Label{
						Title:    label.Name,
						HexColor: color,
					})

					log.Debugf("[Trello Migration] Converted label %s from card %s", label.ID, card.ID)
				}

				// Attachments
				if len(card.Attachments) > 0 {
					log.Debugf("[Trello Migration] Downloading %d card attachments from card %s", len(card.Attachments), card.ID)
				}
				for _, attachment := range card.Attachments {
					if attachment.MimeType == "" { // Attachments can also be not downloadable - the mime type is empty in that case.
						log.Debugf("[Trello Migration] Attachment %s does not have a mime type, not downloading", attachment.ID)
						continue
					}

					log.Debugf("[Trello Migration] Downloading card attachment %s", attachment.ID)

					buf, err := migration.DownloadFileWithHeaders(attachment.URL, map[string][]string{
						"Authorization": {`OAuth oauth_consumer_key="` + config.MigrationTrelloKey.GetString() + `", oauth_token="` + token + `"`},
					})
					if err != nil {
						return nil, err
					}

					task.Attachments = append(task.Attachments, &models.TaskAttachment{
						File: &files.File{
							Name:        attachment.Name,
							Mime:        attachment.MimeType,
							Size:        uint64(buf.Len()),
							FileContent: buf.Bytes(),
						},
					})

					log.Debugf("[Trello Migration] Downloaded card attachment %s", attachment.ID)
				}

				list.Tasks = append(list.Tasks, &models.TaskWithComments{Task: *task})
			}

			list.Buckets = append(list.Buckets, bucket)
			bucketID++
		}

		log.Debugf("[Trello Migration] Converted all cards to tasks for board %s", board.ID)

		fullVikunjaHierachie[0].Lists = append(fullVikunjaHierachie[0].Lists, list)
	}

	return
}

// Migrate gets all tasks from trello for a user and puts them into vikunja
// @Summary Migrate all lists, tasks etc. from trello
// @Description Migrates all projects, tasks, notes, reminders, subtasks and files from trello to vikunja.
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationCode body trello.Migration true "The auth token previously obtained from the auth url. See the docs for /migration/trello/auth."
// @Success 200 {object} models.Message "A message telling you everything was migrated successfully."
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/trello/migrate [post]
func (m *Migration) Migrate(u *user.User) (err error) {
	log.Debugf("[Trello Migration] Starting migration for user %d", u.ID)
	log.Debugf("[Trello Migration] Getting all trello data for user %d", u.ID)

	trelloData, err := getTrelloData(m.Token)
	if err != nil {
		return
	}

	log.Debugf("[Trello Migration] Got all trello data for user %d", u.ID)
	log.Debugf("[Trello Migration] Start converting trello data for user %d", u.ID)

	fullVikunjaHierachie, err := convertTrelloDataToVikunja(trelloData, m.Token)
	if err != nil {
		return
	}

	log.Debugf("[Trello Migration] Done migrating trello data for user %d", u.ID)
	log.Debugf("[Trello Migration] Start inserting trello data for user %d", u.ID)

	err = migration.InsertFromStructure(fullVikunjaHierachie, u)
	if err != nil {
		return
	}

	log.Debugf("[Trello Migration] Done inserting trello data for user %d", u.ID)
	log.Debugf("[Trello Migration] Migration done for user %d", u.ID)

	return nil
}
