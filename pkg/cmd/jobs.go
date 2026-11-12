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
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/jobs"
	"github.com/spf13/cobra"
	"xorm.io/builder"
)

func init() {
	command := &cobra.Command{Use: "jobs", Short: "Inspect and replay parked background jobs"}
	command.AddCommand(&cobra.Command{
		Use:   "failed",
		Short: "List up to 100 parked jobs without their payloads",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			initialize.FullInitWithoutAsync()
			s := db.NewReadSession()
			defer s.Close()
			db.SetSessionContext(cmd.Context(), s)
			var rows []jobs.Job
			if err := s.Where(builder.Eq{"state": jobs.StateFailed}).OrderBy("id").Limit(100).Find(&rows); err != nil {
				return err
			}
			for _, job := range rows {
				if _, err := fmt.Fprintf(cmd.OutOrStdout(), "%d\t%s\t%s\tattempt=%d\t%s\n", job.ID, job.Queue, job.Kind, job.Attempt, job.LastError); err != nil {
					return err
				}
			}
			return nil
		},
	})
	command.AddCommand(&cobra.Command{
		Use:   "replay <job-id>",
		Short: "Give a parked job five more attempts",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			id, err := strconv.ParseInt(args[0], 10, 64)
			if err != nil || id <= 0 {
				return fmt.Errorf("job ID must be a positive integer")
			}
			initialize.FullInitWithoutAsync()
			engine, err := db.CreateDBEngine()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			return jobs.NewQueue(engine, "operator").Replay(ctx, id)
		},
	})
	rootCmd.AddCommand(command)
}
