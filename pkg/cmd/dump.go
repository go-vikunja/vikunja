// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
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

package cmd

import (
	"time"

	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/dump"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dumpCmd)
}

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all vikunja data into a zip file. Includes config, files and db.",
	PreRun: func(cmd *cobra.Command, args []string) {
		initialize.FullInit()
	},
	Run: func(cmd *cobra.Command, args []string) {
		filename := "vikunja-dump_" + time.Now().Format("2006-01-02_15-03-05") + ".zip"
		if err := dump.Dump(filename); err != nil {
			log.Critical(err.Error())
		}
	},
}
