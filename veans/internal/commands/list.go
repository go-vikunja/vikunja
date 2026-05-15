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
	"strings"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
	"code.vikunja.io/veans/internal/status"
)

type listFlags struct {
	ready    bool
	mine     bool
	branch   string
	filter   string
	statuses []string
}

func newListCmd() *cobra.Command {
	f := &listFlags{}
	cmd := &cobra.Command{
		Use:     "list",
		Aliases: []string{"ls"},
		Short:   "List tasks in the configured project",
		Long: `List tasks in the project configured in .veans.yml.

Filters can be combined; they're AND-ed together:
  --ready          ready to start: in Todo with done=false (incomplete-blocker
                   detection is best-effort, see veans/README.md)
  --mine           only tasks assigned to the veans bot
  --branch [name]  only tasks tagged 'veans:branch:<name>' (defaults to the
                   current git branch when used without a value)
  --filter <expr>  raw Vikunja filter expression (see Vikunja docs); applied
                   server-side
  --status <s>     filter by status (todo|in-progress|in-review|completed|scrapped),
                   may be repeated`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			tasks, err := runList(cmd, rt, f)
			if err != nil {
				return err
			}
			return json.NewEncoder(cmd.OutOrStdout()).Encode(tasks)
		},
	}
	cmd.Flags().BoolVar(&f.ready, "ready", false, "only ready-to-start tasks (Todo bucket, not done)")
	cmd.Flags().BoolVar(&f.mine, "mine", false, "only tasks assigned to the veans bot")
	cmd.Flags().StringVar(&f.branch, "branch", "", "only tasks tagged 'veans:branch:<name>' (omit value for current branch)")
	cmd.Flags().Lookup("branch").NoOptDefVal = "__auto__"
	cmd.Flags().StringVar(&f.filter, "filter", "", "raw Vikunja filter expression, applied server-side")
	cmd.Flags().StringSliceVar(&f.statuses, "status", nil, "filter by status (repeatable)")
	return cmd
}

func runList(cmd *cobra.Command, rt *runtime, f *listFlags) ([]*client.Task, error) {
	opts := &client.TaskListOptions{
		Filter: f.filter,
		Expand: []string{"reactions"},
	}
	tasks, err := rt.client.ListProjectTasks(cmd.Context(), rt.cfg.ProjectID, opts)
	if err != nil {
		return nil, err
	}

	// Apply client-side filters AND-style.
	var out []*client.Task
	for _, t := range tasks {
		taskBucket := t.CurrentBucketID(rt.cfg.ViewID)
		if f.ready {
			if t.Done || taskBucket != rt.cfg.Buckets.Todo {
				continue
			}
		}
		if f.mine {
			if !taskAssignedTo(t, rt.cfg.Bot.UserID) {
				continue
			}
		}
		if f.branch != "" {
			want := f.branch
			if want == "__auto__" {
				want = currentGitBranch(cmd.Context())
				if want == "" {
					return nil, output.New(output.CodeValidation,
						"--branch given without a value but no current git branch detected")
				}
			}
			label := branchLabel(want)
			if !taskHasLabel(t, label) {
				continue
			}
		}
		if len(f.statuses) > 0 {
			ok := false
			for _, raw := range f.statuses {
				s, err := status.Parse(raw)
				if err != nil {
					return nil, err
				}
				wantBucket, _ := status.BucketID(s, rt.cfg.Buckets)
				if taskBucket == wantBucket {
					ok = true
					break
				}
			}
			if !ok {
				continue
			}
		}
		out = append(out, t)
	}
	return out, nil
}

func taskAssignedTo(t *client.Task, userID int64) bool {
	for _, a := range t.Assignees {
		if a != nil && a.ID == userID {
			return true
		}
	}
	return false
}

func taskHasLabel(t *client.Task, title string) bool {
	for _, l := range t.Labels {
		if l != nil && l.Title == title {
			return true
		}
	}
	return false
}

func branchLabel(branch string) string {
	return "veans:branch:" + branch
}

// currentGitBranch returns the current git branch as reported by
// `git rev-parse --abbrev-ref HEAD`, or "" if we're not in a git repo or
// HEAD is detached. Failures are silent so callers can decide.
func currentGitBranch(ctx context.Context) string {
	out, err := runGit(ctx, "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return ""
	}
	out = strings.TrimSpace(out)
	if out == "HEAD" {
		return ""
	}
	return out
}
