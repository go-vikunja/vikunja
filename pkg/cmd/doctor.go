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

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/doctor"
	"code.vikunja.io/api/pkg/log"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(doctorCmd)
}

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Run diagnostic checks on your Vikunja installation",
	Long: `The doctor command runs a series of diagnostic checks to help troubleshoot
issues with your Vikunja installation. It checks:

- System information (version, user, working directory)
- Configuration (config file, public URL, JWT secret, CORS)
- Database connectivity and version
- File storage (local or S3)
- Optional services (Redis, Typesense, Mailer, LDAP, OpenID)

Exit codes:
  0 - All checks passed
  1 - One or more checks failed`,
	PreRun: func(_ *cobra.Command, _ []string) {
		// Minimal init - just config and logger
		// Each check will initialize and test its own components
		log.InitLogger()
		config.InitConfig()
	},
	Run: func(_ *cobra.Command, _ []string) {
		results := doctor.Run()

		doctor.PrintResults(os.Stdout, results)

		failed := doctor.CountFailed(results)
		if failed > 0 {
			fmt.Printf("%d check(s) failed\n", failed)
			os.Exit(1)
		}

		fmt.Println("All checks passed")
	},
}
