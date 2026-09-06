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

package cmd

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/initialize"
	"code.vikunja.io/api/pkg/license"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"github.com/spf13/cobra"
	"xorm.io/xorm"
)

var (
	botFlagAdmin   bool
	botFlagScopes  string
	botFlagPreset  string
	botFlagExpires string
	botFlagTitle   string
)

// Presets deliberately leave out users_set_password and users_set_admin
// (root-equivalent) and projects_list.
var botScopePresets = map[string][]string{
	"provisioning": {"users_list", "users_create", "users_set_status", "users_delete"},
}

const botDefaultExpiry = "1y"

func init() {
	for _, c := range []*cobra.Command{userBotCreateCmd, userBotTokenCreateCmd} {
		c.Flags().StringVar(&botFlagScopes, "scopes", "", "Comma-separated group:permission scopes, e.g. admin:users_list,admin:users_create. Only admin scopes are allowed.")
		c.Flags().StringVar(&botFlagPreset, "preset", "", "Scope preset to add: provisioning (users_list, users_create, users_set_status, users_delete).")
		c.Flags().StringVar(&botFlagExpires, "expires", botDefaultExpiry, "Token lifetime as days (90d), years (1y) or an RFC3339 timestamp.")
		c.Flags().StringVar(&botFlagTitle, "title", "", "Title of the token.")
	}
	userBotCreateCmd.Flags().BoolVar(&botFlagAdmin, "admin", false, "Create an instance admin bot. Required; there is no other kind of instance bot yet.")

	userBotTokenCmd.AddCommand(userBotTokenCreateCmd, userBotTokenListCmd, userBotTokenRevokeCmd)
	userBotCmd.AddCommand(userBotCreateCmd, userBotListCmd, userBotDeleteCmd, userBotTokenCmd)
	userCmd.AddCommand(userBotCmd)
}

func fullInit(_ *cobra.Command, _ []string) {
	initialize.FullInit()
}

// Errors past flag parsing are domain errors; Execute prints them, without usage.
func botRun(fn func(cmd *cobra.Command, args []string) error) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		cmd.SilenceUsage = true
		cmd.SilenceErrors = true
		return fn(cmd, args)
	}
}

var userBotCmd = &cobra.Command{
	Use:   "bot",
	Short: "Manage instance-owned admin bots and their API tokens.",
}

var userBotCreateCmd = &cobra.Command{
	Use:    "create [username]",
	Short:  "Create an instance admin bot and print its first API token.",
	Args:   cobra.ExactArgs(1),
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, args []string) error {
		if !botFlagAdmin {
			return fmt.Errorf("instance bots must be admin bots; pass --admin")
		}
		if err := requireAdminPanelLicense(); err != nil {
			return err
		}
		perms, err := parseBotScopes(botFlagScopes, botFlagPreset)
		if err != nil {
			return err
		}
		expires, err := parseBotExpiry(botFlagExpires, time.Now())
		if err != nil {
			return err
		}

		s := db.NewSession()
		defer s.Close()

		bot, err := user.CreateInstanceBotUser(s, &user.User{Username: args[0]})
		if err != nil {
			_ = s.Rollback()
			return fmt.Errorf("could not create bot: %w", err)
		}
		token, err := mintBotToken(s, bot, perms, expires, botFlagTitle)
		if err != nil {
			_ = s.Rollback()
			return err
		}
		if err := commitAndDispatch(s); err != nil {
			return err
		}

		out := cmd.OutOrStdout()
		fmt.Fprintf(out, "Created instance admin bot %q (ID %d).\n", bot.Username, bot.ID)
		printBotToken(out, token)
		return nil
	}),
}

var userBotListCmd = &cobra.Command{
	Use:    "list",
	Short:  "List instance bots.",
	Args:   cobra.NoArgs,
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, _ []string) error {
		s := db.NewSession()
		defer s.Close()

		bots := []*user.User{}
		if err := s.Where("is_instance_bot = ?", true).OrderBy("id").Find(&bots); err != nil {
			return fmt.Errorf("could not list bots: %w", err)
		}

		table := tablewriter.NewTable(cmd.OutOrStdout(),
			tablewriter.WithHeader([]string{"ID", "Username", "Status", "Created"}),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft}),
		)
		for _, b := range bots {
			_ = table.Append([]string{strconv.FormatInt(b.ID, 10), b.Username, b.Status.String(), b.Created.Format(time.RFC3339)})
		}
		return table.Render()
	}),
}

