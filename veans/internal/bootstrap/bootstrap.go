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

// Package bootstrap orchestrates `veans init`. It chains together the steps
// outlined in the plan: probe /info, acquire the human's transient token,
// pick or create a project, designate a Kanban view, bootstrap canonical
// buckets, create the bot user, share the project with the bot, mint the
// bot's API token, and write .veans.yml.
//
// The flow is split into small functions so e2e tests can drive it with
// scripted answers without going through the cobra command surface.
package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"

	"code.vikunja.io/veans/internal/auth"
	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/config"
	"code.vikunja.io/veans/internal/credentials"
	"code.vikunja.io/veans/internal/output"
	"code.vikunja.io/veans/internal/status"
)

// Options configures Init. All fields are optional unless noted; missing
// values are filled in interactively from Prompter.
type Options struct {
	// ConfigPath is where .veans.yml will be written. Required.
	ConfigPath string

	// Server is the Vikunja base URL (e.g. https://vikunja.example.com).
	// If empty, the prompter asks.
	Server string

	// HumanToken short-circuits all auth when set.
	HumanToken string
	// HumanUsePassword forces POST /login instead of the default OAuth flow.
	HumanUsePassword bool
	// HumanUsername / HumanPassword feed POST /login (used when set).
	HumanUsername string
	HumanPassword string
	HumanTOTP     string

	// BotUsername overrides the bot-<reponame> default. The "bot-" prefix is
	// auto-prepended if missing — Vikunja will reject otherwise.
	BotUsername string

	// ProjectID, when non-zero, skips the interactive project picker.
	ProjectID int64

	// ViewID, when non-zero, skips the interactive view picker.
	ViewID int64

	// Bucket bootstrap behavior:
	//   AutoApproveBuckets — skip the prompt, create missing canonical buckets.
	//   SkipBucketBootstrap — neither prompt nor create.
	AutoApproveBuckets  bool
	SkipBucketBootstrap bool

	// Store and Prompter are dependency-injected for testing.
	Store    credentials.Store
	Prompter auth.Prompter

	// Out is where progress is written.
	Out io.Writer

	// RepoRoot, if empty, is detected via git rev-parse from cwd.
	RepoRoot string
}

// Result is returned on success. The caller (cobra command) prints
// hook snippets and the bot username for the user.
type Result struct {
	Config  *config.Config
	Info    *client.Info
	BotUser *client.BotUser
	Token   *client.APIToken
}

