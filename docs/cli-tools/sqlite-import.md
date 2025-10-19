# SQLite Database Import

## Overview

The `import-db` command provides a robust way to migrate data from one Vikunja instance to another by importing directly from a SQLite database file. This is the **primary migration path** for Vikunja, especially when migrating from SQLite to PostgreSQL or MySQL.

## Use Cases

- **Database Migration**: Migrate from SQLite to PostgreSQL or MySQL
- **Instance Transfer**: Move your Vikunja installation to a new server
- **Database Consolidation**: Combine data from multiple instances (with manual ID coordination)
- **Disaster Recovery**: Restore from SQLite backup files
- **Development/Testing**: Import production data into development environments

## Command Syntax

```bash
vikunja import-db --sqlite-file=<path> [options]
```

### Required Flags

| Flag | Short | Description |
|------|-------|-------------|
| `--sqlite-file` | `-s` | Path to the SQLite database file to import (e.g., `/backup/vikunja.db`) |

### Optional Flags

| Flag | Short | Default | Description |
|------|-------|---------|-------------|
| `--files-dir` | `-f` | (none) | Path to the files directory containing attachments, backgrounds, and avatars |
| `--dry-run` | `-d` | `false` | Perform a validation run without making any database changes |
| `--quiet` | `-q` | `false` | Suppress progress output (useful for automation) |

## Basic Usage

### Import Database Only

Import just the database without file attachments:

```bash
vikunja import-db --sqlite-file=/backup/vikunja.db
```

### Import Database with Files

Import both database and files (recommended for complete migration):

```bash
vikunja import-db \
  --sqlite-file=/backup/vikunja.db \
  --files-dir=/backup/files
```

### Dry Run (Validation)

Test the import without making any changes:

```bash
vikunja import-db \
  --sqlite-file=/backup/vikunja.db \
  --dry-run
```

### Quiet Mode (Automation)

Suppress progress output for use in scripts or automation:

```bash
vikunja import-db \
  --sqlite-file=/backup/vikunja.db \
  --quiet
```

## Migration Workflow

### Step 1: Prepare Source Database

1. **Stop the source Vikunja instance** to ensure data consistency
2. **Locate the SQLite database file** (usually `vikunja.db` in the data directory)
3. **Locate the files directory** (usually `files/` in the data directory)
4. **Create backups** of both the database and files directory

```bash
# Example backup commands
cp /var/lib/vikunja/vikunja.db /backup/vikunja.db
cp -r /var/lib/vikunja/files /backup/files
```

### Step 2: Prepare Target Instance

1. **Install Vikunja** on the target server with your desired database backend
2. **Configure the database** in `config.yml`:

```yaml
database:
  type: "postgres"  # or "mysql"
  host: "localhost"
  database: "vikunja"
  user: "vikunja"
  password: "your-password"
```

3. **Initialize the database** (Vikunja will create tables on first run):

```bash
vikunja migrate
```

4. **Ensure the target database is empty** or contains only schema (no data)

### Step 3: Transfer Files

Copy the SQLite database and files directory to the target server:

```bash
# From source server to target server
scp /backup/vikunja.db user@target-server:/tmp/vikunja.db
scp -r /backup/files user@target-server:/tmp/files
```

### Step 4: Run Import

On the target server, run the import command:

```bash
# Test with dry run first
vikunja import-db \
  --sqlite-file=/tmp/vikunja.db \
  --files-dir=/tmp/files \
  --dry-run

# If validation passes, run the actual import
vikunja import-db \
  --sqlite-file=/tmp/vikunja.db \
  --files-dir=/tmp/files
```

### Step 5: Verify Import

1. **Check the import report** for entity counts and any errors
2. **Start Vikunja** and log in to verify data is accessible
3. **Test critical functionality**: projects, tasks, attachments, team access
4. **Verify file attachments** open correctly

### Step 6: Clean Up

After successful migration:

1. **Update DNS/URLs** to point to the new instance
2. **Decommission the old instance** (keep backups)
3. **Remove temporary files** from the target server

```bash
rm /tmp/vikunja.db
rm -rf /tmp/files
```

## Data Transformation

The import service automatically transforms data from the old schema format to the new format. Here's what gets transformed:

### Entity Mapping

