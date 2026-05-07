// Package auth handles the human's transient authentication during init and
// login. v0 uses POST /login (username + password) to mint a JWT we hold in
// memory only — Vikunja's OAuth provider flow requires a registered client
// and an existing JWT to authorize, which adds friction we don't need yet.
//
// Pre-existing JWTs and personal API tokens may be passed via --token, which
// short-circuits the prompt entirely; this is the path SSO/OIDC users take
// since they cannot log in with a local password.
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
	// Token short-circuits the prompt. May be a JWT or a personal API token.
	Token string
	// Username is optional — if empty, the prompter asks. Required for
	// password-based login.
	Username string
	// Password is optional — if empty, the prompter asks (masked).
	Password string
	// TOTP, if set, is sent with the login request.
	TOTP string
}

// AcquireHumanToken returns a bearer token to act as the human during init.
// Order of resolution:
//  1. opts.Token (paste-in or --token flag)
//  2. POST /login with opts.Username/Password (prompts to fill missing parts)
func AcquireHumanToken(ctx context.Context, c *client.Client, opts LoginOptions, p Prompter) (string, error) {
	if opts.Token != "" {
		return opts.Token, nil
	}
	if p == nil {
		p = NewStdPrompter()
	}
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
		return "", output.New(output.CodeAuth, "username and password are required")
	}

	// Vikunja's local /login takes either a username or an email; we let the
	// server decide. LongToken=true requests a longer-lived JWT, useful since
	// init may take a few seconds.
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
