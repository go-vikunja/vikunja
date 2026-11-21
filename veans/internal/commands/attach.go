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
	"os"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/output"
)

func addAttachFlag(cmd *cobra.Command, paths *[]string) {
	cmd.Flags().StringArrayVar(paths, "attach", nil, "path of a file to upload as attachment (repeatable)")
}

// checkAttachPaths runs before any mutation so a typo can't leave a half-applied change.
func checkAttachPaths(paths []string) error {
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			return output.Wrap(output.CodeValidation, err, "--attach %s: %v", p, err)
		}
		if !fi.Mode().IsRegular() {
			return output.New(output.CodeValidation, "--attach %s: not a regular file", p)
		}
	}
	return nil
}

func attachFiles(ctx context.Context, rt *runtime, taskID int64, paths []string, progress *steps) error {
	for _, p := range paths {
		step := "attach " + p
		if _, err := rt.client.UploadTaskAttachment(ctx, taskID, p); err != nil {
			return progress.fail(step, err)
		}
		progress.done(step)
	}
	return nil
}
