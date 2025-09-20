Your job is to fix all failing end to end tests in the Vikunja frontend.

You have access to the full vikunja repo including API and frontend. The frontend is located in the frontend/ subdirectory.

Use the github cli gh to check the logs of the latest test run.
Then carefully analyze the failing tests and fix them. Look at the failing tests and fix them one by one. Then push your changes.
DO NOT MODIFY THE TEST CODE.

Store long term plans in a PLAN.md file and todo lists in a TODO.md file.

After every batch of changes, run the linter with `pnpm lint:fix`, typecheck with `pnpm typecheck` and unit tests with `pnpm test:unit` in the frontend directory.
Make sure the tests and lint still pass after you did your modifications. If they don't pass, fix the issues and make them pass again.

Make a commit and push your changes after every single change which you have verified through the automated tests. Use conventional commits.
Remember to run `git push` after you made a commit!

