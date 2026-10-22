---
name: sentry-triage
description: Survey unresolved Sentry issues for Vikunja, classify them (already fixed, real bug, reporting noise, external), then investigate and fix each cluster in its own worktree via a subagent that opens a PR, and update Sentry statuses afterwards. Use when asked "what sentry issues do we have", "triage sentry", "fix the sentry errors", or "check sentry again".
user-invocable: true
---

# Sentry triage → fix → PR

Survey Sentry, classify, dispatch one subagent per issue cluster (investigate, fix, test, `/open-pr`), then mark issues in Sentry. Derived from a full pass in Sept 2026 (~35 subagents, 33 PRs); the gotchas below are all things that actually bit.

## 0. Preflight (every run)

```bash
git pull --ff-only origin main     # in the main checkout. Worktrees branch from LOCAL main; stale main = conflicts + duplicate fixes
gh auth status
```

Load Sentry MCP tools via `ToolSearch` (`select:mcp__sentry__search_issues,mcp__sentry__get_sentry_resource,mcp__sentry__update_issue,mcp__sentry__execute_sentry_tool,mcp__sentry__search_sentry_tools`).

Sentry: org `vikunja`, `regionUrl: https://us.sentry.io`. Projects: `frontend-oss`, `api-oss`, `api-cloud`, `frontend-cloud`, `cloud-portal`, `app`. The `vikunja-eu` org is empty. `app` is the Flutter client in another repo: list its top issues in the survey, never dispatch fixes for it.

## 1. Sweep snoozed issues (every run)

Snoozed issues are invisible to `is:unresolved`. Sweep them before the survey, otherwise a fix that did not work hides for 14 days and then reappears with no context.

- `search_issues` with `query: "is:ignored !project:app"`, `period: 90d`, `limit: 100`; also `query: "is:regressed"` and `query: "is:escalating"` for ones Sentry already woke up.
- For each: `execute_sentry_tool get_issue_activity` to read our own comment (it names the fix PR or commit), then `get_issue_tag_values(tagKey: release)` to see which releases produced events after the snooze started.
- Decide per issue by comparing the fix commit against the releases still producing events (`git merge-base --is-ancestor <fix-sha> <release-sha>`; release names are `v2.6.0-51-c9ab8c33`, the trailing hash is the commit):
  - No events since the snooze → `resolved`, reason "no events since <PR> deployed".
  - Events only from releases that predate the fix → deployment lag; extend the snooze 14 days once, mention it in the report.
  - Events from a release that contains the fix → the fix did not work or covered only part. Set `unresolved`, treat as a fresh real bug in the survey, and tell the user which PR missed.
- Regressed/escalating issues get the same check; Sentry already flipped their status, so only the classification step applies.
- Report the outcome as a table: issue, PR, verdict (resolved / extended / reopened).

## 2. Survey

- `search_issues` with `query: "is:unresolved !project:app"`, `sort: freq`, `limit: 100`, `period: 90d`. The org-wide list caps at 100; when it does, query per project.
- Group by message shape + culprit route, not by issue id. Different browsers word the same bug differently (`Cannot read properties of null (reading 'x')` / `null is not an object (evaluating 'a.x')` / `can't access property "x", a is null`).
- API issues all share the `reportToSentry` frame; before `pkg/errorreport` fingerprinting deployed, one issue could hold several unrelated errors. Check `search_issue_events` on the issue before trusting its title.
- Classify each cluster:
  - **Already fixed**: `git log --since=<first seen> -- <suspect files>` finds the fix and last-seen predates it. Confirm with `get_issue_tag_values(tagKey: release)`.
  - **Covered by an open PR**: note the PR.
  - **Real bug**: dispatch.
  - **Reporting noise**: expected failures (401 on token refresh, stale chunks after deploy, webhook target 4xx, upstream provider 4xx), extension/webview injections, Sentry's own detections (rage click, http_client, ai_detected), empty events, scanner traffic. Fix the *filter* (`frontend/src/helpers/sentryFilters.ts`, `pkg/events` poison logger, migration listener) rather than the symptom.
  - **External / infra**: demo-instance nightly reset (errors at exactly 04:00:01 UTC on try.vikunja.io), disk full, plugin repos (cloud-sync), user content.
- Present the survey to the user grouped this way with event counts and a recommended batch. Wait for go-ahead unless told to run autonomously.

## 3. Dispatch

