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
	"context"
	"strings"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/status"
)

type createFlags struct {
	description string
	statusName  string
	priority    int64
	labels      []string
	parent      string
	blockedBy   []string
}

func newCreateCmd() *cobra.Command {
	f := &createFlags{}
	var fieldsRaw string
	cmd := &cobra.Command{
		Use:     "create <title>",
		Aliases: []string{"c"},
		Short:   "Create a new task",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseFields(fieldsRaw)
			if err != nil {
				return err
			}
			rt, err := loadRuntime(cmd.Context())
			if err != nil {
				return err
			}
			task, err := runCreate(cmd.Context(), rt, args[0], f)
			if err != nil {
				return err
			}
			return rt.printTask(cmd.OutOrStdout(), fields, task)
		},
	}
	cmd.Flags().StringVarP(&f.description, "description", "d", "", "task description (HTML; see `veans prime` for canonical TipTap shapes)")
	cmd.Flags().StringVarP(&f.statusName, "status", "s", "todo", "initial status (defaults to todo)")
	cmd.Flags().Int64Var(&f.priority, "priority", 0, "priority (0=unset, 1=low, 5=DO_NOW)")
	cmd.Flags().StringSliceVar(&f.labels, "label", nil, "labels to attach (repeatable; veans: prefix added if missing)")
	cmd.Flags().StringVar(&f.parent, "parent", "", "parent task ID (creates parenttask relation)")
	cmd.Flags().StringSliceVar(&f.blockedBy, "blocked-by", nil, "task IDs that block this one (repeatable)")
	addFieldsFlag(cmd, &fieldsRaw)
	return cmd
}

func runCreate(ctx context.Context, rt *runtime, title string, f *createFlags) (*client.Task, error) {
	st, err := status.Parse(f.statusName)
	if err != nil {
		return nil, err
	}
	bucketID, err := status.BucketID(st, rt.cfg.Buckets)
	if err != nil {
		return nil, err
	}

	// Resolve refs first so a typo fails before anything is created.
	var parentID int64
	if f.parent != "" {
		if parentID, err = rt.resolveTaskID(ctx, f.parent); err != nil {
			return nil, err
		}
	}
	blockerIDs, err := resolveRefs(ctx, rt, f.blockedBy)
	if err != nil {
		return nil, err
	}

	created, err := rt.client.CreateTask(ctx, rt.cfg.ProjectID, &client.Task{
		Title:       strings.TrimSpace(title),
		Description: f.description,
		Priority:    f.priority,
		ProjectID:   rt.cfg.ProjectID,
		BucketID:    bucketID,
		Done:        st.Done(),
	})
	if err != nil {
		return nil, err
	}

	progress := &steps{task: rt.cfg.FormatTaskID(created.Index), completed: []string{"create"}}

	// Recent servers honor bucket_id on create; older ones put the task in
	// the default bucket, so move explicitly.
	if st != status.Todo {
		if err := rt.client.MoveTaskToBucket(ctx,
			rt.cfg.ProjectID, rt.cfg.ViewID, bucketID, created.ID); err != nil {
			return nil, progress.fail("set bucket", err)
		}
		progress.done("set bucket")
	}

	for _, raw := range f.labels {
		title := normalizeLabelTitle(raw)
		step := "label " + title
		l, err := getOrCreateLabelByTitle(ctx, rt.client, title)
		if err != nil {
			return nil, progress.fail(step, err)
		}
		if err := rt.client.AddLabelToTask(ctx, created.ID, l.ID); err != nil {
			return nil, progress.fail(step, err)
		}
		progress.done(step)
	}

	if f.parent != "" {
		step := "parent " + f.parent
		if _, err := rt.client.CreateRelation(ctx, created.ID, parentID, "parenttask"); err != nil {
			return nil, progress.fail(step, err)
		}
		progress.done(step)
	}

	for i, ref := range f.blockedBy {
		step := "blocked-by " + ref
		if _, err := rt.client.CreateRelation(ctx, created.ID, blockerIDs[i], "blocked"); err != nil {
			return nil, progress.fail(step, err)
		}
		progress.done(step)
	}

	// Re-fetch so the response reflects the labels and any post-create state.
	// The refetch is best-effort: if it fails (token expired mid-operation,
	// transient network blip), we still return the create result so callers
	// see the new task ID rather than a confusing error.
	final, err := rt.client.GetTask(ctx, created.ID)
	if err != nil {
		return created, nil //nolint:nilerr // intentional best-effort refetch
	}
	return final, nil
}
