# Improve Permission Debugging for Rootless Docker (Issue #2173)

## Problem

When running Vikunja in **rootless Docker**, the container fails to start with:

```
Could not init file handler: storage validation failed: failed to create test file
at /app/vikunja/files/.vikunja-check-...: permission denied
```

The user sets `-u 1001:1001` and `chown 1001:1001` on the bind mount — UIDs appear
to match, but the write still fails. The reason is invisible: rootless Docker uses
**Linux user namespaces**, which remap UIDs. Container UID 1001 is **not** host
UID 1001 — it maps to a high subordinate UID (e.g., 101001) via `/etc/subuid`.

Current diagnostics don't surface this. The `doctor` command now shows directory
ownership and an ownership-match check, but the ownership match compares
**namespaced** UIDs (both appear as 1001 inside the container), which is misleading
when a user namespace is active. The real mismatch is invisible.

## Goal

Make it obvious to the user when a user namespace is active, so that the
UID-match check in doctor can warn that the UIDs it sees are remapped, and
so that the startup error message contains enough context to diagnose the issue
without a GitHub issue.

## Current State (after latest main)

### `pkg/files/filehandling.go` — `ValidateFileStorage()`
- Checks directory exists, creates it if missing, validates it's a directory
- Writes a test file, removes it
- On failure: `"failed to create test file at %s: %w"` — no UID/GID, no ownership info

### `pkg/doctor/files.go` — `checkLocalStorage()`
- Reports: Path, Directory exists, Directory permissions (octal), Directory owner
  (via `checkDirectoryOwnership`), Writable, Disk space, Stored files

### `pkg/doctor/files_unix.go` — `checkDirectoryOwnership()`
- Uses `syscall.Stat_t` to get directory UID/GID
- Compares against `os.Getuid()` — but both are namespaced values
- Reports ownership mismatch if UIDs differ

### `pkg/initialize/init.go:83-85`
- Calls `files.InitFileHandler()`, fatals on error with only the error string

## Implementation Plan

### Direction 1: Detect User Namespace Remapping

**What:** Add a Linux-specific function that reads `/proc/self/uid_map` to detect
whether UID remapping is active, and extract the mapping.

**Why:** This is the root cause of the confusion. Without surfacing this, all other
UID comparisons are misleading inside a user namespace.

**How `/proc/self/uid_map` works:**

The file contains lines of the form:
```
<inside_uid> <outside_uid> <count>
```

A trivial (non-remapped) mapping is a single line:
```
         0          0 4294967295
```

A rootless Docker mapping typically looks like:
```
         0       1001          1
         1     101001      65536
```

This means: container UID 0 → host UID 1001, container UIDs 1–65536 → host UIDs
101001–166536.

**Files to create/modify:**

1. **New file: `pkg/utils/userns_linux.go`** (build tag: `linux`)
   - `func IsUserNamespaceActive() bool` — reads `/proc/self/uid_map`, returns
     `true` if the mapping is non-trivial (not a single `0 0 4294967295` line)
   - `func GetUIDMapping() ([]UIDMapEntry, error)` — parses `/proc/self/uid_map`
     into a slice of `{InsideUID, OutsideUID, Count}` structs
   - `func MapToHostUID(containerUID int) (hostUID int, mapped bool)` — given a
     container UID, returns the corresponding host UID using the mapping. Returns
     `mapped=false` if no mapping entry covers that UID.
   - `func UIDMappingSummary() string` — returns a human-readable one-line summary
     of the mapping, e.g., `"0→1001, 1-65536→101001-166536"`

2. **New file: `pkg/utils/userns_other.go`** (build tag: `!linux`)
   - Stub implementations returning `false` / empty / not-available for all
     the above functions. User namespaces are a Linux-specific feature.

3. **New file: `pkg/utils/userns_linux_test.go`**
   - Unit tests that parse sample `/proc/self/uid_map` content (trivial mapping,
     rootless mapping, multi-line mapping)
   - Test `MapToHostUID` with various container UIDs

**Design notes:**
- Put these in `pkg/utils/` rather than `pkg/doctor/` because Direction 2 needs
  them from `pkg/files/` as well.
- Read `/proc/self/uid_map` only once and cache the result (sync.Once), since
  the mapping doesn't change during process lifetime.
- Keep the parsing minimal — the file format is simple and stable (Linux ABI).

---

### Direction 2: Enrich `ValidateFileStorage()` Errors

**What:** When the test file write fails with a permission error, include process
identity, directory ownership, and user namespace status in the error message.