// Init runs the full onboarding flow. Steps are deliberately sequential and
// each prints a one-line progress note to opts.Out; failures are wrapped
// with output.Error codes so cobra's error handler renders them cleanly.
func Init(ctx context.Context, opts *Options) (*Result, error) {
	if opts == nil {
		opts = &Options{}
	}
	if opts.Out == nil {
		opts = &Options{}
	}
	if opts.Out == nil {
		opts = &Options{Out: io.Discard}
	}
	if opts.ConfigPath == "" {
		return nil, output.New(output.CodeValidation, "ConfigPath is required")
	}

	prompter := opts.Prompter
	if prompter == nil {
		prompter = auth.NewStdPrompter()
	}
	store := opts.Store
	if store == nil {
		store = credentials.Default()
	}

	// 1. Repo root + suggested bot username.
	repoRoot := opts.RepoRoot
	if repoRoot == "" {
		var err error
		repoRoot, err = config.RepoRoot(ctx, "")
		if err != nil {
			return nil, output.Wrap(output.CodeUnknown, err, "detect repo root: %v", err)
		}
	}
	suggested := config.SuggestedBotUsername(repoRoot)
	botUsername := normalizeBotUsername(opts.BotUsername, suggested)
	progress(opts.Out, "Bot username will be %q", botUsername)

	// 2. Server URL.
	if opts.Server == "" {
		v, err := prompter.ReadLine("Vikunja server URL: ")
		if err != nil {
			return nil, err
		}
		opts.Server = strings.TrimSpace(v)
	}
	opts.Server = strings.TrimRight(opts.Server, "/")
	if opts.Server == "" {
		return nil, output.New(output.CodeValidation, "server URL is required")
	}

	// 3. Probe /info.
	human := client.New(opts.Server, "")
	info, err := human.Info(ctx)
	if err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "GET /info on %s: %v", opts.Server, err)
	}
	progress(opts.Out, "Connected to Vikunja %s", info.Version)

	// 4. Acquire human JWT (transient — used until step 11). Default is the
	// OAuth flow; --token / --use-password / --username+--password override.
	tok, err := auth.AcquireHumanToken(ctx, human, auth.LoginOptions{
		Token:       opts.HumanToken,
		UsePassword: opts.HumanUsePassword,
		Username:    opts.HumanUsername,
		Password:    opts.HumanPassword,
		TOTP:        opts.HumanTOTP,
		Out:         opts.Out,
	}, prompter)
	if err != nil {
		return nil, err
	}
	human.Token = tok

	// 5. Pick (or accept passed) project.
	project, err := pickProject(ctx, human, opts.ProjectID, prompter, opts.Out)
	if err != nil {
		return nil, err
	}
	progress(opts.Out, "Using project #%d %q (identifier=%q)", project.ID, project.Title, project.Identifier)

	// 6. Pick (or accept passed) Kanban view.
	view, err := pickKanbanView(ctx, human, project.ID, opts.ViewID, prompter, opts.Out)
	if err != nil {
		return nil, err
	}
	progress(opts.Out, "Using view #%d %q", view.ID, view.Title)

	// 7. Bucket bootstrap (with strict-with-override prompt).
	buckets, err := bootstrapBuckets(ctx, human, project.ID, view.ID, opts, prompter)
	if err != nil {
		return nil, err
	}

	// 8. Resolve the bot user: reuse one we already own if the name is
	// taken by us, prompt for a fresh name (with a petname suggestion)
	// if the name is taken by someone else, otherwise create new.
	bot, err := resolveBotUser(ctx, human, botUsername, project.Title, prompter, opts.Out)
	if err != nil {
		return nil, err
	}

	// 9. Share the project with the bot.
	if _, err := human.ShareProjectWithUser(ctx, project.ID, &client.ProjectUser{
		Username:   bot.Username,
		Permission: client.PermissionReadWrite,
	}); err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "share project with bot: %v", err)
	}
	progress(opts.Out, "Shared project with %q (read+write)", bot.Username)

	// 10. Discover available API permission scopes, mint the bot's token.
	routes, err := human.Routes(ctx)
	if err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "fetch /routes: %v", err)
	}
	perms := client.PermissionsForBot(routes)
	if len(perms) == 0 {
		return nil, output.New(output.CodeUnknown, "no API token permissions available — Vikunja /routes returned no matching groups")
	}
	mintedToken, err := human.CreateToken(ctx, &client.APIToken{
		Title:       "veans for " + project.Title,
		Permissions: perms,
		ExpiresAt:   client.FarFuture,
		OwnerID:     bot.ID,
	})
	if err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "mint bot token: %v", err)
	}
	if mintedToken.Token == "" {
		return nil, output.New(output.CodeUnknown, "PUT /tokens did not return a token plaintext — cannot continue")
	}

	// 11. Persist credentials. Discard human JWT immediately after.
	if err := store.Set(opts.Server, bot.Username, mintedToken.Token); err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "store bot token: %v", err)
	}
	human.Token = ""

	// 12. Write .veans.yml.
	cfg := &config.Config{
		Server:            opts.Server,
		ProjectID:         project.ID,
		ProjectIdentifier: project.Identifier,
		ViewID:            view.ID,
		Buckets:           buckets,
		Bot: config.Bot{
			Username: bot.Username,
			UserID:   bot.ID,
		},
	}
	if err := cfg.SaveAs(opts.ConfigPath); err != nil {
		return nil, output.Wrap(output.CodeUnknown, err, "write %s: %v", opts.ConfigPath, err)
	}
	progress(opts.Out, "Wrote %s", opts.ConfigPath)

	return &Result{Config: cfg, Info: info, BotUser: bot, Token: mintedToken}, nil
}

func normalizeBotUsername(override, suggested string) string {
	if override == "" {
		return suggested
	}
	if !strings.HasPrefix(override, "bot-") {
		return "bot-" + override
	}
	return override
}

func pickProject(ctx context.Context, c *client.Client, id int64, p auth.Prompter, out io.Writer) (*client.Project, error) {
	if id != 0 {
		return c.GetProject(ctx, id)
	}
	projects, err := c.ListProjects(ctx)
	if err != nil {
		return nil, err
	}
	// Filter out archived projects to keep the list short.
	var active []*client.Project
	for _, pr := range projects {
		if pr.IsArchived {
			continue
		}
		active = append(active, pr)
	}
	if len(active) == 0 {
		return nil, output.New(output.CodeNotFound, "no projects visible to this user — create one in the Vikunja UI first")
	}
	sort.Slice(active, func(i, j int) bool { return active[i].Title < active[j].Title })

	fmt.Fprintln(out, "Available projects:")
	for i, pr := range active {
		ident := pr.Identifier
		if ident == "" {
			ident = "(no identifier)"
		}
		fmt.Fprintf(out, "  [%d] #%d %s — %s\n", i+1, pr.ID, pr.Title, ident)
	}
	choice, err := p.ReadLine("Pick a project [1]: ")
	if err != nil {
		return nil, err
	}
	choice = strings.TrimSpace(choice)
	idx := 1
	if choice != "" {
		v, err := strconv.Atoi(choice)
		if err != nil || v < 1 || v > len(active) {
			return nil, output.New(output.CodeValidation, "invalid project choice %q", choice)
		}
		idx = v
	}
	return active[idx-1], nil
}