One worktree + one subagent per cluster. From the main checkout:

```bash
mage dev:prepare-worktree fix-<slug> ""
```

Then `Agent` (model `opus`, `general-purpose`, background) with the brief from [brief-template.md](brief-template.md). Fill in: worktree path, Sentry ids, exact messages, event counts, routes, suspected files, and which other agents own neighbouring files.

Rules that must be in every brief (they are in the template, keep them):
- Work only in the named worktree. Never `git stash` (shared across all worktrees of the repo; agents have popped each other's).
- No amending. Atomic conventional commits.
- Verify the bug is not already fixed on `origin/main` before changing code; stop and report instead of forcing a PR.
- Reproduce with a failing test first where feasible; state which tests fail without the fix.
- Fix at the boundary (make the bad state unrepresentable), not by sprinkling `?.` in consumers.
- Backend: `mage test:filter` only, narrowest filter, `mage test:web` at most once. Frontend: touched test files only, no full suite, no `pnpm typecheck` (broken on main, ~1600 errors, 40+ min under load).
- New world-error codes: `git fetch origin` and grep `origin/main` for the highest code first; three agents picked the same "next free" code in one day. `pkg/web/error_codes_test.go` now fails on duplicates, but only after rebase.
- End with the `open-pr` skill and a report: PR URL, root cause, files, tests, what was skipped and why.

Concurrency: **max 3 agents that run tests at once** on this box. Read-only investigation agents don't count. Queue the rest.

Investigation-only agents (unclear root cause, e.g. deadlocks, perf issues): same brief minus the fix section, deliverable is a report with options ranked. Decide with the user before dispatching the fix.

## 4. While agents run

- Relay each completion to the user in a few bullets: root cause, fix, tests, deviations. Do not paraphrase away caveats the agent raised.
- When two PRs touch the same file (check the reports), decide merge order and tell the second agent to rebase after the first merges (`SendMessage` to the agent by name). Duplicate fixes for one root cause: keep the merged one, have the other agent drop its commit.
- CI failure on a PR: read the job log with `gh api repos/go-vikunja/vikunja/actions/jobs/<id>/logs`, fix small things yourself in the worktree, else send the agent back.
- Before declaring a PR set done: `git merge-tree --write-tree origin/main origin/<branch>` per branch to catch conflicts GitHub has not computed yet.

## 5. Sentry status updates

"Resolve in next release" is **broken** for this org: release names come from `git describe` (`v2.6.0-51-c9ab8c33`), which Sentry parses as a prerelease of 2.6.0 and sorts lexically, so it picks a stale release and everything regresses on the next event. Confirmed identical from the UI. Until release naming is fixed, use:

| Situation | Status | Reason text |
|---|---|---|
| Fixed by an old commit, no events since | `resolved` | commit hash + one line |
| Fixed by a PR merged but not yet deployed | `ignored`, `forDuration`, `20160` min (14 d) | `Fixed in go-vikunja/vikunja#N (...). Snoozed 14 days until deployed.` |
| Noise that a filter PR will drop | same 14 d snooze | name the filter PR |
| Noise with no filter, external, one-offs | `ignored`, `untilEscalating` | why it is not ours |
| Scanner / third-party test events | `ignored`, `forever` | note the source host |
| Blocked on another PR or repo | leave open | say so in the summary |

Always pass `reason`; it becomes a comment. Batch ~30 `update_issue` calls per message. When snoozes wake, whatever still fires is either undeployed or not actually fixed.

## 6. Cleanup

After PRs merge: `bash ~/.claude/skills/cleanup-worktrees/cleanup-worktrees.sh`. It only removes worktrees whose PR is merged and whose tree is clean. If it exits 141 with no output (broken pipe, seen once), remove by hand:

```bash
git -C ../<wt> status --porcelain   # must be empty
git worktree remove ../<wt> && git branch -D <wt>
```

Investigation-only worktrees with no PR: remove the same way once the report is in.

## Known standing items (update when they change)

- Release naming fix for Sentry semver ordering: not done.
- `pnpm typecheck` red on main: not done.
- Sentry inbound filter for allowed domains (scanner events from `panel.secureky.eu`): settings change, not done.
- `en.json` maps 1025 to a timezone string, Go has 1025 as TOTP passcode used.
- API-CLOUD-1G lives in the cloud-sync plugin repo (not checked out); needs `%w` wrapping there.
- API-CLOUD-1X waits on #3697.
