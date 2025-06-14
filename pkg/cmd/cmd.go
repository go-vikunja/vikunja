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

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "vikunja",
	Short: "Vikunja is the to-do app to organize your life.",
	Long: `Vikunja (/vɪˈkuːnjə/)
The to-do app to organize your life.

Also one of the two wild South American camelids which live in the high
alpine areas of the Andes and a relative of the llama.

Vikunja is a self-hosted To-Do list application with a web app and mobile apps for all platforms. It is licensed under the AGPL-3.0-or-later.

Find out more at vikunja.io.`,
	PreRun: webCmd.PreRun,
	Run:    webCmd.Run,
}

// Execute starts the application
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