**Why:** The startup `log.Fatalf` is the first (and often only) thing users see.
Currently it says `permission denied` with no context.

**Files to modify:**

1. **`pkg/files/filehandling.go`** — `ValidateFileStorage()` function

   At lines 223-225, when `writeToStorage` returns an error, and separately at
   lines 210-212 when `MkdirAll` fails:

   - Import `os`, `syscall`, and the new `pkg/utils` userns functions.
   - Build a diagnostic suffix string that includes:
     - Process UID/GID: `os.Getuid()`, `os.Getgid()`
     - Directory owner UID/GID: `os.Stat(basePath)` → `syscall.Stat_t`
     - User namespace status: `utils.IsUserNamespaceActive()`
     - If namespace is active: the UID mapping summary and what the process
       container UID maps to on the host
   - Append this to the error, e.g.:
     ```
     failed to create test file at /app/vikunja/files/.vikunja-check-...: permission denied
       [process uid=1001 gid=1001, dir owner uid=1001 gid=1001, user namespace ACTIVE (0→1001, 1-65536→101001-166536), process host uid=102001]
       Hint: User namespace is active (common in rootless Docker). The process appears as uid 1001 inside the container but maps to uid 102001 on the host. To fix: either run with -u 0:0, or chown the directory to the mapped host uid.
     ```

   To avoid duplicating the platform-specific stat logic, extract a helper:

2. **New file: `pkg/files/diagnostics_unix.go`** (build tag: `!windows`)
   - `func storageDiagnosticInfo(basePath string) string` — gathers process
     UID/GID, directory UID/GID via `syscall.Stat_t`, user namespace info via
     `utils.IsUserNamespaceActive()` / `utils.UIDMappingSummary()` /
     `utils.MapToHostUID()`, and returns a formatted diagnostic string.

3. **New file: `pkg/files/diagnostics_windows.go`** (build tag: `windows`)
   - Stub returning `""` (no additional diagnostics on Windows).

4. **`pkg/files/filehandling.go`** — call `storageDiagnosticInfo(basePath)` in
   the three error returns within `ValidateFileStorage()`:
   - Line 206: `"failed to access file storage directory at %s: %w\n%s"`
   - Line 212: `"failed to create file storage directory at %s: %w\n%s"`
   - Line 225: `"failed to create test file at %s: %w\n%s"`

**Design notes:**
- The diagnostic string is best-effort — if `Stat` or uid_map reading fails,
  just omit that piece. Never let diagnostic gathering cause a new error.
- The hint text about rootless Docker should only appear when
  `IsUserNamespaceActive()` returns true, to avoid confusing non-Docker users.
- Keep the hint concise (2-3 lines max). This appears in log output.

---

### Direction 3: Add User Namespace Check to `vikunja doctor`

**What:** Add a new check to the doctor System group (or Files group) that detects
and reports user namespace status, and improve the existing ownership-match check
to account for it.

**Why:** The current ownership-match check in `checkDirectoryOwnership()` compares
namespaced UIDs. In a user namespace, both the process UID and the directory's
apparent owner UID are remapped to the same namespace — so the check may **pass**
(both show 1001) even though the kernel will deny access because the *host* UIDs
don't match. The doctor output needs to surface this so users understand the
mismatch is at the host level.

**Files to modify:**

1. **`pkg/doctor/system_unix.go`** (existing, build tag `!windows`)

   Note: since user namespaces are Linux-only, we need a Linux-specific file.
   The `!windows` build tag isn't precise enough — it would also cover macOS/BSD
   where `/proc/self/uid_map` doesn't exist. Two options:

   **Option A (preferred):** Add a new `pkg/doctor/system_linux.go` with build
   tag `linux`, containing `checkUserNamespace()`. Add a stub in
   `pkg/doctor/system_notlinux.go` with build tag `!linux`.

   **Option B:** Put it in `system_unix.go` and guard the `/proc` read with
   an `os.Stat` existence check. Simpler but less clean.

   Go with Option A.

