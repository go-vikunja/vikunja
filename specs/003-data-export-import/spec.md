# Feature Specification: Universal Data Export/Import System

**Feature ID**: 003  
**Created**: 2025-10-18  
**Status**: Planning  
**Priority**: Critical (Migration Blocker)  

## Executive Summary

Design and implement a comprehensive data export/import system to enable seamless migration from the original Vikunja instance (sqlite-based) to the refactored version (supporting PostgreSQL/MySQL). The system must support complete data portability **without requiring email authentication** for data extraction.

**Key Requirements**:
- Export all user data from original instance (any database)
- Import into refactored instance (any database)
- No email requirement for data export
- Support for both authenticated and CLI-based operations
- Database-agnostic design
- Zero data loss guarantee

## Context & Problem Statement

### Current State

**Original Vikunja** (`vikunja_original_main`):
- Running on production server
- SQLite database with live user data
- Existing user export functionality (requires authentication)
- User-scoped exports only (per-user data)

**Refactored Vikunja** (`vikunja`):
- Complete service-layer refactor (spec 001 completed)
- Will run on PostgreSQL or MySQL (not SQLite)
- Existing user export service (already refactored)
- No import functionality yet

### The Migration Challenge

1. **Server Replacement**: Need to replace original instance with refactored version
2. **Database Change**: Moving from SQLite to PostgreSQL/MySQL
3. **Data Preservation**: All user data must be migrated without loss
4. **Email Problem**: Original export requires email/password, but:
   - Some users may have OIDC/LDAP authentication (no password)
   - Admin needs to export all data without per-user authentication
   - Bulk export needed for efficiency

### Current Export Capabilities

The existing user export system (refactored in spec 001-T023) exports:
- ✅ Projects and all sub-entities (tasks, views, buckets)
- ✅ Task comments
- ✅ Task attachments (files)
- ✅ Task positions (kanban/board positions)
- ✅ Saved filters
- ✅ Project background images
- ✅ Version information

**What's Missing**:
- ❌ Import functionality (critical for migration)
- ❌ SQLite database import (critical for migration from original)
- ❌ Admin/bulk export (all users at once)
- ❌ Cross-database import support
- ❌ Migration validation tools

## Proposed Solution

### Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│               Migration Source                               │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Original Vikunja SQLite DB (PRIMARY PATH)               │
│     - vikunja.db file from production                        │
│     - Direct file access                                     │
│     - No code changes to original needed                     │
│                                                              │
│  2. Exported ZIP Archive (BACKUP/PORTABILITY)               │
│     - From refactored instance's export                      │
│     - Standard format for data portability                   │
│                                                              │
└────────────────────┬────────────────────────────────────────┘
                     │
                     ▼
┌─────────────────────────────────────────────────────────────┐
│            Refactored Vikunja (Import Side)                 │
│           (PostgreSQL / MySQL / SQLite)                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. SQLite Database Import (NEW) ⭐ PRIMARY                 │
│     - CLI command: import-db                                 │
│     - Direct SQLite database read                            │
│     - Transform and migrate to target database               │
│     - Handle schema differences                              │
│                                                              │
│  2. ZIP Archive Import (NEW)                                │
│     - CLI command: import                                    │
│     - Full/merge/user modes                                  │
│     - Conflict resolution                                    │
│                                                              │
│  3. Admin Export (NEW)                                      │
│     - CLI command: export                                    │
│     - All users to ZIP archive                               │
│     - For backups and portability                            │
│                                                              │
│  4. User Export (EXISTING)                                  │
│     - API endpoint (authenticated)                           │
│     - Single user data                                       │
│                                                              │
│  5. Validation & Verification (NEW)                         │
│     - Pre-import validation                                  │
│     - Post-import verification                               │
│     - Migration report generation                            │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

### Components to Implement

#### 1. SQLite Database Import Service (NEW) ⭐ **CRITICAL FOR MIGRATION**

**Location**: `pkg/services/sqlite_import.go`

