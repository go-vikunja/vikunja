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
	"context"
	"fmt"
	"os"

	"code.vikunja.io/api/pkg/health"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/modules/auth/openid"

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

		// Unhealthy providers are reported but don't fail the check — this
		// command backs Docker HEALTHCHECK, restarting Vikunja cannot fix an
		// unreachable external identity provider, and a failed registration
		// heals itself through the availability cron without a restart.
		for _, p := range openid.ProbeProvidersAvailability(context.Background()) {
			switch {
			case !p.Registered:
				fmt.Printf("Warning: OpenID provider %q is not registered, logging in with it is unavailable until Vikunja can reach it\n", p.Name)
			case !p.Reachable:
				fmt.Printf("Warning: OpenID provider %q is not reachable\n", p.Name)
			default:
				fmt.Printf("OpenID provider %q is registered and reachable\n", p.Name)
			}
		}

		fmt.Println("API server is healthy")
		os.Exit(0)
	},
}
