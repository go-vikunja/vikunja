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

	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/migration"
	"github.com/spf13/cobra"
)

func init() {
	licenseCmd.AddCommand(licenseInstanceIDCmd)
	rootCmd.AddCommand(licenseCmd)
}

var licenseCmd = &cobra.Command{
	Use:   "license",
	Short: "License related commands.",
}

var licenseInstanceIDCmd = &cobra.Command{
	Use:   "instance-id",
	Short: "Print the instance ID, needed to generate an offline license.",
	// Not FullInit: that runs license.Init, which would contact the license server.
	PreRun: func(_ *cobra.Command, _ []string) {
		initialize.LightInit()
		migration.Migrate(nil)
		initialize.InitEngines()
	},
	Run: func(_ *cobra.Command, _ []string) {
		id, err := license.InstanceID()
		if err != nil {
			log.Fatalf("Could not get instance ID: %s", err)
		}
		fmt.Println(id)
	},
}
