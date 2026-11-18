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
	parent           string
	noParent         bool
	addBlockedBy     []string
	removeBlockedBy  []string
}

func newUpdateCmd() *cobra.Command {
	f := &updateFlags{}
	var (
		fieldsRaw string
		taskIDArg int64
	)
	cmd := &cobra.Command{
		Use:     "update <PROJ-NN|#NN|NN> | update --id <task id>",
		Aliases: []string{"u"},
		Short:   "Update a task by PROJ-NN, #NN, NN (project index), or --id <task id>",
		RunE: func(cmd *cobra.Command, args []string) error {
			fields, err := parseFields(fieldsRaw)
			if err != nil {
				return err
			}
			rt, err := loadRuntime(cmd.Context())
			if err != nil {
				return err
			}
			f.descriptionIsSet = cmd.Flags().Changed("description")
			f.priorityIsSet = cmd.Flags().Changed("priority")

			id, err := rt.taskID(cmd.Context(), args, taskIDArg)
			if err != nil {
				return err
			}
			task, err := runUpdate(cmd.Context(), rt, id, f)
			if err != nil {
				return err
			}
			return rt.printTask(cmd.OutOrStdout(), fields, task)
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
	cmd.Flags().StringVar(&f.parent, "parent", "", "set the parent task (replaces any existing parent)")
	cmd.Flags().BoolVar(&f.noParent, "no-parent", false, "remove the parent task")
	cmd.MarkFlagsMutuallyExclusive("parent", "no-parent")
	cmd.Flags().StringSliceVar(&f.addBlockedBy, "blocked-by-add", nil, "task IDs that block this one (repeatable)")
	cmd.Flags().StringSliceVar(&f.removeBlockedBy, "blocked-by-remove", nil, "task IDs that no longer block this one (repeatable)")
	cmd.Args = addTaskIDFlag(cmd, &taskIDArg)
	addFieldsFlag(cmd, &fieldsRaw)
	return cmd
}

// runUpdate is intentionally a single linear flow — the steps it performs
// (concurrency check → status → field changes → comments → field PATCH →
// bucket move → label add/remove → refetch) all share the same task,
// flag set, and error-handling shape. Splitting them produces five tiny
// functions that each take the same five arguments.
//
//nolint:gocyclo // single-pass orchestration; each branch is one short stanza
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

	var parentID int64
	if f.parent != "" {
		if parentID, err = rt.resolveTaskID(ctx, f.parent); err != nil {
			return nil, err
		}
		if parentID == id {
			return nil, output.New(output.CodeValidation, "a task cannot be its own parent")
		}
	}

	addBlockers, err := resolveRefs(ctx, rt, f.addBlockedBy)
	if err != nil {
		return nil, err
	}
	removeBlockers, err := resolveRefs(ctx, rt, f.removeBlockedBy)
	if err != nil {
		return nil, err
	}

	// Build the merge-patch payload from only the changed fields. PATCH leaves
	// absent fields untouched, so omitting a field preserves it — the id rides
	// in the URL, not the body.
	body := &client.TaskPatch{}
	dirty := false

	if f.title != "" {
		body.Title = &f.title
		dirty = true
	}
	if f.priorityIsSet {
		body.Priority = &f.priority
		dirty = true
	}

	// Description ops are mutually-exclusive layers; --description wins
	// outright, otherwise replace-old/new + append run on the current body.
	newDesc, descChanged, err := composeDescription(current.Description, f)
	if err != nil {
		return nil, err
	}
	if descChanged {
		body.Description = &newDesc
		dirty = true
	}

	// Status transitions: `done` is set on the task body (Update processes
	// it natively), but the bucket move uses the dedicated TaskBucket
	// endpoint after the field update so the change is visible on the
	// Kanban view.
	var bucketTransitionTarget int64
	if newStatus != "" {
		bid, err := status.BucketID(newStatus, rt.cfg.Buckets)
		if err != nil {
			return nil, err
		}
		bucketTransitionTarget = bid
		done := newStatus.Done()
		body.Done = &done
		dirty = true
	}

	progress := &steps{task: rt.cfg.FormatTaskID(current.Index)}

	// Comment first when transitioning to scrapped — the reason is part of
	// the audit trail and should appear before the bucket move in the log.
	if newStatus == status.Scrapped {
		if _, err := rt.client.AddTaskComment(ctx, id, "<strong>Scrapped:</strong> "+strings.TrimSpace(f.reason)); err != nil {
			return nil, progress.fail("scrapped comment", err)
		}
		progress.done("scrapped comment")
	}
	if f.comment != "" {
		if _, err := rt.client.AddTaskComment(ctx, id, f.comment); err != nil {
			return nil, progress.fail("comment", err)
		}
		progress.done("comment")
	}

	// Apply the field update if anything changed.
	updated := current
	if dirty {
		u, err := rt.client.UpdateTask(ctx, id, body)
		if err != nil {
			return nil, progress.fail("update fields", err)
		}
		updated = u
		progress.done("update fields")
	}

	// Move the task between buckets after the field update.
	if bucketTransitionTarget != 0 {
		if err := rt.client.MoveTaskToBucket(ctx,
			rt.cfg.ProjectID, rt.cfg.ViewID, bucketTransitionTarget, id); err != nil {
			return nil, progress.fail("set bucket", err)
		}
		progress.done("set bucket")
	}

	// Label add/remove run after the field update so a status transition
	// can't clobber freshly-attached labels.
	for _, raw := range f.addLabels {
		title := normalizeLabelTitle(raw)
		step := "label-add " + title
		l, err := getOrCreateLabelByTitle(ctx, rt.client, title)
		if err != nil {
			return nil, progress.fail(step, err)
		}
		if err := rt.client.AddLabelToTask(ctx, id, l.ID); err != nil {
			return nil, progress.fail(step, err)
		}
		progress.done(step)
	}
	for _, raw := range f.removeLabels {
		title := normalizeLabelTitle(raw)
		if l := findLabelOnTask(updated, title); l != nil {
			step := "label-remove " + title
			if err := rt.client.RemoveLabelFromTask(ctx, id, l.ID); err != nil {
				return nil, progress.fail(step, err)
			}
			progress.done(step)
		}
	}

	parentChanged := f.parent != "" || f.noParent
	if parentChanged {
		if err := setParent(ctx, rt, current, parentID, progress); err != nil {
			return nil, err
		}
	}

	blocked := map[int64]bool{}
	for _, b := range current.RelatedTasks["blocked"] {
		if b != nil {
			blocked[b.ID] = true
		}
	}
	for i, blockerID := range addBlockers {
		if blocked[blockerID] {
			continue
		}
		step := "blocked-by-add " + f.addBlockedBy[i]
		if _, err := rt.client.CreateRelation(ctx, id, blockerID, "blocked"); err != nil {
			return nil, progress.fail(step, err)
		}
		blocked[blockerID] = true
		progress.done(step)
	}
	for i, blockerID := range removeBlockers {
		if !blocked[blockerID] {
			continue
		}
		step := "blocked-by-remove " + f.removeBlockedBy[i]
		if err := rt.client.DeleteRelation(ctx, id, blockerID, "blocked"); err != nil {
			return nil, progress.fail(step, err)
		}
		delete(blocked, blockerID)
		progress.done(step)
	}
	relationsChanged := parentChanged || len(addBlockers) > 0 || len(removeBlockers) > 0

	if len(f.addLabels) > 0 || len(f.removeLabels) > 0 || bucketTransitionTarget != 0 || relationsChanged {
		fresh, err := rt.client.GetTask(ctx, id)
		if err == nil {
			updated = fresh
		}
	}

	return updated, nil
}

// composeDescription folds --description / --description-replace-* / --description-append
// into the existing body. Returns (new, changed, error).
func resolveRefs(ctx context.Context, rt *runtime, refs []string) ([]int64, error) {
	ids := make([]int64, len(refs))
	for i, ref := range refs {
		var err error
		if ids[i], err = rt.resolveTaskID(ctx, ref); err != nil {
			return nil, err
		}
	}
	return ids, nil
}

// setParent makes parentID the task's only parent; 0 removes all parents.
func setParent(ctx context.Context, rt *runtime, task *client.Task, parentID int64, progress *steps) error {
	hasParent := false
	for _, p := range task.RelatedTasks["parenttask"] {
		if p == nil {
			continue
		}
		if p.ID == parentID {
			hasParent = true
			continue
		}
		step := fmt.Sprintf("remove parent %d", p.ID)
		if p.Identifier != "" {
			step = "remove parent " + p.Identifier
		}
		if err := rt.client.DeleteRelation(ctx, task.ID, p.ID, "parenttask"); err != nil {
			return progress.fail(step, err)
		}
		progress.done(step)
	}
	if parentID == 0 || hasParent {
		return nil
	}
	if _, err := rt.client.CreateRelation(ctx, task.ID, parentID, "parenttask"); err != nil {
		return progress.fail("set parent", err)
	}
	progress.done("set parent")
	return nil
}

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