**Purpose**: Import data directly from original Vikunja's SQLite database file

**Capabilities**:
- Read SQLite database file directly
- Map old schema to new schema (if differences exist)
- Transform data to target database format
- Handle schema version differences
- Import all users and their data
- Import system-wide entities (labels, teams, team memberships)
- Import all files (attachments, backgrounds, avatars)
- Progress reporting for large databases
- Transaction safety with rollback

**Access Control**:
- CLI command only (direct server access required)
- No API endpoint (security consideration)
- Requires read access to SQLite file

**CLI Command**:
```bash
# Primary migration path
./vikunja import-db --sqlite-file=/path/to/vikunja.db --target-database=postgres

# With files directory
./vikunja import-db --sqlite-file=/path/to/vikunja.db --files-dir=/path/to/files
```

**Why This is Critical**:
- Primary migration path from original to refactored
- No need to modify original codebase
- Direct database access (fastest, most reliable)
- No authentication complexity
- Works even if original instance is offline

#### 2. Admin Export Service (NEW)

**Location**: `pkg/services/admin_export.go`

**Purpose**: Export all instance data for backups and portability

**Capabilities**:
- Export all users and their data
- Export system-wide entities (labels, teams, team memberships)
- Export all files (attachments, backgrounds, avatars)
- Generate manifest with metadata (version, export date, user count)
- Output format compatible with import service

**Access Control**:
- CLI command only (direct server access required)
- No API endpoint (security consideration)
- Requires file system access to Vikunja binary

**CLI Command**:
```bash
./vikunja export --output=/path/to/export.zip [--users=1,2,3] [--all]
```

#### 3. ZIP Archive Import Service (NEW)

**Location**: `pkg/services/import.go`

**Purpose**: Import previously exported data into any Vikunja instance

**Capabilities**:
- **Full Import Mode**: Import entire instance (fresh database)
- **Merge Import Mode**: Import into existing instance
- **User Import Mode**: Import single user data
- **Conflict Resolution**: Handle ID conflicts, duplicate emails
- **Validation**: Pre-import checks, post-import verification
- **Rollback**: Transaction-based import with rollback support

**CLI Commands**:
```bash
# Full instance import (fresh database)
./vikunja import --file=/path/to/export.zip --mode=full

# Merge import (existing instance)
./vikunja import --file=/path/to/export.zip --mode=merge --conflict-strategy=rename

# Single user import
./vikunja import --file=/path/to/user_export.zip --mode=user --target-user-id=123
```

#### 4. Migration Validator (NEW)

**Location**: `pkg/services/migration_validator.go`

**Purpose**: Validate exports before import and verify after import

**Capabilities**:
- Schema validation (ensure export format is correct)
- Data integrity checks (reference validation)
- Version compatibility checks
- Generate validation report
- Post-import verification (count entities, verify relationships)

**CLI Commands**:
```bash
# Validate export file
./vikunja validate-export --file=/path/to/export.zip

# Verify migration after import
./vikunja verify-migration --export=/path/to/export.zip --database=current
```

### Export Format Specification

#### Directory Structure

```
export-{timestamp}.zip
├── manifest.json              # Export metadata
├── users.json                 # All users (sanitized passwords)
├── teams.json                 # All teams
├── team_members.json          # Team memberships
├── labels.json                # Global labels
├── user_data/
│   ├── user_{id}/
│   │   ├── projects.json      # User's projects, tasks, views, buckets
│   │   ├── filters.json       # Saved filters
│   │   ├── attachments/       # Task attachment files
│   │   │   ├── {file_id}_{filename}
│   │   │   └── ...
│   │   └── backgrounds/       # Project background images
│   │       ├── {file_id}_{filename}
│   │       └── ...
│   └── ...
└── files_metadata.json        # File records (sizes, mimes, etc.)
```

#### Manifest Format

