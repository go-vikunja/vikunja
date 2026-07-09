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
	"fmt"
	"os"

	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/initialize"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

var healthcheckCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "Preform a healthcheck on the Vikunja api server",
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.FullInitWithoutAsync()
	},
	Run: func(_ *cobra.Command, _ []string) {
		if err := health.Check(); err != nil {
			fmt.Printf("API server is not healthy: %v\n", err)
			os.Exit(1)
			return
		}

		fmt.Println("API server is healthy")
		os.Exit(0)
	},
}
