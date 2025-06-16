# Histoire stories

The frontend still contains a few `.story.vue` files. These were written for the
[Histoire](https://histoire.dev/) component explorer which was previously used
for isolated component development and screenshot testing.  Storybook does not
load these files because it only processes `*.stories.*` globs.

While new stories should be written for Storybook, the existing Histoire files
remain for reference and for screenshot-based tests. They can be safely ignored
when working on Storybook.