```json
{
  "version": "0.24.0",
  "export_type": "full|user|admin",
  "export_date": "2025-10-18T10:30:00Z",
  "database_type": "sqlite|postgres|mysql",
  "counts": {
    "users": 42,
    "teams": 5,
    "projects": 156,
    "tasks": 3421,
    "labels": 89,
    "files": 234
  },
  "options": {
    "include_passwords": false,
    "include_oauth_tokens": false,
    "sanitize_emails": false
  }
}
```

#### Users Export Format

```json
[
  {
    "id": 1,
    "username": "admin",
    "email": "admin@example.com",
    "name": "Administrator",
    "created": "2024-01-01T00:00:00Z",
    "updated": "2025-10-18T00:00:00Z",
    "password_hash": null,  // Excluded for security
    "is_local_user": true,
    "email_reminders_enabled": true,
    "discoverable_by_name": false,
    "discoverable_by_email": false,
    "overdue_tasks_reminders_enabled": true,
    "overdue_tasks_reminders_time": "09:00:00",
    "default_project_id": 3,
    "week_start": 0,
    "language": "en",
    "timezone": "America/New_York",
    "deletion_scheduled_at": null,
    "deletion_last_reminded_at": null
  }
]
```

### Import Strategies

#### Strategy 1: Full Import (Fresh Database)

**Use Case**: New Vikunja instance, empty database

**Process**:
1. Validate export file format and version compatibility
2. Check database is empty (safety check)
3. Import in dependency order:
   - Users (generate new passwords or mark for reset)
   - Teams
   - Team members
   - Labels
   - Projects (per user)
   - Tasks
   - Task relationships
   - Files (attachments, backgrounds)
   - Saved filters
4. Verify counts match manifest
5. Generate import report

**ID Handling**: Preserve original IDs where possible

#### Strategy 2: Merge Import (Existing Data)

**Use Case**: Merging data from another instance into existing instance

**Process**:
1. Validate export file
2. Detect conflicts:
   - Email conflicts (user already exists)
   - ID conflicts (entity ID already used)
   - Team name conflicts
3. Apply conflict resolution strategy:
   - **Skip**: Skip conflicting entities
   - **Rename**: Rename conflicting entities (append suffix)
   - **Merge**: Merge data into existing entities (risky)
   - **Error**: Fail on first conflict
4. Import non-conflicting data
5. Generate report with conflicts and actions taken

**ID Handling**: Remap all IDs to avoid conflicts

#### Strategy 3: User Import (Single User)

**Use Case**: Restore user's data or import from personal export

**Process**:
1. Validate export file
2. Check target user exists or create new user
3. Import user's data:
   - Projects → create new projects
   - Tasks → create in new projects
   - Filters → create for user
   - Files → copy to new locations
4. Remap all references and IDs
5. Generate import report

**ID Handling**: Remap all IDs, create new projects

### Security & Privacy Considerations

#### Data Sanitization Options

**Level 1: Full Export** (default for admin)
- Include all data except passwords
- Include emails (needed for user identification)
- Include OAuth tokens? **NO** (security risk)
- Include team memberships
- Include all files

**Level 2: User Export** (existing functionality)
- Single user's data only
- No other users' information
- No team member details beyond own
- Requires authentication

**Level 3: Anonymized Export** (optional flag)
- Replace emails with hashed versions
- Remove personally identifiable information
- Keep structural data intact
- For testing/development purposes

#### Password Handling

**Export Side**:
- **Never export password hashes** (security risk)
- Mark users as "password_reset_required" in import

**Import Side**:
- Set temporary random passwords for local users
- Mark all accounts for password reset on first login
- OIDC/LDAP users: No password needed (external auth)
- Send password reset emails post-import (optional)

#### Access Control

**Export Commands**:
- CLI only (no API endpoints for admin export)
- Requires server file system access
- Log all export operations
- Optional: Require confirmation for full exports

**Import Commands**:
- CLI only (no API endpoints)
- Requires server file system access
- Dry-run mode available
- Transaction rollback on errors

### Implementation Plan

#### Phase 1: Export Enhancements (Week 1)

