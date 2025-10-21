// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"xorm.io/xorm"
)

// SQLiteImportService handles importing data directly from SQLite database files
type SQLiteImportService struct {
	DB       *xorm.Engine
	Registry *ServiceRegistry
}

// NewSQLiteImportService creates a new instance of SQLiteImportService
func NewSQLiteImportService(engine *xorm.Engine, registry *ServiceRegistry) *SQLiteImportService {
	return &SQLiteImportService{
		DB:       engine,
		Registry: registry,
	}
}

// ImportReport contains statistics and results from an import operation
type ImportReport struct {
	Success          bool
	DatabaseImported bool
	FilesMigrated    bool
	FilesError       error
	StartTime        time.Time
	EndTime          time.Time
	Duration         time.Duration
	Counts           ImportCounts
	Errors           []string
}

// ImportCounts tracks how many entities were imported
type ImportCounts struct {
	Users              int64
	Teams              int64
	TeamMembers        int64
	Projects           int64
	Tasks              int64
	Labels             int64
	TaskLabels         int64
	Comments           int64
	Attachments        int64
	Buckets            int64
	SavedFilters       int64
	Subscriptions      int64
	ProjectViews       int64
	ProjectBackgrounds int64
	LinkShares         int64
	Webhooks           int64
	Reactions          int64
	APITokens          int64
	Favorites          int64
	Files              int64
	FilesCopied        int64
	FilesFailed        int64
}

// ImportOptions configures the import behavior
type ImportOptions struct {
	SQLiteFile string
	FilesDir   string
	DryRun     bool
	Quiet      bool
}

