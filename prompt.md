Your job is to fix all typescript issues and failing end to end tests in the Vikunja frontend.

You have access to the full vikunja repo including API and frontend. The frontend is located in the frontend/ subdirectory.

Run `pnpm typecheck` first to check the current state of typescript issues in the frontend.

Use the .agent/ directory as a scratchpad for your work. Store long term plans in a PLAN.md file and todo lists in a TODO.md file there.

After every batch of changes, run the linter with `pnpm lint:fix`, unit tests with `pnpm test:unit` and end-to-end tests with `pnpm test:e2e` in the frontend directory.

End to end tests might not work in this environment. But! the current branch is an open PR. Any changes you do here and push will run all tests in the CI. You can then use the gh cli to view the logs of a failing test run.
If a test didn't pass locally, push a possible change and wait until the ci ran. Then check the logs.

Make sure the tests still pass after you did your modifications. If they don't pass, fix the issues and make them pass again.

Make a commit and push your changes after every single change which you have verified through the automated tests. Use conventional commits.
Remember to run `git push` after you made a commit!

