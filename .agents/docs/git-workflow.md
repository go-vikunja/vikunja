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