**Tasks**:
1. Implement `AdminExportService` in `pkg/services/admin_export.go`
   - Export all users with their data
   - Export teams and team memberships
   - Export global labels
   - Generate manifest file
   - Output to ZIP archive

2. Implement `export` CLI command in `pkg/cmd/export.go`
   - Parse command-line arguments
   - Validate permissions
   - Call AdminExportService
   - Handle output file creation

3. Add `export-db` CLI command for direct database export
   - Direct database queries (bypass service layer)
   - Support for all database engines
   - Optional filters (user IDs, date ranges)

4. Update existing user export to match new format
   - Ensure compatibility with import service
   - Add manifest generation
   - Document format version

**Deliverables**:
- Working admin export CLI command
- Export format documentation
- Unit tests for export services

#### Phase 2: Import Implementation (Week 2)

**Tasks**:
1. Implement `ImportService` in `pkg/services/import.go`
   - Read and validate ZIP archives
   - Parse manifest and data files
   - Full import mode implementation
   - Transaction management
   - Error handling and rollback

2. Implement `import` CLI command in `pkg/cmd/import.go`
   - Parse command-line arguments
   - Mode selection (full/merge/user)
   - Progress reporting
   - Generate import report

3. Implement ID remapping logic
   - Track old → new ID mappings
   - Update all references
   - Handle foreign key constraints

4. Implement conflict resolution strategies
   - Detect conflicts (email, ID, names)
   - Apply resolution strategy
   - Log all conflicts and resolutions

**Deliverables**:
- Working import CLI command (full mode)
- Import service with transaction safety
- Unit tests for import service

#### Phase 3: Merge & User Import (Week 3)

**Tasks**:
1. Implement merge import mode
   - Conflict detection
   - Resolution strategy application
   - Partial import support

2. Implement user import mode
   - Single user data import
   - Target user mapping
   - Data ownership transfer

3. Add validation and verification tools
   - Pre-import validation
   - Post-import verification
   - Count verification
   - Relationship integrity checks

4. Implement `MigrationValidatorService`
   - Schema validation
   - Data integrity checks
   - Version compatibility checks
   - Generate detailed reports

**Deliverables**:
- Merge and user import modes working
- Validation tools implemented
- Integration tests

#### Phase 4: Testing & Documentation (Week 4)

**Tasks**:
1. Comprehensive testing
   - Unit tests for all services
   - Integration tests for full workflow
   - Test with multiple database engines
   - Test conflict scenarios
   - Performance testing with large datasets

2. Migration guide documentation
   - Step-by-step migration process
   - Common scenarios and solutions
   - Troubleshooting guide
   - FAQ

3. CLI help and examples
   - Complete command reference
   - Usage examples
   - Best practices

4. Validation and edge cases
   - Empty database import
   - Large instance migration
   - Partial import scenarios
   - Error recovery procedures

**Deliverables**:
- Complete test suite
- Migration documentation
- User guide
- Admin documentation

### Testing Strategy

#### Unit Tests

**Export Services**:
- Test admin export with multiple users
- Test database direct export
- Test manifest generation
- Test file packaging
- Test partial exports

**Import Services**:
- Test full import on empty database
- Test merge import with conflicts
- Test user import
- Test ID remapping
- Test conflict resolution strategies
- Test transaction rollback

**Validation**:
- Test pre-import validation
- Test post-import verification
- Test schema validation
- Test version compatibility

#### Integration Tests

**Full Migration Flow**:
1. Export from SQLite instance (original)
2. Import into PostgreSQL instance (refactored)
3. Verify all data transferred correctly
4. Compare counts and relationships
5. Test application functionality

**Merge Scenarios**:
1. Export from two separate instances
2. Import first instance fully
3. Import second instance in merge mode
4. Verify conflict resolution
5. Verify data integrity

**Database Compatibility**:
- Test export from: SQLite, PostgreSQL, MySQL
- Test import to: PostgreSQL, MySQL
- Verify cross-database compatibility

