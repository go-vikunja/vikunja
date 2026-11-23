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
	"github.com/spf13/cobra"
)

func newShowCmd() *cobra.Command {
	var (
		fieldsRaw string
		taskIDArg int64
	)
	cmd := &cobra.Command{
		Use:   "show <PROJ-NN|#NN|NN> | show --id <task id>",
		Short: "Show a task by PROJ-NN, #NN, NN (project index), or --id <task id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseFields(fieldsRaw)
			if err != nil {
				return err
			}
			rt, err := loadRuntime(cmd.Context())
			if err != nil {
				return err
			}
			id, err := rt.taskID(cmd.Context(), args, taskIDArg)
			if err != nil {
				return err
			}
			task, err := rt.client.GetTask(cmd.Context(), id)
			if err != nil {
				return err
			}
			return rt.printTask(cmd.OutOrStdout(), fields, task)
		},
	}
	cmd.Args = addTaskIDFlag(cmd, &taskIDArg)
	addFieldsFlag(cmd, &fieldsRaw)
	return cmd
}
