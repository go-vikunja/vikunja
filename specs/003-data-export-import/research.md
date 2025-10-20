# Research: Universal Data Export/Import System

**Feature**: 003-data-export-import  
**Research Date**: 2025-10-18  

## Overview

This document contains research findings related to implementing a comprehensive data export/import system for Vikunja, focusing on migration between database backends and handling various authentication scenarios.

---

## Existing Export/Import Solutions

### Vikunja Current Implementation

**User Export** (Already Implemented):
- Location: `pkg/services/user_export.go`
- Exports user's own data only
- Requires authentication (password or OAuth)
- Format: ZIP archive with JSON data files
- Includes:
  - Projects, tasks, views, buckets
  - Task comments
  - Task attachments (binary files)
  - Saved filters
  - Project backgrounds
  - Version information

**Limitations**:
- No import functionality
- Cannot export other users' data
- Requires email/password authentication
- No admin/bulk export option

### Similar Projects Analysis

#### 1. GitLab Export/Import

**Features**:
- Project-level exports
- Full instance exports (admin only)
- Import with conflict resolution
- Preserves relationships and metadata

**Format**:
- Tarball archives
- JSON for data
- Separate directories for files/assets
- Manifest file with version info

**Lessons**:
- ✅ Separate concerns: data vs files
- ✅ Manifest file is critical for validation
- ✅ Transaction-based imports
- ✅ Progress reporting for UX

#### 2. Nextcloud Data Export

**Features**:
- User data export via admin interface
- Bulk export tools
- Import via CLI
- Cross-instance migration support

**Format**:
- ZIP archives
- XML for configuration
- Original file structures preserved

**Lessons**:
- ✅ CLI-first for admin operations
- ✅ Dry-run mode is essential
- ✅ File permissions matter
- ✅ Checksums for integrity verification

#### 3. WordPress Export/Import

**Features**:
- WXR (WordPress eXtended RSS) format
- Media library export
- Plugin for imports
- Content mapping UI

**Format**:
- XML for content
- Separate media archives
- Metadata preservation

**Lessons**:
- ✅ User-friendly import mapping
- ✅ Preview before import
- ✅ Support partial imports
- ❌ XML is verbose (JSON preferred)

#### 4. PostgreSQL Logical Backup

**Features**:
- `pg_dump` for exports
- `pg_restore` for imports
- Schema + data separation
- Transaction logs for point-in-time recovery

**Format**:
- Custom binary format
- Plain SQL text format
- Directory format for parallel ops

**Lessons**:
- ✅ Schema version compatibility checks
- ✅ Parallel operations for performance
- ✅ Transaction safety
- ✅ Incremental backups

---

## Technical Challenges & Solutions

### Challenge 1: Cross-Database Migration

**Problem**: Migrating from SQLite to PostgreSQL/MySQL with different feature sets.

**SQLite → PostgreSQL Issues**:
- Auto-increment sequences
- Date/time handling
- Boolean vs integer
- Transaction isolation levels
- BLOB storage

**Solution**:
- Use XORM's database abstraction
- Normalize data during export (canonical JSON format)
- Use service layer (already database-agnostic)
- Test with all database engines
- Handle type conversions during import

**Reference**:
- XORM documentation: https://xorm.io/
- Database-specific considerations documented

### Challenge 2: ID Remapping & Foreign Keys

**Problem**: Merging data from multiple instances requires ID conflict resolution.

**Scenarios**:
- User ID 1 exists in both instances
- Task ID 500 exists in both instances
- Foreign keys reference old IDs

**Solution Approach**:
```go
type IDMapper struct {
    userMap    map[int64]int64  // old -> new
    projectMap map[int64]int64
    taskMap    map[int64]int64
    // ... other entities
}

func (m *IDMapper) RemapTask(task *Task) {
    task.ID = m.taskMap[task.ID]
    task.ProjectID = m.projectMap[task.ProjectID]
    task.CreatedByID = m.userMap[task.CreatedByID]
    // ... remap all foreign keys
}
```

**Implementation Strategy**:
1. Import users first (generate new IDs)
2. Build ID maps as entities are created
3. Update foreign keys in second pass
4. Use SQL UPDATE statements for efficiency

**Performance Consideration**:
- Use maps (O(1) lookup) not linear scans
- Batch updates where possible
- Consider database constraints order