#### Performance Tests

- Export instance with 1000+ users
- Import instance with 10,000+ tasks
- Measure export/import times
- Monitor memory usage
- Test with large attachments (GB of files)

### Error Handling & Recovery

#### Export Errors

**Possible Errors**:
- Database connection failure
- File system write errors
- Insufficient disk space
- Corrupted database records

**Handling**:
- Log detailed error information
- Partial export capability (continue on non-critical errors)
- Validation before ZIP finalization
- Clear error messages with solutions

#### Import Errors

**Possible Errors**:
- Invalid export format
- Version incompatibility
- Database constraints violations
- Foreign key conflicts
- Duplicate entries

**Handling**:
- Pre-import validation (fail fast)
- Transaction-based import (rollback on error)
- Detailed error logging
- Continue import with warnings (for non-critical errors)
- Generate error report with failed entities

#### Recovery Procedures

**Failed Import**:
1. Transaction rollback (automatic)
2. Database restored to pre-import state
3. Review error report
4. Fix issues in export file or database
5. Retry import

**Partial Import Success**:
1. Review import report
2. Identify missing/failed entities
3. Create patch export with failed entities
4. Import patch in merge mode

### CLI Command Reference

#### Export Commands

```bash
# Export all users and data (admin export)
./vikunja export --output=/path/to/export.zip --all

# Export specific users
./vikunja export --output=/path/to/export.zip --users=1,2,3

# Export with database direct access (no auth)
./vikunja export-db --output=/path/to/export.zip --config=/etc/vikunja/config.yml

# User export via API (existing, enhanced)
# POST /api/v1/user/export/request
# GET /api/v1/user/export/download
```

#### Import Commands

```bash
# Full import (fresh database)
./vikunja import --file=/path/to/export.zip --mode=full

# Merge import with conflict resolution
./vikunja import --file=/path/to/export.zip --mode=merge --conflicts=rename

# Merge import (skip conflicts)
./vikunja import --file=/path/to/export.zip --mode=merge --conflicts=skip

# User import
./vikunja import --file=/path/to/user_export.zip --mode=user --user=email@example.com

# Dry run (validation only)
./vikunja import --file=/path/to/export.zip --dry-run
```

#### Validation Commands

```bash
# Validate export file
./vikunja validate-export --file=/path/to/export.zip

# Verify database after import
./vikunja verify-migration --export=/path/to/export.zip

# Compare two databases (original vs imported)
./vikunja compare-instances --source-config=/path/to/source.yml --target-config=/path/to/target.yml
```

### Migration Workflow Example

#### Scenario: Migrating Production Server (Primary Path)

**Current State**:
- Original Vikunja on server A (SQLite - `vikunja.db` file)
- 50 users, 500 projects, 5000 tasks
- Mix of local and OIDC users

**Target State**:
- Refactored Vikunja on server B (PostgreSQL)
- All data preserved
- All users can login immediately

**Migration Steps**:

1. **Preparation** (Day 1)
   ```bash
   # On server A (original instance)
   # Stop Vikunja to ensure consistent backup
   systemctl stop vikunja
   
   # Backup database and files
   cp /var/lib/vikunja/vikunja.db /backup/vikunja.db.backup
   cp -r /var/lib/vikunja/files /backup/files_backup
   
   # Restart original if needed for parallel operation
   systemctl start vikunja
   ```

2. **Setup Target** (Day 1)
   ```bash
   # On server B (refactored instance)
   # Setup PostgreSQL database
   createdb vikunja_production
   
   # Configure Vikunja
   vim /etc/vikunja/config.yml
   # Set database connection to PostgreSQL
   # Set files path
   
   # Initialize empty database with schema
   ./vikunja migrate
   ```