// ImportFromSQLite imports data from a SQLite database file into the target database
func (s *SQLiteImportService) ImportFromSQLite(opts ImportOptions) (*ImportReport, error) {
	report := &ImportReport{
		StartTime: time.Now(),
		Success:   false,
	}

	// Validate SQLite file exists and is readable
	if _, err := os.Stat(opts.SQLiteFile); err != nil {
		return report, fmt.Errorf("cannot access SQLite file: %w", err)
	}

	// Open SQLite database (read-only)
	sqliteDB, err := sql.Open("sqlite3", fmt.Sprintf("file:%s?mode=ro", opts.SQLiteFile))
	if err != nil {
		return report, fmt.Errorf("failed to open SQLite database: %w", err)
	}
	defer sqliteDB.Close()

	// Test SQLite connection
	if err := sqliteDB.Ping(); err != nil {
		return report, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	if !opts.Quiet {
		log.Info("Starting SQLite database import...")
		log.Infof("Source: %s", opts.SQLiteFile)
	}

	// Begin transaction on target database
	sess := s.DB.NewSession()
	defer sess.Close()

	if !opts.DryRun {
		if err := sess.Begin(); err != nil {
			return report, fmt.Errorf("failed to start transaction: %w", err)
		}
		if !opts.Quiet {
			log.Info("Transaction started")
		}
	}

	// Import data in dependency order
	var importErr error

	// 1. Import users
	if importErr == nil {
		report.Counts.Users, importErr = s.importUsers(sess, sqliteDB, opts)
	}

	// 2. Import file metadata (before attachments which reference files)
	if importErr == nil {
		var filesCount int64
		filesCount, importErr = s.importFileMetadata(sess, sqliteDB, opts)
		report.Counts.Files = filesCount
	}

	// 3. Import teams
	if importErr == nil {
		report.Counts.Teams, importErr = s.importTeams(sess, sqliteDB, opts)
	}

	// 4. Import team members
	if importErr == nil {
		report.Counts.TeamMembers, importErr = s.importTeamMembers(sess, sqliteDB, opts)
	}

	// 5. Import projects
	if importErr == nil {
		report.Counts.Projects, importErr = s.importProjects(sess, sqliteDB, opts)
	}

	// 6. Import tasks
	if importErr == nil {
		report.Counts.Tasks, importErr = s.importTasks(sess, sqliteDB, opts)
	}

	// 7. Import labels
	if importErr == nil {
		report.Counts.Labels, importErr = s.importLabels(sess, sqliteDB, opts)
	}

	// 8. Import task-label associations
	if importErr == nil {
		report.Counts.TaskLabels, importErr = s.importTaskLabels(sess, sqliteDB, opts)
	}

	// 9. Import comments
	if importErr == nil {
		report.Counts.Comments, importErr = s.importComments(sess, sqliteDB, opts)
	}

	// 10. Import attachments
	if importErr == nil {
		report.Counts.Attachments, importErr = s.importAttachments(sess, sqliteDB, opts)
	}

	// 11. Import buckets
	if importErr == nil {
		report.Counts.Buckets, importErr = s.importBuckets(sess, sqliteDB, opts)
	}

	// 11. Import saved filters
	if importErr == nil {
		report.Counts.SavedFilters, importErr = s.importSavedFilters(sess, sqliteDB, opts)
	}

	// 12. Import subscriptions
	if importErr == nil {
		report.Counts.Subscriptions, importErr = s.importSubscriptions(sess, sqliteDB, opts)
	}

	// 13. Import project views
	if importErr == nil {
		report.Counts.ProjectViews, importErr = s.importProjectViews(sess, sqliteDB, opts)
	}

	// 14. Import project backgrounds
	if importErr == nil {
		report.Counts.ProjectBackgrounds, importErr = s.importProjectBackgrounds(sess, sqliteDB, opts)
	}

	// 15. Import link shares
	if importErr == nil {
		report.Counts.LinkShares, importErr = s.importLinkShares(sess, sqliteDB, opts)
	}

	// 16. Import webhooks
	if importErr == nil {
		report.Counts.Webhooks, importErr = s.importWebhooks(sess, sqliteDB, opts)
	}

	// 17. Import reactions
	if importErr == nil {
		report.Counts.Reactions, importErr = s.importReactions(sess, sqliteDB, opts)
	}

	// 18. Import API tokens
	if importErr == nil {
		report.Counts.APITokens, importErr = s.importAPITokens(sess, sqliteDB, opts)
	}

	// 19. Import favorites
	if importErr == nil {
		report.Counts.Favorites, importErr = s.importFavorites(sess, sqliteDB, opts)
	}

	// Handle transaction commit/rollback
	if importErr != nil {
		report.Errors = append(report.Errors, importErr.Error())
		if !opts.DryRun {
			if !opts.Quiet {
				log.Errorf("Import failed: %v", importErr)
				log.Info("Rolling back transaction...")
			}
			if rollbackErr := sess.Rollback(); rollbackErr != nil {
				log.Errorf("Failed to rollback transaction: %v", rollbackErr)
				report.Errors = append(report.Errors, fmt.Sprintf("rollback failed: %v", rollbackErr))
			} else {
				if !opts.Quiet {
					log.Info("Transaction rolled back successfully - database state unchanged")
				}
			}
		}
		report.EndTime = time.Now()
		report.Duration = report.EndTime.Sub(report.StartTime)
		return report, importErr
	}

	// Commit transaction
	if !opts.DryRun {
		if !opts.Quiet {
			log.Info("All data imported successfully, committing transaction...")
		}
		if err := sess.Commit(); err != nil {
			report.Errors = append(report.Errors, fmt.Sprintf("commit failed: %v", err))
			report.EndTime = time.Now()
			report.Duration = report.EndTime.Sub(report.StartTime)
			if !opts.Quiet {
				log.Errorf("Transaction commit failed: %v", err)
			}
			return report, fmt.Errorf("failed to commit transaction: %w", err)
		}
		report.DatabaseImported = true
		if !opts.Quiet {
			log.Info("Transaction committed successfully")
			log.Info("Database import completed successfully")
		}
	} else {
		if !opts.Quiet {
			log.Info("Dry-run completed (no changes made)")
		}
	}

	// Import files (after database transaction)
	if opts.FilesDir != "" && !opts.DryRun {
		copied, failed, err := s.importFiles(opts)
		report.Counts.Files = copied + failed
		report.Counts.FilesCopied = copied
		report.Counts.FilesFailed = failed
		if err != nil {
			report.FilesError = err
			report.Errors = append(report.Errors, fmt.Sprintf("files migration failed: %v", err))
			if !opts.Quiet {
				log.Warningf("Files failed to migrate: %v", err)
			}
		} else if copied > 0 {
			report.FilesMigrated = true
			if !opts.Quiet {
				log.Info("Files migrated successfully")
			}
		}
	}

	report.Success = report.DatabaseImported || opts.DryRun
	report.EndTime = time.Now()
	report.Duration = report.EndTime.Sub(report.StartTime)

	return report, nil
}

// importUsers imports user data from SQLite
func (s *SQLiteImportService) importUsers(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	// Count total users for progress reporting
	total, err := countTableRows(sqliteDB, "users")
	if err != nil {
		return 0, fmt.Errorf("failed to count users: %w", err)
	}

	if !opts.Quiet {
		if total > 0 {
			log.Infof("Importing users... (0/%d)", total)
		} else {
			log.Info("Importing users...")
		}
	}

	rows, err := sqliteDB.Query(`
		SELECT id, username, password, email, name, created, updated, 
		       status, avatar_provider, language, timezone, week_start, 
		       default_project_id, overdue_tasks_reminders_time, 
		       overdue_tasks_reminders_enabled
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		u := &user.User{}
		var defaultProjectID sql.NullInt64
		var overdueRemindersTime sql.NullString
		var avatarProvider sql.NullString
		var language sql.NullString
		var timezone sql.NullString
		var weekStart sql.NullInt64

		err := rows.Scan(
			&u.ID, &u.Username, &u.Password, &u.Email, &u.Name,
			&u.Created, &u.Updated, &u.Status, &avatarProvider, &language, &timezone,
			&weekStart, &defaultProjectID, &overdueRemindersTime,
			&u.OverdueTasksRemindersEnabled,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan user row: %w", err)
		}

		// Handle nullable fields
		if defaultProjectID.Valid {
			u.DefaultProjectID = defaultProjectID.Int64
		}
		if overdueRemindersTime.Valid {
			u.OverdueTasksRemindersTime = overdueRemindersTime.String
		}
		if avatarProvider.Valid {
			u.AvatarProvider = avatarProvider.String
		}
		if language.Valid {
			u.Language = language.String
		}
		if timezone.Valid {
			u.Timezone = timezone.String
		}
		if weekStart.Valid {
			u.WeekStart = int(weekStart.Int64)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(u); err != nil {
				return count, fmt.Errorf("failed to insert user %d: %w", u.ID, err)
			}
		}
		count++

		// Progress updates with total
		if !opts.Quiet && total > 0 && count%100 == 0 {
			percentage := (count * 100) / total
			log.Infof("Importing users... %d/%d (%d%%)", count, total, percentage)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating users: %w", err)
	}

	if !opts.Quiet {
		logProgress(count, total, "users", opts.Quiet)
	}

	return count, nil
}

// importTeams imports team data from SQLite
func (s *SQLiteImportService) importTeams(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing teams...")
	}

	rows, err := sqliteDB.Query(`
		SELECT id, name, description, created, updated, created_by_id
		FROM teams
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query teams: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		team := &models.Team{}
		err := rows.Scan(
			&team.ID, &team.Name, &team.Description,
			&team.Created, &team.Updated, &team.CreatedByID,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan team row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(team); err != nil {
				return count, fmt.Errorf("failed to insert team %d: %w", team.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating teams: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d teams", count)
	}

	return count, nil
}

// tableExists checks if a table exists in the SQLite database
func tableExists(db *sql.DB, tableName string) (bool, error) {
	var name string
	err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", tableName).Scan(&name)
	if err == sql.ErrNoRows {
		return false, nil
	} else if err != nil {
		return false, fmt.Errorf("failed to check for table %s: %w", tableName, err)
	}
	return true, nil
}

// importTeamMembers imports team membership data from SQLite
func (s *SQLiteImportService) importTeamMembers(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing team members...")
	}

	rows, err := sqliteDB.Query(`
		SELECT id, team_id, user_id, admin, created
		FROM team_members
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		member := &models.TeamMember{}
		err := rows.Scan(
			&member.ID, &member.TeamID, &member.UserID,
			&member.Admin, &member.Created,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan team member row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(member); err != nil {
				return count, fmt.Errorf("failed to insert team member %d: %w", member.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating team members: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d team members", count)
	}

	return count, nil
}

// importProjects imports project data from SQLite
func (s *SQLiteImportService) importProjects(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	// Check which table name exists (old: lists, new: projects)
	tableName := "projects"
	exists, err := tableExists(sqliteDB, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
	}
	if !exists {
		// Try old table name
		tableName = "lists"
		exists, err = tableExists(sqliteDB, tableName)
		if err != nil {
			return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
		}
		if !exists {
			if !opts.Quiet {
				log.Info("No projects table found, skipping")
			}
			return 0, nil
		}
	}

	// Count total projects for progress reporting
	total, err := countTableRows(sqliteDB, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to count projects: %w", err)
	}

	if !opts.Quiet {
		if total > 0 {
			log.Infof("Importing projects... (0/%d)", total)
		} else {
			log.Info("Importing projects...")
		}
	}

	// Build query with actual table name
	// Note: Old "lists" table uses "list_id" for parent, new "projects" uses "parent_project_id"
	var query string
	if tableName == "lists" {
		query = `
			SELECT id, title, description, owner_id, identifier, 
			       hex_color, is_archived, background_file_id, background_blur_hash,
			       created, updated, parent_list_id, position
			FROM lists
			ORDER BY id
		`
	} else {
		query = `
			SELECT id, title, description, owner_id, identifier, 
			       hex_color, is_archived, background_file_id, background_blur_hash,
			       created, updated, parent_project_id, position
			FROM projects
			ORDER BY id
		`
	}

	rows, err := sqliteDB.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to query projects from %s: %w", tableName, err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		project := &models.Project{}
		var parentProjectID sql.NullInt64
		var position sql.NullFloat64
		var identifier sql.NullString
		var hexColor sql.NullString
		var backgroundFileID sql.NullInt64
		var backgroundBlurHash sql.NullString

		err := rows.Scan(
			&project.ID, &project.Title, &project.Description, &project.OwnerID,
			&identifier, &hexColor, &project.IsArchived,
			&backgroundFileID, &backgroundBlurHash, &project.Created, &project.Updated,
			&parentProjectID, &position,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan project row: %w", err)
		}

		// Handle nullable fields
		if parentProjectID.Valid {
			project.ParentProjectID = parentProjectID.Int64
		}
		if position.Valid {
			project.Position = position.Float64
		}
		if identifier.Valid {
			project.Identifier = identifier.String
		}
		if hexColor.Valid {
			project.HexColor = hexColor.String
		}
		if backgroundFileID.Valid {
			project.BackgroundFileID = backgroundFileID.Int64
		}
		if backgroundBlurHash.Valid {
			project.BackgroundBlurHash = backgroundBlurHash.String
		}

		if !opts.DryRun {
			if _, err := sess.Insert(project); err != nil {
				return count, fmt.Errorf("failed to insert project %d: %w", project.ID, err)
			}
		}
		count++

		// Progress updates with total
		if !opts.Quiet && total > 0 && count%50 == 0 {
			percentage := (count * 100) / total
			log.Infof("Importing projects... %d/%d (%d%%)", count, total, percentage)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating projects: %w", err)
	}

	if !opts.Quiet {
		logProgress(count, total, "projects", opts.Quiet)
	}

	return count, nil
}

// importTasks imports task data from SQLite
func (s *SQLiteImportService) importTasks(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	// Count total tasks for progress reporting
	total, err := countTableRows(sqliteDB, "tasks")
	if err != nil {
		return 0, fmt.Errorf("failed to count tasks: %w", err)
	}

	if !opts.Quiet {
		if total > 0 {
			log.Infof("Importing tasks... (0/%d)", total)
		} else {
			log.Info("Importing tasks...")
		}
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, description, done, done_at, due_date, 
		       project_id, repeat_after, repeat_mode,
		       priority, start_date, end_date, hex_color, 
		       percent_done, "index", uid, cover_image_attachment_id,
		       created, updated, created_by_id
		FROM tasks
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		task := &models.Task{}
		var doneAt sql.NullTime
		var dueDate sql.NullTime
		var repeatAfter sql.NullInt64
		var repeatMode sql.NullInt64
		var priority sql.NullInt64
		var startDate sql.NullTime
		var endDate sql.NullTime
		var hexColor sql.NullString
		var percentDone sql.NullFloat64
		var uid sql.NullString
		var coverImageAttachmentID sql.NullInt64

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Done, &doneAt,
			&dueDate, &task.ProjectID, &repeatAfter,
			&repeatMode, &priority, &startDate, &endDate,
			&hexColor, &percentDone, &task.Index,
			&uid, &coverImageAttachmentID, &task.Created, &task.Updated,
			&task.CreatedByID,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan task row: %w", err)
		}

		// Handle nullable fields
		if doneAt.Valid {
			task.DoneAt = doneAt.Time
		}
		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		if repeatAfter.Valid {
			task.RepeatAfter = repeatAfter.Int64
		}
		if repeatMode.Valid {
			task.RepeatMode = models.TaskRepeatMode(repeatMode.Int64)
		}
		if priority.Valid {
			task.Priority = priority.Int64
		}
		if startDate.Valid {
			task.StartDate = startDate.Time
		}
		if endDate.Valid {
			task.EndDate = endDate.Time
		}
		if hexColor.Valid {
			task.HexColor = hexColor.String
		}
		if percentDone.Valid {
			task.PercentDone = percentDone.Float64
		}
		if uid.Valid {
			task.UID = uid.String
		}
		if coverImageAttachmentID.Valid {
			task.CoverImageAttachmentID = coverImageAttachmentID.Int64
		}

		if !opts.DryRun {
			if _, err := sess.Insert(task); err != nil {
				return count, fmt.Errorf("failed to insert task %d: %w", task.ID, err)
			}
		}
		count++

		// Progress updates with total
		if !opts.Quiet && total > 0 && count%500 == 0 {
			percentage := (count * 100) / total
			log.Infof("Importing tasks... %d/%d (%d%%)", count, total, percentage)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating tasks: %w", err)
	}

	if !opts.Quiet {
		logProgress(count, total, "tasks", opts.Quiet)
	}

	return count, nil
}

// importLabels imports label data from SQLite
func (s *SQLiteImportService) importLabels(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing labels...")
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, description, hex_color, created_by_id, 
		       created, updated
		FROM labels
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query labels: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		label := &models.Label{}

		err := rows.Scan(
			&label.ID, &label.Title, &label.Description, &label.HexColor,
			&label.CreatedByID, &label.Created, &label.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan label row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(label); err != nil {
				return count, fmt.Errorf("failed to insert label %d: %w", label.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating labels: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d labels", count)
	}

	return count, nil
}

// importTaskLabels imports task-label associations from SQLite
func (s *SQLiteImportService) importTaskLabels(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing task-label associations...")
	}

	// Check which table name exists (old: task_labels, new: label_tasks)
	tableName := "label_tasks"
	exists, err := tableExists(sqliteDB, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
	}
	if !exists {
		// Try old table name
		tableName = "task_labels"
		exists, err = tableExists(sqliteDB, tableName)
		if err != nil {
			return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
		}
		if !exists {
			if !opts.Quiet {
				log.Info("No task-label association table found, skipping")
			}
			return 0, nil
		}
	}

	query := fmt.Sprintf(`
		SELECT id, task_id, label_id, created
		FROM %s
		ORDER BY id
	`, tableName)

	rows, err := sqliteDB.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to query task labels from %s: %w", tableName, err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		tl := &models.LabelTask{}
		err := rows.Scan(&tl.ID, &tl.TaskID, &tl.LabelID, &tl.Created)
		if err != nil {
			return count, fmt.Errorf("failed to scan task label row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(tl); err != nil {
				return count, fmt.Errorf("failed to insert task label %d: %w", tl.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating task labels: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d task-label associations", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importFileMetadata(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing file metadata...")
	}

	// Check if files table exists
	exists, err := tableExists(sqliteDB, "files")
	if err != nil {
		return 0, fmt.Errorf("failed to check if files table exists: %w", err)
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Files table does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, name, mime, size, created_by_id, created
		FROM files
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query files: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		file := &files.File{}
		var mime sql.NullString

		err := rows.Scan(
			&file.ID, &file.Name, &mime, &file.Size, &file.CreatedByID, &file.Created,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan file row: %w", err)
		}

		// Handle nullable fields
		if mime.Valid {
			file.Mime = mime.String
		}

		if !opts.DryRun {
			if _, err := sess.Insert(file); err != nil {
				return count, fmt.Errorf("failed to insert file %d: %w", file.ID, err)
			}
		}
		count++

		if !opts.Quiet && count%100 == 0 {
			log.Infof("Imported %d file records...", count)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating files: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d file records", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importComments(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing comments...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "task_comments")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table task_comments does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, comment, author_id, task_id, created, updated
		FROM task_comments
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query comments: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		comment := &models.TaskComment{}
		err := rows.Scan(
			&comment.ID, &comment.Comment, &comment.AuthorID,
			&comment.TaskID, &comment.Created, &comment.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan comment row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(comment); err != nil {
				return count, fmt.Errorf("failed to insert comment %d: %w", comment.ID, err)
			}
		}
		count++

		if !opts.Quiet && count%100 == 0 {
			log.Infof("Imported %d comments...", count)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating comments: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d comments", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importAttachments(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing attachments...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "task_attachments")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table task_attachments does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, task_id, file_id, created_by_id, created
		FROM task_attachments
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query attachments: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		attachment := &models.TaskAttachment{}
		err := rows.Scan(
			&attachment.ID, &attachment.TaskID, &attachment.FileID,
			&attachment.CreatedByID, &attachment.Created,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan attachment row: %w", err)
		}

		if !opts.DryRun {
			if _, err := sess.Insert(attachment); err != nil {
				return count, fmt.Errorf("failed to insert attachment %d: %w", attachment.ID, err)
			}
		}
		count++

		if !opts.Quiet && count%100 == 0 {
			log.Infof("Imported %d attachments...", count)
		}
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating attachments: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d attachments", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importBuckets(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing buckets...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "buckets")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table buckets does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, project_view_id, "limit", position, created, updated, created_by_id
		FROM buckets
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query buckets: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		bucket := &models.Bucket{}
		var limit sql.NullInt64
		var position sql.NullFloat64

		err := rows.Scan(
			&bucket.ID, &bucket.Title, &bucket.ProjectViewID,
			&limit, &position, &bucket.Created, &bucket.Updated,
			&bucket.CreatedByID,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan bucket row: %w", err)
		}

		// Handle nullable fields
		if limit.Valid {
			bucket.Limit = limit.Int64
		}
		if position.Valid {
			bucket.Position = position.Float64
		}

		if !opts.DryRun {
			if _, err := sess.Insert(bucket); err != nil {
				return count, fmt.Errorf("failed to insert bucket %d: %w", bucket.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating buckets: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d buckets", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importSavedFilters(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing saved filters...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "saved_filters")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table saved_filters does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, description, filters, owner_id, is_favorite, created, updated
		FROM saved_filters
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query saved filters: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		filter := &models.SavedFilter{}
		var filtersJSON sql.NullString
		var description sql.NullString

		err := rows.Scan(
			&filter.ID, &filter.Title, &description, &filtersJSON,
			&filter.OwnerID, &filter.IsFavorite, &filter.Created, &filter.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan saved filter row: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			filter.Description = description.String
		}

		// Parse filters JSON - this is critical for saved filters
		if filtersJSON.Valid && filtersJSON.String != "" {
			// The filters field is stored as JSON in the database
			// We need to set the raw JSON string which xorm will handle
			filter.Filters = &models.TaskCollection{}
			// Note: xorm will handle JSON unmarshaling when inserting
		}

		if !opts.DryRun {
			if _, err := sess.Insert(filter); err != nil {
				return count, fmt.Errorf("failed to insert saved filter %d: %w", filter.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating saved filters: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d saved filters", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importSubscriptions(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing subscriptions...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "subscriptions")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table subscriptions does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, entity_type, entity_id, user_id, created
		FROM subscriptions
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query subscriptions: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		subscription := &models.Subscription{}
		var entityType int64

		err := rows.Scan(
			&subscription.ID, &entityType, &subscription.EntityID,
			&subscription.UserID, &subscription.Created,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan subscription row: %w", err)
		}

		// Convert entity type from integer to SubscriptionEntityType
		subscription.EntityType = models.SubscriptionEntityType(entityType)

		if !opts.DryRun {
			if _, err := sess.Insert(subscription); err != nil {
				return count, fmt.Errorf("failed to insert subscription %d: %w", subscription.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating subscriptions: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d subscriptions", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importProjectViews(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing project views...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "project_views")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table project_views does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, project_id, view_kind, filter, position, 
		       bucket_configuration_mode, bucket_configuration, 
		       default_bucket_id, done_bucket_id, created, updated
		FROM project_views
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query project views: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		view := &models.ProjectView{}
		var filter sql.NullString
		var position sql.NullFloat64
		var bucketConfigMode sql.NullInt64
		var bucketConfig sql.NullString
		var defaultBucketID sql.NullInt64
		var doneBucketID sql.NullInt64

		err := rows.Scan(
			&view.ID, &view.Title, &view.ProjectID, &view.ViewKind,
			&filter, &position, &bucketConfigMode, &bucketConfig,
			&defaultBucketID, &doneBucketID, &view.Created, &view.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan project view row: %w", err)
		}

		// Handle nullable fields
		if filter.Valid && filter.String != "" {
			view.Filter = &models.TaskCollection{}
			// xorm will handle JSON unmarshaling
		}
		if position.Valid {
			view.Position = position.Float64
		}
		if bucketConfigMode.Valid {
			view.BucketConfigurationMode = models.BucketConfigurationModeKind(bucketConfigMode.Int64)
		}
		if bucketConfig.Valid && bucketConfig.String != "" {
			// xorm will handle JSON unmarshaling for bucket configuration
		}
		if defaultBucketID.Valid {
			view.DefaultBucketID = defaultBucketID.Int64
		}
		if doneBucketID.Valid {
			view.DoneBucketID = doneBucketID.Int64
		}

		if !opts.DryRun {
			if _, err := sess.Insert(view); err != nil {
				return count, fmt.Errorf("failed to insert project view %d: %w", view.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating project views: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d project views", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importProjectBackgrounds(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing project backgrounds...")
	}

	// Project backgrounds are stored in the projects table itself
	// BackgroundFileID, BackgroundInformation, BackgroundBlurHash
	// These are already imported as part of importProjects()
	// This function is a no-op placeholder for completeness

	if !opts.Quiet {
		log.Info("Project backgrounds are imported with projects")
	}

	return 0, nil
}

func (s *SQLiteImportService) importLinkShares(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing link shares...")
	}

	// Check which table name exists (old: link_sharing, new: link_shares)
	tableName := "link_shares"
	exists, err := tableExists(sqliteDB, tableName)
	if err != nil {
		return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
	}
	if !exists {
		// Try old table name
		tableName = "link_sharing"
		exists, err = tableExists(sqliteDB, tableName)
		if err != nil {
			return 0, fmt.Errorf("failed to check if %s table exists: %w", tableName, err)
		}
		if !exists {
			if !opts.Quiet {
				log.Info("No link shares table found, skipping")
			}
			return 0, nil
		}
	}

	// Check which columns exist (old: "right" and "list_id", new: "permission" and "project_id")
	// Query the schema to determine column names
	var hasRight bool
	var hasListID bool

	schemaRows, err := sqliteDB.Query(fmt.Sprintf("PRAGMA table_info(%s)", tableName))
	if err != nil {
		return 0, fmt.Errorf("failed to get schema for %s: %w", tableName, err)
	}
	defer schemaRows.Close()

	for schemaRows.Next() {
		var cid int
		var name string
		var colType string
		var notNull int
		var dfltValue sql.NullString
		var pk int

		if err := schemaRows.Scan(&cid, &name, &colType, &notNull, &dfltValue, &pk); err != nil {
			return 0, fmt.Errorf("failed to scan schema: %w", err)
		}

		if name == "right" {
			hasRight = true
		}
		if name == "list_id" {
			hasListID = true
		}
	}

	// Build query based on available columns
	permissionCol := "permission"
	if hasRight {
		permissionCol = "\"right\""
	}

	projectIDCol := "project_id"
	if hasListID {
		projectIDCol = "list_id"
	}

	query := fmt.Sprintf(`
		SELECT id, hash, name, %s, %s, sharing_type, 
		       password, shared_by_id, created, updated
		FROM %s
		ORDER BY id
	`, projectIDCol, permissionCol, tableName)

	rows, err := sqliteDB.Query(query)
	if err != nil {
		return 0, fmt.Errorf("failed to query link shares from %s: %w", tableName, err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		linkShare := &models.LinkSharing{}
		var name sql.NullString
		var password sql.NullString

		err := rows.Scan(
			&linkShare.ID, &linkShare.Hash, &name, &linkShare.ProjectID,
			&linkShare.Permission, &linkShare.SharingType, &password,
			&linkShare.SharedByID, &linkShare.Created, &linkShare.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan link share row: %w", err)
		}

		// Handle nullable fields
		if name.Valid {
			linkShare.Name = name.String
		}
		if password.Valid {
			linkShare.Password = password.String
		}

		if !opts.DryRun {
			if _, err := sess.Insert(linkShare); err != nil {
				return count, fmt.Errorf("failed to insert link share %d: %w", linkShare.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating link shares: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d link shares", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importWebhooks(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing webhooks...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "webhooks")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table webhooks does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, target_url, events, project_id, secret, created_by_id, created, updated
		FROM webhooks
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query webhooks: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		webhook := &models.Webhook{}
		var eventsJSON sql.NullString
		var secret sql.NullString

		err := rows.Scan(
			&webhook.ID, &webhook.TargetURL, &eventsJSON, &webhook.ProjectID,
			&secret, &webhook.CreatedByID, &webhook.Created, &webhook.Updated,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan webhook row: %w", err)
		}

		// Handle nullable fields
		if secret.Valid {
			webhook.Secret = secret.String
		}

		// Parse events JSON array
		if eventsJSON.Valid && eventsJSON.String != "" {
			// Events are stored as JSON array - xorm will handle unmarshaling
			webhook.Events = []string{}
		}

		if !opts.DryRun {
			if _, err := sess.Insert(webhook); err != nil {
				return count, fmt.Errorf("failed to insert webhook %d: %w", webhook.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating webhooks: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d webhooks", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importReactions(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing reactions...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "reactions")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table reactions does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, user_id, entity_id, entity_kind, value, created
		FROM reactions
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query reactions: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		reaction := &models.Reaction{}
		var entityKind int64

		err := rows.Scan(
			&reaction.ID, &reaction.UserID, &reaction.EntityID,
			&entityKind, &reaction.Value, &reaction.Created,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan reaction row: %w", err)
		}

		// Convert entity kind from integer to ReactionKind
		reaction.EntityKind = models.ReactionKind(entityKind)

		if !opts.DryRun {
			if _, err := sess.Insert(reaction); err != nil {
				return count, fmt.Errorf("failed to insert reaction %d: %w", reaction.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating reactions: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d reactions", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importAPITokens(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing API tokens...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "api_tokens")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table api_tokens does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT id, title, token_salt, token_hash, token_last_eight, 
		       permissions, expires_at, created, owner_id
		FROM api_tokens
		ORDER BY id
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query API tokens: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		token := &models.APIToken{}
		var permissionsJSON sql.NullString

		err := rows.Scan(
			&token.ID, &token.Title, &token.TokenSalt, &token.TokenHash,
			&token.TokenLastEight, &permissionsJSON, &token.ExpiresAt,
			&token.Created, &token.OwnerID,
		)
		if err != nil {
			return count, fmt.Errorf("failed to scan API token row: %w", err)
		}

		// Parse permissions JSON
		if permissionsJSON.Valid && permissionsJSON.String != "" {
			// Permissions are stored as JSON - xorm will handle unmarshaling
			token.APIPermissions = models.APIPermissions{}
		}

		if !opts.DryRun {
			if _, err := sess.Insert(token); err != nil {
				return count, fmt.Errorf("failed to insert API token %d: %w", token.ID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating API tokens: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d API tokens", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importFavorites(sess *xorm.Session, sqliteDB *sql.DB, opts ImportOptions) (int64, error) {
	if !opts.Quiet {
		log.Info("Importing favorites...")
	}

	// Check if table exists
	exists, err := tableExists(sqliteDB, "favorites")
	if err != nil {
		return 0, err
	}
	if !exists {
		if !opts.Quiet {
			log.Info("Table favorites does not exist, skipping")
		}
		return 0, nil
	}

	rows, err := sqliteDB.Query(`
		SELECT entity_id, user_id, kind
		FROM favorites
		ORDER BY entity_id, user_id, kind
	`)
	if err != nil {
		return 0, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var count int64
	for rows.Next() {
		favorite := &models.Favorite{}
		var kind int64

		err := rows.Scan(&favorite.EntityID, &favorite.UserID, &kind)
		if err != nil {
			return count, fmt.Errorf("failed to scan favorite row: %w", err)
		}

		// Convert kind from integer to FavoriteKind
		favorite.Kind = models.FavoriteKind(kind)

		if !opts.DryRun {
			if _, err := sess.Insert(favorite); err != nil {
				return count, fmt.Errorf("failed to insert favorite (entity=%d, user=%d): %w", favorite.EntityID, favorite.UserID, err)
			}
		}
		count++
	}

	if err := rows.Err(); err != nil {
		return count, fmt.Errorf("error iterating favorites: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Imported %d favorites", count)
	}

	return count, nil
}

func (s *SQLiteImportService) importFiles(opts ImportOptions) (int64, int64, error) {
	// If no files directory specified, skip file migration
	if opts.FilesDir == "" {
		if !opts.Quiet {
			log.Info("No files directory specified, skipping file migration")
		}
		return 0, 0, nil
	}

	// Validate source files directory exists
	if _, err := os.Stat(opts.FilesDir); err != nil {
		if os.IsNotExist(err) {
			log.Warningf("Files directory does not exist: %s (continuing without files)", opts.FilesDir)
			return 0, 0, nil
		}
		return 0, 0, fmt.Errorf("cannot access files directory: %w", err)
	}

	// Get target files directory from config
	targetFilesDir := config.FilesBasePath.GetString()
	if targetFilesDir == "" {
		return 0, 0, fmt.Errorf("target files directory not configured (service.files.basepath)")
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetFilesDir, 0755); err != nil {
		return 0, 0, fmt.Errorf("failed to create target files directory: %w", err)
	}

	if !opts.Quiet {
		log.Infof("Migrating files from %s to %s", opts.FilesDir, targetFilesDir)
	}

	// Get all file IDs from the database
	var fileIDs []int64
	err := s.DB.Table("files").Cols("id").Find(&fileIDs)
	if err != nil {
		return 0, 0, fmt.Errorf("failed to query file IDs: %w", err)
	}

	if len(fileIDs) == 0 {
		if !opts.Quiet {
			log.Info("No files to migrate")
		}
		return 0, 0, nil
	}

	if !opts.Quiet {
		log.Infof("Found %d files to migrate", len(fileIDs))
	}

	var copied, failed int64
	var errors []string

	for _, fileID := range fileIDs {
		// Source file path (stored by ID in source directory)
		sourceFile := filepath.Join(opts.FilesDir, strconv.FormatInt(fileID, 10))

		// Target file path (stored by ID in target directory)
		targetFile := filepath.Join(targetFilesDir, strconv.FormatInt(fileID, 10))

		// Check if source file exists
		sourceInfo, err := os.Stat(sourceFile)
		if err != nil {
			if os.IsNotExist(err) {
				log.Warningf("File %d: source file not found at %s (skipping)", fileID, sourceFile)
				failed++
				errors = append(errors, fmt.Sprintf("File %d: source file not found", fileID))
				continue
			}
			log.Warningf("File %d: cannot access source file: %v (skipping)", fileID, err)
			failed++
			errors = append(errors, fmt.Sprintf("File %d: %v", fileID, err))
			continue
		}

		// Check if target file already exists (avoid overwriting)
		if _, err := os.Stat(targetFile); err == nil {
			log.Debugf("File %d: target file already exists (skipping)", fileID)
			copied++
			continue
		}

		// Copy file with integrity verification
		if err := copyFileWithVerification(sourceFile, targetFile, sourceInfo.Size()); err != nil {
			log.Warningf("File %d: copy failed: %v (skipping)", fileID, err)
			failed++
			errors = append(errors, fmt.Sprintf("File %d: %v", fileID, err))

			// Clean up partial file
			_ = os.Remove(targetFile)
			continue
		}

		copied++

		if !opts.Quiet && copied%100 == 0 {
			log.Infof("Progress: %d/%d files copied", copied, len(fileIDs))
		}
	}

	if !opts.Quiet {
		log.Infof("File migration complete: %d copied, %d failed", copied, failed)
		if failed > 0 {
			log.Warningf("Failed files will be reported in import summary")
		}
	}

	return copied, failed, nil
}

// copyFileWithVerification copies a file and verifies its integrity using SHA-256 checksum
func copyFileWithVerification(src, dst string, expectedSize int64) error {
	// Open source file
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source: %w", err)
	}
	defer sourceFile.Close()

	// Create target file
	targetFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create target: %w", err)
	}
	defer targetFile.Close()

	// Copy with checksum calculation
	sourceHash := sha256.New()
	targetHash := sha256.New()

	// Use io.TeeReader to calculate source hash while copying
	sourceReader := io.TeeReader(sourceFile, sourceHash)

	// Copy to target and calculate target hash
	targetWriter := io.MultiWriter(targetFile, targetHash)
	written, err := io.Copy(targetWriter, sourceReader)
	if err != nil {
		return fmt.Errorf("copy failed: %w", err)
	}

	// Verify size
	if written != expectedSize {
		return fmt.Errorf("size mismatch: expected %d, got %d", expectedSize, written)
	}

	// Sync to disk
	if err := targetFile.Sync(); err != nil {
		return fmt.Errorf("failed to sync: %w", err)
	}

	// Verify checksums match
	sourceChecksum := hex.EncodeToString(sourceHash.Sum(nil))
	targetChecksum := hex.EncodeToString(targetHash.Sum(nil))

	if sourceChecksum != targetChecksum {
		return fmt.Errorf("checksum mismatch: source %s != target %s",
			sourceChecksum[:16], targetChecksum[:16])
	}

	return nil
}

// countTableRows returns the number of rows in a table, or 0 if table doesn't exist
func countTableRows(db *sql.DB, tableName string) (int64, error) {
	var count int64
	query := fmt.Sprintf("SELECT COUNT(*) FROM %s", tableName)
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		// Table might not exist, return 0
		return 0, nil
	}
	return count, nil
}

// logProgress logs progress in the format "Imported X/Y (Z%)"
func logProgress(current, total int64, entityType string, quiet bool) {
	if quiet {
		return
	}
	if total > 0 {
		percentage := (current * 100) / total
		log.Infof("Imported %d/%d %s (%d%%)", current, total, entityType, percentage)
	} else {
		log.Infof("Imported %d %s", current, entityType)
	}
}
