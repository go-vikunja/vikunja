# Migration Strategy: SQLite Database Import

**Key Decision**: Implement import-only in refactored branch, no changes to original codebase.

## Why This Approach?

### Original Plan (Discarded)
- Implement export in original codebase
- Implement import in refactored codebase
- Complex: Changes to two codebases
- Risky: Modifying production code

### New Plan (Simplified) ⭐
- **Access `vikunja.db` SQLite file directly**
- Implement SQLite database import in refactored branch only
- No changes to original codebase needed
- Simpler, safer, faster

## Primary Migration Path

```
┌─────────────────────────┐
│  Original Vikunja       │
│  (Production Server)    │
│                         │
│  /var/lib/vikunja/      │
│    ├── vikunja.db      │ ← COPY THIS
│    └── files/         │ ← COPY THIS
└─────────────────────────┘
          │
          │ scp/rsync
          ▼
┌─────────────────────────┐
│  New Server             │
│                         │
│  /tmp/import/           │
│    ├── vikunja.db       │
│    └── files/           │
└─────────────────────────┘
          │
          │ ./vikunja import-db
          ▼
┌─────────────────────────┐
│  Refactored Vikunja     │
│  (PostgreSQL/MySQL)     │
│                         │
│  All data migrated ✓    │
└─────────────────────────┘
```

## Implementation Focus

### Phase 1: SQLite Import (Week 1) ⭐ **CRITICAL**

**File**: `pkg/services/sqlite_import.go`

```go
type SQLiteImportService struct {
    DB *xorm.Engine // Target database (PostgreSQL/MySQL)
}

func (s *SQLiteImportService) ImportFromSQLite(
    sqliteFile string,
    filesDir string,
) (*ImportReport, error) {
    // 1. Open SQLite database file
    sqliteDB, err := sql.Open("sqlite3", sqliteFile)
    
    // 2. Read data from SQLite
    users := s.readUsers(sqliteDB)
    projects := s.readProjects(sqliteDB)
    tasks := s.readTasks(sqliteDB)
    // ... etc
    
    // 3. Transform data (handle schema differences)
    transformedUsers := s.transformUsers(users)
    
    // 4. Begin transaction on target database
    sess := s.DB.NewSession()
    sess.Begin()
    
    // 5. Insert into target database
    err = s.insertUsers(sess, transformedUsers)
    err = s.insertProjects(sess, projects)
    // ... etc
    
    // 6. Copy files to new location
    err = s.migrateFiles(filesDir)
    
    // 7. Commit transaction
    sess.Commit()
    
    return &ImportReport{...}, nil
}
```

**CLI Command**:
```bash
./vikunja import-db --sqlite-file=/path/to/vikunja.db \
                    --files-dir=/path/to/files \
                    --target-database=postgres
```

### Phase 2: Export & ZIP Import (Week 2)

For backups and data portability:
- Admin export: `./vikunja export --output=backup.zip --all`
- ZIP import: `./vikunja import --file=backup.zip --mode=full`

### Phase 3 & 4: Advanced Features

Merge mode, user import, validation, documentation.

## Why This is Better

### Advantages

1. **No Original Codebase Changes**
   - Don't touch production code
   - No risk of breaking original instance
   - No need to deploy changes to original server

2. **Simpler Architecture**
   - One-way migration (import only)
   - Direct file access (fastest method)
   - No network transfer during export
   - No authentication complexity

3. **Faster Implementation**
   - Only one codebase to modify
   - No export format negotiation
   - Direct database access

4. **Safer Migration**
   - Original database file is read-only
   - No risk to source data
   - Can retry unlimited times
   - Original instance can stay running

5. **Better Performance**
   - Direct SQLite reading (no serialization)
   - Single transaction on target
   - Streaming file copies

### Disadvantages (Mitigated)

1. **Requires File System Access**
   - Mitigation: Standard for admin operations
   - Mitigation: Same requirement for backups
   - Not a significant limitation

2. **SQLite-Specific Code**
   - Mitigation: Only for import service
   - Mitigation: Well-isolated module
   - Worth it for primary migration path

## Schema Mapping

### Handling Schema Differences

If original and refactored schemas differ:

```go
type SchemaMapper struct {
    version string
}

func (sm *SchemaMapper) MapUser(oldUser *SQLiteUser) (*User, error) {
    // Handle field renames
    // Handle new fields (defaults)
    // Handle removed fields (skip)
    return &User{
        ID:       oldUser.ID,
        Username: oldUser.Username,
        Email:    oldUser.Email,
        // New field in refactored version
        EmailVerified: false, // Default value
    }, nil
}
```

