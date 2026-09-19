# Git, plans, worktrees

## Commits

- Conventional Commits (`feat: add foo`, `fix: correct bar`). Changelogs are generated from them.
- Lint before committing and fix what it reports:
  - backend changes: `mage lint:fix`
  - frontend changes: `cd frontend && pnpm lint:fix`
  - CSS/SCSS or Vue `<style>` changes: also `pnpm lint:styles:fix`
- Never commit edits to `pkg/swagger/`; CI regenerates it after commit.

## Plans

When asked to create a plan, write it to `plans/<kebab-case-name>.md` (e.g. `fix-position-healing.md`). `plans/` is gitignored; never force-add it.

## Worktrees

To implement a plan in isolation, invoke the `prepare-worktree` skill.

## Stacked PRs

- One session owns a stack and does every rebase from its own worktree. Don't check out a stack branch in another worktree: git refuses to rebase a branch checked out elsewhere, and `gh stack rebase` aborts.
- Cascade with recorded tips: note each layer's tip before a fix round, then `git rebase --onto <new lower tip> <recorded old lower tip>` per layer, and push with `--force-with-lease=<branch>:<old tip>`. Never derive the old base from `merge-base` once a lower layer was rebased; it replays the lower layer's commits into the upper one.
- After rebasing a stack onto `main`, search for imports of modules the stack deletes. New files on `main` can reference them without causing a conflict.
- When a stack merges as a whole, intermediate PRs may depend on later ones. Say so in each PR body so reviewers don't report those states as bugs.
- Add a layer with `gh stack add` or `gh stack link`, never plain `gh pr create`; that PR doesn't join the stack.
- A new helper shared across resources gets its own PR at the bottom of the stack, and that PR moves every existing copy onto it. Never add it inside a resource PR or leave the old and new pattern side by side.

## Reviewable PRs

- Keep mechanical changes and behaviour changes in separate commits. A mechanical commit (renames, `projectId` to `project_id`, swapping legacy types for generated ones) contains no logic change; a line that needs different behaviour waits for the next commit, even if that leaves a type error in between. Mark the subject: `refactor(project-views): switch consumers to generated types (mechanical)`.
- Commits in a PR don't need to build individually; the history shows the real sequence of work.
- Produce mechanical changes with a script where possible (sed, ts-morph) and put the exact command in the commit body, so the reviewer can check the script and re-run it instead of reading the diff.
- Renames of data fields must not touch i18n keypaths or `$t()` keys. Grep the diff for `keypath=` and `t('` before committing; the `sharedBy` to `shared_by` rename rendered raw keys.
- Order commits the same way in every PR: tests pinning current behaviour, shared test harness, new query module, mechanical consumer switch, behaviour changes, deletion of the legacy files.
- While a PR is a draft, record corrections as `git commit --fixup <sha>`. Before marking it ready, fold them in with `git rebase -i --autosquash` so reviewers read the final commits, not the iteration history. This is the one exception to never amending or rewriting commits, and it applies only to PRs nobody has reviewed yet.
- A fixup that touches a test file targets the source commit owning the module, not the test commit. Autosquash moves fixups next to their target, so a fixup of the test commit lands before later source fixups on the same file and conflicts. Dry-run `GIT_SEQUENCE_EDITOR=true git rebase -i --autosquash <base>` in a throwaway worktree after each fix round.
- Before merge: rebase the whole stack onto `origin/main` first, then autosquash each layer onto the squashed layer below. Check `git diff --numstat <base>..<tip>` per layer is unchanged by the rebase and the tree is identical before and after the squash.
- Start the PR body with a short review guide: which commits are mechanical and can be skimmed, the few decisions that need real review, deviations from the plan, and any intermediate state that is knowingly broken.
- One resource per PR. When a PR starts carrying a neighbouring migration, split that into its own PR in the stack.

## Refactors

- Find the consumers of a module by import path (`rg "components/misc/Subscription"`), not by component tag or local name; components are often imported under an alias.
