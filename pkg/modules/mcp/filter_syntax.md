# Vikunja filter syntax

Filters select tasks: the `filter` argument of task list tools, project view filters and saved filters. Time entry filters use the same grammar with their own fields.

## Grammar

A filter is one or more conditions of the form `field operator value`.

- Combine conditions with `&&` (and) or `||` (or). The words `and` / `or` are **not** operators.
- Group with parentheses: `(priority >= 4 || labels = 3) && done = false`
- Quote values that contain spaces or `&`, `|`, `(`, `)`: `title like 'call mum'`

## Operators

| Operator | Meaning |
| --- | --- |
| `=` `!=` | equal, not equal |
| `>` `>=` `<` `<=` | comparison (numbers, dates) |
| `like` | substring match on text fields; `%` inside the value is a wildcard |
| `in` `not in` | comma-separated list: `project in 1, 2, 3` |

## Fields

| Field | Type | Notes |
| --- | --- | --- |
| `done` | bool | `true` / `false` |
| `priority` | number | 0 unset, 1 low … 5 do now |
| `percent_done` | number | 0 to 1 |
| `due_date`, `start_date`, `end_date`, `done_at`, `created`, `updated` | date | see Dates |
| `reminders` | date | matches if any reminder matches |
| `title`, `description` | text | use `like` for partial matches |
| `project` / `project_id` | number | project ID |
| `labels` | number | label ID, not the label title |
| `assignees` | text | username |
| `created_by` | text | username |
| `bucket_id` | number | Kanban bucket ID |
| `id`, `index`, `repeat_after`, `hex_color`, `uid` | | |

## Dates

Date fields accept absolute dates or date math. Both resolve in the `filter_timezone` argument (IANA name, e.g. `Europe/Berlin`), or the server timezone if unset.

### Absolute

- `2026-01-31`: 00:00 that day
- `'2026-01-31 14:00'`: quote it, the value contains a space
- `2026-01-31T14:00:00+01:00`: RFC 3339, carries its own offset

### Date math

An expression starts at `now`, followed by any number of steps, applied left to right:

- `+<n><unit>` / `-<n><unit>`: add or subtract, e.g. `+1d`, `-2w`
- `/<unit>`: round down to the start of that unit, e.g. `/d` is 00:00 today

| Unit | Meaning |
| --- | --- |
| `s` | seconds |
| `m` | minutes |
| `h`, `H` | hours |
| `d` | days |
| `b` | business days (Mon–Fri) |
| `w` | weeks (start Monday) |
| `M` | months (capital M; lowercase `m` is minutes) |
| `y` | years |

| Expression | Resolves to |
| --- | --- |
| `now` | right now |
| `now+24h` | 24 hours from now |
| `now+30d` | 30 days from now |
| `now/d` | today 00:00 |
| `now/d+1d` | tomorrow 00:00 |
| `now/w` | Monday 00:00 of this week |
| `now/w+1w` | Monday 00:00 of next week (exclusive end of this week) |
| `now/M` | first day of this month 00:00 |
| `now-1M/M` | first day of last month 00:00 |
| `now+5b` | five business days from now |

For a range, pair `>=` start with `<` end: `due_date >= now/d && due_date < now/d+1d` is "due today".

Anchoring math on a fixed date (`2026-01-31||+1M`) is not supported in filters; compute the absolute date instead.

## Nulls

Tasks without a value for a filtered field (e.g. no due date) are excluded. Pass `filter_include_nulls: true` to include them.

## Examples

- Open tasks due this week: `done = false && due_date >= now/w && due_date < now/w+1w`
- Overdue: `done = false && due_date < now`
- Due in the next 7 days: `done = false && due_date >= now && due_date < now+7d`
- Done last month: `done = true && done_at >= now-1M/M && done_at < now/M`
- High priority in two projects: `priority >= 4 && project in 5, 8`
- Tagged with label 3 or 7, assigned to alice: `labels in 3, 7 && assignees = alice`
- Title search: `title like invoice && done = false`
