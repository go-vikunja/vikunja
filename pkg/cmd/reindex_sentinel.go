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
	"github.com/schollz/progressbar/v3"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(reindexSentinelCmd)
}

var reindexSentinelCmd = &cobra.Command{
	Use:   "reindex-sentinel",
	Short: "Reindex all tasks with sentinel dates",
	Long:  "Rebuilds the Typesense index using the -1 sentinel for tasks without a due date.",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		if config.TypesenseURL.GetString() == "" {
			log.Error("Typesense not configured")
			return
		}

		if err := models.CreateTypesenseCollections(); err != nil {
			log.Criticalf("Could not create Typesense collections: %s", err.Error())
			return
		}

		bar := progressbar.Default(-1)
		if err := models.ReindexAllTasks(bar); err != nil {
			log.Criticalf("Could not reindex all tasks: %s", err.Error())
			return
		}
		log.Infof("Done!")
	},
}
