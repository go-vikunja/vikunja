// Package auth handles the human's transient authentication during init and
// login. The default interactive flow is OAuth 2.0 Authorization Code + PKCE
// against Vikunja's built-in authorization server (no client registration
// needed; PKCE/S256 mandatory). The user opens the authorize URL in their
// browser, signs in, and pastes the resulting `vikunja-veans-cli://callback`
// URL back into the CLI — that side-steps custom-scheme handler registration
// entirely.
//
// For non-interactive contexts (CI scripts, paste-in tokens, accounts on
// instances without OAuth), pass --token, --username + --password, or
// --use-password. Personal API tokens via --token also let SSO/OIDC users
// onboard without exercising local password login.
package auth

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"

	"code.vikunja.io/veans/internal/client"
	"code.vikunja.io/veans/internal/output"
)

// Prompter abstracts stdin / TTY reads so tests can inject scripted answers.
type Prompter interface {
	ReadLine(prompt string) (string, error)
	ReadPassword(prompt string) (string, error)
}

// StdPrompter reads from os.Stdin and uses term.ReadPassword for masked
// input. It's the production default.
type StdPrompter struct {
	In  io.Reader
	Out io.Writer
}

func NewStdPrompter() *StdPrompter {
	return &StdPrompter{In: os.Stdin, Out: os.Stderr}
}

func (p *StdPrompter) ReadLine(prompt string) (string, error) {
	if _, err := fmt.Fprint(p.Out, prompt); err != nil {
		return "", err
	}
	r := bufio.NewReader(p.In)
	line, err := r.ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (p *StdPrompter) ReadPassword(prompt string) (string, error) {
	if _, err := fmt.Fprint(p.Out, prompt); err != nil {
		return "", err
	}
	if f, ok := p.In.(*os.File); ok && term.IsTerminal(int(f.Fd())) {
		buf, err := term.ReadPassword(int(f.Fd()))
		fmt.Fprintln(p.Out)
		if err != nil {
			return "", err
		}
		return string(buf), nil
	}
	// Non-TTY (CI, scripted test) — read a plain line.
	line, err := p.ReadLine("")
	return line, err
}

// LoginOptions controls how AcquireHumanToken obtains a JWT.
type LoginOptions struct {
	// Token short-circuits all flows. May be a JWT or a personal API token.
	Token string
	// UsePassword forces the legacy POST /login flow even when no password
	// is set yet (the prompter will ask for it). Useful on instances where
	// OAuth is disabled or the user prefers entering a password.
	UsePassword bool
	// Username / Password / TOTP feed POST /login. If both Username and
	// Password are non-empty, AcquireHumanToken uses /login non-interactively
	// regardless of UsePassword.
	Username string
	Password string
	TOTP     string
	// Out is where progress / OAuth instructions are written. Defaults to
	// os.Stderr in production via NewStdPrompter; tests can pass any writer.
	Out io.Writer
}

// AcquireHumanToken returns a bearer token to act as the human during init.
// Resolution order:
//  1. opts.Token (paste-in or --token flag)
//  2. POST /login with Username + Password (used non-interactively when both
//     are set, or when --use-password is passed)
//  3. OAuth Authorization Code + PKCE flow with manual callback paste-back
//     (the default for interactive use)
func AcquireHumanToken(ctx context.Context, c *client.Client, opts LoginOptions, p Prompter) (string, error) {
	if opts.Token != "" {
		return opts.Token, nil
	}
	if p == nil {
		p = NewStdPrompter()
	}
	w := opts.Out
	if w == nil {
		w = os.Stderr
	}

	usePassword := opts.UsePassword || (opts.Username != "" && opts.Password != "")
	if usePassword {
		return loginWithPassword(ctx, c, opts, p)
	}

	return runOAuthFlow(ctx, c, p, w)
}

// loginWithPassword runs the legacy POST /login path. Kept for instances
// that have OAuth disabled or for non-interactive `--username` + `--password`
// invocations in CI.
func loginWithPassword(ctx context.Context, c *client.Client, opts LoginOptions, p Prompter) (string, error) {
	if opts.Username == "" {
		u, err := p.ReadLine("Vikunja username: ")
		if err != nil {
			return "", output.Wrap(output.CodeAuth, err, "read username: %v", err)
		}
		opts.Username = strings.TrimSpace(u)
	}
	if opts.Password == "" {
		pw, err := p.ReadPassword("Vikunja password: ")
		if err != nil {
			return "", output.Wrap(output.CodeAuth, err, "read password: %v", err)
		}
		opts.Password = pw
	}
	if opts.Username == "" || opts.Password == "" {
		return "", output.New(output.CodeAuth, "username and password are required for password login")
	}
	resp, err := c.Login(ctx, &client.LoginRequest{
		Username:     opts.Username,
		Password:     opts.Password,
		TOTPPasscode: opts.TOTP,
		LongToken:    true,
	})
	if err != nil {
		return "", err
	}
	if resp.Token == "" {
		return "", output.New(output.CodeAuth, "login returned empty token")
	}
	return resp.Token, nil
}

// silenceLinter suppresses the unused syscall import on platforms where
// term.ReadPassword inlines its own platform call. We keep the import to
// document that masked input is expected to use POSIX-level terminal modes.
var _ = syscall.Stdin
