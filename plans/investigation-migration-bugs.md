# Investigation: Database Migrations Recorded But Not Applied

**Issue Reference:** https://github.com/go-vikunja/vikunja/issues/2172
**Date:** 2026-02-04

## Problem Statement

Users report that some database migrations appear in the `migration` table but their effects are not present in the database. Specifically:

1. **Reporter 1 (SamBouwer):** Duplicate entries in migrations table — all migrations appear twice, with two `SCHEMA_INIT` rows. The `bucket_configuration` column contains old-format string filters instead of the expected JSON object format.

2. **Reporter 2 (freeware-superman):** All migrations after `20240919130957` are completely missing from the migrations table (roughly 11 migrations), despite being known to the binary. No duplicate entries.

---

## Root Causes Identified

Three distinct problems were found:

### Problem 1: Race Condition — No Locking, No Unique Constraint

**Location:** `xormigrate@v1.7.1/xormigrate.go:287-308` and `:358-370`

The `Migration` struct definition:

```go
type Migration struct {
    ID string `xorm:"id"`  // Just a column name, NOT a primary key!
    // ...
}
```

The tag `xorm:"id"` is **only a column name mapping**, not a primary key declaration. XORM only auto-detects primary keys for untagged `int64` fields named `ID`. The `migration` table has **no UNIQUE constraint and no PRIMARY KEY** on the `id` column.

Combined with zero locking in the migration process:

1. `canInitializeSchema()` checks `COUNT(*) == 0` on the migrations table
2. `runInitSchema()` inserts `SCHEMA_INIT` + all migration IDs
3. There is no advisory lock, no `SELECT FOR UPDATE`, no mutex between steps 1 and 2

**Race scenario:**

If two Vikunja instances start concurrently (Docker restart, Kubernetes scaling, systemd `Restart=always`):

1. Instance A creates migration table, sees it's empty
2. Instance B sees table exists, also sees it's empty (A hasn't inserted yet)
3. Instance A runs `initSchema`, inserts all migration IDs
4. Instance B runs `initSchema`, inserts all migration IDs again

Result: Every migration ID appears twice, including two `SCHEMA_INIT` rows.

**This explains Reporter 1's duplicate entries.**

---

### Problem 2: `initSchema` Records Migrations Without Executing Them

**Location:** `xormigrate@v1.7.1/xormigrate.go:287-308`

The `runInitSchema` function:

```go
func (x *Xormigrate) runInitSchema() error {
    // 1. Run initSchema - creates tables via Sync2 (DDL only)
    if err := x.initSchema(x.db); err != nil {
        return err
    }

    // 2. Insert SCHEMA_INIT marker
    if err := x.insertMigration(initSchemaMigrationId); err != nil {
        return err
    }

    // 3. Insert ALL migration IDs WITHOUT executing migration functions
    for _, migration := range x.migrations {
        if err := x.insertMigration(migration.ID); err != nil {
            return err
        }
    }
    return nil
}
```

This is **by design** for fresh databases (no data to migrate). But it causes problems when:

- The migrations table gets corrupted/emptied while data still exists
- Concurrent starts cause the race condition above

In the race scenario, if Instance A completes `initSchema` and inserts all IDs, then Instance B's `canInitializeSchema()` check sees `count > 0` and falls through to the per-migration loop. There, `migrationDidRun()` returns `true` for every migration (IDs exist from Instance A), so **no migration function ever runs**.

But `initSchema` only creates schema (DDL) — it doesn't perform data transformations. Any existing data with old-format filters stays unconverted.

**This explains why Reporter 1's `bucket_configuration` has old-format strings despite migrations being "recorded".**

---

### Problem 3: Non-Transactional Migration Execution

**Location:** `xormigrate@v1.7.1/xormigrate.go:310-339`

All Vikunja migrations use the `Migrate` field (engine path):

```go
func (x *Xormigrate) runMigration(migration *Migration) error {
    if !x.migrationDidRun(migration) {
        // Execute migration - NO TRANSACTION WRAPPING
        if migration.Migrate != nil {
            if err := migration.Migrate(x.db); err != nil {
                return fmt.Errorf("migration %s failed: %s", migration.ID, err.Error())
            }
        }

        // Record migration ID AFTER execution
        if err := x.insertMigration(migration.ID); err != nil {
            return fmt.Errorf("inserting migration %s failed: %s", migration.ID, err.Error())
        }
    }
    return nil
}
```

When `migration.Migrate(x.db)` is called, the migration receives a `*xorm.Engine`. Every operation (`tx.Exec()`, `tx.Find()`, `tx.Update()`) creates a fresh auto-commit session — **there is no transaction wrapping**.