### Challenge 3: Authentication Without Email

**Problem**: OIDC/LDAP users have no local password, but export requires auth.

**Current Export Flow**:
```
User → API Request → Password Check → Export
```

**Solutions**:

**Option A: Admin Export (Chosen)**
```
Admin → CLI Command → Direct DB/Service Access → Export
```
- Pros: No authentication needed, can export all users
- Cons: Requires server access
- Use case: Full instance migration

**Option B: Database Direct Export**
```
Admin → CLI Command → Direct DB Queries → Export
```
- Pros: Bypasses all authentication, works offline
- Cons: Bypasses service layer (less safe)
- Use case: Emergency exports, corrupted instances

**Option C: Temporary Token Generation**
```
Admin → Generate Token → Give to User → User Uses Token → Export
```
- Pros: Secure, user-controlled
- Cons: Complex workflow, token management
- Use case: Individual user exports

**Decision**: Implement Options A & B for maximum flexibility.

### Challenge 4: Transaction Management

**Problem**: Import must be atomic (all-or-nothing).

**Requirements**:
- Multi-table inserts
- Foreign key constraints
- File operations (not transactional)
- Performance (avoid holding transaction too long)

**Solution Design**:

```go
func (is *ImportService) Import(exportFile string, mode ImportMode) error {
    // Phase 1: Validation (no transaction)
    if err := is.validateExport(exportFile); err != nil {
        return err
    }
    
    // Phase 2: Database import (transaction)
    sess := is.DB.NewSession()
    defer sess.Close()
    
    if err := sess.Begin(); err != nil {
        return err
    }
    
    // Import data in dependency order
    if err := is.importUsers(sess, data); err != nil {
        sess.Rollback()
        return err
    }
    
    if err := is.importProjects(sess, data); err != nil {
        sess.Rollback()
        return err
    }
    
    // ... more imports
    
    // Commit transaction
    if err := sess.Commit(); err != nil {
        sess.Rollback()
        return err
    }
    
    // Phase 3: File operations (post-commit)
    // If this fails, database is still consistent
    // User can retry file import separately
    if err := is.importFiles(data); err != nil {
        log.Warn("Files failed to import, but data is imported")
        return err
    }
    
    return nil
}
```

**Key Decisions**:
- Import files AFTER database transaction
- If files fail, database is still consistent
- Provide separate file retry mechanism
- Log all operations for debugging

### Challenge 5: Conflict Resolution Strategies

**Problem**: Merging data from multiple instances creates conflicts.

**Conflict Types**:

1. **Email Conflicts** (User with same email exists)
   - Strategy: Skip, Rename (email+suffix), Error
   - Default: Error (safety first)

2. **ID Conflicts** (Entity ID already used)
   - Strategy: Remap (always)
   - Default: Remap all IDs

3. **Name Conflicts** (Team/Project with same name)
   - Strategy: Skip, Rename (name+suffix), Merge, Error
   - Default: Rename

4. **Identifier Conflicts** (Project identifier already used)
   - Strategy: Rename (append number), Error
   - Default: Rename

**Implementation**:

```go
type ConflictStrategy string

const (
    ConflictSkip   ConflictStrategy = "skip"
    ConflictRename ConflictStrategy = "rename"
    ConflictMerge  ConflictStrategy = "merge"
    ConflictError  ConflictStrategy = "error"
)

type ConflictResolver struct {
    strategy ConflictStrategy
    log      *ConflictLog
}

func (cr *ConflictResolver) ResolveUserConflict(
    existingUser *User, 
    importUser *User,
) (*User, error) {
    switch cr.strategy {
    case ConflictSkip:
        cr.log.Add("Skipped user: %s", importUser.Email)
        return existingUser, nil
    case ConflictRename:
        importUser.Email = fmt.Sprintf("%s+import", importUser.Email)
        cr.log.Add("Renamed user email: %s", importUser.Email)
        return importUser, nil
    case ConflictError:
        return nil, ErrUserConflict{Email: importUser.Email}
    }
}
```

### Challenge 6: Large File Handling

**Problem**: Instances with many/large attachments can create huge export files.

**Considerations**:
- Export file size (can be GBs)
- Memory usage during export/import
- Disk I/O performance
- Network transfer time

**Solutions**:

