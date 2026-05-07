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
	"time"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
	"code.vikunja.io/veans/internal/status"
)

type updateFlags struct {
	statusName       string
	title            string
	priority         int64
	priorityIsSet    bool
	addLabels        []string
	removeLabels     []string
	description      string
	descriptionIsSet bool
	replaceOld       string
	replaceNew       string
	descriptionApp   string
	comment          string
	reason           string
	ifUnchangedSince string
}

func newUpdateCmd() *cobra.Command {
	f := &updateFlags{}
	cmd := &cobra.Command{
		Use:     "update <id>",
		Aliases: []string{"u"},
		Short:   "Update a task by PROJ-NN, #NN, or numeric ID",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			f.descriptionIsSet = cmd.Flags().Changed("description")
			f.priorityIsSet = cmd.Flags().Changed("priority")

			id, err := rt.resolveTaskID(cmd.Context(), args[0])
			if err != nil {
				return err
			}
			task, err := runUpdate(cmd.Context(), rt, id, f)
			if err != nil {
				return err
			}
			if globals.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(task)
			}
			s := status.FromBucketID(task.BucketID, rt.cfg.Buckets)
			fmt.Fprintf(cmd.OutOrStdout(), "Updated %s  [%s]  %s\n",
				rt.cfg.FormatTaskID(task.Index), s, task.Title)
			return nil
		},
	}
	cmd.Flags().StringVarP(&f.statusName, "status", "s", "", "transition to a status")
	cmd.Flags().StringVarP(&f.title, "title", "t", "", "new title")
	cmd.Flags().Int64Var(&f.priority, "priority", 0, "new priority")
	cmd.Flags().StringSliceVar(&f.addLabels, "label-add", nil, "labels to attach (repeatable; veans: prefix added if missing)")
	cmd.Flags().StringSliceVar(&f.removeLabels, "label-remove", nil, "labels to detach (repeatable)")
	cmd.Flags().StringVar(&f.description, "description", "", "replace the entire description")
	cmd.Flags().StringVar(&f.replaceOld, "description-replace-old", "", "exact-match string to replace in description (must be unique)")
	cmd.Flags().StringVar(&f.replaceNew, "description-replace-new", "", "replacement for --description-replace-old")
	cmd.Flags().StringVar(&f.descriptionApp, "description-append", "", "append text to the existing description")
	cmd.Flags().StringVarP(&f.comment, "comment", "c", "", "post a comment as part of this update")
	cmd.Flags().StringVar(&f.reason, "reason", "", "rationale (required when --status scrapped)")
	cmd.Flags().StringVar(&f.ifUnchangedSince, "if-unchanged-since", "", "RFC3339 timestamp; abort if the task has changed since")
	return cmd
}

func runUpdate(ctx context.Context, rt *runtime, id int64, f *updateFlags) (*client.Task, error) {
	current, err := rt.client.GetTask(ctx, id)
	if err != nil {
		return nil, err
	}

	// Optimistic concurrency.
	if f.ifUnchangedSince != "" {
		ts, err := time.Parse(time.RFC3339, f.ifUnchangedSince)
		if err != nil {
			return nil, output.Wrap(output.CodeValidation, err, "parse --if-unchanged-since: %v", err)
		}
		if current.Updated.After(ts) {
			return nil, output.New(output.CodeConflict,
				"task %s changed at %s, after --if-unchanged-since %s",
				rt.cfg.FormatTaskID(current.Index), current.Updated.Format(time.RFC3339), ts.Format(time.RFC3339))
		}
	}

	// Resolve new status / done flag if --status is set.
	var newStatus status.Status
	if f.statusName != "" {
		s, err := status.Parse(f.statusName)
		if err != nil {
			return nil, err
		}
		newStatus = s
		if s == status.Scrapped && strings.TrimSpace(f.reason) == "" {
			return nil, output.New(output.CodeValidation, "--reason is required when --status scrapped")
		}
	}

	// Build the update payload incrementally so we don't clobber unmentioned
	// fields. The base must include the ID; bucket/done are conditional.
	body := &client.Task{ID: id}
	dirty := false

	if f.title != "" {
		body.Title = f.title
		dirty = true
	}
	if f.priorityIsSet {
		body.Priority = f.priority
		dirty = true
	}

	// Description ops are mutually-exclusive layers; --description wins
	// outright, otherwise replace-old/new + append run on the current body.
	newDesc, descChanged, err := composeDescription(current.Description, f)
	if err != nil {
		return nil, err
	}
	if descChanged {
		body.Description = newDesc
		dirty = true
	}

	if newStatus != "" {
		bid, err := status.BucketID(newStatus, rt.cfg.Buckets)
		if err != nil {
			return nil, err
		}
		body.BucketID = bid
		body.Done = newStatus.Done()
		dirty = true
	}

	// Comment first when transitioning to scrapped — the reason is part of
	// the audit trail and should appear before the bucket move in the log.
	if newStatus == status.Scrapped {
		if _, err := rt.client.AddTaskComment(ctx, id, "**Scrapped:** "+strings.TrimSpace(f.reason)); err != nil {
			return nil, err
		}
	}
	if f.comment != "" {
		if _, err := rt.client.AddTaskComment(ctx, id, f.comment); err != nil {
			return nil, err
		}
	}

	// Apply the field update if anything changed.
	updated := current
	if dirty {
		u, err := rt.client.UpdateTask(ctx, id, body)
		if err != nil {
			return nil, err
		}
		updated = u
	}

	// Label add/remove run after the field update so a status transition
	// can't clobber freshly-attached labels.
	for _, raw := range f.addLabels {
		title := normalizeLabelTitle(raw)
		l, err := getOrCreateLabelByTitle(ctx, rt.client, title)
		if err != nil {
			return nil, err
		}
		if err := rt.client.AddLabelToTask(ctx, id, l.ID); err != nil {
			return nil, err
		}
	}
	for _, raw := range f.removeLabels {
		title := normalizeLabelTitle(raw)
		if l := findLabelOnTask(updated, title); l != nil {
			if err := rt.client.RemoveLabelFromTask(ctx, id, l.ID); err != nil {
				return nil, err
			}
		}
	}

	if len(f.addLabels) > 0 || len(f.removeLabels) > 0 {
		fresh, err := rt.client.GetTask(ctx, id)
		if err == nil {
			updated = fresh
		}
	}

	return updated, nil
}

// composeDescription folds --description / --description-replace-* / --description-append
// into the existing body. Returns (new, changed, error).
func composeDescription(existing string, f *updateFlags) (string, bool, error) {
	if f.descriptionIsSet {
		// --description replaces wholesale.
		return f.description, true, nil
	}

	out := existing
	changed := false

	if f.replaceOld != "" || f.replaceNew != "" {
		if f.replaceOld == "" {
			return "", false, output.New(output.CodeValidation, "--description-replace-new requires --description-replace-old")
		}
		count := strings.Count(out, f.replaceOld)
		switch {
		case count == 0:
			return "", false, output.New(output.CodeValidation,
				"--description-replace-old not found in description")
		case count > 1:
			return "", false, output.New(output.CodeValidation,
				"--description-replace-old matched %d times — make it unique", count)
		}
		out = strings.Replace(out, f.replaceOld, f.replaceNew, 1)
		changed = true
	}

	if f.descriptionApp != "" {
		if out != "" && !strings.HasSuffix(out, "\n") {
			out += "\n"
		}
		out += f.descriptionApp
		changed = true
	}

	return out, changed, nil
}
