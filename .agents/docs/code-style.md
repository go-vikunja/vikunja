# Code style

## General

- Before adding a file, function, or helper, `rg` for existing code that does the same thing, in Go and in `frontend/src/`. Extend it rather than duplicate.

## Go

- Write struct literals with multiple fields and slice/map literals with multiple entries across multiple lines, one field or entry per line, with trailing commas. Apply this to nested literals and tests too; do not pack fields or entries together just because `gofmt` permits it.
- Wrap errors: `fmt.Errorf("...: %w", err)`.
- No raw SQL anywhere, migrations and tests included. Use XORM's builder (`s.Where(...)`, `builder.In`, `.Cols().Update()`), never hand-built strings via `s.Exec`, `s.Query`, or `builder.Expr`.
  Gotcha: an argument-less `builder.In("col")` is silently dropped by `Where` and matches every row. Pass an empty typed slice (`[]int64{}`) to get `0=1`.
- High-entropy random tokens are stored as plain SHA-256 (`utils.Sha256Hex`), never a slow KDF — see `sessions.go`. User passwords stay bcrypt.
- Never log secrets.

## Frontend

- Same literal rule as Go: object and type literals with more than one property, long import lists and long argument lists go one entry per line with trailing commas, in tests too. Nothing over ~120 columns; split long ternaries and chains at the operator. `frontend/src/client/queries/projects.ts` is the reference.
- One ref per purpose; a delete target must not share the search selection's ref. After an awaited mutation, clear or close UI state only if it still equals the value captured when the action started, because users keep interacting while requests are pending.
- Never mutate a parent's `v-model` array or object in place; emit the new value. Editing a copied prop broke assignee removal: the removed user came back on the next save.
- Don't bind native `disabled` to a control that can hold focus during or after a mutation, including an empty state reached after success; focus drops to `<body>`. Use `aria-disabled` and guard the handler.

## Comments

Document the *why*, not the *what*. Default to no comment. Only a non-obvious gotcha, invariant, rejected alternative, or cross-file constraint earns one tight line. Cut comments that restate the code, a name, or a signature.