### Version Detection

```go
func (s *SQLiteImportService) detectVersion(db *sql.DB) (string, error) {
    // Query version from SQLite metadata
    var version string
    err := db.QueryRow("SELECT value FROM settings WHERE key = 'version'").Scan(&version)
    
    return version, err
}
```

## Error Handling

### Transaction Rollback

```go
// All imports use transactions
sess := s.DB.NewSession()
defer sess.Close()

if err := sess.Begin(); err != nil {
    return err
}

// Import data...
if err != nil {
    sess.Rollback()
    return err
}

// Success
if err := sess.Commit(); err != nil {
    sess.Rollback()
    return err
}
```

### File Migration Errors

```go
// Files are copied AFTER database transaction
// If file copy fails, database is still consistent
if err := s.migrateFiles(filesDir); err != nil {
    log.Warn("Files failed to migrate, but database is imported")
    // User can manually copy files or retry
    return ImportReport{
        Success: true,
        DatabaseImported: true,
        FilesMigrated: false,
        FilesError: err,
    }
}
```

## Testing Strategy

### Unit Tests

```go
func TestSQLiteImport_BasicMigration(t *testing.T) {
    // Create test SQLite database
    sqliteDB := createTestSQLite(t)
    defer sqliteDB.Close()
    
    // Insert test data
    insertTestData(sqliteDB)
    
    // Setup target PostgreSQL
    targetDB := setupTestPostgres(t)
    defer targetDB.Close()
    
    // Import
    service := NewSQLiteImportService(targetDB)
    report, err := service.ImportFromSQLite(sqliteDB.Path, testFilesDir)
    
    require.NoError(t, err)
    assert.True(t, report.Success)
    
    // Verify data
    verifyUsers(t, targetDB, expectedUsers)
    verifyProjects(t, targetDB, expectedProjects)
}
```

### Integration Tests

```go
func TestSQLiteImport_RealWorldData(t *testing.T) {
    // Use fixture: copy of real vikunja.db
    sqliteFile := "fixtures/vikunja_production_sample.db"
    
    // Import to PostgreSQL
    targetDB := setupTestPostgres(t)
    service := NewSQLiteImportService(targetDB)
    report, err := service.ImportFromSQLite(sqliteFile, testFilesDir)
    
    require.NoError(t, err)
    
    // Verify counts
    assert.Equal(t, 50, report.Counts.Users)
    assert.Equal(t, 234, report.Counts.Projects)
    assert.Equal(t, 1523, report.Counts.Tasks)
}
```

## Production Deployment

### Pre-Migration Checklist

- [ ] Backup original vikunja.db
- [ ] Test import on staging with real data
- [ ] Verify all functionality after test import
- [ ] Document rollback procedure
- [ ] Schedule maintenance window

### Migration Day

1. **Stop original instance** (ensure consistent backup)
2. **Copy database file** (vikunja.db)
3. **Copy files directory** (attachments, backgrounds)
4. **Transfer to new server**
5. **Run import with dry-run** (validate)
6. **Run actual import**
7. **Verify migration** (counts, integrity)
8. **Start refactored instance**
9. **Test functionality** (logins, data access)
10. **Update DNS/proxy** (switch traffic)

### Rollback Procedure

If migration fails:
1. **Rollback is automatic** (transaction-based)
2. **Database unchanged** (transaction rolled back)
3. **Restart original instance** (still has all data)
4. **Review logs** (identify issue)
5. **Fix and retry**

## Timeline

**Phase 1 (Week 1)**: SQLite Import - **HIGHEST PRIORITY**
- Implement SQLiteImportService
- Implement import-db CLI command
- Test with real data
- **Result: Can migrate production**

**Phase 2 (Week 2)**: Export & ZIP Import
- Implement AdminExportService
- Implement ZipImportService
- **Result: Can do backups and portability**

**Phases 3-4 (Weeks 3-4)**: Advanced features and documentation

## Success Metrics

✅ **Week 1 Complete**: Can migrate production from SQLite to PostgreSQL  
✅ **Week 2 Complete**: Can create portable backups  
✅ **Week 3 Complete**: Can merge data from multiple sources  
✅ **Week 4 Complete**: Full documentation and testing  

---

**Key Takeaway**: By leveraging direct SQLite file access, we've simplified the migration from a two-codebase problem to a one-codebase solution, making it faster, safer, and more reliable.