var userBotDeleteCmd = &cobra.Command{
	Use:    "delete [username]",
	Short:  "Delete an instance bot and all of its tokens.",
	Args:   cobra.ExactArgs(1),
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, args []string) error {
		s := db.NewSession()
		defer s.Close()

		bot, err := getInstanceBot(s, args[0])
		if err != nil {
			return err
		}
		if err := models.DeleteUser(s, bot); err != nil {
			_ = s.Rollback()
			return fmt.Errorf("could not delete bot: %w", err)
		}
		if err := commitAndDispatch(s); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Deleted instance bot %q.\n", bot.Username)
		return nil
	}),
}

var userBotTokenCmd = &cobra.Command{
	Use:   "token",
	Short: "Manage API tokens of instance bots.",
}

var userBotTokenCreateCmd = &cobra.Command{
	Use:    "create [bot-username]",
	Short:  "Mint a new API token for an instance bot and print it.",
	Args:   cobra.ExactArgs(1),
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, args []string) error {
		if err := requireAdminPanelLicense(); err != nil {
			return err
		}
		perms, err := parseBotScopes(botFlagScopes, botFlagPreset)
		if err != nil {
			return err
		}
		expires, err := parseBotExpiry(botFlagExpires, time.Now())
		if err != nil {
			return err
		}

		s := db.NewSession()
		defer s.Close()

		bot, err := getInstanceBot(s, args[0])
		if err != nil {
			return err
		}
		token, err := mintBotToken(s, bot, perms, expires, botFlagTitle)
		if err != nil {
			_ = s.Rollback()
			return err
		}
		if err := commitAndDispatch(s); err != nil {
			return err
		}
		printBotToken(cmd.OutOrStdout(), token)
		return nil
	}),
}

var userBotTokenListCmd = &cobra.Command{
	Use:    "list [bot-username]",
	Short:  "List the API tokens of an instance bot.",
	Args:   cobra.ExactArgs(1),
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, args []string) error {
		s := db.NewSession()
		defer s.Close()

		bot, err := getInstanceBot(s, args[0])
		if err != nil {
			return err
		}
		tokens := []*models.APIToken{}
		if err := s.Where("owner_id = ?", bot.ID).OrderBy("id").Find(&tokens); err != nil {
			return fmt.Errorf("could not list tokens: %w", err)
		}

		table := tablewriter.NewTable(cmd.OutOrStdout(),
			tablewriter.WithHeader([]string{"ID", "Title", "Scopes", "Expires", "Created"}),
			tablewriter.WithAlignment(tw.Alignment{tw.AlignLeft}),
		)
		for _, t := range tokens {
			_ = table.Append([]string{
				strconv.FormatInt(t.ID, 10),
				t.Title,
				formatBotScopes(t.APIPermissions),
				t.ExpiresAt.Format(time.RFC3339),
				t.Created.Format(time.RFC3339),
			})
		}
		return table.Render()
	}),
}

var userBotTokenRevokeCmd = &cobra.Command{
	Use:    "revoke [token-id]",
	Short:  "Revoke an instance bot API token.",
	Args:   cobra.ExactArgs(1),
	PreRun: fullInit,
	RunE: botRun(func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid token id %q", args[0])
		}

		s := db.NewSession()
		defer s.Close()

		token, err := models.GetAPITokenByID(s, id)
		if err != nil {
			return fmt.Errorf("could not load token: %w", err)
		}
		if token.ID == 0 {
			return fmt.Errorf("token %d does not exist", id)
		}
		bot, err := user.GetUserByID(s, token.OwnerID)
		if err != nil && !user.IsErrUserStatusError(err) {
			return err
		}
		if !bot.IsInstanceBot {
			return fmt.Errorf("token %d does not belong to an instance bot", id)
		}
		if err := token.RevokeInstanceBotToken(s, bot); err != nil {
			_ = s.Rollback()
			return fmt.Errorf("could not revoke token: %w", err)
		}
		if err := commitAndDispatch(s); err != nil {
			return err
		}
		fmt.Fprintf(cmd.OutOrStdout(), "Revoked token %d of bot %q.\n", token.ID, bot.Username)
		return nil
	}),
}