2. **New file: `pkg/doctor/system_linux.go`** (build tag: `linux`)
   - `func checkUserNamespace() CheckResult` — uses `utils.IsUserNamespaceActive()`
     and `utils.UIDMappingSummary()`:
     - If active: `Passed: true` (it's not an error per se), `Value: "active (0→1001, 1-65536→101001-166536)"`
       with a `Lines` entry explaining: `"UIDs inside this container are remapped. See directory ownership check for details."`
     - If not active: `Passed: true`, `Value: "not active"`

3. **New file: `pkg/doctor/system_notlinux.go`** (build tag: `!linux`)
   - Stub: `func checkUserNamespace() CheckResult` returns
     `Value: "not applicable (Linux only)"`

4. **`pkg/doctor/system.go`** — add `checkUserNamespace()` to the `CheckSystem()`
   results slice, after `checkUser()`.

5. **`pkg/doctor/files_unix.go`** — modify `checkDirectoryOwnership()`:
   - After the existing ownership-match check (lines 106-124), if
     `utils.IsUserNamespaceActive()` is true, add an additional result or
     modify the existing one:
     - If UIDs match (inside the namespace) but namespace is active, change the
       ownership-match result to a **warning** (Passed: true but with a Lines
       entry explaining):
       ```
       Name:  "Ownership match"
       Passed: true
       Value: "uid 1001 matches"
       Lines: ["WARNING: user namespace active — the matching uid 1001 maps to host uid 102001, which may not match the host directory owner. Consider using -u 0:0 in rootless Docker."]
       ```
     - If UIDs don't match and namespace is active, enhance the error message:
       ```
       Error: "directory owned by uid X but Vikunja runs as uid Y (user namespace active, host uid = Z)"
       ```
   - Use `utils.MapToHostUID()` to compute the host UID for display.

**Expected doctor output in a rootless Docker scenario:**

```
Vikunja Doctor
==============

System
  ✓ Version           1.0.0
  ✓ Go version        go1.25.0
  ✓ OS                linux/amd64
  ✓ User              ? (uid=1001)
  ✓ Working directory  /app/vikunja
  ✓ User namespace    active (0→1001, 1-65536→101001-166536)

Files (local)
  ✓ Path              /app/vikunja/files
  ✓ Directory exists  yes
  ✓ Dir permissions   0755
  ✓ Directory owner   1001:1001 (uid=1001, gid=1001)
  ✓ Ownership match   uid 1001 matches
      WARNING: user namespace active — uid 1001 maps to host uid 102001,
      which may differ from the actual host directory owner.
      Consider using -u 0:0 in rootless Docker.
  ✗ Writable          failed to create test file at ...: permission denied
  ✓ Disk space        50.0 GB available
```

This makes it immediately clear why the write fails despite the UIDs "matching."

---

## File Summary

| File | Action | Direction |
|------|--------|-----------|
| `pkg/utils/userns_linux.go` | **Create** | 1 |
| `pkg/utils/userns_other.go` | **Create** | 1 |
| `pkg/utils/userns_linux_test.go` | **Create** | 1 |
| `pkg/files/filehandling.go` | **Modify** — call diagnostics in error paths | 2 |
| `pkg/files/diagnostics_unix.go` | **Create** | 2 |
| `pkg/files/diagnostics_windows.go` | **Create** | 2 |
| `pkg/doctor/system.go` | **Modify** — add `checkUserNamespace()` to results | 3 |
| `pkg/doctor/system_linux.go` | **Create** | 3 |
| `pkg/doctor/system_notlinux.go` | **Create** | 3 |
| `pkg/doctor/files_unix.go` | **Modify** — enhance ownership check with userns awareness | 3 |

## Implementation Order

1. **Direction 1 first** — the `pkg/utils/userns_*.go` functions are dependencies
   for both Direction 2 and Direction 3.
2. **Direction 2 next** — enriches the fatal startup error, which is the primary
   user-facing symptom.
3. **Direction 3 last** — enhances the doctor command, which builds on all prior work.

## Testing Strategy

- **Unit tests** for `pkg/utils/userns_linux.go`: parse sample uid_map content,
  test mapping lookups. These can run without an actual user namespace by testing
  the parsing logic with string inputs.
- **Manual test**: run the built binary inside rootless Docker with a
  permission-denied scenario, verify the enriched error message includes the
  namespace mapping and host UID.
- **Doctor test**: run `vikunja doctor` inside the same rootless Docker setup,
  verify the user namespace check appears and the ownership-match warning fires.
- **Non-Linux**: verify the stubs compile and return sensible defaults on
  macOS/Windows (cross-compile check via `GOOS=windows go build ./...`).

## Out of Scope

- Dockerfile changes (switching from `scratch` to a base image with proper
  directory setup) — separate discussion.
- Automatically fixing permissions or adjusting UIDs at startup — too risky and
  opinionated.
- S3 storage — not affected by user namespaces.
