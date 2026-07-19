# CI scaffolding — invoked from `/16-eyes init`

Triggered by `init-flow.md` Phase 5, only when the user answered yes to the Phase 3 "CI
template?" question. Writes files into the target repo's `.github/workflows/` and
`.claude/`, which is why it only ever runs after the user has explicitly agreed (Phase 4
already showed them this was coming — never reach this flow silently).

## 1. Confirm the skill itself is committed to the repo

CI needs `.claude/skills/16-eyes/` present in the checked-out repo — a headless CI job
can't `npx 16-eyes install` from the registry reliably for every run (version drift,
registry dependency, extra latency). Check whether `.claude/skills/16-eyes/` exists at
the repo root:

- **Missing**: tell the user, then run `npx 16-eyes@latest install --project` yourself
  via `Bash` (a local, reversible file copy — safe to just do, but say what you're
  about to do first). Remind them to `git add .claude/skills/16-eyes && git commit` —
  you never commit on their behalf.
- **Present**: continue.

## 2. Write the GitHub Actions workflow(s)

Copy `../assets/ci/pr-audit-diff.yml` (via `Read` then `Write`) to
`.github/workflows/16-eyes-audit-diff.yml`. If a file already exists at that path,
**never overwrite silently** — show the user both versions and ask.

If the user also wants the optional scheduled full-repo sweep (ask this explicitly,
it's opt-in, not implied by the Phase 3 answer), also copy
`../assets/ci/nightly-full-audit.yml` to `.github/workflows/16-eyes-full-audit.yml`, same
overwrite-protection rule.

## 3. Write or merge `.claude/settings.json`

Read `../assets/ci/settings.json` (the `dontAsk`-mode allowlist scoped to exactly what
`audit-diff` needs).

- **No `.claude/settings.json` exists yet**: write the template as-is.
- **One already exists**: `Read` it first. **Merge, never overwrite**:
  - Union the template's `permissions.allow` entries into the existing array
    (skip duplicates).
  - If `permissions.defaultMode` is already set to something other than `dontAsk`,
    **don't silently change it** — tell the user their existing mode may cause the CI
    job to hang waiting for approval it'll never get, and ask whether to switch it to
    `dontAsk` for this repo.
  - Leave every other existing key in the file untouched.

## 4. Tell the user what's still manual

You cannot set repository secrets yourself (that's a GitHub Actions/repo-settings
action, out of reach for this skill). Tell them plainly:

- Add `ANTHROPIC_API_KEY` as a repository (or organization) secret — Settings → Secrets
  and variables → Actions, on GitHub's own UI.
- The workflow **comments on every PR by default and never blocks merge**
  (`.16-eyes/config.json`'s `ci.failOn` defaults to `"none"`). To make it a required,
  merge-blocking check: set `ci.failOn` to `"risky"` (fail if any RISKY finding
  survives) or `"any"` (fail on any confirmed finding), **and** mark the
  `audit-diff` job as a required status check in the repo's branch protection rules
  (also a GitHub UI action, not something this skill can do for you).
- Point them at `docs/ci.md` in this package for the full writeup if they want more
  detail than this summary.

Set `.16-eyes/config.json`'s `ci.enabled` to `true` (already done in `init-flow.md`
Phase 5 step 2, just confirm it landed) so `/16-eyes init`, re-run later, knows CI is
already wired and won't ask again by default.
