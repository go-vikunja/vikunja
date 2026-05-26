package commands

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/bootstrap"
	"code.vikunja.io/veans/internal/config"
)

type initFlags struct {
	server        string
	token         string
	username      string
	password      string
	totp          string
	botUsername   string
	projectID     int64
	viewID        int64
	yesBuckets    bool
	skipBuckets   bool
	configPath    string
}

func newInitCmd() *cobra.Command {
	f := &initFlags{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Provision a Vikunja bot user and write .veans.yml",
		Long: `Onboards veans into the current repository:

  1. Authenticate as you (--token, or username/password)
  2. Pick a Vikunja project and Kanban view
  3. Bootstrap canonical buckets (Todo / In Progress / In Review / Done / Scrapped)
  4. Create a 'bot-<repo>' user, share the project with it, mint its API token
  5. Store the bot's token in your keychain (or ~/.config/veans/credentials.yml)
  6. Write .veans.yml to the repository root

The token stored locally belongs to the bot, not to you — you can rotate or
revoke it at any time without affecting your own session.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path := f.configPath
			if path == "" {
				root, err := config.RepoRoot("")
				if err != nil {
					return err
				}
				path = filepath.Join(root, config.Filename)
			}
			res, err := bootstrap.Init(cmd.Context(), &bootstrap.Options{
				ConfigPath:           path,
				Server:               f.server,
				HumanToken:           f.token,
				HumanUsername:        f.username,
				HumanPassword:        f.password,
				HumanTOTP:            f.totp,
				BotUsername:          f.botUsername,
				ProjectID:            f.projectID,
				ViewID:               f.viewID,
				AutoApproveBuckets:   f.yesBuckets,
				SkipBucketBootstrap:  f.skipBuckets,
				Out:                  os.Stderr,
			})
			if err != nil {
				return err
			}
			printPostInitSummary(cmd.OutOrStdout(), res)
			return nil
		},
	}

	cmd.Flags().StringVar(&f.server, "server", "", "Vikunja server URL")
	cmd.Flags().StringVar(&f.token, "token", "", "JWT or personal API token (skips password prompt; useful for SSO/OIDC instances)")
	cmd.Flags().StringVar(&f.username, "username", "", "Vikunja username (prompted if empty)")
	cmd.Flags().StringVar(&f.password, "password", "", "Vikunja password (prompted if empty; usually safer to omit)")
	cmd.Flags().StringVar(&f.totp, "totp", "", "TOTP code if your account requires 2FA")
	cmd.Flags().StringVar(&f.botUsername, "bot-username", "", "override the bot-<repo> default")
	cmd.Flags().Int64Var(&f.projectID, "project", 0, "skip the interactive project picker")
	cmd.Flags().Int64Var(&f.viewID, "view", 0, "skip the interactive view picker")
	cmd.Flags().BoolVar(&f.yesBuckets, "yes-buckets", false, "auto-approve canonical bucket bootstrap")
	cmd.Flags().BoolVar(&f.skipBuckets, "skip-buckets", false, "do not prompt or create buckets (assumes they exist)")
	cmd.Flags().StringVar(&f.configPath, "config", "", "where to write .veans.yml (defaults to the repo root)")

	return cmd
}

func printPostInitSummary(w fmtWriter, res *bootstrap.Result) {
	fmt.Fprintf(w, "\nveans is ready. Bot user: %s\n", res.BotUser.Username)
	fmt.Fprintf(w, "Config:    %s\n", res.Config.Path())
	fmt.Fprintf(w, "Project:   #%d %s\n", res.Config.ProjectID, identOrFallback(res.Config.ProjectIdentifier))

	fmt.Fprintln(w, `
To wire veans into your coding agent, paste one of these snippets:

Claude Code (.claude/settings.json):
  {
    "hooks": {
      "SessionStart": [{ "hooks": [{ "type": "command", "command": "veans prime" }] }],
      "PreCompact":   [{ "hooks": [{ "type": "command", "command": "veans prime" }] }]
    }
  }

OpenCode (.opencode/plugin/veans-prime.ts): see veans/README.md`)
}

func identOrFallback(s string) string {
	if s == "" {
		return "(no identifier — task IDs render as #NN)"
	}
	return s
}

// fmtWriter is what cobra.Cmd.OutOrStdout returns — type aliased to keep the
// import surface minimal.
type fmtWriter = interface {
	Write(p []byte) (int, error)
}
