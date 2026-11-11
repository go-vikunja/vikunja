# Subagent brief template

Replace every `<...>`. Keep all rules; they exist because each one failed once.

---

Investigate, fix, open a PR. Work ONLY in the worktree `/home/clawd/projects/vikunja/vikunja/<worktree>` (branch `<worktree>`, created from current origin/main, <frontend initialized | config.yml adjusted>). Never touch `/home/clawd/projects/vikunja/vikunja/main` or sibling worktrees. Read the worktree's CLAUDE.md and follow it: conventional commits, small atomic commits, no amending, NEVER use `git stash` (shared across worktrees), lint before committing (`mage lint:fix` for Go, retry if another agent holds the golangci-lint lock; `pnpm lint:fix` in frontend/ for TS/Vue), comments only for a non-obvious why.

Resource limits, other agents share this machine: backend tests only via `mage test:filter <Filter>` with the narrowest filter, `mage test:web` at most once at the end and only if a route is affected; frontend `pnpm test:unit --run <touched files>` only, never the full suite, never `pnpm typecheck`. Save every test run to a file (`2>&1 | tee /tmp/<name>.log`) and read the file instead of re-running.

## Sentry: <cluster title>

Org `vikunja`, region `https://us.sentry.io`, project `<project>`. Use `mcp__sentry__get_sentry_resource` on each issue, then via `mcp__sentry__search_sentry_tools` + `mcp__sentry__execute_sentry_tool` use `search_issue_events`, `get_issue_tag_values` (release, browser, url) and replays. Minified frame context is in the events; map it to source by the surrounding identifiers.

Issues:
- <ID> (<n> events, `<route>`, last seen <date>): `<exact message>`
- ...

Suspects: <files, functions, hypotheses>. <What other agents are working on nearby, so this agent keeps its diff confined.>

## Deliverable

1. Root cause with file:line. If the same minified frame appears across several issues, say so; if they differ, treat as separate bugs.
2. Check first whether it is already fixed on `origin/main` (`git log --since=<first seen> -- <files>`, compare the event release against the fix commit with `git merge-base --is-ancestor`). If yes, stop and report; no PR.
3. Fix at the source: make the bad state unrepresentable (types, parse at the boundary, correct lifecycle) rather than guarding every consumer. Reuse existing helpers; grep before adding one.
4. Tests: a failing-first reproduction where feasible; state in the report which tests fail without the fix. Backend `mage test:filter <names>`; frontend touched test files.
5. If you add a world-error code: `git fetch origin && git grep -n "= 1[0-9]\{4\}" origin/main -- pkg/` and pick a code not present anywhere; add the English string to `frontend/src/i18n/lang/en.json` and validate with `node -e "JSON.parse(require('fs').readFileSync('frontend/src/i18n/lang/en.json','utf8'))"`.
6. Lint, then invoke the `open-pr` skill (Skill tool, name `open-pr`) from the worktree. PR body: Sentry ids, mechanism in two or three sentences, how to verify.

If you cannot pin the cause with reasonable confidence, do not ship speculative guards; report what you found, what you ruled out, and what would settle it (replay link, extra instrumentation). Report back: PR URL or reason for none, root cause, files changed, tests run with results, anything pending or deviating from this brief.
