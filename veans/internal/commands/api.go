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

			status, respBody, err := rt.client.DoRaw(cmd.Context(), method, path, query, body)
			if err != nil {
				return err
			}
			// On non-2xx, write the body to stderr and exit non-zero so
			// shell pipelines see the failure clearly.
			if status >= 400 {
				fmt.Fprintf(cmd.ErrOrStderr(), "HTTP %d %s %s\n", status, method, path)
				cmd.OutOrStdout().Write(respBody)
				return output.New(mapStatusToCode(status), "HTTP %d", status)
			}
			cmd.OutOrStdout().Write(respBody)
			return nil
		},
	}
	cmd.Flags().StringVar(&dataFlag, "data", "", "request body (raw)")
	cmd.Flags().StringVar(&dataFile, "data-file", "", "read request body from file (`-` = stdin)")
	cmd.Flags().StringSliceVar(&queryFlag, "query", nil, "query parameter, key=value (repeatable)")
	return cmd
}

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
