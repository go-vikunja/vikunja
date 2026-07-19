---
name: 16-eyes
description: >-
  Security audits with four subcommands — `/16-eyes init` (configure: detect
  quality gates, exclude patterns, output location, and design the repo's
  tailored investigation lenses — auto-runs non-interactively if skipped),
  `/16-eyes audit` (run every persisted lens across the WHOLE repo, verify
  every finding skeptically, adversarially re-check high-impact ones, produce
  a classified safe/risky report), `/16-eyes audit-diff` (the identical
  engine, scoped to a diff/PR instead of the whole repo — for reviewing a
  change before merge), `/16-eyes fix` (apply the findings — safe ones
  directly, risky ones with your confirmation, never commits or pushes).
  `audit` is the deliberate, occasional deep sweep (dozens of subagent calls,
  several minutes) that diff-scoped tools miss — a vulnerability sitting
  untouched in the codebase for months is invisible to them. `audit-diff` is
  the fast, per-PR counterpart, built on the same persisted-lens + skeptical
  verify + adversarial multi-vote pipeline (unlike Claude Code's built-in
  `/security-review` or most CI-wired scanners, which do a single pass).
  Trigger on: full security audit, complete security review of a repo, scan
  the whole codebase for vulnerabilities, review this PR/diff for security
  issues, "16-eyes init/audit/audit-diff/fix"; "auditoria de segurança
  completa do repositório", "varredura de segurança full-repo", "audita esse
  repo inteiro por vulnerabilidades", "revisão de segurança deste PR/diff";
  "auditoría de seguridad completa del repositorio", "escanea todo el código
  en busca de vulnerabilidades", "revisa este PR/diff en busca de problemas
  de seguridad".
---

# 16 Eyes — security audits, full-repo or diff-scoped

Sixteen independent eyes look at every finding before it reaches you: the
lens that found it, a skeptical verifier that re-reads the real code, and —
for high-impact findings — several adversarial reviewers actively trying to
disprove it. Nothing reaches the report on one agent's word alone. The same
rigor applies whether you're running a full-repo sweep or reviewing a single
diff — see `/16-eyes audit` vs `/16-eyes audit-diff` below.

## Why this is different from `/security-review` and CI scanners

Diff-scoped tools (Claude Code's built-in `/security-review`, most CI-wired
scanners) do a single pass over what changed and only see the current
PR/branch. `/16-eyes audit-diff` is also diff-scoped, but runs the repo's own
tailored investigation lenses, verifies every finding skeptically against
the real code, and adversarially re-checks high-impact ones with multiple
independent reviewers before anything reaches the report — a heavier, more
skeptical pipeline than a single-pass reviewer. `/16-eyes audit` goes further
still: it scans the **whole repository**, regardless of recent changes — a
vulnerability that's been sitting untouched in the codebase for months is
invisible to any diff-scoped tool, `audit-diff` included. Use `audit-diff`
for "does this PR introduce a problem?"; use `audit` for "does this codebase,
as a whole, have problems?" `audit` is expensive (dozens of subagent calls,
several minutes) — a deliberate, occasional deep sweep, not a per-commit
check.

## When the user runs `/16-eyes init`

Read `references/init-flow.md` and follow it end to end. Detects the repo's
quality-gate commands (test/lint/typecheck/build, across ecosystems, not just
Node), interviews briefly about exclude patterns / output location / audit
depth / language / whether to gitignore reports / whether to scaffold a CI
template, then **profiles the repo and designs its investigation lenses**,
persisting both to `.16-eyes/config.json` and `.16-eyes/lenses.json`. If
skipped, `/16-eyes audit` and `/16-eyes audit-diff` both auto-bootstrap this
non-interactively (sane defaults, no questions asked — safe to run headlessly
in CI) the first time they need lenses that don't exist yet. Running it
explicitly first just lets you customize things before that happens.

## When the user runs `/16-eyes audit`

Read-only — never edits code. Read `references/audit-flow.md` and follow it
end to end: load the repo's persisted investigation lenses (bootstrapping
them via `init` first if none exist), run them across the whole repo via the
`Workflow` tool, verify every finding, adversarially re-check high-impact
ones, and write a classified markdown + JSON report.

## When the user runs `/16-eyes audit-diff`

Read-only — never edits code. Read `references/audit-diff-flow.md` and
follow it end to end: same persisted lenses and same verify/adversarial
pipeline as `audit`, but each lens investigates only a diff's changed hunks
(default: against the repo's default branch, or an explicit PR/base ref the
user named). Use this to review a PR before merge, including as a CI check
(see `references/ci-flow.md`).

## When the user runs `/16-eyes fix`

Read `references/fix-flow.md` and follow it end to end. Applies the most
recent `/16-eyes audit` or `/16-eyes audit-diff` findings (from this
conversation if fresh, otherwise the last saved report of either kind) —
`safe` findings directly, `risky` ones with your explicit confirmation one at
a time. **Never runs any git write command** (no commit, no push) — changes
are always left in the working tree for you to review.