3. **Test Migration** (Day 2)
   ```bash
   # Copy database and files to server B
   scp /backup/vikunja.db.backup user@serverB:/tmp/vikunja.db
   scp -r /backup/files_backup user@serverB:/tmp/vikunja_files
   
   # Test import in staging environment (dry-run)
   ./vikunja import-db --sqlite-file=/tmp/vikunja.db \
     --files-dir=/tmp/vikunja_files \
     --dry-run
   
   # Actual test import
   ./vikunja import-db --sqlite-file=/tmp/vikunja.db \
     --files-dir=/tmp/vikunja_files \
     --target-database=postgres
   
   # Verify migration
   ./vikunja verify-migration --sqlite-file=/tmp/vikunja.db
   
   # Test application functionality
   # - Login as test users
   # - Check projects and tasks
   # - Verify attachments load
   # - Test team functionality
   
   # Reset database for actual migration
   dropdb vikunja_production
   createdb vikunja_production
   ./vikunja migrate
   ```

4. **Production Migration** (Day 3 - Maintenance Window)
   ```bash
   # Server A: Announce maintenance
   # Stop original Vikunja
   systemctl stop vikunja
   
   # Final database and files backup
   cp /var/lib/vikunja/vikunja.db /backup/vikunja_final.db
   rsync -av /var/lib/vikunja/files/ /backup/vikunja_files_final/
   
   # Copy to server B
   scp /backup/vikunja_final.db user@serverB:/var/vikunja/import/
   rsync -av /backup/vikunja_files_final/ user@serverB:/var/vikunja/import_files/
   
   # Server B: Import
   ./vikunja import-db --sqlite-file=/var/vikunja/import/vikunja_final.db \
     --files-dir=/var/vikunja/import_files \
     --target-database=postgres
   
   # Verify
   ./vikunja verify-migration --sqlite-file=/var/vikunja/import/vikunja_final.db
   
   # Start refactored Vikunja
   systemctl start vikunja
   
   # Post-migration checks
   # - Verify user count
   # - Test logins (local and OIDC)
   # - Spot-check projects and tasks
   # - Monitor logs for errors
   
   # Update DNS/load balancer to point to server B
   ```

5. **Post-Migration** (Day 4+)
   ```bash
   # Monitor for issues
   tail -f /var/log/vikunja/vikunja.log
   
   # Send password reset emails to local users
   ./vikunja send-password-resets --all-local-users
   
   # Keep server A running for 1 week as backup
   # After verification period: decommission server A
   ```

**Alternative: Migration Using Export/Import** (if SQLite file access is difficult)
   ```bash
   # This is a fallback method - use SQLite import if possible
   
   # On refactored instance: Export data
   ./vikunja export --output=/backup/export.zip --all
   
   # On target instance: Import
   ./vikunja import --file=/backup/export.zip --mode=full
   ```

### Database Schema Changes

No database schema changes required! The import/export system works with existing schemas by:
- Using service layer (already refactored)
- Working with current data models
- Handling relationships through existing foreign keys
- Supporting all current database engines

### API Changes

No API changes required for core functionality. The system is primarily CLI-based for security reasons.

**Optional Enhancement** (Future):
- Admin API endpoint for triggering exports (requires admin auth)
- API endpoint for import status/progress monitoring
- Webhook notifications for import completion

### Configuration Options

Add to `config.yml`:

```yaml
service:
  # Export configuration
  export:
    # Maximum export file size (in MB, 0 = unlimited)
    max_file_size: 0
    
    # Include password hashes in admin exports (NOT RECOMMENDED)
    include_passwords: false
    
    # Temporary directory for export generation
    temp_dir: "/tmp/vikunja-exports"
    
    # Export compression level (0-9, 9 = best compression)
    compression_level: 6

  # Import configuration  
  import:
    # Allow imports that overwrite existing data
    allow_full_import: true
    
    # Allow merge imports
    allow_merge_import: true
    
    # Default conflict resolution strategy (skip|rename|error)
    default_conflict_strategy: "rename"
    
    # Send password reset emails after import
    send_password_resets: true
    
    # Validate export format before import
    validate_before_import: true
    
    # Maximum import file size (in MB, 0 = unlimited)
    max_file_size: 0
```

