# Supported databases

Nothing in the code checks the server version (`vikunja doctor` only prints it). The real minimums come from the hand-rolled SQL, not from xorm.

| DB | Minimum | Why |
|---|---|---|
| MySQL | 8.0 | `WITH RECURSIVE` + `ROW_NUMBER() OVER` in `pkg/models/project_access.go` and `pkg/models/subscription.go`, with no fallback. 5.7 breaks permission checks. |
| MariaDB | 10.2 | Same CTE and window functions. |
| PostgreSQL | 9.5 | `ON CONFLICT` upserts (`kanban_task_bucket.go`, `task_position.go`, `project_view.go`). `jsonb ?|` in `saved_filters.go` needs 9.4. |
| SQLite | 3.25 | Window functions, upsert with `excluded.` (3.24), `RENAME COLUMN` in `renameColumn`. The bundled go-sqlite3 ships a newer SQLite; only matters with the `libsqlite3` build tag. |

CI tests only the newest versions: MySQL 8, MariaDB 12, Postgres 18, bundled SQLite (`.github/workflows/test.yml`). The minimums above are not tested.

## Writing SQL

- Before using a SQL feature in raw SQL or a migration, check it exists on every minimum in the table. If not, branch on `db.Type()`.
- If raising a minimum is unavoidable: update the table here, call it out as a breaking change in the PR description, and flag that the website docs (supported databases section) need the same change.
- MySQL has no partial indexes and no `CREATE INDEX IF NOT EXISTS`. Skip it there (see `pkg/migration/20260906010000.go`).
- `ON DUPLICATE KEY UPDATE x = VALUES(x)` is deprecated since MySQL 8.0.20. Keep it anyway: the replacement row alias syntax (`AS new ... new.x`) doesn't exist on MariaDB, and both share the `mysql` dialect.
- Case-insensitive search goes through `db.ILIKE` (ILIKE on Postgres, LIKE elsewhere).
- ParadeDB `bm25` indexes (`pkg/db/db.go`) are optional and only created when the extension is present.
