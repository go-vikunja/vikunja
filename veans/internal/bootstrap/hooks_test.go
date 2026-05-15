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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureClaudeHook_FreshFile(t *testing.T) {
	s := map[string]any{}
	if !ensureClaudeHook(s, "SessionStart") {
		t.Fatal("expected change on empty settings")
	}
	hooks, ok := s["hooks"].(map[string]any)
	if !ok {
		t.Fatalf("hooks key missing or wrong type: %v", s)
	}
	ss, ok := hooks["SessionStart"].([]any)
	if !ok || len(ss) != 1 {
		t.Fatalf("SessionStart shape: %v", hooks["SessionStart"])
	}
	entry := ss[0].(map[string]any)
	inner := entry["hooks"].([]any)
	if len(inner) != 1 {
		t.Fatalf("inner hooks: %v", inner)
	}
	h := inner[0].(map[string]any)
	if h["command"] != "veans prime" || h["type"] != "command" {
		t.Fatalf("hook shape: %v", h)
	}
}

func TestEnsureClaudeHook_Idempotent(t *testing.T) {
	s := map[string]any{}
	if !ensureClaudeHook(s, "SessionStart") {
		t.Fatal("first call should change")
	}
	if ensureClaudeHook(s, "SessionStart") {
		t.Fatal("second call should NOT change")
	}
	ss := s["hooks"].(map[string]any)["SessionStart"].([]any)
	if len(ss) != 1 {
		t.Fatalf("expected exactly one entry, got %d: %v", len(ss), ss)
	}
}

func TestEnsureClaudeHook_PreservesOtherHooks(t *testing.T) {
	// Existing settings have an unrelated PreToolUse hook and a SessionStart
	// entry running a different command. The veans entry should be appended,
	// not replace the existing structure.
	raw := []byte(`{
	  "hooks": {
	    "PreToolUse": [
	      { "matcher": "Bash", "hooks": [ { "type": "command", "command": "echo hi" } ] }
	    ],
	    "SessionStart": [
	      { "hooks": [ { "type": "command", "command": "other-tool init" } ] }
	    ]
	  },
	  "permissions": { "allow": ["Bash"] }
	}`)
	var s map[string]any
	if err := json.Unmarshal(raw, &s); err != nil {
		t.Fatal(err)
	}
	if !ensureClaudeHook(s, "SessionStart") {
		t.Fatal("expected change")
	}
	// PreToolUse + permissions untouched.
	if _, ok := s["permissions"]; !ok {
		t.Error("permissions key dropped")
	}
	if pt := s["hooks"].(map[string]any)["PreToolUse"].([]any); len(pt) != 1 {
		t.Errorf("PreToolUse perturbed: %v", pt)
	}
	// SessionStart now has BOTH the original and the veans entry.
	ss := s["hooks"].(map[string]any)["SessionStart"].([]any)
	if len(ss) != 2 {
		t.Fatalf("SessionStart should have 2 entries, got %d", len(ss))
	}
	gotVeans := false
	for _, e := range ss {
		inner := e.(map[string]any)["hooks"].([]any)
		for _, h := range inner {
			if h.(map[string]any)["command"] == "veans prime" {
				gotVeans = true
			}
		}
	}
	if !gotVeans {
		t.Errorf("veans prime not found in merged SessionStart: %v", ss)
	}
}

func TestInstallClaudeCodeHook_CreatesFile(t *testing.T) {
	dir := t.TempDir()
	path, action, err := installClaudeCodeHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	if action != "Wrote" {
		t.Errorf("first install should say Wrote, got %q", action)
	}
	if !strings.HasSuffix(path, ".claude/settings.json") {
		t.Errorf("unexpected path: %s", path)
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(buf), `"veans prime"`) {
		t.Errorf("written file missing veans prime command:\n%s", buf)
	}
	// Two-space indent + trailing newline.
	if !strings.HasSuffix(string(buf), "\n") {
		t.Error("written file missing trailing newline")
	}
	if !strings.Contains(string(buf), "  \"hooks\"") {
		t.Error("expected 2-space indent")
	}
}

