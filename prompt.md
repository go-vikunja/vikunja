Your job is to fix all typescript issues in the Vikunja frontend.

You have access to the full vikunja repo including API and frontend. The frontend is located in the frontend/ subdirectory.

Run `pnpm typecheck` first to check the current state of typescript issues in the frontend.

Use the .agent/ directory as a scratchpad for your work. Store long term plans and todo lists there.

After every batch of changes, run the linter with `pnpm lint:fix`, unit tests with `pnpm test:unit` and end-to-end tests with `pnpm test:e2e` in the frontend directory.

Make sure the tests still pass after you did your modifications. If they don't pass, fix the issues and make them pass again.

Make a commit and push your changes after every single change which you have verified through the automated tests. Use conventional commits.
Remember to run `git push` after you made a commit!

