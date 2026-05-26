package commands

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/status"
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
			if globals.JSON {
				return json.NewEncoder(cmd.OutOrStdout()).Encode(task)
			}
			renderTaskHuman(cmd.OutOrStdout(), task, rt.cfg)
			return nil
		},
	}
	return cmd
}

func renderTaskHuman(w fmtWriter, t *client.Task, cfg *config.Config) {
	s := status.FromBucketID(t.BucketID, cfg.Buckets)
	fmt.Fprintf(w, "%s  %s  [%s]\n", cfg.FormatTaskID(t.Index), t.Title, s)
	if t.Priority > 0 {
		fmt.Fprintf(w, "Priority: %d\n", t.Priority)
	}
	if len(t.Assignees) > 0 {
		fmt.Fprintf(w, "Assignees: ")
		for i, a := range t.Assignees {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, a.Username)
		}
		fmt.Fprintln(w)
	}
	if len(t.Labels) > 0 {
		fmt.Fprintf(w, "Labels: ")
		for i, l := range t.Labels {
			if i > 0 {
				fmt.Fprint(w, ", ")
			}
			fmt.Fprint(w, l.Title)
		}
		fmt.Fprintln(w)
	}
	if t.Description != "" {
		fmt.Fprintln(w)
		fmt.Fprintln(w, t.Description)
	}
}