**Streaming ZIP Creation**:
```go
// Don't load entire file in memory
func writeFileToZip(file *os.File, zipWriter *zip.Writer) error {
    writer, err := zipWriter.Create(filename)
    if err != nil {
        return err
    }
    
    // Stream copy (buffered)
    _, err = io.Copy(writer, file)
    return err
}
```

**Chunked Import**:
```go
// Import files in batches
func (is *ImportService) importFiles(files []FileRecord) error {
    const batchSize = 100
    
    for i := 0; i < len(files); i += batchSize {
        end := i + batchSize
        if end > len(files) {
            end = len(files)
        }
        
        batch := files[i:end]
        if err := is.importFileBatch(batch); err != nil {
            return err
        }
    }
    return nil
}
```

**Progress Reporting**:
```go
type ImportProgress struct {
    Stage       string  // "users", "projects", "files"
    Current     int
    Total       int
    Percentage  float64
}

func (is *ImportService) reportProgress(stage string, current, total int) {
    progress := ImportProgress{
        Stage:      stage,
        Current:    current,
        Total:      total,
        Percentage: float64(current) / float64(total) * 100,
    }
    
    fmt.Printf("\r[%s] %d/%d (%.1f%%)", 
        progress.Stage, progress.Current, progress.Total, progress.Percentage)
}
```

---

## Security Considerations

### Password Handling

**Problem**: Should we export password hashes?

**Analysis**:

**Arguments FOR exporting hashes**:
- Users can login immediately after import
- No password reset workflow needed
- Seamless migration experience

**Arguments AGAINST exporting hashes** (Chosen):
- Security risk if export file is compromised
- Passwords may be weak (breach risk)
- Encourages password resets (better security)
- GDPR/compliance concerns

**Decision**: DO NOT export password hashes by default.

**Implementation**:
- Set all imported users to `password_reset_required = true`
- Generate random temporary passwords (or null)
- Send password reset emails post-import
- Document this behavior clearly

### OAuth Token Handling

**Problem**: Should we export OAuth tokens/refresh tokens?

**Decision**: NEVER export OAuth tokens.

**Reasoning**:
- Security risk (tokens grant account access)
- Tokens are instance-specific
- Users can re-authenticate with OAuth provider
- Compliance issues (tokens are sensitive data)

**Implementation**:
- Skip OAuth token fields during export
- Document that OAuth users need to re-authenticate
- Provide clear instructions for OAuth re-setup

### Access Control for Export/Import

**Export Operations**:
- User export: Requires user authentication (existing)
- Admin export: Requires CLI access (no API)
- DB export: Requires server file system access

**Import Operations**:
- All imports: CLI only (no API endpoint)
- Requires server file system access
- Requires database access
- Log all operations

**Reasoning**:
- CLI access implies trusted administrator
- No risk of web-based attacks
- Audit trail through server logs
- Standard practice for administrative tools

### Data Sanitization Options

**Levels of Export**:

1. **Full Export** (default for admin)
   - All data except passwords/tokens
   - Real emails (for user identification)
   - All file attachments
   - All relationships

2. **User Export** (existing)
   - Single user's data
   - No other users' information
   - Requires authentication

3. **Anonymized Export** (future enhancement)
   - Hashed emails
   - Fake usernames
   - Sample data only
   - For development/testing

**Implementation**:
```go
type ExportOptions struct {
    IncludePasswords bool // default: false
    IncludeTokens    bool // default: false
    Anonymize        bool // default: false
    SanitizeEmails   bool // default: false
}
```

---

## Performance Benchmarks

### Expected Performance Targets

Based on research of similar systems:

**Export Performance**:
- 1,000 users, 10,000 tasks: ~2-5 minutes
- 10,000 users, 100,000 tasks: ~20-30 minutes
- Memory usage: < 500 MB
- Bottleneck: File I/O (attachments)

**Import Performance**:
- 1,000 users, 10,000 tasks: ~5-15 minutes
- 10,000 users, 100,000 tasks: ~45-90 minutes
- Memory usage: < 500 MB
- Bottleneck: Database inserts, foreign key constraints

