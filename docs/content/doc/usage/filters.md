---
title: "Filters"
date: 2024-03-09T19:51:32+02:00
draft: false
type: doc
menu:
  sidebar:
    parent: "usage"
---

# Filter Syntax

To filter tasks via the api, you can use a query syntax similar to SQL. 

This document is about filtering via the api. To filter in Vikunja's web ui, check out the help text below the filter query input.

{{< table_of_contents >}}

## Available fields

The available fields for filtering include:

*   `done`: Whether the task is completed or not
*   `priority`: The priority level of the task (1-5)
*   `percentDone`: The percentage of completion for the task (0-100)
*   `dueDate`: The due date of the task
*   `startDate`: The start date of the task
*   `endDate`: The end date of the task
*   `doneAt`: The date and time when the task was completed
*   `assignees`: The assignees of the task
*   `labels`: The labels associated with the task
*   `project`: The project the task belongs to (only available for saved filters, not on a project level)

You can date math to set relative dates. Click on the date value in a query to find out more.

All strings must be either single-word or enclosed in `"` or `'`. This extends to date values like `2024-03-11`.

## Operators

The available operators for filtering include:

*   `!=`: Not equal to
*   `=`: Equal to
*   `>`: Greater than
*   `>=`: Greater than or equal to
*   `<`: Less than
*   `<=`: Less than or equal to
*   `like`: Matches a pattern (using wildcard `%`)
*   `in`: Matches any value in a comma-seperated list of values

To combine multiple conditions, you can use the following logical operators:

*   `&&`: AND operator, matches if all conditions are true
*   `||`: OR operator, matches if any of the conditions are true
*   `(` and `)`: Parentheses for grouping conditions

## Examples

Here are some examples of filter queries:

*   `priority = 4`: Matches tasks with priority level 4
*   `dueDate < now`: Matches tasks with a due date in the past
*   `done = false && priority >= 3`: Matches undone tasks with priority level 3 or higher
*   `assignees in user1, user2`: Matches tasks assigned to either "user1" or "user2
*   `(priority = 1 || priority = 2) && dueDate <= now`: Matches tasks with priority level 1 or 2 and a due date in the past