### Success Criteria

#### Functional Requirements

- ✅ Export all users from original instance without authentication
- ✅ Import into refactored instance with any database backend
- ✅ Zero data loss during migration
- ✅ All relationships preserved (tasks→projects, users→teams, etc.)
- ✅ All files transferred (attachments, backgrounds, avatars)
- ✅ OIDC/LDAP users can login after migration
- ✅ Local users can reset passwords and login
- ✅ CLI commands are intuitive and well-documented

#### Non-Functional Requirements

- ✅ Export/import completes in reasonable time (< 1 hour for 1000 users)
- ✅ Memory efficient (streams data, doesn't load everything in memory)
- ✅ Disk space efficient (compressed exports)
- ✅ Transaction safety (rollback on errors)
- ✅ Comprehensive error messages
- ✅ Detailed logging for troubleshooting
- ✅ Support for partial exports/imports
- ✅ Dry-run validation before actual import

#### Testing Requirements

- ✅ Unit tests for all services (>80% coverage)
- ✅ Integration tests for full migration flow
- ✅ Tested with all supported databases
- ✅ Tested with large datasets (1000+ users)
- ✅ Tested with all conflict scenarios
- ✅ Performance benchmarks documented

#### Documentation Requirements

- ✅ Migration guide for admins
- ✅ CLI command reference
- ✅ Troubleshooting guide
- ✅ Export format specification
- ✅ Architecture documentation
- ✅ Example migration workflows

### Risks & Mitigations

#### Risk 1: Data Loss During Migration

**Impact**: High  
**Probability**: Low  
**Mitigation**:
- Transaction-based imports with rollback
- Validation before and after import
- Comprehensive testing with real data
- Keep original database as backup
- Dry-run capability

#### Risk 2: Performance Issues with Large Datasets

**Impact**: Medium  
**Probability**: Medium  
**Mitigation**:
- Stream data instead of loading all in memory
- Batch processing for large imports
- Progress reporting to show system is working
- Performance testing with large datasets
- Optimize database queries

#### Risk 3: Version Incompatibilities

**Impact**: Medium  
**Probability**: Low  
**Mitigation**:
- Version checking in manifest
- Schema validation
- Support for format versioning
- Clear error messages for incompatibilities
- Migration path documentation

#### Risk 4: Authentication Issues Post-Import

**Impact**: High  
**Probability**: Medium  
**Mitigation**:
- Clear documentation on password reset process
- Automated password reset email sending
- Support for admin password resets
- Test OIDC/LDAP authentication thoroughly
- Emergency admin access procedure

#### Risk 5: File System Issues (Attachments)

**Impact**: Medium  
**Probability**: Low  
**Mitigation**:
- Verify file permissions during import
- Check disk space before import
- Validate file integrity
- Support for external file storage (S3, etc.)
- Clear error messages for file issues

### Future Enhancements

#### Phase 5: Advanced Features (Post-MVP)

1. **Incremental Exports**
   - Export only changes since last export
   - Delta synchronization
   - Reduced export file sizes

2. **Scheduled Exports**
   - Automatic daily/weekly exports
   - Backup retention policies
   - Email notifications on completion

3. **Cloud Storage Integration**
   - Export directly to S3/Azure/GCS
   - Import from cloud storage
   - Encrypted storage support

4. **Multi-Instance Sync**
   - Bidirectional synchronization
   - Conflict resolution for concurrent changes
   - Distributed Vikunja deployments

5. **Import Mapping UI**
   - Web interface for import configuration
   - Visual conflict resolution
   - User/team mapping interface
   - Progress monitoring dashboard

6. **Selective Import**
   - Choose which users to import
   - Choose which projects to import
   - Exclude specific data types
   - Date range filtering

7. **Export Templates**
   - Pre-configured export scenarios
   - Anonymization profiles
   - Compliance-friendly exports (GDPR)
   - Development/testing exports

### Appendix A: Export File Format Examples

#### Manifest Example

```json
{
  "version": "0.24.0",
  "export_type": "admin",
  "export_date": "2025-10-18T14:30:00Z",
  "exporter_version": "1.0.0",
  "database_type": "sqlite",
  "source_instance": {
    "url": "https://vikunja.example.com",
    "installation_id": "550e8400-e29b-41d4-a716-446655440000"
  },
  "counts": {
    "users": 50,
    "teams": 8,
    "projects": 234,
    "tasks": 5432,
    "labels": 156,
    "attachments": 89,
    "backgrounds": 23,
    "filters": 67,
    "comments": 1234
  },
  "options": {
    "include_passwords": false,
    "include_oauth_tokens": false,
    "sanitize_emails": false,
    "anonymize": false
  },
  "schema_version": "2.0",
  "checksum": "sha256:abcdef1234567890..."
}
```

#### User Data Example

```json
{
  "user": {
    "id": 5,
    "username": "alice",
    "email": "alice@example.com",
    "name": "Alice Smith",
    "created": "2024-06-15T10:00:00Z",
    "updated": "2025-10-18T09:00:00Z",
    "is_local_user": true,
    "email_reminders_enabled": true
  },
  "projects": [
    {
      "id": 42,
      "title": "Website Redesign",
      "description": "Q4 2025 website refresh project",
      "is_archived": false,
      "hex_color": "#3498db",
      "identifier": "WEB",
      "position": 1,
      "created": "2024-08-01T00:00:00Z",
      "updated": "2025-10-17T15:30:00Z",
      "views": [
        {
          "id": 100,
          "title": "Board View",
          "view_kind": 3,
          "position": 1,
          "bucket_configuration_mode": 1
        }
      ],
      "tasks": [
        {
          "id": 500,
          "title": "Design mockups",
          "description": "Create design mockups for homepage",
          "done": true,
          "priority": 3,
          "due_date": "2024-09-15T00:00:00Z",
          "created": "2024-08-05T10:00:00Z",
          "attachments": ["file_1.pdf", "file_2.png"]
        }
      ]
    }
  ],
  "filters": [
    {
      "id": 10,
      "title": "High Priority Tasks",
      "filters": "priority >= 3 && done = false"
    }
  ]
}
```

### Appendix B: CLI Command Examples

See "CLI Command Reference" section above for comprehensive examples.

### Appendix C: Migration Checklist

**Pre-Migration**:
- [ ] Backup original database
- [ ] Test export on staging environment
- [ ] Validate export file
- [ ] Setup target instance and database
- [ ] Test import on staging environment
- [ ] Verify test import data
- [ ] Document rollback procedure
- [ ] Schedule maintenance window
- [ ] Notify users of migration

**During Migration**:
- [ ] Stop original instance
- [ ] Perform final export
- [ ] Transfer export file to target server
- [ ] Import into target instance
- [ ] Verify migration (counts, integrity)
- [ ] Test authentication (local + OIDC/LDAP)
- [ ] Spot-check critical data
- [ ] Start target instance
- [ ] Monitor logs for errors

**Post-Migration**:
- [ ] Update DNS/load balancer
- [ ] Send password reset emails
- [ ] Monitor for user issues
- [ ] Verify all functionality working
- [ ] Keep original instance as backup (1 week)
- [ ] Document any issues encountered
- [ ] Update documentation with lessons learned
- [ ] Decommission original instance (after verification period)

---

## Conclusion

This specification provides a comprehensive plan for implementing a universal data export/import system that enables seamless migration from the original Vikunja instance to the refactored version. The system is designed to be database-agnostic, secure, and reliable, with multiple export options including email-free admin exports for OIDC/LDAP scenarios.

The implementation follows a phased approach over 4 weeks, with clear deliverables, testing requirements, and success criteria. The system will enable smooth production migrations with zero data loss and minimal downtime.

**Next Steps**:
1. Review and approve specification
2. Begin Phase 1 implementation (Export Enhancements)
3. Create detailed technical design for services
4. Setup development environment for testing
