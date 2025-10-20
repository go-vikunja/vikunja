# Contracts: Data Export/Import System

This directory contains interface contracts and API specifications for the data export/import system.

## Service Interfaces

### AdminExportService Interface

```go
type AdminExportService interface {
    // ExportAllUsers exports all users and their data to a ZIP archive
    ExportAllUsers(outputPath string) error
    
    // ExportUsers exports specific users by ID to a ZIP archive
    ExportUsers(outputPath string, userIDs []int64) error
    
    // GenerateManifest creates a manifest file for an export
    GenerateManifest(data *ExportData) (*Manifest, error)
}
```

### ImportService Interface

```go
type ImportService interface {
    // ImportFull imports into an empty database (full mode)
    ImportFull(exportPath string) (*ImportReport, error)
    
    // ImportMerge imports into existing database with conflict resolution
    ImportMerge(exportPath string, strategy ConflictStrategy) (*ImportReport, error)
    
    // ImportUser imports a single user's data
    ImportUser(exportPath string, targetUserID int64) (*ImportReport, error)
    
    // ValidateExport validates an export file before import
    ValidateExport(exportPath string) (*ValidationReport, error)
}
```

### MigrationValidatorService Interface

```go
type MigrationValidatorService interface {
    // ValidateExportFormat validates the export file format
    ValidateExportFormat(exportPath string) error
    
    // ValidateDataIntegrity validates data integrity within export
    ValidateDataIntegrity(exportPath string) error
    
    // VerifyImport verifies imported data matches export
    VerifyImport(exportPath string, database Database) (*VerificationReport, error)
    
    // CompareInstances compares two database instances
    CompareInstances(source, target Database) (*ComparisonReport, error)
}
```

## Data Structures

### Manifest

```go
type Manifest struct {
    Version            string                 `json:"version"`
    ExportType         string                 `json:"export_type"`
    ExportDate         time.Time              `json:"export_date"`
    ExporterVersion    string                 `json:"exporter_version"`
    DatabaseType       string                 `json:"database_type"`
    SourceInstance     SourceInstanceInfo     `json:"source_instance"`
    Counts             EntityCounts           `json:"counts"`
    Options            ExportOptions          `json:"options"`
    SchemaVersion      string                 `json:"schema_version"`
    Checksum           string                 `json:"checksum"`
}

type SourceInstanceInfo struct {
    URL            string `json:"url"`
    InstallationID string `json:"installation_id"`
}

type EntityCounts struct {
    Users       int `json:"users"`
    Teams       int `json:"teams"`
    Projects    int `json:"projects"`
    Tasks       int `json:"tasks"`
    Labels      int `json:"labels"`
    Attachments int `json:"attachments"`
    Backgrounds int `json:"backgrounds"`
    Filters     int `json:"filters"`
    Comments    int `json:"comments"`
}

type ExportOptions struct {
    IncludePasswords bool `json:"include_passwords"`
    IncludeTokens    bool `json:"include_oauth_tokens"`
    SanitizeEmails   bool `json:"sanitize_emails"`
    Anonymize        bool `json:"anonymize"`
}
```

### Import Report

```go
type ImportReport struct {
    Success           bool                    `json:"success"`
    Mode              ImportMode              `json:"mode"`
    StartTime         time.Time               `json:"start_time"`
    EndTime           time.Time               `json:"end_time"`
    Duration          time.Duration           `json:"duration"`
    EntitiesImported  EntityCounts            `json:"entities_imported"`
    EntitiesSkipped   EntityCounts            `json:"entities_skipped"`
    Conflicts         []Conflict              `json:"conflicts"`
    Errors            []ImportError           `json:"errors"`
    IDMappings        IDMappings              `json:"id_mappings"`
}

type ImportMode string

const (
    ImportModeFull  ImportMode = "full"
    ImportModeMerge ImportMode = "merge"
    ImportModeUser  ImportMode = "user"
)

type Conflict struct {
    EntityType string           `json:"entity_type"`
    EntityID   int64            `json:"entity_id"`
    Field      string           `json:"field"`
    OldValue   interface{}      `json:"old_value"`
    NewValue   interface{}      `json:"new_value"`
    Resolution ConflictStrategy `json:"resolution"`
    Resolved   bool             `json:"resolved"`
}

type ConflictStrategy string

const (
    ConflictSkip   ConflictStrategy = "skip"
    ConflictRename ConflictStrategy = "rename"
    ConflictMerge  ConflictStrategy = "merge"
    ConflictError  ConflictStrategy = "error"
)

type ImportError struct {
    Stage      string      `json:"stage"`
    EntityType string      `json:"entity_type"`
    EntityID   int64       `json:"entity_id"`
    Error      string      `json:"error"`
    Timestamp  time.Time   `json:"timestamp"`
}

type IDMappings struct {
    Users    map[int64]int64 `json:"users"`
    Teams    map[int64]int64 `json:"teams"`
    Projects map[int64]int64 `json:"projects"`
    Tasks    map[int64]int64 `json:"tasks"`
    Labels   map[int64]int64 `json:"labels"`
    Files    map[int64]int64 `json:"files"`
}
```

