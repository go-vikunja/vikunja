# Specification 003: Universal Data Export/Import System

**Status**: Planning  
**Priority**: Critical (Migration Blocker)  
**Created**: 2025-10-18  
**Estimated Duration**: 4 weeks  

## Quick Summary

A comprehensive data export/import system enabling seamless migration from the original Vikunja instance to the refactored version, supporting different database backends and authentication methods **without requiring email authentication** for data extraction.

## Problem Statement

Need to migrate from original Vikunja (SQLite) to refactored Vikunja (PostgreSQL/MySQL) with:
- All user data preserved
- Support for OIDC/LDAP users (no local passwords)
- Direct SQLite database import (primary migration path)
- Admin-level exports (for backups and portability)
- Cross-database compatibility
- Zero data loss guarantee

**Key Simplification**: Since the original Vikunja's `vikunja.db` SQLite file can be directly accessed, we only need to implement import functionality in the refactored branch. No changes to the original codebase are required.

## Solution Overview

Implement import and export functionality in the refactored branch only:

### Primary Migration Path
1. **SQLite Database Import** (new) ⭐ **CRITICAL**
   - Direct import from `vikunja.db` file
   - No changes to original codebase needed
   - Maps old schema to new schema
   - Imports to any database backend (PostgreSQL/MySQL/SQLite)

### Export Layer (for backups and portability)
1. **User Export** (existing) - Authenticated user exports own data
2. **Admin Export** (new) - CLI-based export of all users to ZIP

### Import Layer (for portability)
1. **ZIP Archive Import** (new) - Import from exported ZIP files
2. **Full Import Mode** - Import into empty database
3. **Merge Import Mode** - Import with conflict resolution
4. **User Import Mode** - Import single user's data

### Validation Layer
1. **Pre-Import Validation** - Verify database/ZIP format
2. **Post-Import Verification** - Verify data integrity
3. **Migration Comparison** - Compare source vs target

## Key Features

✅ **Direct SQLite database import** (primary migration path)  
✅ No changes to original codebase required  
✅ Support SQLite, PostgreSQL, MySQL (SQLite → any)  
✅ Transaction-safe imports with automatic rollback  
✅ Admin export for backups and portability  
✅ Conflict resolution strategies (skip, rename, error)  
✅ ID remapping for merge scenarios  
✅ Progress reporting for long operations  
✅ Comprehensive validation and verification  
✅ CLI-first design (security)  
✅ Password-free (security - require reset)  

## Files in This Specification

- **[spec.md](./spec.md)** - Complete feature specification (main document)
- **[plan.md](./plan.md)** - Implementation plan with phases and timeline
- **[tasks.md](./tasks.md)** - Detailed task breakdown (35 tasks)
- **[research.md](./research.md)** - Technical research and analysis
- **[contracts/](./contracts/)** - Interface contracts and API specs

## Implementation Phases

### Phase 1: SQLite Database Import (Week 1) ⭐ **CRITICAL**
- SQLiteImportService implementation
- Direct SQLite database file reading
- Schema mapping and data transformation
- Files migration
- CLI command: `import-db`
- **PRIMARY MIGRATION PATH COMPLETE**

### Phase 2: Export & ZIP Import (Week 2)
- AdminExportService implementation
- ZipImportService implementation
- Manifest generation
- Transaction management
- CLI commands: `export`, `import`

### Phase 3: Merge & User Import (Week 3)
- Conflict detection and resolution
- Merge import mode
- User import mode
- Validation service

### Phase 4: Testing & Documentation (Week 4)
- Comprehensive testing (unit, integration, performance)
- Cross-database testing
- Migration guide documentation
- CLI reference documentation

## Export Format

```
export-{timestamp}.zip
├── manifest.json              # Metadata, counts, version
├── users.json                 # All users (no passwords)
├── teams.json                 # All teams
├── team_members.json          # Team memberships
├── labels.json                # Global labels
├── user_data/
│   ├── user_{id}/
│   │   ├── projects.json      # Projects, tasks, views
│   │   ├── filters.json       # Saved filters
│   │   ├── attachments/       # Task files
│   │   └── backgrounds/       # Project backgrounds
└── files_metadata.json        # File records
```

