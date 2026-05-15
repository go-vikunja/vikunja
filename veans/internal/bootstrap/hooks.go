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

package bootstrap

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"code.vikunja.io/veans/internal/auth"
	"code.vikunja.io/veans/internal/output"
)

// veansPrimeCommand is the literal command line every hook ends up invoking.
// Centralising it here keeps the install logic and the duplicate-detection
// reading the same string.
const veansPrimeCommand = "veans prime"

// AgentHookChoice captures the user's per-agent install decision so the
// orchestration in bootstrap.Init can hand the per-repo set of choices
// off to the install routines below.
type AgentHookChoice struct {
	ClaudeCode bool
	OpenCode   bool
}

// offerAgentHooks asks the user — one yes/no per agent — which integrations
// they want veans to wire up. Callers pre-populate `choices` from CLI flags
// (--install-claude / --install-opencode); only the unset slots get
// prompted. When `noHooks` is true we skip everything and return the empty
// choice, mirroring the old "just print the snippets" behaviour.
func offerAgentHooks(p auth.Prompter, w io.Writer, choices AgentHookChoice, claudeFlagSet, opencodeFlagSet, noHooks bool) (AgentHookChoice, error) {
	if noHooks {
		return AgentHookChoice{}, nil
	}
	if !claudeFlagSet {
		yes, err := promptYesNo(p, w,
			"Wire `veans prime` into Claude Code (.claude/settings.json)?", true)
		if err != nil {
			return choices, err
		}
		choices.ClaudeCode = yes
	}
	if !opencodeFlagSet {
		yes, err := promptYesNo(p, w,
			"Wire `veans prime` into OpenCode (.opencode/plugin/veans-prime.ts)?", false)
		if err != nil {
			return choices, err
		}
		choices.OpenCode = yes
	}
	return choices, nil
}

// installAgentHooks writes the requested integrations to disk relative to
// repoRoot. Each install is idempotent: if the hook entry is already there,
// it's left alone; if the settings file is missing, it's created with a
// fresh skeleton.
func installAgentHooks(repoRoot string, choices AgentHookChoice, w io.Writer) error {
	if choices.ClaudeCode {
		path, action, err := installClaudeCodeHook(repoRoot)
		if err != nil {
			return output.Wrap(output.CodeUnknown, err, "install Claude Code hook: %v", err)
		}
		progress(w, "%s Claude Code hook in %s", action, path)
	}
	if choices.OpenCode {
		path, action, err := installOpenCodeHook(repoRoot)
		if err != nil {
			return output.Wrap(output.CodeUnknown, err, "install OpenCode hook: %v", err)
		}
		progress(w, "%s OpenCode hook in %s", action, path)
	}
	return nil
}

// installClaudeCodeHook merges (or creates) `<repoRoot>/.claude/settings.json`
// so SessionStart and PreCompact invoke `veans prime`. Returns the path,
// a human verb describing what happened ("Wrote", "Updated", "Already
// configured"), and any error.
func installClaudeCodeHook(repoRoot string) (string, string, error) {
	path := filepath.Join(repoRoot, ".claude", "settings.json")
	settings, existed, err := readJSONOrEmpty(path)
	if err != nil {
		return path, "", err
	}
	changed := false
	for _, event := range []string{"SessionStart", "PreCompact"} {
		if ensureClaudeHook(settings, event) {
			changed = true
		}
	}
	if !changed {
		return path, "Already configured", nil
	}
	if err := writeJSON(path, settings); err != nil {
		return path, "", err
	}
	if existed {
		return path, "Updated", nil
	}
	return path, "Wrote", nil
}