### Validation Report

```go
type ValidationReport struct {
    Valid             bool              `json:"valid"`
    ManifestValid     bool              `json:"manifest_valid"`
    SchemaValid       bool              `json:"schema_valid"`
    IntegrityValid    bool              `json:"integrity_valid"`
    VersionCompatible bool              `json:"version_compatible"`
    Errors            []ValidationError `json:"errors"`
    Warnings          []string          `json:"warnings"`
}

type ValidationError struct {
    Field   string `json:"field"`
    Message string `json:"message"`
    Fatal   bool   `json:"fatal"`
}
```

### Verification Report

```go
type VerificationReport struct {
    Success              bool                  `json:"success"`
    CountsMatch          bool                  `json:"counts_match"`
    IntegrityValid       bool                  `json:"integrity_valid"`
    ExpectedCounts       EntityCounts          `json:"expected_counts"`
    ActualCounts         EntityCounts          `json:"actual_counts"`
    Discrepancies        []Discrepancy         `json:"discrepancies"`
    OrphanedRecords      []OrphanedRecord      `json:"orphaned_records"`
    MissingFiles         []string              `json:"missing_files"`
}

type Discrepancy struct {
    EntityType string `json:"entity_type"`
    Expected   int    `json:"expected"`
    Actual     int    `json:"actual"`
    Difference int    `json:"difference"`
}

type OrphanedRecord struct {
    Table      string `json:"table"`
    RecordID   int64  `json:"record_id"`
    ReferenceTo string `json:"reference_to"`
}
```

## CLI Command Contracts

### Export Command

```bash
vikunja export --output=<path> [--all] [--users=<ids>] [--config=<path>]

Options:
  --output PATH     Output ZIP file path (required)
  --all             Export all users (mutually exclusive with --users)
  --users IDS       Comma-separated user IDs to export
  --config PATH     Config file path (default: ./config.yml)
  --help            Show help

Exit Codes:
  0   Success
  1   Invalid arguments
  2   Export failed
  3   Permission denied
  4   Disk space error
```

### Export-DB Command

```bash
vikunja export-db --output=<path> --config=<path> [--users=<ids>]

Options:
  --output PATH     Output ZIP file path (required)
  --config PATH     Config file path (required)
  --users IDS       Comma-separated user IDs to export (optional)
  --help            Show help

Exit Codes:
  0   Success
  1   Invalid arguments
  2   Database connection failed
  3   Export failed
  4   Disk space error
```

### Import Command

```bash
vikunja import --file=<path> --mode=<mode> [--conflicts=<strategy>] [--dry-run]

Options:
  --file PATH              Import ZIP file path (required)
  --mode MODE              Import mode: full, merge, user (required)
  --conflicts STRATEGY     Conflict resolution: skip, rename, error (default: rename)
  --user-id ID             Target user ID (for user mode)
  --user-email EMAIL       Target user email (for user mode)
  --dry-run                Validate without importing
  --help                   Show help

Exit Codes:
  0   Success
  1   Invalid arguments
  2   Validation failed
  3   Import failed (transaction rolled back)
  4   Conflict resolution failed
```

### Validate Export Command

```bash
vikunja validate-export --file=<path>

Options:
  --file PATH     Export ZIP file path (required)
  --strict        Enable strict validation
  --help          Show help

Exit Codes:
  0   Valid export
  1   Invalid arguments
  2   Validation failed
  3   File not found or corrupt
```

### Verify Migration Command

```bash
vikunja verify-migration --export=<path> [--database=<current>]

Options:
  --export PATH       Export file path (required)
  --database TYPE     Database to verify (default: current)
  --help              Show help

Exit Codes:
  0   Verification passed
  1   Invalid arguments
  2   Verification failed
  3   Database connection failed
```