## CLI Commands

```bash
# PRIMARY MIGRATION PATH: Import SQLite database directly ⭐
./vikunja import-db --sqlite-file=/path/to/vikunja.db --files-dir=/path/to/files

# Export all users (for backups and portability)
./vikunja export --output=/path/to/export.zip --all

# Import from ZIP archive
./vikunja import --file=/path/to/export.zip --mode=full

# Import with merge and conflict resolution
./vikunja import --file=/path/to/export.zip --mode=merge --conflicts=rename

# Validate before import
./vikunja validate-export --file=/path/to/export.zip
./vikunja validate-db --sqlite-file=/path/to/vikunja.db

# Verify migration after import
./vikunja verify-migration --sqlite-file=/path/to/vikunja.db
# or
./vikunja verify-migration --export=/path/to/export.zip
```

## Success Criteria

### Functional
- ✅ Export all users without authentication
- ✅ Import into any database backend
- ✅ Zero data loss during migration
- ✅ All relationships preserved
- ✅ All files transferred

### Non-Functional
- ✅ Export 1000 users in < 5 minutes
- ✅ Import 1000 users in < 15 minutes
- ✅ Memory usage < 500 MB
- ✅ Transaction safety with rollback
- ✅ >80% test coverage

## Dependencies

### Internal
- ✅ Spec 001 (Service Layer Refactor) - **COMPLETE**
- ✅ Existing user export service - **COMPLETE**

### External
- XORM (database abstraction) - already in use
- Go stdlib (archive/zip, encoding/json)
- Database drivers (already in use)

## Risk Assessment

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Data Loss | High | Low | Transaction rollback, extensive testing |
| Auth Issues Post-Import | High | Medium | Password reset flow, OIDC testing |
| Performance Issues | Medium | Medium | Stream processing, benchmarking |
| Version Incompatibility | Medium | Low | Version checking, format versioning |

## Migration Example

**Scenario**: Migrate from SQLite (original) to PostgreSQL (refactored)

```bash
# 1. On original server: Stop Vikunja and copy database file
systemctl stop vikunja
cp /var/lib/vikunja/vikunja.db /backup/vikunja.db
cp -r /var/lib/vikunja/files /backup/vikunja_files

# 2. Transfer files to new server
scp /backup/vikunja.db newserver:/tmp/
scp -r /backup/vikunja_files newserver:/tmp/

# 3. On new server: Setup PostgreSQL
createdb vikunja_production
./vikunja migrate

# 4. Test import (dry-run)
./vikunja import-db --sqlite-file=/tmp/vikunja.db \
  --files-dir=/tmp/vikunja_files --dry-run

# 5. Actual import ⭐ PRIMARY MIGRATION PATH
./vikunja import-db --sqlite-file=/tmp/vikunja.db \
  --files-dir=/tmp/vikunja_files --target-database=postgres

# 6. Verify migration
./vikunja verify-migration --sqlite-file=/tmp/vikunja.db

# 7. Start new instance
systemctl start vikunja
```

## Timeline

**Start Date**: TBD  
**Estimated Completion**: 4 weeks from start  

```
Week 1: Export Enhancements
Week 2: Import Implementation
Week 3: Merge & User Import
Week 4: Testing & Documentation
```

## Team

**Required**:
- 1 Senior Backend Developer (full-time, 4 weeks)
- 1 QA Engineer (part-time, week 4)

**Estimated Effort**: ~200 hours total

## Next Steps

1. ✅ Specification complete
2. [ ] Review and approval
3. [ ] Create development branch
4. [ ] Setup test infrastructure
5. [ ] Begin Phase 1 implementation

## Questions?

See the detailed documents:
- Technical details → [spec.md](./spec.md)
- Implementation plan → [plan.md](./plan.md)
- Task breakdown → [tasks.md](./tasks.md)
- Research findings → [research.md](./research.md)
- API contracts → [contracts/](./contracts/)

---

**Specification Version**: 1.0.0  
**Last Updated**: 2025-10-18  
**Status**: Draft (awaiting approval)
