# `/16-eyes fix` — apply the audit's findings

`fix` is linear and high-stakes (it edits code) — unlike `audit` (parallel, read-only,
low-stakes). It does **not** use the `Workflow` tool. Work directly in this session with
`Read`/`Edit`, one locus of accountability, no subagent fan-out.

## Phase 1 — Locate the findings

Try, in order. Both a full-repo `audit` report and a diff-scoped `audit-diff` report are
valid sources — they carry the identical `findings.safe[]`/`findings.risky[]` shape —
so consider both, and use whichever is actually the freshest/relevant one:

1. **This conversation** — if `/16-eyes audit` or `/16-eyes audit-diff` already ran
   earlier in this same session, use its `findings.safe[]`/`findings.risky[]` directly.
   Cheapest, guaranteed fresh. If the user's `/16-eyes fix` invocation doesn't say which
   one they mean and both ran this session, ask.
2. **`.16-eyes/last-run.json`** (full audit) and/or **`.16-eyes/last-diff-run.json`**
   (diff audit) — `Read` whichever exist, then `Read` the `.json` path each points to.
   If both exist and the user didn't specify, prefer the more recent `generatedAt`, but
   say so explicitly rather than silently picking one.
3. **No pointer, or it's stale/missing** — glob `SECURITY_AUDIT*.json` (matches both
   `SECURITY_AUDIT_<date>.json` and `SECURITY_AUDIT_DIFF_<date>.json`) in the configured
   (or default) output directory, take the newest by mtime. Tell the user this was a
   best-effort recovery, not a guaranteed-fresh source.
4. **Nothing found anywhere** — say so plainly and run `/16-eyes audit` (or `audit-diff`,
   whichever fits what the user actually wants fixed) first, then continue here with its
   output. Never invent findings.

Whichever source you used, **state it and its age** to the user before doing anything
else (e.g. "using the audit from this session, run 2 minutes ago" vs. "using
`docs/SECURITY_AUDIT_DIFF_2026-07-10.json` (PR #42), generated 6 days ago — code may
have changed since").

## Phase 2 — Apply `safe` findings

Group `findings.safe[]` by `file`. For each file, one coherent edit pass covering every
finding in that file — not one subagent per finding (two editors racing on the same file
with a stale line number is a real correctness risk). Match the surrounding code's own
conventions; reuse existing helpers/constants the finding's `suggested_fix` names rather
than inventing parallel ones. If a file has multiple findings, `Read` it again between
edits within that same file rather than trusting line numbers that may have shifted.

After all safe findings are applied, run gates **only for the workspace(s) whose files
were actually touched** — map each touched path to `config.gates.workspaces[]` (from
`.16-eyes/config.json`; if no config, skip gates and say so). If a gate fails, try to
localize which specific finding's fix caused it and repair or revert just that one —
never leave the working tree in a broken state.

## Phase 3 — Apply `risky` findings

Present each `risky` finding **one at a time**: title, `file:line`, the concrete diff
you're proposing, `why`, and `exploit_scenario`. Require an explicit yes before writing
anything. This is the most portable confirmation shape across arbitrary repos/sessions —
it doesn't assume any particular batching UI exists.

You may *propose* grouping multiple risky findings into one confirmation only when
they're genuinely the same fix repeated at near-identical call sites ("these 3 are the
same missing check at different call sites — apply all 3 with one confirmation?") — show
the proposal, then wait for the yes. Never group silently.

A declined finding goes in the ledger as **declined by user** — distinct from **skipped by
agent** (you looked closer and decided the finding was wrong/inapplicable on inspection;
say why).

## Hard invariant: never commit, never push

`fix` never runs any git write command — no `git commit`, no `git push`, no `git add`
even. All changes are left in the working tree for the user to review and commit
themselves. This is not configurable, by design:

- The tool runs in arbitrary repositories whose commit conventions (message format,
  signing, hooks, ticket references) it cannot know.
- It removes the one human checkpoint before a security-relevant change enters history.
- The verify+adversarial-review pipeline in `audit` is good but not infallible — an
  auto-commit would make a subtly-wrong "safe" classification permanent before anyone
  looks at it.
- This is a public tool installed by strangers — even offering auto-commit as an opt-in
  config risks someone flipping it on without weighing the risk.

## Phase 4 — Verify and summarize

Run the gates for every workspace touched by either safe or (confirmed) risky changes.
Close with a ledger, same shape regardless of repo:

- **Applied** — by finding id, by file.
- **Gated** — risky findings, confirmed vs. declined.
- **Skipped** — findings the agent chose not to apply on closer inspection, with why.
- **Verification** — gate results (pass/fail per workspace, or "no gates configured").
- **Follow-ups** — anything that needs a human decision beyond what `fix` could resolve
  on its own (e.g. a gate that's still red after a best-effort repair attempt).
