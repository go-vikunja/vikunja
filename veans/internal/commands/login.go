package commands

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"code.vikunja.io/veans/internal/auth"
	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/credentials"
	"code.vikunja.io/veans/internal/output"
)

func newLoginCmd() *cobra.Command {
	var (
		token       string
		username    string
		password    string
		totp        string
		usePassword bool
	)
	cmd := &cobra.Command{
		Use:   "login",
		Short: "Mint a fresh API token for the bot user (rotation)",
		Long: `Re-authenticates as you (the bot's owner) and mints a new API token
for the bot configured in .veans.yml. The new token replaces the
existing one in the credential store.

The default flow is OAuth 2.0 Authorization Code + PKCE — open the
URL veans prints, sign in, and paste the callback URL back. Use
--token to paste in a personal API token, or --use-password / --username
to force POST /login instead.

Use this after revoking the bot's token in Vikunja's UI, or any time
you want to rotate.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := config.Find("")
			if err != nil {
				if errors.Is(err, config.ErrNotFound) {
					return output.Wrap(output.CodeNotConfigured, err,
						"no .veans.yml found — run `veans init` first")
				}
				return err
			}
			cfg, err := config.Load(path)
			if err != nil {
				return err
			}

			human := client.New(cfg.Server, "")
			tok, err := auth.AcquireHumanToken(cmd.Context(), human, auth.LoginOptions{
				Token:       token,
				UsePassword: usePassword,
				Username:    username,
				Password:    password,
				TOTP:        totp,
				Out:         os.Stderr,
			}, auth.NewStdPrompter())
			if err != nil {
				return err
			}
			human.Token = tok

			routes, err := human.Routes(cmd.Context())
			if err != nil {
				return output.Wrap(output.CodeUnknown, err, "fetch /routes: %v", err)
			}
			perms := client.PermissionsForBot(routes)
			if len(perms) == 0 {
				return output.New(output.CodeUnknown, "no API token permissions available")
			}

			minted, err := human.CreateToken(cmd.Context(), &client.APIToken{
				Title:       "veans (rotated)",
				Permissions: perms,
				ExpiresAt:   client.FarFuture,
				OwnerID:     cfg.Bot.UserID,
			})
			if err != nil {
				return err
			}
			if minted.Token == "" {
				return output.New(output.CodeUnknown, "PUT /tokens did not return token plaintext")
			}

			if err := credentials.Default().Set(cfg.Server, cfg.Bot.Username, minted.Token); err != nil {
				return err
			}

			fmt.Fprintf(os.Stderr, "Rotated token for %s on %s\n", cfg.Bot.Username, cfg.Server)
			return nil
		},
	}
	cmd.Flags().StringVar(&token, "token", "", "JWT or personal API token (skips OAuth/password)")
	cmd.Flags().BoolVar(&usePassword, "use-password", false, "use POST /login instead of the default OAuth flow")
	cmd.Flags().StringVar(&username, "username", "", "your Vikunja username (implies --use-password)")
	cmd.Flags().StringVar(&password, "password", "", "your Vikunja password (implies --use-password; prompted if empty)")
	cmd.Flags().StringVar(&totp, "totp", "", "TOTP code if your account requires 2FA")
	return cmd
}