func pickKanbanView(ctx context.Context, c *client.Client, projectID int64, viewID int64, p auth.Prompter, out io.Writer) (*client.ProjectView, error) {
	views, err := c.ListProjectViews(ctx, projectID)
	if err != nil {
		return nil, err
	}
	var kanban []*client.ProjectView
	for _, v := range views {
		if v.ViewKind == client.ViewKindKanban {
			kanban = append(kanban, v)
		}
	}
	if len(kanban) == 0 {
		return nil, output.New(output.CodeNotFound, "no Kanban views on this project — create one in the Vikunja UI first")
	}
	if viewID != 0 {
		for _, v := range kanban {
			if v.ID == viewID {
				return v, nil
			}
		}
		return nil, output.New(output.CodeNotFound, "view %d is not a Kanban view on this project", viewID)
	}
	if len(kanban) == 1 {
		return kanban[0], nil
	}
	fmt.Fprintln(out, "Available Kanban views:")
	for i, v := range kanban {
		fmt.Fprintf(out, "  [%d] #%d %s\n", i+1, v.ID, v.Title)
	}
	choice, err := p.ReadLine("Pick a view [1]: ")
	if err != nil {
		return nil, err
	}
	choice = strings.TrimSpace(choice)
	idx := 1
	if choice != "" {
		v, err := strconv.Atoi(choice)
		if err != nil || v < 1 || v > len(kanban) {
			return nil, output.New(output.CodeValidation, "invalid view choice %q", choice)
		}
		idx = v
	}
	return kanban[idx-1], nil
}

func bootstrapBuckets(ctx context.Context, c *client.Client, projectID, viewID int64, opts *Options, p auth.Prompter) (config.Buckets, error) {
	existing, err := c.ListBuckets(ctx, projectID, viewID)
	if err != nil {
		return config.Buckets{}, err
	}

	// Resolve canonical statuses to existing buckets via the alias table.
	// Vikunja's default Kanban view ships with "To-Do" / "Doing" / "Done";
	// matching them as Todo / InProgress / Done avoids creating a parallel
	// set of buckets every time veans runs against a vanilla project.
	matched := map[status.Status]*client.Bucket{}
	for _, s := range status.All() {
		for _, b := range existing {
			if b == nil {
				continue
			}
			if status.MatchBucketTitle(s, b.Title) {
				matched[s] = b
				break
			}
		}
	}

	var missing []string
	for _, s := range status.All() {
		if _, ok := matched[s]; !ok {
			missing = append(missing, s.BucketTitle())
		}
	}

	if len(missing) > 0 && !opts.SkipBucketBootstrap {
		approve := opts.AutoApproveBuckets
		if !approve {
			fmt.Fprintf(opts.Out, "Missing canonical buckets: %s\n", strings.Join(missing, ", "))
			ans, err := p.ReadLine("Bootstrap missing buckets? [Y/n/abort]: ")
			if err != nil {
				return config.Buckets{}, err
			}
			ans = strings.ToLower(strings.TrimSpace(ans))
			switch ans {
			case "", "y", "yes":
				approve = true
			case "n", "no":
				approve = false
			case "a", "abort":
				return config.Buckets{}, output.New(output.CodeValidation, "user aborted bucket bootstrap")
			}
		}
		if approve {
			for _, s := range status.All() {
				if _, ok := matched[s]; ok {
					continue
				}
				title := s.BucketTitle()
				b, err := c.CreateBucket(ctx, projectID, viewID, &client.Bucket{Title: title})
				if err != nil {
					return config.Buckets{}, output.Wrap(output.CodeUnknown, err, "create bucket %q: %v", title, err)
				}
				matched[s] = b
				progress(opts.Out, "Created bucket %q (id=%d)", title, b.ID)
			}
		}
	}

	for _, s := range status.All() {
		if b, ok := matched[s]; ok && b != nil && b.Title != s.BucketTitle() {
			progress(opts.Out, "Reusing existing bucket %q as %s (id=%d)", b.Title, s.BucketTitle(), b.ID)
		}
	}

	out := config.Buckets{
		Todo:       bucketID(matched, status.Todo),
		InProgress: bucketID(matched, status.InProgress),
		InReview:   bucketID(matched, status.InReview),
		Done:       bucketID(matched, status.Completed),
		Scrapped:   bucketID(matched, status.Scrapped),
	}
	if out.Todo == 0 || out.InProgress == 0 || out.InReview == 0 || out.Done == 0 || out.Scrapped == 0 {
		return config.Buckets{}, output.New(output.CodeValidation,
			"canonical buckets missing — re-run with bucket bootstrap approved or create them manually")
	}
	return out, nil
}

func bucketID(m map[status.Status]*client.Bucket, s status.Status) int64 {
	if b, ok := m[s]; ok && b != nil {
		return b.ID
	}
	return 0
}

func progress(w io.Writer, format string, args ...any) {
	if w == nil {
		return
	}
	fmt.Fprintf(w, "  → "+format+"\n", args...)
}

// silence the unused-import linter when errors isn't used elsewhere.
var _ = errors.New