**Failure scenarios:**

1. **Migration succeeds but insert fails:** The migration's effects are committed but not recorded. Next startup re-attempts the migration, which may fail with "duplicate column" errors.

2. **Migration partially succeeds then fails:** For multi-statement migrations, early statements are auto-committed. The migration is not recorded, and the next attempt sees a partially-modified database.

3. **Early migration failure halts entire chain:** If migration N fails, migrations N+1 through M are never attempted. The process calls `log.Fatalf` at `pkg/migration/migration.go:78`.

**This could explain Reporter 2's missing migrations after `20240919130957`.**

---

### Problem 4: Logic Bug in Migration 20241028131622

**Location:** `pkg/migration/20241028131622.go:57`

```go
if err != nil && (!strings.Contains(err.Error(), "Error 1061") || !strings.Contains(err.Error(), "Duplicate key name")) {
    return err
}
```

This condition uses `||` (OR) when it should use `&&` (AND).

**Current behavior:** "Return error if it does NOT contain 'Error 1061' OR does NOT contain 'Duplicate key name'"

Since MySQL error messages typically contain one pattern but not both simultaneously, at least one `!strings.Contains` is almost always `true`, meaning this **returns the error instead of suppressing it**.

**Correct logic should be:**
```go
if err != nil && !strings.Contains(err.Error(), "Error 1061") && !strings.Contains(err.Error(), "Duplicate key name") {
    return err
}
```

On MySQL 8 (which doesn't support `IF NOT EXISTS` for `CREATE INDEX`), if indexes already exist, this migration would fail and halt the entire migration chain at `20241028131622`, leaving all subsequent migrations unrecorded.

**This could specifically explain Reporter 2's case on MySQL 8.**

---

## Evidence Summary

| Finding | Location | Impact |
|---------|----------|--------|
| No UNIQUE/PK on migration.id | xormigrate Migration struct | Allows duplicate rows |
| No locking in migration process | xormigrate migrate() | Race condition on concurrent starts |
| initSchema inserts IDs without running migrations | xormigrate runInitSchema() | Data transformations skipped |
| Auto-commit per statement, no transaction | xormigrate runMigration() | Partial failures leave inconsistent state |
| `\|\|` vs `&&` logic bug | 20241028131622.go:57 | False failures on MySQL 8 |

---

## Existing Mitigations

The team has already added "catch-up" migrations that re-check and fix data:

- `20250323212553.go` — Re-converts any `filter` values still in old string format
- `20251001113831.go` — Re-converts any `bucket_configuration` filters still in string format

These migrations are idempotent and handle mixed-format data, which suggests awareness of this pattern.

---

## Recommended Fixes

### Fix 1: Add UNIQUE constraint to migration table

Either patch xormigrate or add a startup check:

```go
// Option A: Modify xormigrate's Migration struct
type Migration struct {
    ID string `xorm:"pk"`  // or `xorm:"unique"`
    // ...
}

// Option B: Add constraint via raw SQL at startup
ALTER TABLE migration ADD CONSTRAINT migration_id_unique UNIQUE (id);
```

### Fix 2: Add database-level advisory locking

Wrap the entire migration process in an advisory lock:

```go
// MySQL
SELECT GET_LOCK('vikunja_migrations', 30);
// ... run migrations ...
SELECT RELEASE_LOCK('vikunja_migrations');

// PostgreSQL
SELECT pg_advisory_lock(12345);
// ... run migrations ...
SELECT pg_advisory_unlock(12345);
```

### Fix 3: Fix the `||` vs `&&` logic bug

In `pkg/migration/20241028131622.go:57`:

```go
// Before (buggy)
if err != nil && (!strings.Contains(err.Error(), "Error 1061") || !strings.Contains(err.Error(), "Duplicate key name")) {

// After (correct)
if err != nil && !strings.Contains(err.Error(), "Error 1061") && !strings.Contains(err.Error(), "Duplicate key name") {
```

### Fix 4: Make data migrations idempotent

Every data-transformation migration should check current state before modifying:

```go
// Example: Only convert if not already in new format
if !strings.HasPrefix(view.Filter, "{") {
    // Convert to new format
}
```

### Fix 5: Consider a "repair" command

Add a CLI command that re-runs data migrations regardless of the tracking table, useful for recovering from inconsistent states.

---

## Questions for Further Investigation

1. Can we reproduce the race condition in a test environment?
2. What is the exact MySQL version and error message for Reporter 2?
3. Should we upstream fixes to xormigrate or fork/vendor it?
4. Is there telemetry showing how common concurrent starts are?
