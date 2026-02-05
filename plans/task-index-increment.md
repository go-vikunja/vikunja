# Design Doc: Project-Scoped Auto-Incrementing Task Index

**Status:** Draft  
**Date:** 2026-01-30

---

## Problem Statement

Our application has a **tasks** table where each task belongs to a project. Tasks have two identifiers:

- **id** — A globally unique, auto-incrementing primary key
- **index** — A human-friendly, project-scoped sequential number (e.g., PROJ-1, PROJ-2)

The **index** field must increment independently per project, starting at 1 for each new project.

Currently, the application code computes the next index by querying `MAX(index) + 1` before inserting a new task. This approach fails under concurrent inserts—two simultaneous requests can read the same MAX value and produce duplicate indexes, violating data integrity.

---

## Requirements

1. **Uniqueness:** Each (project_id, index) pair must be unique
2. **Sequential:** Indexes should be sequential within each project (no gaps under normal operation)
3. **Concurrent-safe:** Must handle multiple simultaneous inserts correctly
4. **Performance:** Should scale with the number of projects and tasks

---

## Options

### Option 1: Counter Table with Row-Level Locking

Maintain a separate table that tracks the last assigned index for each project. A trigger atomically increments the counter and assigns the value to new tasks.

**Implementation:**

```sql
-- Counter table
CREATE TABLE project_task_counters (
    project_id INTEGER PRIMARY KEY REFERENCES projects(id),
    last_index INTEGER NOT NULL DEFAULT 0
);

-- Trigger function
CREATE OR REPLACE FUNCTION set_task_index()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO project_task_counters (project_id, last_index)
    VALUES (NEW.project_id, 1)
    ON CONFLICT (project_id) DO UPDATE
        SET last_index = project_task_counters.last_index + 1
    RETURNING last_index INTO NEW.index;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER task_set_index
BEFORE INSERT ON tasks
FOR EACH ROW EXECUTE FUNCTION set_task_index();
```

**How it works:**

- The UPSERT (INSERT ... ON CONFLICT DO UPDATE) acquires a row-level lock on the counter row
- Concurrent transactions block until the lock is released
- Each transaction gets the next sequential value guaranteed

---

### Option 2: Advisory Locks with MAX Query

Use PostgreSQL advisory locks to serialize inserts per project, then compute MAX(index) + 1 safely.

**Implementation:**

```sql
CREATE OR REPLACE FUNCTION set_task_index()
RETURNS TRIGGER AS $$
BEGIN
    -- Lock based on project_id (released at transaction end)
    PERFORM pg_advisory_xact_lock('tasks'::regclass::integer, NEW.project_id);
    
    SELECT COALESCE(MAX(index), 0) + 1 INTO NEW.index
    FROM tasks
    WHERE project_id = NEW.project_id;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER task_set_index
BEFORE INSERT ON tasks
FOR EACH ROW EXECUTE FUNCTION set_task_index();
```

**How it works:**

- pg_advisory_xact_lock acquires an exclusive lock scoped to the project
- Lock is automatically released when the transaction commits or rolls back
- No additional tables required

---

### Option 3: Optimistic Locking with Retry

Add a unique constraint on (project_id, index) and retry on conflict at the application level.

**Implementation:**

```sql
-- Database constraint
ALTER TABLE tasks ADD CONSTRAINT tasks_project_index_unique 
    UNIQUE (project_id, index);
```

```python
# Application code (pseudocode)
def create_task(project_id, title):
    for attempt in range(MAX_RETRIES):
        try:
            next_index = db.query(
                "SELECT COALESCE(MAX(index),0)+1 FROM tasks WHERE project_id=$1", 
                project_id
            )
            db.execute(
                "INSERT INTO tasks (project_id, index, title) VALUES ($1, $2, $3)", 
                project_id, next_index, title
            )
            return
        except UniqueViolation:
            continue
    raise Exception("Failed after retries")
```

---

## Trade-off Comparison

| Criteria | Option 1: Counter Table | Option 2: Advisory Locks | Option 3: Optimistic Retry |
|----------|------------------------|-------------------------|---------------------------|
| **Performance** | O(1) — single row update | O(n) — scans tasks table | O(n) per attempt — may retry multiple times |
| **Complexity** | Medium — extra table to manage | Low — no schema changes | Medium — retry logic in app |
| **Scalability** | Excellent — constant time regardless of task count | Degrades with task count per project | Degrades under high contention |
| **Contention Behavior** | Serializes at counter row — predictable | Serializes via advisory lock — predictable | Retries under collision — unpredictable latency |
| **Gap-Free Guarantee** | Yes (if transactions commit) | Yes (if transactions commit) | Yes (if transactions commit) |
| **External Dependencies** | None — pure PostgreSQL | None — pure PostgreSQL | Application retry logic |

---

## Recommendation

**Option 1 (Counter Table)** is recommended for most use cases because:

- **Constant-time performance** regardless of how many tasks exist per project
- **Fully contained in PostgreSQL** — no application-level coordination required
- **Predictable behavior** under concurrent load
- **Simple operational model** — counter table can be inspected for debugging

Consider **Option 2 (Advisory Locks)** if you have a small number of tasks per project and want to avoid schema changes.

**Option 3 (Optimistic Retry)** is not recommended for this use case because it shifts complexity to the application layer and has unpredictable performance under contention.

---

## Migration Notes

If adopting Option 1 with existing data, initialize the counter table:

```sql
INSERT INTO project_task_counters (project_id, last_index)
SELECT project_id, MAX(index)
FROM tasks
GROUP BY project_id;
```