// ensureClaudeHook walks the settings object and appends a `veans prime`
// command entry under hooks.<event> if one isn't already present. Returns
// true iff the structure was modified.
//
// Claude Code's settings shape:
//
//	{
//	  "hooks": {
//	    "SessionStart": [
//	      { "hooks": [ { "type": "command", "command": "veans prime" } ] }
//	    ]
//	  }
//	}
func ensureClaudeHook(settings map[string]any, event string) bool {
	hooks := mapAt(settings, "hooks")
	entries, _ := hooks[event].([]any)

	for _, entry := range entries {
		entryMap, ok := entry.(map[string]any)
		if !ok {
			continue
		}
		inner, _ := entryMap["hooks"].([]any)
		for _, h := range inner {
			hmap, ok := h.(map[string]any)
			if !ok {
				continue
			}
			if str(hmap, "type") == "command" && str(hmap, "command") == veansPrimeCommand {
				return false
			}
		}
	}

	entries = append(entries, map[string]any{
		"hooks": []any{
			map[string]any{"type": "command", "command": veansPrimeCommand},
		},
	})
	hooks[event] = entries
	settings["hooks"] = hooks
	return true
}

// installOpenCodeHook writes `<repoRoot>/.opencode/plugin/veans-prime.ts`
// if missing. Existing files are left alone (TypeScript merging is out of
// scope; the user can edit by hand).
func installOpenCodeHook(repoRoot string) (string, string, error) {
	path := filepath.Join(repoRoot, ".opencode", "plugin", "veans-prime.ts")
	if _, err := os.Stat(path); err == nil {
		return path, "Already configured", nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return path, "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return path, "", err
	}
	if err := os.WriteFile(path, []byte(openCodeHookSource), 0o644); err != nil {
		return path, "", err
	}
	return path, "Wrote", nil
}

const openCodeHookSource = `// Auto-generated by 'veans init'. Re-emits the veans agent prompt at the
// start of every OpenCode session and before every compaction. See
// https://github.com/go-vikunja/vikunja/tree/main/veans for context.
export const VeansPrime = {
	event: ["session.start", "compact.before"],
	handler: async ({ exec }: { exec: (cmd: string) => Promise<unknown> }) =>
		exec("veans prime"),
}
`

// readJSONOrEmpty reads `path` as JSON or returns an empty object if the
// file doesn't exist. The `existed` flag tells the caller whether the
// resulting object was loaded from disk (so it can decide between
// "Wrote" and "Updated").
func readJSONOrEmpty(path string) (out map[string]any, existed bool, err error) {
	buf, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return map[string]any{}, false, nil
		}
		return nil, false, err
	}
	out = map[string]any{}
	if len(buf) == 0 {
		return out, true, nil
	}
	if err := json.Unmarshal(buf, &out); err != nil {
		return nil, true, fmt.Errorf("parse %s: %w", path, err)
	}
	return out, true, nil
}

// writeJSON encodes `data` with two-space indent (Claude Code's house
// style) and a trailing newline, creating parent directories as needed.
func writeJSON(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	buf, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	buf = append(buf, '\n')
	return os.WriteFile(path, buf, 0o644)
}

// mapAt returns the map at key `k` on `m`, creating it if missing or if
// the existing value is the wrong type. Lets ensureClaudeHook treat the
// JSON object tree as if it were always well-shaped.
func mapAt(m map[string]any, k string) map[string]any {
	if v, ok := m[k].(map[string]any); ok {
		return v
	}
	v := map[string]any{}
	m[k] = v
	return v
}

func str(m map[string]any, k string) string {
	s, _ := m[k].(string)
	return s
}

// promptYesNo reads a Y/n (or y/N) answer with the given default.
func promptYesNo(p auth.Prompter, w io.Writer, question string, defaultYes bool) (bool, error) {
	tag := "[Y/n]"
	if !defaultYes {
		tag = "[y/N]"
	}
	fmt.Fprintln(w, question)
	ans, err := p.ReadLine(tag + " ")
	if err != nil {
		return defaultYes, err
	}
	switch strings.ToLower(strings.TrimSpace(ans)) {
	case "":
		return defaultYes, nil
	case "y", "yes":
		return true, nil
	case "n", "no":
		return false, nil
	}
	return defaultYes, nil
}