| Source Entity | Target Entity | Notes |
|--------------|---------------|-------|
| `users` | `users` | Preserves user IDs, sanitizes data |
| `teams` | `teams` | Preserves team structure |
| `team_members` | `team_members` | Maintains memberships |
| `lists` | `projects` | **Old terminology → new terminology** |
| `tasks` | `tasks` | Full field mapping with NULL handling |
| `labels` | `labels` | Project-level and global labels |
| `task_labels` | `label_tasks` | **Table name change** |
| `task_comments` | `comments` | Preserved with relationships |
| `task_attachments` | `attachments` | Links to migrated files |
| `buckets` | `buckets` | Kanban board buckets |
| `saved_filters` | `saved_filters` | User-defined filters |
| `subscriptions` | `subscriptions` | Entity subscriptions |
| `project_views` | `project_views` | View configurations |
| `link_shares` | `link_shares` | Public sharing links |
| `webhooks` | `webhooks` | Webhook configurations |
| `reactions` | `reactions` | Task reactions |
| `api_tokens` | `api_tokens` | API authentication tokens |
| `favorites` | `favorites` | User favorites |

### Field Transformations

**Date/Time Fields**:
- Source: `DATETIME` strings or Unix timestamps
- Target: Proper database timestamp types (PostgreSQL `TIMESTAMP`, MySQL `DATETIME`)

**Boolean Fields**:
- Source: SQLite integers (`0` or `1`)
- Target: Native boolean types (PostgreSQL `BOOLEAN`, MySQL `TINYINT(1)`)

**Enum Fields**:
- Source: String values
- Target: Validated enum values (e.g., `SubscriptionEntityType`, `ReactionKind`)

**JSON Fields**:
- Source: JSON strings
- Target: Native JSON types (PostgreSQL `JSONB`, MySQL `JSON`)

**NULL Handling**:
- All nullable fields properly handle `NULL` values
- Empty strings converted to `NULL` where appropriate
- Zero values preserved or converted based on field semantics

### File Migration

When `--files-dir` is provided, files are migrated with integrity verification:

1. **Files are copied by ID** to match Vikunja's storage structure
2. **SHA-256 checksums verify** each file's integrity
3. **Missing files are logged** but don't block the import
4. **File metadata** (name, size, MIME type) is imported from the database

File storage structure:
```
files/
├── 1/
│   └── <file-id-1>
├── 2/
│   └── <file-id-2>
└── ...
```

## Transaction Safety

The import process uses database transactions to ensure **atomicity**:

- ✅ **All-or-nothing**: If any error occurs, the entire import is rolled back
- ✅ **No partial imports**: Your database is never left in an inconsistent state
- ✅ **Safe to retry**: Failed imports leave the database unchanged
- ✅ **Works across database engines**: PostgreSQL, MySQL, and SQLite all support transactions

### Import Order

Entities are imported in dependency order to satisfy foreign key constraints:

1. Users
2. Teams & Team Members
3. Projects & Project Backgrounds
4. Buckets (Kanban boards)
5. Tasks
6. Labels & Task-Label Links
7. Comments
8. Attachments (with file migration)
9. Saved Filters
10. Subscriptions
11. Project Views
12. Link Shares
13. Webhooks
14. Reactions
15. API Tokens
16. Favorites

## Import Report

After a successful import, you'll see a detailed report:

```
========================================
  Import Report
========================================
✓ Import completed successfully!
Duration: 15.2s

Entity Counts:
  Users:              100
  Teams:              10
  Team Members:       150
  Projects:           50
  Tasks:              1000
  Labels:             20
  Task-Label Links:   500
  Comments:           200
  Attachments:        75
  Buckets:            25
  Saved Filters:      30
  Subscriptions:      120
  Project Views:      100
  Project Backgrounds:10
  Link Shares:        5
  Webhooks:           3
  Reactions:          50
  API Tokens:         15
  Favorites:          80

File Migration:
  Files Processed:    75
  Files Copied:       75
  Files Failed:       0

========================================
✓ Successfully imported 1705 entities in 15.2s
✓ Migrated 75 files
```

## Progress Reporting

During import, progress is displayed for major entities:

```
Importing Users... (0/100)
Importing Users... 50/100 (50%)
Imported 100/100 Users (100%)

Importing Projects... (0/50)
Imported 50/50 Projects (100%)

Importing Tasks... (0/1000)
Importing Tasks... 500/1000 (50%)
Imported 1000/1000 Tasks (100%)
```

Progress updates are shown:
- **Every 100 users**
- **Every 50 projects**
- **Every 500 tasks**

Use `--quiet` to disable progress output.

## Troubleshooting

### Common Issues

#### Issue: "SQLite file not found or not accessible"

**Cause**: The SQLite file path is incorrect or the file doesn't exist.

**Solution**:
```bash
# Check if the file exists
ls -l /path/to/vikunja.db

# Check file permissions
chmod 644 /path/to/vikunja.db

# Use absolute path
vikunja import-db --sqlite-file=/absolute/path/to/vikunja.db
```