**Optimization Strategies**:
1. Batch inserts (100-1000 records per transaction)
2. Disable foreign key checks during import (re-enable after)
3. Stream file operations (don't load in memory)
4. Parallel processing where possible (Go goroutines)
5. Use prepared statements for repeated queries

### Test Data Sets

**Small**: 10 users, 50 projects, 500 tasks, 50 files
**Medium**: 100 users, 500 projects, 5,000 tasks, 500 files
**Large**: 1,000 users, 5,000 projects, 50,000 tasks, 5,000 files
**Huge**: 10,000 users, 50,000 projects, 500,000 tasks, 50,000 files

---

## Alternative Approaches Considered

### Alternative 1: Database-Level Replication

**Approach**: Use PostgreSQL logical replication or MySQL binlog replication.

**Pros**:
- Native database feature
- Very fast
- Real-time sync possible
- No custom code needed

**Cons**:
- Requires same database engine (can't migrate SQLite → PostgreSQL)
- Complex setup
- Network requirements
- Not suitable for one-time migrations
- Doesn't solve OIDC/LDAP export issue

**Decision**: Rejected (doesn't meet requirements)

### Alternative 2: SQL Dump/Restore

**Approach**: Use `sqlite3 .dump` or `pg_dump` for exports.

**Pros**:
- Standard tools
- Fast
- Reliable
- Well-documented

**Cons**:
- Doesn't support cross-database migrations
- Includes password hashes (security risk)
- No data sanitization
- No conflict resolution
- Requires database admin access

**Decision**: Rejected (doesn't meet requirements)

### Alternative 3: ETL Tool (e.g., Apache NiFi, Talend)

**Approach**: Use existing ETL tools for data migration.

**Pros**:
- Powerful transformation capabilities
- Visual workflow designer
- Many connectors available
- Production-proven

**Cons**:
- Adds external dependency
- Overkill for our use case
- Requires learning new tool
- Not integrated with Vikunja
- Harder for users to use

**Decision**: Rejected (too complex)

### Alternative 4: REST API-Based Migration

**Approach**: Export via API calls, import via API calls.

**Pros**:
- Uses existing API
- Can be done remotely
- Language-agnostic client

**Cons**:
- Very slow (one request per entity)
- Rate limiting issues
- Authentication required
- No transaction support
- Complex client implementation

**Decision**: Rejected (too slow, unreliable)

---

## Database Schema Considerations

### Current Schema (Relevant Tables)

```sql
-- Users
users (id, username, email, password, created, updated, ...)

-- Teams
teams (id, name, description, created, updated)
team_members (id, team_id, user_id, admin, created)

-- Projects
projects (id, title, description, owner_user_id, created, updated, ...)
project_users (id, project_id, user_id, right, created, updated)
project_teams (id, project_id, team_id, right, created, updated)

-- Tasks
tasks (id, title, description, project_id, created_by_id, created, updated, ...)
task_assignees (task_id, user_id)
task_comments (id, task_id, author_id, comment, created, updated)
task_attachments (id, task_id, file_id, created, updated)

-- Labels
labels (id, title, hex_color, created_by_id, created, updated)
label_tasks (label_id, task_id)

-- Files
files (id, name, mime, size, created, ...)

-- Views & Buckets
project_views (id, project_id, title, view_kind, ...)
buckets (id, project_view_id, title, position, ...)
task_buckets (bucket_id, task_id, project_view_id)
task_positions (task_id, project_view_id, position)

-- Filters
saved_filters (id, user_id, title, filters, created, updated)
```

### Import Order (Dependency Resolution)

Must import in this order to satisfy foreign key constraints:

1. **Users** (no dependencies)
2. **Teams** (no dependencies)
3. **Team Members** (depends on: users, teams)
4. **Labels** (depends on: users - created_by)
5. **Projects** (depends on: users - owner)
6. **Project Users** (depends on: projects, users)
7. **Project Teams** (depends on: projects, teams)
8. **Project Views** (depends on: projects)
9. **Buckets** (depends on: project_views)
10. **Tasks** (depends on: projects, users - created_by)
11. **Task Assignees** (depends on: tasks, users)
12. **Task Comments** (depends on: tasks, users - author)
13. **Label Tasks** (depends on: labels, tasks)
14. **Files** (depends on: users - uploader)
15. **Task Attachments** (depends on: tasks, files)
16. **Task Buckets** (depends on: buckets, tasks, project_views)
17. **Task Positions** (depends on: tasks, project_views)
18. **Saved Filters** (depends on: users)

### Circular Dependencies

**None identified** in current schema. This is good for import logic.

If circular dependencies are added in future:
- Import with NULL foreign keys first
- Update foreign keys in second pass
- Use deferred constraint checking (PostgreSQL)

---

## Testing Strategy Research

### Test Pyramid for Data Migration

```
         /\
        /  \
       / UI \
      /______\
     /  API   \
    /__________\
   / Integration\
  /______________\
 /   Unit Tests   \
/_________________ \
```

**Unit Tests** (80%):
- Export service methods
- Import service methods
- Validation logic
- Conflict resolution
- ID remapping
- Data transformations

**Integration Tests** (15%):
- Full export → import workflow
- Cross-database migrations
- Large dataset handling
- Error recovery
- Transaction rollback

**API Tests** (5%):
- User export endpoint (existing)
- Download export endpoint

**UI Tests** (0%):
- None needed (CLI-only feature)

### Test Data Generation

Use existing fixtures + generate larger datasets:

```go
func generateTestData(userCount, projectCount, taskCount int) *TestDataSet {
    // Generate users
    users := make([]*User, userCount)
    for i := 0; i < userCount; i++ {
        users[i] = &User{
            Username: fmt.Sprintf("user%d", i),
            Email:    fmt.Sprintf("user%d@test.com", i),
            // ...
        }
    }
    
    // Generate projects per user
    // Generate tasks per project
    // Generate relationships
    
    return &TestDataSet{
        Users:    users,
        Projects: projects,
        Tasks:    tasks,
    }
}
```

### Performance Testing

**Tools**:
- Go built-in benchmarks
- `pprof` for profiling
- Database query logging
- Custom metrics collection

**Metrics to Track**:
- Export time per user
- Import time per entity type
- Memory usage (heap)
- Disk I/O throughput
- Database connection count
- Transaction duration

---

## Documentation Research

### Best Practices for Migration Docs

**Audience**: System administrators, DevOps engineers

**Key Sections** (from industry research):
1. **Prerequisites** - What you need before starting
2. **Preparation** - Backup procedures, test plans
3. **Step-by-Step Guide** - Detailed instructions
4. **Verification** - How to verify success
5. **Troubleshooting** - Common issues and solutions
6. **Rollback** - How to recover from failures
7. **FAQ** - Common questions

**Best Practices**:
- Include exact commands (copy-pasteable)
- Use screenshots sparingly (they go out of date)
- Provide multiple examples (small vs large instance)
- Document time estimates
- Include checkpoint verification steps
- Provide emergency contact/support info

### CLI Help Text Standards

**Format**:
```
vikunja export - Export Vikunja data

USAGE:
    vikunja export [OPTIONS]

OPTIONS:
    --output FILE     Path to output ZIP file (required)
    --all             Export all users (default: false)
    --users IDS       Comma-separated user IDs to export
    --config FILE     Config file path (default: ./config.yml)
    --help, -h        Show help

EXAMPLES:
    # Export all users
    vikunja export --output=/backup/export.zip --all

    # Export specific users
    vikunja export --output=/backup/users.zip --users=1,2,3

NOTES:
    - Requires file system access to Vikunja installation
    - Export files can be large (check disk space)
    - Passwords are NOT included in exports
```

---

## Conclusion

Based on this research, the proposed implementation is:

✅ **Feasible** - Similar systems exist and work well  
✅ **Secure** - Following best practices for sensitive data  
✅ **Performant** - Achievable performance targets  
✅ **Maintainable** - Clean architecture, well-tested  
✅ **User-Friendly** - Clear documentation, good UX  

**Next Steps**:
1. Review and approve specification
2. Begin Phase 1 implementation
3. Setup test infrastructure
4. Create development branch

---

## References

### External Resources
- PostgreSQL Dump/Restore: https://www.postgresql.org/docs/current/backup-dump.html
- XORM Documentation: https://xorm.io/
- Go Archive/Zip: https://pkg.go.dev/archive/zip
- Database Migration Best Practices: Various industry blogs

### Internal Resources
- Spec 001: Complete Service-Layer Refactor
- Existing user export service: `pkg/services/user_export.go`
- Vikunja database schema: `pkg/migration/*.go`

### Tools & Libraries
- XORM: Database abstraction (already in use)
- Go stdlib `archive/zip`: ZIP file handling
- Go stdlib `encoding/json`: JSON serialization
- Existing CLI framework: Already in use for other commands
