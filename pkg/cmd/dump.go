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
	"path/filepath"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/modules/dump"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(dumpCmd)
}

var (
	dumpPathFlag     string
	dumpFilenameFlag string
)

var dumpCmd = &cobra.Command{
	Use:   "dump",
	Short: "Dump all vikunja data into a zip file. Includes config, files and db.",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		filename := "vikunja-dump_" + time.Now().Format("2006-01-02_15-03-05") + ".zip"
		if dumpFilenameFlag != "" {
			filename = dumpFilenameFlag
			if !strings.HasSuffix(filename, ".zip") {
				filename += ".zip"
			}
		}

		path := config.ServiceRootpath.GetString()
		if dumpPathFlag != "" {
			path = dumpPathFlag
		}

		if err := dump.Dump(filepath.Join(path, filename)); err != nil {
			log.Critical(err.Error())
		}
	},
}

func init() {
	dumpCmd.Flags().StringVarP(&dumpPathFlag, "path", "p", "", "The folder path where the dump file should be saved. Vikunja will use the configured root path or the binary location if the flag is not provided.")
	dumpCmd.Flags().StringVarP(&dumpFilenameFlag, "filename", "f", "", "The filename of the dump file. If it does not end in '.zip', it will be added as a file extension. Defaults to 'vikunja-dump_YYYY-MM-DD_HH-II-SS.zip'.")
}
