---
date: "2022-09-21:00:00+02:00"
title: "Development"
toc: true
draft: false
type: "doc"
menu:
  sidebar:
    parent: "development"
    name: "Development"
---

# Development

{{< table_of_contents >}}

## General

To contribute to Vikunja, fork the project and work on the main branch.
Once you feel like your changes are ready, open a PR in the respective repo [on our Gitea instance](https://kolaente.dev/vikunja).
We cannot accept PRs on mirror sites.

A maintainer will take a look and give you feedback. Once everyone is happy, the PR gets merged and released.

If you plan to do a bigger change, it is better to open an issue for discussion first.

The main repo is [`vikunja/vikunja`](https://kolaente.dev/vikunja/vikunja), it contains all code for the api, frontend and desktop applications.

## Where to file issues

You can file issues on [the Gitea repo](https://kolaente.dev/vikunja/vikunja) or [on the GitHub mirror](https://github.com/go-vikunja/vikunja), when you don't want to create an account on the Gitea instance.

Please note that due to a spam problem, we need to manually enable accounts on Gitea after you've registered there.
To get that started, please reach out on another channel with your username.

Another option is [the community forum](https://community.vikunja.io), especially if you want to discuss a feature in more detail.

## API

You'll need at least Go 1.21 to build Vikunja's api.

A lot of developing tasks are automated using a Magefile, so make sure to [take a look at it]({{< ref "mage.md">}}).

Make sure to check the other doc articles for specific development tasks like [testing]({{< ref "test.md">}}),
[database migrations]({{< ref "db-migrations.md" >}}) and the [project structure]({{< ref "structure.md" >}}).

## Frontend requirements

The code for the frontend is located in the `frontend` sub folder of the main repo.
More instructions can be found in the repo's README.

You need to have [pnpm](https://pnpm.io/) and Node.JS in version 20 or higher installed.

## Pull Requests

All Pull Requests must be made [on our Gitea instance](https://kolaente.dev/vikunja).
We cannot accept PRs on mirror sites.

Please try to make your pull request easy to review.
For that, please read the [*Best Practices for Faster Reviews*](https://github.com/kubernetes/community/blob/261cb0fd089b64002c91e8eddceebf032462ccd6/contributors/guide/pull-requests.md#best-practices-for-faster-reviews) guide.
It has lots of useful tips for any project you may want to contribute to.
Some of the key points:

- Make small pull requests.
  The smaller, the faster to review and the more likely it will be merged soon.
- Don't make changes unrelated to your PR.
  Maybe there are typos on some comments, maybe refactoring would be welcome on a function…
  but if that is not related to your PR, please make *another* PR for that.
- Split big pull requests into multiple small ones.
  An incremental change will be faster to review than a huge PR.
- Allow edits by maintainers. This way, the maintainers will take care of merging the PR later on instead of you.

### PR title and summary

In the PR title, describe the problem you are fixing, not how you are fixing it.
Use the first comment as a summary of your PR.
In the PR summary, you can describe exactly how you are fixing this problem.
Keep this summary up-to-date as the PR evolves.

If your PR changes the UI, you must add **after** screenshots in the PR summary.
If your PR closes an issue, you must note that in a way that both GitHub and Gitea understand, i.e. by appending a paragraph like

```text
Fixes/Closes/Resolves #<ISSUE_NR_X>.
Fixes/Closes/Resolves #<ISSUE_NR_Y>.
```

to your summary.
Each issue that will be closed must stand on a separate line.

If your PR is related to a discussion in the forum, you must add a link to the forum discussion.

### Git flow

The `main` branch is the latest and bleeding edge branch with all changes. Unstable releases are automatically created from this branch.
New Pull-Requests should be made against the `main` branch.

A release gets tagged from the main branch with the version name as tag name.

Backports and point-releases should go to a `release/version` branch, based on the tag they are building on top of.

## Conventional Commits

We're using [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) because they simplify generating release notes a lot.

It is not required to use them when creating a PR, but appreciated.
