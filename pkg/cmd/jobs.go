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
	command.AddCommand(newJobReplayCommand())
	rootCmd.AddCommand(command)
}

func newJobReplayCommand() *cobra.Command {
	filter := jobs.ReplayFilter{}
	var parkedAfter, parkedBefore string
	var dryRun bool
	command := &cobra.Command{
		Use:   "replay [job-id]",
		Short: "Replay parked jobs selected by ID, queue, kind, or parked-at time",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			var err error
			if filter.Limit < 1 {
				return fmt.Errorf("replay limit must be between 1 and 1000")
			}
			if len(args) == 1 {
				filter.ID, err = strconv.ParseInt(args[0], 10, 64)
				if err != nil || filter.ID <= 0 {
					return fmt.Errorf("job ID must be a positive integer")
				}
			}
			if parkedAfter != "" {
				filter.ParkedAfter, err = time.Parse(time.RFC3339, parkedAfter)
				if err != nil {
					return fmt.Errorf("invalid parked-after: %w", err)
				}
			}
			if parkedBefore != "" {
				filter.ParkedBefore, err = time.Parse(time.RFC3339, parkedBefore)
				if err != nil {
					return fmt.Errorf("invalid parked-before: %w", err)
				}
			}
			if err = filter.Validate(); err != nil {
				return err
			}
			initialize.FullInitWithoutAsync()
			engine, err := db.CreateDBEngine()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithTimeout(cmd.Context(), 30*time.Second)
			defer cancel()
			result, err := jobs.NewQueue(engine, "operator").ReplaySelected(ctx, filter, dryRun)
			if err != nil {
				return fmt.Errorf("replay stopped after %d changes (%d matched): %w", result.Replayed, result.Matched, err)
			}
			if filter.ID != 0 && result.Replayed == 0 && !dryRun {
				return fmt.Errorf("parked job not found")
			}
			_, err = fmt.Fprintf(cmd.OutOrStdout(), "matched=%d replayed=%d limit=%d dry-run=%t\n", result.Matched, result.Replayed, filter.Limit, dryRun)
			return err
		},
	}
	command.Flags().StringVar(&filter.Queue, "queue", "", "Queue to replay")
	command.Flags().StringVar(&filter.Kind, "kind", "", "Job kind to replay")
	command.Flags().StringVar(&parkedAfter, "parked-after", "", "Inclusive parked-at lower bound (RFC3339)")
	command.Flags().StringVar(&parkedBefore, "parked-before", "", "Exclusive parked-at upper bound (RFC3339)")
	command.Flags().IntVar(&filter.Limit, "limit", 100, "Maximum jobs to replay (1–1000)")
	command.Flags().BoolVar(&dryRun, "dry-run", false, "Count matching parked jobs without changing them")
	return command
}