#### Issue: "Files directory not found or not accessible"

**Cause**: The files directory path is incorrect or doesn't exist.

**Solution**:
```bash
# Check if directory exists
ls -ld /path/to/files

# Use absolute path
vikunja import-db \
  --sqlite-file=/path/to/vikunja.db \
  --files-dir=/absolute/path/to/files
```

#### Issue: "Duplicate key error" or "Constraint violation"

**Cause**: Target database already contains data with conflicting IDs.

**Solution**:
```bash
# Option 1: Import into an empty database
# Drop and recreate the database, then run migrations
dropdb vikunja && createdb vikunja
vikunja migrate

# Option 2: Clean specific tables (if you know what you're doing)
# This is risky - backup first!
psql vikunja -c "TRUNCATE TABLE users, teams, projects, tasks CASCADE;"
```

#### Issue: "Transaction rollback - database unchanged"

**Cause**: An error occurred during import, and the transaction was rolled back.

**Solution**:
1. **Check the error message** in the output
2. **Review the import report** for specific errors
3. **Fix the underlying issue** (e.g., database permissions, disk space)
4. **Re-run the import** - the database is unchanged

#### Issue: "Files copied: 50/75, Files failed: 25"

**Cause**: Some files couldn't be copied (missing, permissions, disk space).

**Solution**:
- **Check the logs** for specific file IDs that failed
- **Verify source files exist**: `ls -l /path/to/files/<file-id>`
- **Check target disk space**: `df -h`
- **Check file permissions**: `chmod -R 644 /path/to/files`
- **Re-run import** - successfully copied files won't be re-copied

#### Issue: "Import is very slow (> 1 hour for 1000 tasks)"

**Cause**: Database performance issues or large files.

**Solution**:
- **Check database performance**: Ensure indexes exist (created by migrations)
- **Monitor disk I/O**: `iostat -x 5`
- **Check network latency**: If database is remote, consider local database first
- **Split import**: Import database first, then files separately
- **Increase database resources**: More RAM, faster disk

#### Issue: "Cannot connect to database"

**Cause**: Database configuration is incorrect or database is not running.

**Solution**:
```bash
# Check database is running
systemctl status postgresql  # or mysql

# Test connection manually
psql -h localhost -U vikunja -d vikunja  # PostgreSQL
mysql -h localhost -u vikunja -p vikunja  # MySQL

# Verify config.yml
cat config.yml | grep -A 10 database
```

### Validation Before Import

Always run a dry run first to validate:

```bash
vikunja import-db \
  --sqlite-file=/path/to/vikunja.db \
  --files-dir=/path/to/files \
  --dry-run
```

This will:
- ✅ Validate SQLite file is readable
- ✅ Validate target database is accessible
- ✅ Check schema compatibility
- ✅ Verify files directory exists
- ✅ Report expected entity counts
- ❌ **NOT** make any database changes

## Known Limitations

### 1. Authentication & Passwords

**Limitation**: User passwords are imported as-is from the source database.

**Impact**:
- Local users can log in with the same password
- OIDC/LDAP users will continue to authenticate through their providers
- API tokens are imported and remain valid

**Workaround**: None needed - authentication works as expected after import.

### 2. ID Conflicts in Merge Scenarios

**Limitation**: This command does **not** support merging data from multiple instances (e.g., combining two separate Vikunja installations).

**Impact**: If the target database already contains data, duplicate IDs will cause constraint violations.

**Workaround**:
- Import into an **empty database** only
- For merge scenarios, wait for Phase 3 (merge import mode) to be implemented

### 3. Schema Version Compatibility

**Limitation**: Import is designed for migrating between Vikunja instances of the **same version**.

**Impact**: Importing from significantly older versions may fail due to schema differences.

**Workaround**:
1. Upgrade source instance to the latest version
2. Run migrations: `vikunja migrate`
3. Export the updated database
4. Import into target instance

### 4. External Service Integrations

**Limitation**: External integrations (webhooks, OAuth providers) are imported but may need reconfiguration.

**Impact**:
- Webhook URLs may be incorrect if instance URL changed
- OAuth client secrets are not included in export (security)
- Email settings must be reconfigured

**Workaround**:
- Review webhook configurations after import
- Reconfigure OAuth providers if needed
- Update email settings in `config.yml`

### 5. File Checksums Not Stored

**Limitation**: File checksums are calculated during import but not stored for later verification.

**Impact**: No built-in way to verify file integrity after import.

