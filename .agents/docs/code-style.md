# Code style

## Go

- Write struct literals with multiple fields and slice/map literals with multiple entries across multiple lines, one field or entry per line, with trailing commas. Apply this to nested literals and tests too; do not pack fields or entries together just because `gofmt` permits it.
- Wrap errors: `fmt.Errorf("...: %w", err)`.
- No raw SQL anywhere, migrations and tests included. Use XORM's builder (`s.Where(...)`, `builder.In`, `.Cols().Update()`), never hand-built strings via `s.Exec`, `s.Query`, or `builder.Expr`.
  Gotcha: an argument-less `builder.In("col")` is silently dropped by `Where` and matches every row. Pass an empty typed slice (`[]int64{}`) to get `0=1`.
- High-entropy random tokens are stored as plain SHA-256 (`utils.Sha256Hex`), never a slow KDF — see `sessions.go`. User passwords stay bcrypt.
- Never log secrets.
- Before adding a file, function, or helper, `rg` for existing code that does the same thing. Extend it rather than duplicate.

## Comments

Document the *why*, not the *what*. Default to no comment. Only a non-obvious gotcha, invariant, rejected alternative, or cross-file constraint earns one tight line. Cut comments that restate the code, a name, or a signature.