func TestInstallClaudeCodeHook_IdempotentRerun(t *testing.T) {
	dir := t.TempDir()
	if _, _, err := installClaudeCodeHook(dir); err != nil {
		t.Fatal(err)
	}
	path, action, err := installClaudeCodeHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	if action != "Already configured" {
		t.Errorf("second install should report Already configured, got %q", action)
	}
	// File hasn't grown duplicate entries.
	buf, _ := os.ReadFile(path)
	if c := strings.Count(string(buf), `"veans prime"`); c != 2 {
		// 2 because both SessionStart and PreCompact reference it once.
		t.Errorf("expected exactly 2 references to veans prime, got %d:\n%s", c, buf)
	}
}

func TestInstallClaudeCodeHook_MergesWithUserSettings(t *testing.T) {
	dir := t.TempDir()
	settingsPath := filepath.Join(dir, ".claude", "settings.json")
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		t.Fatal(err)
	}
	existing := `{
	  "model": "claude-opus-4-7",
	  "hooks": {
	    "SessionStart": [
	      { "hooks": [ { "type": "command", "command": "other-tool" } ] }
	    ]
	  }
	}`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatal(err)
	}
	_, action, err := installClaudeCodeHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	if action != "Updated" {
		t.Errorf("merging into existing file should say Updated, got %q", action)
	}
	buf, _ := os.ReadFile(settingsPath)
	out := string(buf)
	for _, want := range []string{`"model": "claude-opus-4-7"`, `"other-tool"`, `"veans prime"`} {
		if !strings.Contains(out, want) {
			t.Errorf("merged file missing %q:\n%s", want, out)
		}
	}
}

func TestInstallOpenCodeHook_CreatesAndIdempotent(t *testing.T) {
	dir := t.TempDir()
	path, action, err := installOpenCodeHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	if action != "Wrote" {
		t.Errorf("first install should say Wrote, got %q", action)
	}
	buf, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"VeansPrime", "veans prime", "session.start", "compact.before"} {
		if !strings.Contains(string(buf), want) {
			t.Errorf("opencode file missing %q:\n%s", want, buf)
		}
	}

	// Re-run leaves the file alone — we don't merge TS by hand.
	_, action2, err := installOpenCodeHook(dir)
	if err != nil {
		t.Fatal(err)
	}
	if action2 != "Already configured" {
		t.Errorf("rerun should say Already configured, got %q", action2)
	}
}

func TestOfferAgentHooks_NoHooks(t *testing.T) {
	choices, err := offerAgentHooks(nil, nil, AgentHookChoice{}, false, false, true)
	if err != nil {
		t.Fatal(err)
	}
	if choices.ClaudeCode || choices.OpenCode {
		t.Errorf("NoHooks should return empty: %+v", choices)
	}
}

func TestOfferAgentHooks_FlagsBypassPrompt(t *testing.T) {
	// Both flags set explicitly — no prompts.
	p := &scriptedPrompter{} // would panic with out-of-range on any ReadLine
	choices, err := offerAgentHooks(p, nopWriter{}, AgentHookChoice{ClaudeCode: true, OpenCode: false}, true, true, false)
	if err != nil {
		t.Fatal(err)
	}
	if !choices.ClaudeCode || choices.OpenCode {
		t.Errorf("expected ClaudeCode=true, OpenCode=false; got %+v", choices)
	}
}

func TestOfferAgentHooks_PromptsWhenFlagsUnset(t *testing.T) {
	// User accepts Claude default (Y), declines OpenCode.
	p := &scriptedPrompter{answers: []string{"", "n"}}
	choices, err := offerAgentHooks(p, nopWriter{}, AgentHookChoice{}, false, false, false)
	if err != nil {
		t.Fatal(err)
	}
	if !choices.ClaudeCode || choices.OpenCode {
		t.Errorf("expected ClaudeCode=true OpenCode=false, got %+v", choices)
	}
}

// nopWriter discards everything; lets tests run prompts without console noise.
type nopWriter struct{}

func (nopWriter) Write(p []byte) (int, error) { return len(p), nil }
