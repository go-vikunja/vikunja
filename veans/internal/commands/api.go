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
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/output"
)

func newAPICmd() *cobra.Command {
	var (
		dataFlag  string
		queryFlag []string
		dataFile  string
	)
	cmd := &cobra.Command{
		Use:   "api <METHOD> <PATH>",
		Short: "Raw REST passthrough — escape hatch for endpoints veans doesn't wrap",
		Long: `Sends a request to /api/v1<PATH> as the bot. Use this when curated
commands don't shape the data the way you need. The response body is
written to stdout verbatim.

Examples:
  veans api GET /projects
  veans api GET /tasks/123
  veans api POST /tasks/123 --data '{"description":"updated"}'
  veans api GET /tasks --query expand=reactions --query per_page=100`,
		Args: cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			rt, err := loadRuntime()
			if err != nil {
				return err
			}
			method := strings.ToUpper(args[0])
			path := args[1]
			if !strings.HasPrefix(path, "/") {
				path = "/" + path
			}

			query := url.Values{}
			for _, kv := range queryFlag {
				eq := strings.Index(kv, "=")
				if eq < 0 {
					return output.New(output.CodeValidation, "--query must be key=value: %q", kv)
				}
				query.Add(kv[:eq], kv[eq+1:])
			}

			var body []byte
			switch {
			case dataFile == "-":
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}
				body = b
			case dataFile != "":
				b, err := os.ReadFile(dataFile)
				if err != nil {
					return err
				}
				body = b
			case dataFlag != "":
				body = []byte(dataFlag)
			}

			status, respBody, retryAfter, err := rt.client.DoRaw(cmd.Context(), method, path, query, body)
			if err != nil {
				return err
			}
			// On non-2xx, do NOT write the body to stdout — the agent
			// contract is "stdout is the success payload". Fold a short
			// snippet of the upstream error into the envelope message so
			// the agent gets actionable context without a separate channel
			// to parse.
			if status >= 400 {
				snippet := strings.TrimSpace(string(respBody))
				if len(snippet) > maxAPIErrorSnippet {
					snippet = snippet[:maxAPIErrorSnippet] + "…(truncated)"
				}
				msg := fmt.Sprintf("HTTP %d %s %s", status, method, path)
				if snippet != "" {
					msg = fmt.Sprintf("%s: %s", msg, snippet)
				}
				if retryAfter > 0 {
					msg = fmt.Sprintf("%s (retry-after %s)", msg, retryAfter)
				}
				return output.New(mapStatusToCode(status), "%s", msg)
			}
			if _, werr := cmd.OutOrStdout().Write(respBody); werr != nil {
				return werr
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&dataFlag, "data", "", "request body (raw)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "read request body from file (`-` = stdin)")
	cmd.Flags().StringSliceVar(&queryFlag, "query", nil, "query parameter, key=value (repeatable)")
	return cmd
}

// maxAPIErrorSnippet caps how much upstream-error body we fold into the
// `error` envelope field. Anything longer is almost always an HTML page.
const maxAPIErrorSnippet = 512

func mapStatusToCode(status int) output.Code {
	switch {
	case status == 401, status == 403:
		return output.CodeAuth
	case status == 404:
		return output.CodeNotFound
	case status == 409:
		return output.CodeConflict
	case status == 429:
		return output.CodeRateLimited
	case status >= 400 && status < 500:
		return output.CodeValidation
	default:
		return output.CodeUnknown
	}
}