## Error Codes

### Export Errors

```go
const (
    ErrExportInvalidArgs       = 1000
    ErrExportDatabaseConnection = 1001
    ErrExportQueryFailed       = 1002
    ErrExportFileCreation      = 1003
    ErrExportDiskSpace         = 1004
    ErrExportPermission        = 1005
    ErrExportUserNotFound      = 1006
)
```

### Import Errors

```go
const (
    ErrImportInvalidArgs       = 2000
    ErrImportFileNotFound      = 2001
    ErrImportInvalidFormat     = 2002
    ErrImportVersionMismatch   = 2003
    ErrImportValidationFailed  = 2004
    ErrImportDatabaseConnection = 2005
    ErrImportTransactionFailed = 2006
    ErrImportConflict          = 2007
    ErrImportRollbackFailed    = 2008
)
```

### Validation Errors

```go
const (
    ErrValidationInvalidArgs    = 3000
    ErrValidationFileNotFound   = 3001
    ErrValidationCorruptZip     = 3002
    ErrValidationMissingManifest = 3003
    ErrValidationInvalidManifest = 3004
    ErrValidationMissingData     = 3005
    ErrValidationInvalidSchema   = 3006
    ErrValidationChecksumFailed  = 3007
)
```

## Configuration Contract

### Config File Section

```yaml
service:
  export:
    max_file_size: 0                    # MB, 0 = unlimited
    include_passwords: false            # Security: never true in production
    temp_dir: "/tmp/vikunja-exports"
    compression_level: 6                # 0-9, 9 = best compression
    
  import:
    allow_full_import: true
    allow_merge_import: true
    default_conflict_strategy: "rename" # skip|rename|error
    send_password_resets: true
    validate_before_import: true
    max_file_size: 0                    # MB, 0 = unlimited
```

## Event Hooks

### Export Events

```go
type ExportStartedEvent struct {
    UserIDs   []int64
    OutputPath string
    Timestamp time.Time
}

type ExportCompletedEvent struct {
    OutputPath string
    FileSize   int64
    Duration   time.Duration
    EntityCounts EntityCounts
    Timestamp  time.Time
}

type ExportFailedEvent struct {
    Error     error
    Stage     string
    Timestamp time.Time
}
```

### Import Events

```go
type ImportStartedEvent struct {
    FilePath  string
    Mode      ImportMode
    Timestamp time.Time
}

type ImportProgressEvent struct {
    Stage      string
    Current    int
    Total      int
    Percentage float64
    Timestamp  time.Time
}

type ImportCompletedEvent struct {
    Report    *ImportReport
    Timestamp time.Time
}

type ImportFailedEvent struct {
    Error     error
    Stage     string
    Report    *ImportReport
    Timestamp time.Time
}
```

## Testing Contracts

### Test Data Requirements

All services must be tested with:
- ✅ Empty database (0 users, 0 projects)
- ✅ Small dataset (10 users, 50 projects, 500 tasks)
- ✅ Medium dataset (100 users, 500 projects, 5,000 tasks)
- ✅ Large dataset (1,000 users, 5,000 projects, 50,000 tasks)
- ✅ Edge cases (orphaned records, missing foreign keys)
- ✅ All database engines (SQLite, PostgreSQL, MySQL)

### Performance Requirements

- Export 1,000 users: < 5 minutes
- Import 1,000 users: < 15 minutes
- Memory usage: < 500 MB
- Disk I/O: Efficient streaming (no full load)

### Test Coverage Requirements

- Unit test coverage: > 80%
- Integration test coverage: > 60%
- All error paths tested
- All conflict scenarios tested
- All database engines tested

## API Stability

**Version**: 1.0.0 (following semver)

**Compatibility Promise**:
- Export format is versioned in manifest
- Backward compatibility for 1.x format versions
- Forward compatibility warnings
- Migration tools for format upgrades

**Breaking Changes**:
- Will be major version bumps (2.0.0)
- Will be announced with migration guide
- Will provide conversion tools

## Documentation Requirements

Each service must have:
- ✅ Godoc comments on all public methods
- ✅ Usage examples in documentation
- ✅ Error handling documented
- ✅ Performance characteristics documented
- ✅ Migration guide for users

---

**Contract Version**: 1.0.0  
**Last Updated**: 2025-10-18  
**Status**: Draft (for review)
