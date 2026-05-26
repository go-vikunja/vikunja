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
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
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
	cmd := &cobra.Command{
		Use:     "create <title>",
		Aliases: []string{"c"},
		Short:   "Create a new task",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			task, err := runCreate(cmd.Context(), rt, args[0], f)
			if err != nil {
				return err
			}
			if globals.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(task)
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created %s  %s\n",
				rt.cfg.FormatTaskID(task.Index), task.Title)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.description, "description", "d", "", "task description (markdown)")
	cmd.Flags().StringVarP(&f.statusName, "status", "s", "todo", "initial status (defaults to todo)")
	cmd.Flags().Int64Var(&f.priority, "priority", 0, "priority (0=unset, 1=low, 5=DO_NOW)")
	cmd.Flags().StringSliceVar(&f.labels, "label", nil, "labels to attach (repeatable; veans: prefix added if missing)")
	cmd.Flags().StringVar(&f.parent, "parent", "", "parent task ID (creates parenttask relation)")
	cmd.Flags().StringSliceVar(&f.blockedBy, "blocked-by", nil, "task IDs that block this one (repeatable)")
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

	// Vikunja places newly-created tasks in the view's default bucket
	// regardless of bucket_id in the create payload — move it explicitly
	// when the requested status isn't Todo.
	if st != status.Todo {
		if err := rt.client.MoveTaskToBucket(ctx,
			rt.cfg.ProjectID, rt.cfg.ViewID, bucketID, created.ID); err != nil {
			return nil, output.Wrap(output.CodeUnknown, err, "set initial bucket: %v", err)
		}
	}

	// Attach labels (lazily creating them under veans: namespace).
	for _, raw := range f.labels {
		title := normalizeLabelTitle(raw)
		l, err := getOrCreateLabelByTitle(ctx, rt.client, title)
		if err != nil {
			return nil, output.Wrap(output.CodeUnknown, err, "label %q: %v", title, err)
		}
		if err := rt.client.AddLabelToTask(ctx, created.ID, l.ID); err != nil {
			return nil, err
		}
	}

	// Parent relation.
	if f.parent != "" {
		parentID, err := rt.resolveTaskID(ctx, f.parent)
		if err != nil {
			return nil, err
		}
		if _, err := rt.client.CreateRelation(ctx, created.ID, parentID, "parenttask"); err != nil {
			return nil, err
		}
	}

	// Blocked-by relations.
	for _, ref := range f.blockedBy {
		blockerID, err := rt.resolveTaskID(ctx, ref)
		if err != nil {
			return nil, err
		}
		if _, err := rt.client.CreateRelation(ctx, created.ID, blockerID, "blocked"); err != nil {
			return nil, err
		}
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