func requireAdminPanelLicense() error {
	if !license.IsFeatureEnabled(license.FeatureAdminPanel) {
		return fmt.Errorf("the admin-panel license feature is not active; refusing to create an admin bot token")
	}
	return nil
}

func getInstanceBot(s *xorm.Session, username string) (*user.User, error) {
	bot, err := user.GetUserByUsername(s, username)
	if err != nil && !user.IsErrUserStatusError(err) {
		return nil, fmt.Errorf("could not load bot: %w", err)
	}
	if !bot.IsInstanceBot {
		return nil, fmt.Errorf("%q is not an instance bot", username)
	}
	return bot, nil
}

func mintBotToken(s *xorm.Session, bot *user.User, perms models.APIPermissions, expires time.Time, title string) (*models.APIToken, error) {
	if title == "" {
		title = bot.Username
	}
	token := &models.APIToken{Title: title, APIPermissions: perms, ExpiresAt: expires}
	if err := token.CreateInstanceBotToken(s, bot); err != nil {
		return nil, fmt.Errorf("could not create token: %w", err)
	}
	return token, nil
}

func commitAndDispatch(s *xorm.Session) error {
	if err := s.Commit(); err != nil {
		return fmt.Errorf("could not commit: %w", err)
	}
	events.DispatchPending(context.Background(), s)
	return nil
}

// The token is the last line so scripts can `tail -n1` it.
func printBotToken(out io.Writer, token *models.APIToken) {
	fmt.Fprintf(out, "Token %d (%s) expires at %s. It is shown only once:\n", token.ID, formatBotScopes(token.APIPermissions), token.ExpiresAt.Format(time.RFC3339))
	fmt.Fprintln(out, token.Token)
}

func formatBotScopes(perms models.APIPermissions) string {
	scopes := []string{}
	for group, list := range perms {
		for _, p := range list {
			scopes = append(scopes, group+":"+p)
		}
	}
	sort.Strings(scopes)
	return strings.Join(scopes, ",")
}

// parseBotScopes unions --scopes (group:perm, the frontend deep-link syntax)
// with --preset; validity beyond the admin-only rule is checked on create.
func parseBotScopes(scopes, preset string) (models.APIPermissions, error) {
	perms := models.APIPermissions{}
	add := func(group, perm string) {
		for _, existing := range perms[group] {
			if existing == perm {
				return
			}
		}
		perms[group] = append(perms[group], perm)
	}

	for _, scope := range strings.Split(scopes, ",") {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		group, perm, ok := strings.Cut(scope, ":")
		if !ok || group == "" || perm == "" {
			return nil, fmt.Errorf("invalid scope %q, expected group:permission", scope)
		}
		if group != "admin" {
			return nil, &models.ErrInstanceBotScopeNotAllowed{Group: group}
		}
		add(group, perm)
	}

	if preset != "" {
		list, ok := botScopePresets[preset]
		if !ok {
			return nil, fmt.Errorf("unknown preset %q", preset)
		}
		for _, perm := range list {
			add("admin", perm)
		}
	}

	if len(perms) == 0 {
		return nil, fmt.Errorf("at least one scope is required; pass --scopes or --preset")
	}
	return perms, nil
}

// parseBotExpiry accepts Nd, Ny or an RFC3339 timestamp. There is no "never".
func parseBotExpiry(value string, now time.Time) (time.Time, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, fmt.Errorf("expiry is required")
	}

	if unit := value[len(value)-1]; unit == 'd' || unit == 'y' {
		n, err := strconv.Atoi(value[:len(value)-1])
		if err == nil && n > 0 {
			if unit == 'd' {
				return now.AddDate(0, 0, n), nil
			}
			return now.AddDate(n, 0, 0), nil
		}
	}

	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid expiry %q, use e.g. 90d, 1y or an RFC3339 timestamp", value)
	}
	if !at.After(now) {
		return time.Time{}, fmt.Errorf("expiry %s is in the past", value)
	}
	return at, nil
}
