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

package cmd

import (
	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(indexCmd)
}

var indexPartialFlag bool

var indexCmd = &cobra.Command{
	Use:   "index",
	Short: "Reindex all of Vikunja's data into Typesense. This will remove any existing index.",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		if config.TypesenseURL.GetString() == "" {
			log.Error("Typesense not configured")
			return
		}

		err := models.CreateTypesenseCollections()
		if err != nil {
			log.Criticalf("Could not create Typesense collections: %s", err.Error())
			return
		}
		if indexPartialFlag {
			log.Infof("Indexing changed tasks… This may take a while.")
			err = models.SyncUpdatedTasksIntoTypesense()
			if err != nil {
				log.Criticalf("Could not reindex all changed tasks into Typesense: %s", err.Error())
				return
			}
		} else {
			log.Infof("Indexing all tasks… This may take a while.")
			err = models.ReindexAllTasks()
			if err != nil {
				log.Criticalf("Could not reindex all tasks into Typesense: %s", err.Error())
				return
			}
		}

		log.Infof("Done!")
	},
}

func init() {
	indexCmd.Flags().BoolVarP(&indexPartialFlag, "partial", "p", false, "If provided, Vikunja will only index those tasks which are not present in the index. It will not remove any existing tasks.")
}
