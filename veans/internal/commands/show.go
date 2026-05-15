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

package commands

import (
	"encoding/json"

	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show <id>",
		Short: "Show a task by PROJ-NN, #NN, or numeric ID",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			id, err := rt.resolveTaskID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			task, err := rt.client.GetTask(cmd.Context(), id)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(task)
		},
	}
	return cmd
}