**Workaround**:
- Run SHA-256 checksums manually if needed
- Verify files open correctly in the UI

### 6. Background Tasks & Scheduled Jobs

**Limitation**: Scheduled tasks (reminders, recurring tasks) may not fire immediately after import.

**Impact**: Reminders scheduled before import may be missed.

**Workaround**:
- Review tasks with reminders after import
- Restart Vikunja to ensure background workers are running

### 7. No Incremental Imports

**Limitation**: Each import is a full import - no support for importing only changes.

**Impact**: Re-running an import will attempt to re-insert all data.

**Workaround**:
- Clean the target database before re-importing
- Use backup/restore for incremental updates

## Performance

### Expected Import Times

| Dataset Size | Approximate Time | Notes |
|-------------|------------------|-------|
| 10 users, 100 tasks | < 5 seconds | Small personal instance |
| 100 users, 1,000 tasks | 10-15 seconds | Medium team instance |
| 500 users, 5,000 tasks | 30-60 seconds | Large team instance |
| 1,000 users, 10,000 tasks | 1-2 minutes | Enterprise instance |

**Factors affecting performance**:
- Database backend (PostgreSQL generally faster than MySQL)
- Database location (local vs. remote)
- Disk I/O speed (SSD vs. HDD)
- Number of files to migrate
- Network latency (if database is remote)

### Memory Usage

- **Peak memory**: < 500 MB for most datasets
- **Streaming**: Data is processed in batches, not loaded entirely into memory
- **Large files**: Files are streamed during copy, not loaded into memory

## FAQ

### Q: Can I import multiple times?

**A**: No, importing into a database that already contains data will fail with constraint violations. The target database must be empty (except for the schema).

### Q: Can I import into a different Vikunja version?

**A**: Best practice is to use the **same version** for source and target. If versions differ, upgrade the source instance first, then export and import.

### Q: What happens if the import fails?

**A**: The entire import is wrapped in a transaction. If any error occurs, **all changes are rolled back**, and your database remains unchanged. You can safely retry after fixing the issue.

### Q: Can I skip file migration?

**A**: Yes, simply omit the `--files-dir` flag. The database will be imported, but file attachments won't be accessible. You can copy files manually later to the correct location.

### Q: Do I need to stop Vikunja during import?

**A**: **Yes, strongly recommended**. Stop both the source instance (to ensure data consistency) and the target instance (to avoid conflicts during import).

### Q: Can I import from PostgreSQL or MySQL?

**A**: Currently, only SQLite imports are supported. For PostgreSQL/MySQL sources, use the export/import feature (Phase 2, coming soon).

### Q: Are passwords migrated?

**A**: Yes, password hashes are migrated as-is. Users can log in with their existing passwords after import.

### Q: Are OIDC/LDAP users migrated?

**A**: Yes, user records are migrated. OIDC/LDAP users will authenticate through their external providers after import (ensure providers are configured on the target instance).

### Q: What about API tokens and OAuth tokens?

**A**: API tokens are migrated and remain valid. OAuth tokens are **not** migrated for security reasons - users will need to re-authenticate with external services.

### Q: Can I automate imports?

**A**: Yes, use the `--quiet` flag to suppress progress output, and check the exit code (`0` for success, `1` for failure):

```bash
#!/bin/bash
vikunja import-db --sqlite-file=/backup/vikunja.db --quiet
if [ $? -eq 0 ]; then
  echo "Import successful"
else
  echo "Import failed"
  exit 1
fi
```

### Q: How do I verify the import was successful?

**A**: 
1. Check the import report for entity counts
2. Log in to the UI and verify projects/tasks are visible
3. Test opening file attachments
4. Verify team access and permissions
5. Check that all users can log in

### Q: Can I import from a backup older than my current version?

**A**: Not recommended. Always upgrade the source instance to match the target version before exporting/importing.

## Support

If you encounter issues not covered in this guide:

1. **Check the logs**: Look for detailed error messages in the Vikunja logs
2. **Search GitHub Issues**: [https://kolaente.dev/vikunja/vikunja/issues](https://kolaente.dev/vikunja/vikunja/issues)
3. **Ask in the Forum**: [https://community.vikunja.io](https://community.vikunja.io)
4. **Report bugs**: If you've found a bug, open an issue with:
   - Vikunja version
   - Source and target database types
   - Import command and flags used
   - Error messages from logs
   - Import report (sanitize any sensitive data)

## See Also

- [Database Configuration](../setup/database.md)
- [Migration Guide](../migration/index.md)
- [Backup & Restore](../admin/backup.md)
- [CLI Commands Overview](index.md)
