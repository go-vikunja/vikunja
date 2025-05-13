// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package caldav

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	userpkg "code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// DavBasePath is the base url path
const DavBasePath = `/dav/`

// ProjectBasePath is the base path for all projects resources
const ProjectBasePath = DavBasePath + `projects`

const (
	davNS    = "DAV:"
	caldavNS = "urn:ietf:params:xml:ns:caldav"
)

// listCalendars handles PROPFIND on /caldav/ and lists project calendars
func ListCalendars(c echo.Context, currentUser *userpkg.User) error {
	s := db.NewSession()
	defer s.Close()

	// TODO: Use currentUser to filter projects
	// For now, assuming a method to get all projects for the user exists or can be created
	// Placeholder for project retrieval
	projectModel := models.Project{}
	projectsInterface, _, _, err := projectModel.ReadAll(s, currentUser, "", -1, 0) // Adjust limit as needed
	if err != nil {
		log.Errorf("Error reading projects for user %d: %v", currentUser.ID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	projects, ok := projectsInterface.([]*models.Project)
	if !ok {
		log.Errorf("Failed to cast projects to []*models.Project")
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	_ = enc.EncodeToken(xml.StartElement{
		Name: xml.Name{Local: "d:multistatus"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:d"}, Value: davNS},
			{Name: xml.Name{Local: "xmlns:c"}, Value: caldavNS}, // Note: sketch uses "cal", common is "c" or "cs"
		}})

	for _, p := range projects {
		projectURL := fmt.Sprintf("%s/%s/", ProjectBasePath, strconv.FormatInt(p.ID, 10))
		addPropfindResponse(enc, projectURL, p.Title, true, p.Updated, "") // Added LastModified and Etag
	}
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:multistatus"}})
	_ = enc.Flush()
	return c.XMLBlob(http.StatusMultiStatus, buf.Bytes())
}

// listTasksInProject handles PROPFIND on a project calendar and lists its tasks
func ListTasksInProject(c echo.Context, currentUser *userpkg.User, projectID int64) error {
	s := db.NewSession()
	defer s.Close()

	proj := models.Project{ID: projectID}
	canRead, _, err := proj.CanRead(s, currentUser)
	if err != nil {
		log.Errorf("Error checking read permission for project %d by user %d: %v", projectID, currentUser.ID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if !canRead {
		return c.NoContent(http.StatusForbidden)
	}
	// Read project details, including tasks
	projectWithTasks := models.ProjectWithTasksAndBuckets{Project: proj}
	err = projectWithTasks.ReadOne(s, currentUser)
	if err != nil {
		log.Errorf("Error reading project %d with tasks: %v", projectID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	
	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	_ = enc.EncodeToken(xml.StartElement{
		Name: xml.Name{Local: "d:multistatus"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:d"}, Value: davNS},
			{Name: xml.Name{Local: "xmlns:c"}, Value: caldavNS},
		}})
	
	// Add the calendar itself to the response
	calendarURL := fmt.Sprintf("%s/%s/", ProjectBasePath, strconv.FormatInt(projectID, 10))
	addPropfindResponse(enc, calendarURL, projectWithTasks.Project.Title, true, projectWithTasks.Project.Updated, projectEtag(&projectWithTasks.Project))


	for _, taskItem := range projectWithTasks.Tasks {
		task := taskItem.Task // Assuming TaskWithComments has a Task field
		href := fmt.Sprintf("%s/%s/%s.ics", ProjectBasePath, strconv.FormatInt(projectID, 10), task.UID)
		addPropfindResponse(enc, href, task.Title, false, task.Updated, taskEtag(&task))
	}
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:multistatus"}})
	_ = enc.Flush()
	return c.XMLBlob(http.StatusMultiStatus, buf.Bytes())
}

// addPropfindResponse writes a single d:response XML element
// Extended to include LastModified and Etag
func addPropfindResponse(enc *xml.Encoder, href, displayName string, isCalendar bool, lastModified time.Time, etag string) {
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:response"}})
	_ = enc.EncodeElement(href, xml.StartElement{Name: xml.Name{Local: "d:href"}})

	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:propstat"}})
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:prop"}})

	// Resource Type
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:resourcetype"}})
	if isCalendar {
		_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:collection"}})
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:collection"}})
		_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Space: caldavNS, Local: "calendar"}})
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Space: caldavNS, Local: "calendar"}})
	} else {
		// For individual task .ics files, resourcetype is empty or just <d:resource/>
		// Or, if we want to be specific, it's a VTODO component
	}
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:resourcetype"}})

	_ = enc.EncodeElement(displayName, xml.StartElement{Name: xml.Name{Local: "d:displayname"}})
	
	// Add supported-calendar-component-set for calendars
	if isCalendar {
		_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Space: caldavNS, Local: "supported-calendar-component-set"}})
		_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Space: caldavNS, Local: "comp"}, Attr: []xml.Attr{{Name: xml.Name{Local: "name"}, Value: "VTODO"}}})
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Space: caldavNS, Local: "comp"}})
		_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Space: caldavNS, Local: "supported-calendar-component-set"}})
	}

	// Last Modified
	if !lastModified.IsZero() {
		_ = enc.EncodeElement(lastModified.UTC().Format(http.TimeFormat), xml.StartElement{Name: xml.Name{Local: "d:getlastmodified"}})
	}

	// Etag
	if etag != "" {
		_ = enc.EncodeElement(etag, xml.StartElement{Name: xml.Name{Local: "d:getetag"}})
	}
	
	// ContentType for .ics files
	if !isCalendar {
		_ = enc.EncodeElement("text/calendar; component=vtodo", xml.StartElement{Name: xml.Name{Local: "d:getcontenttype"}})
	} else {
		_ = enc.EncodeElement("httpd/unix-directory", xml.StartElement{Name: xml.Name{Local: "d:getcontenttype"}})
	}


	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:prop"}})
	_ = enc.EncodeElement("HTTP/1.1 200 OK", xml.StartElement{Name: xml.Name{Local: "d:status"}}) // Status per property
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:propstat"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:response"}})
}

// fetchTaskAsICS serves a VTODO as .ics for the specified task
func FetchTaskAsICS(c echo.Context, currentUser *userpkg.User, projectID int64, taskUID string) error {
	s := db.NewSession()
	defer s.Close()

	tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("Error getting task by UID %s: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(tasks) == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	task := tasks[0]

	// Verify task belongs to the project and user has access
	if task.ProjectID != projectID {
		return c.NoContent(http.StatusNotFound) // Or Forbidden
	}
	// CanRead check on project should suffice if GetTasksByUIDs already filters by user access.
	// If not, an explicit task.CanRead(s, currentUser) might be needed.

	// Use Vikunja's existing ICS generation logic
	// We need a ProjectWithTasksAndBuckets and []*TaskWithComments for GetCaldavTodosForTasks
	// For a single task, we construct these.
	projectForICS := &models.ProjectWithTasksAndBuckets{
		Project: models.Project{ID: task.ProjectID, Title: "Dummy Project Title"}, // Title might be needed by GetCaldavTodosForTasks
	}
	taskWithComments := []*models.TaskWithComments{{Task: *task}}

	icsData := caldav.GetCaldavTodosForTasks(projectForICS, taskWithComments)
	if icsData == "" {
		log.Warningf("Generated empty ICS for task UID %s", taskUID)
		// This might happen if the task has no serializable fields or GetCaldavTodosForTasks expects more project context.
		// Fallback to simpler ICS generation if needed, or ensure GetCaldavTodosForTasks handles single tasks.
		// For now, let's assume it works or returns a minimal valid calendar for an empty task.
	}
	
	c.Response().Header().Set(echo.HeaderContentType, "text/calendar; charset=utf-8")
	c.Response().Header().Set("ETag", taskEtag(task))
	return c.String(http.StatusOK, icsData)
}

// upsertTaskFromICS parses a VTODO .ics to create or update a task
func UpsertTaskFromICS(c echo.Context, currentUser *userpkg.User, projectID int64, taskUID string) error {
	s := db.NewSession()
	defer s.Close()

	// Check If-Match header for ETags if present
	ifMatch := c.Request().Header.Get("If-Match")
	var existingTask *models.Task

	tasks, _ := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
	if len(tasks) > 0 {
		existingTask = tasks[0]
		if existingTask.ProjectID != projectID { // Task exists but in wrong project for this URL
			return c.NoContent(http.StatusConflict) // Or create new if UID collision policy allows
		}
		if ifMatch != "" && ifMatch != taskEtag(existingTask) && ifMatch != "*" {
			return c.NoContent(http.StatusPreconditionFailed)
		}
	}


	bodyBytes, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		log.Errorf("Error reading request body: %v", err)
		return c.NoContent(http.StatusBadRequest)
	}

	// Use Vikunja's existing ICS parsing logic
	parsedTask, err := caldav.ParseTaskFromVTODO(string(bodyBytes))
	if err != nil {
		log.Warningf("Failed to parse VTODO from ICS for UID %s: %v", taskUID, err)
		return c.String(http.StatusBadRequest, "Invalid ICS data")
	}

	// Ensure UID from path matches UID in ICS, or is new
	if parsedTask.UID == "" { // ICS might not contain UID for new tasks, client might expect server to generate
		parsedTask.UID = taskUID // Use UID from URL
	} else if parsedTask.UID != taskUID {
		// UID mismatch, client might be trying to move/copy, or it's an error
		log.Warningf("Task UID in path (%s) does not match UID in ICS (%s)", taskUID, parsedTask.UID)
		return c.String(http.StatusBadRequest, "Task UID in path does not match UID in ICS")
	}


	parsedTask.ProjectID = projectID // Assign to the current project

	var httpStatus = http.StatusNoContent // Default for update

	if existingTask == nil { // Task does not exist, create it
		canCreate, err := parsedTask.CanCreate(s, currentUser)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error checking create permission for task %s: %v", parsedTask.UID, err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if !canCreate {
			_ = s.Rollback()
			return c.NoContent(http.StatusForbidden)
		}

		err = parsedTask.Create(s, currentUser)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error creating task %s: %v", parsedTask.UID, err)
			// Could be models.ErrTaskAlreadyExists if UID collision, handle appropriately
			return c.NoContent(http.StatusInternalServerError)
		}
		httpStatus = http.StatusCreated
		// Location header for new resource? CalDAV spec might require.
		// c.Response().Header().Set("Location", getTaskURL(parsedTask))
		// For now, relying on ETag for newly created resource.
	} else { // Task exists, update it
		// Ensure ID is set for update
		parsedTask.ID = existingTask.ID
		
		// Check If-None-Match for creation if client uses it that way (e.g. PUT if-none-match: *)
		ifNoneMatch := c.Request().Header.Get("If-None-Match")
		if ifNoneMatch == "*" && existingTask != nil { // trying to create but it exists
			return c.NoContent(http.StatusPreconditionFailed)
		}


		canUpdate, err := parsedTask.CanUpdate(s, currentUser) // Use parsedTask as it has new data with old ID
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error checking update permission for task %s: %v", parsedTask.UID, err)
			return c.NoContent(http.StatusInternalServerError)
		}
		if !canUpdate {
			_ = s.Rollback()
			return c.NoContent(http.StatusForbidden)
		}

		err = parsedTask.Update(s, currentUser)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error updating task %s: %v", parsedTask.UID, err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	// Persist labels and relations (similar to old implementation)
	if parsedTask.Labels != nil {
		err = persistLabels(s, currentUser, parsedTask, parsedTask.Labels)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error persisting labels for task %s: %v", parsedTask.UID, err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	if parsedTask.RelatedTasks != nil {
		err = persistRelations(s, currentUser, parsedTask, parsedTask.RelatedTasks)
		if err != nil {
			_ = s.Rollback()
			log.Errorf("Error persisting relations for task %s: %v", parsedTask.UID, err)
			return c.NoContent(http.StatusInternalServerError)
		}
	}


	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction for task %s: %v", parsedTask.UID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	
	// Return ETag of the (newly) created/updated resource
	// Need to re-fetch the task to get its latest 'Updated' timestamp for ETag
	finalTasks, err := models.GetTasksByUIDs(db.NewSession(), []string{parsedTask.UID}, currentUser) // Use new session for fresh data
	if err == nil && len(finalTasks) > 0 {
		c.Response().Header().Set("ETag", taskEtag(finalTasks[0]))
	}

	return c.NoContent(httpStatus)
}

// removeTaskICS deletes the task resource
func RemoveTaskICS(c echo.Context, currentUser *userpkg.User, projectID int64, taskUID string) error {
	s := db.NewSession()
	defer s.Close()

	tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		_ = s.Rollback() // Rollback on error before checking IsErrTaskDoesNotExist
		log.Errorf("Error getting task by UID %s for deletion: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(tasks) == 0 {
		_ = s.Rollback() // Ensure rollback if task not found
		return c.NoContent(http.StatusNotFound)
	}
	taskToDelete := tasks[0]

	if taskToDelete.ProjectID != projectID {
		_ = s.Rollback()
		return c.NoContent(http.StatusNotFound) // Or Forbidden
	}
	
	// Check If-Match header for ETag
	ifMatch := c.Request().Header.Get("If-Match")
	if ifMatch != "" && ifMatch != taskEtag(taskToDelete) && ifMatch != "*" {
		_ = s.Rollback()
		return c.NoContent(http.StatusPreconditionFailed)
	}


	canDelete, err := taskToDelete.CanDelete(s, currentUser)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error checking delete permission for task %s: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if !canDelete {
		_ = s.Rollback()
		return c.NoContent(http.StatusForbidden)
	}

	err = taskToDelete.Delete(s, currentUser)
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Error deleting task %s: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction for deleting task %s: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}

	return c.NoContent(http.StatusNoContent)
}

// Helper function to generate ETag for a task
func taskEtag(task *models.Task) string {
	if task == nil {
		return ""
	}
	return fmt.Sprintf(`"%d-%d"`, task.ID, task.Updated.UnixNano())
}

// Helper function to generate ETag for a project (calendar)
func projectEtag(project *models.Project) string {
	if project == nil {
		return ""
	}
	// ETag for a calendar could be based on its last modification time
	// or a hash of its contents' ETags if more precision is needed.
	// For simplicity, using project's own update timestamp.
	return fmt.Sprintf(`"project-%d-%d"`, project.ID, project.Updated.UnixNano())
}

// GetTaskPropertiesAsXML handles PROPFIND for a single task resource.
func GetTaskPropertiesAsXML(c echo.Context, currentUser *userpkg.User, projectID int64, taskUID string) error {
	s := db.NewSession()
	defer s.Close()

	tasks, err := models.GetTasksByUIDs(s, []string{taskUID}, currentUser)
	if err != nil {
		if models.IsErrTaskDoesNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		log.Errorf("Error getting task by UID %s for PROPFIND: %v", taskUID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	if len(tasks) == 0 {
		return c.NoContent(http.StatusNotFound)
	}
	task := tasks[0]

	// Verify task belongs to the project
	if task.ProjectID != projectID {
		log.Warningf("Task UID %s found but belongs to project %d, expected %d", taskUID, task.ProjectID, projectID)
		return c.NoContent(http.StatusNotFound) // Or Forbidden, but NotFound seems appropriate for wrong path
	}

	// Commit session if any read operations were performed and were successful before this point.
	// For GetTasksByUIDs, it's a read, so explicit commit isn't strictly necessary unless s was used for writes before.
	// However, to be safe and consistent if other operations were added to the session:
	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction for task PROPFIND %s: %v", taskUID, err)
        // Not returning error here as main data is already fetched, but logging is important.
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	_ = enc.EncodeToken(xml.StartElement{
		Name: xml.Name{Local: "d:multistatus"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:d"}, Value: davNS},
			{Name: xml.Name{Local: "xmlns:c"}, Value: caldavNS},
		}})

	href := fmt.Sprintf("%s/%s/%s.ics", ProjectBasePath, strconv.FormatInt(projectID, 10), task.UID)
	addPropfindResponse(enc, href, task.Title, false, task.Updated, taskEtag(task))

	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:multistatus"}})
	_ = enc.Flush()

	return c.XMLBlob(http.StatusMultiStatus, buf.Bytes())
}

// The following functions (persistLabels, persistRelations, removeStaleRelations, getTaskURL)
// are adapted from the original VikunjaCaldavProjectStorage.
// They might need adjustments to fit the new handler structure, particularly how 'task' and 'user' are passed.

func getTaskURL(task *models.Task) string {
	return fmt.Sprintf("%s/%s/%s.ics", ProjectBasePath, strconv.FormatInt(task.ProjectID, 10), task.UID)
}

// persistLabels persists labels for a task.
// Requires web.Auth interface, which userpkg.User should satisfy if it has GetID().
func persistLabels(s *xorm.Session, authUser web.Auth, task *models.Task, labels []*models.Label) (err error) {
	labelTitles := make([]string, 0, len(labels))
	for _, label := range labels {
		labelTitles = append(labelTitles, label.Title)
	}

	// Ensure the user object for GetLabelsByTaskIDsOptions is correctly initialized
	var u *userpkg.User
	ifConcreteUser, ok := authUser.(*userpkg.User)
	if ok {
		u = ifConcreteUser
	} else {
		// If authUser is not *userpkg.User, we might need to fetch the user by ID
		// For now, assuming authUser is *userpkg.User or has necessary fields/methods.
		// This part might need adjustment based on how web.Auth is implemented by currentUser.
		u = &userpkg.User{ID: authUser.GetID()} // Minimal user for query
	}


	existingLabelsResult, _, _, err := models.GetLabelsByTaskIDs(s, &models.LabelByTaskIDsOptions{
		Search:              labelTitles,
		User:                u, // This User might need more fields than just ID depending on GetLabelsByTaskIDs internals
		GetForUser:          true,
		GetUnusedLabels:     true, // This might create labels if they don't exist and are "unused" globally? Or specific to user?
		GroupByLabelIDsOnly: true, // Check if this option is appropriate
	})
	if err != nil {
		return err
	}

	labelMap := make(map[string]*models.Label)
	for i := range existingLabelsResult {
		// existingLabelsResult is []models.LabelUser, which has a Label field
		labelMap[existingLabelsResult[i].Label.Title] = &existingLabelsResult[i].Label
	}

	for _, label := range labels {
		if l, has := labelMap[label.Title]; has {
			*label = *l // Use existing label
			continue
		}
		// Label does not exist, create it
		// label.Create will set CreatedByID using authUser
		err = label.Create(s, authUser) // authUser must be web.Auth
		if err != nil {
			return err
		}
	}
	return task.UpdateTaskLabels(s, authUser, labels)
}

// removeStaleRelations removes relations that are no longer in the VTODO.
func removeStaleRelations(s *xorm.Session, authUser web.Auth, task *models.Task, newRelations map[models.RelationKind][]*models.Task) (err error) {
	existingTask := &models.Task{ID: task.ID}
	// Read existing relations for the task
	// ReadOne might be heavy; if there's a lighter way to get just RelatedTasks, use it.
	err = existingTask.ReadOne(s, authUser) // authUser must be web.Auth
	if err != nil {
		return
	}

	for relationKind, relatedTasksInDB := range existingTask.RelatedTasks {
		for _, relatedTaskInDB := range relatedTasksInDB {
			stillExists := false
			if newRelationsForKind, ok := newRelations[relationKind]; ok {
				for _, newRelation := range newRelationsForKind {
					if newRelation.UID == relatedTaskInDB.UID {
						stillExists = true
						break
					}
				}
			}
			if !stillExists {
				rel := models.TaskRelation{
					TaskID:       task.ID,
					OtherTaskID:  relatedTaskInDB.ID, // Need ID of related task
					RelationKind: relationKind,
				}
				// Deleting relation might need OtherTaskID, ensure relatedTaskInDB has ID
				if relatedTaskInDB.ID == 0 {
					// If related task from DB doesn't have ID, try to fetch it by UID
					// This situation should ideally not happen if ReadOne populates RelatedTasks correctly.
					tempRelated, tempErr := models.GetTaskSimpleByUUID(s, relatedTaskInDB.UID)
					if tempErr == nil && tempRelated != nil {
						rel.OtherTaskID = tempRelated.ID
					} else {
						log.Warningf("Cannot find ID for related task UID %s to remove relation", relatedTaskInDB.UID)
						continue // Skip deleting this relation if we can't identify it properly
					}
				}
				err = rel.Delete(s, authUser) // authUser must be web.Auth
				if err != nil {
					return
				}
			}
		}
	}
	return
}

// persistRelations persists new relations from VTODO.
func persistRelations(s *xorm.Session, authUser web.Auth, task *models.Task, newRelations map[models.RelationKind][]*models.Task) (err error) {
	err = removeStaleRelations(s, authUser, task, newRelations)
	if err != nil {
		return err
	}

	for relationType, relatedTasksInVTODO := range newRelations {
		for _, relatedTaskFromVTODO := range relatedTasksInVTODO {
			var targetRelatedTask *models.Task
			createDummy := false

			relatedTaskInDB, errGT := models.GetTaskSimpleByUUID(s, relatedTaskFromVTODO.UID)
			if errGT != nil {
				if models.IsErrTaskDoesNotExist(errGT) {
					targetRelatedTask = relatedTaskFromVTODO
					createDummy = true
				} else {
					return errGT // Other error fetching task
				}
			} else {
				targetRelatedTask = relatedTaskInDB
			}

			if createDummy {
				targetRelatedTask.ProjectID = task.ProjectID // Create dummy in the same project
				if targetRelatedTask.Title == "" { // Ensure dummy task has a title
					targetRelatedTask.Title = "DUMMY-UID-" + targetRelatedTask.UID
				}
				// Set CreatedBy for the dummy task if needed by Task.Create
				// targetRelatedTask.CreatedBy = authUser.GetID() // Or however CreatedBy is determined
				err = targetRelatedTask.Create(s, authUser) // authUser must be web.Auth
				if err != nil {
					return err
				}
			}

			rel := models.TaskRelation{
				TaskID:       task.ID,
				OtherTaskID:  targetRelatedTask.ID,
				RelationKind: relationType,
			}
			err = rel.Create(s, authUser) // authUser must be web.Auth
			if err != nil && !models.IsErrRelationAlreadyExists(err) { // Ignore if relation already exists
				return err
			}
		}
	}
	return nil
}

// ListPrincipalPropertiesAndCalendars handles PROPFIND for principal discovery.
// It lists principal-specific properties and then the user's calendars.
func ListPrincipalPropertiesAndCalendars(c echo.Context, currentUser *userpkg.User) error {
	s := db.NewSession()
	defer s.Close()

	// Get user's projects to list as calendars
	projectModel := models.Project{}
	projectsInterface, _, _, err := projectModel.ReadAll(s, currentUser, "", -1, 0)
	if err != nil {
		log.Errorf("Error reading projects for user %d: %v", currentUser.ID, err)
		return c.NoContent(http.StatusInternalServerError)
	}
	projects, ok := projectsInterface.([]*models.Project)
	if !ok {
		log.Errorf("Failed to cast projects to []*models.Project")
		return c.NoContent(http.StatusInternalServerError)
	}

	if err := s.Commit(); err != nil {
		log.Errorf("Error committing transaction: %v", err)
		return c.NoContent(http.StatusInternalServerError)
	}

	var buf bytes.Buffer
	buf.WriteString(xml.Header)
	enc := xml.NewEncoder(&buf)
	start := xml.StartElement{
		Name: xml.Name{Local: "d:multistatus"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "xmlns:d"}, Value: davNS},
			{Name: xml.Name{Local: "xmlns:c"}, Value: caldavNS},
			{Name: xml.Name{Local: "xmlns:cs"}, Value: "http://calendarserver.org/ns/"}, // Common CalDAV server namespace
		},
	}
	_ = enc.EncodeToken(start)

	// 1. Response for the principal itself (e.g., /dav/user/)
	principalURL := DavBasePath + "user/" // Assuming a fixed principal path structure
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:response"}})
	_ = enc.EncodeElement(principalURL, xml.StartElement{Name: xml.Name{Local: "d:href"}})
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:propstat"}})
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:prop"}})

	// Principal properties
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:resourcetype"}})
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:principal"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:principal"}})
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Local: "d:collection"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:collection"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:resourcetype"}})

	_ = enc.EncodeElement(currentUser.Username, xml.StartElement{Name: xml.Name{Local: "d:displayname"}})

	// cs:calendar-home-set
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Space: "http://calendarserver.org/ns/", Local: "calendar-home-set"}})
	_ = enc.EncodeElement(ProjectBasePath + "/", xml.StartElement{Name: xml.Name{Local: "d:href"}}) // Path to the collection of calendars
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Space: "http://calendarserver.org/ns/", Local: "calendar-home-set"}})

	// c:calendar-user-address-set
	_ = enc.EncodeToken(xml.StartElement{Name: xml.Name{Space: caldavNS, Local: "calendar-user-address-set"}})
	_ = enc.EncodeElement(fmt.Sprintf("mailto:%s", currentUser.Email), xml.StartElement{Name: xml.Name{Local: "d:href"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Space: caldavNS, Local: "calendar-user-address-set"}})

	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:prop"}})
	_ = enc.EncodeElement("HTTP/1.1 200 OK", xml.StartElement{Name: xml.Name{Local: "d:status"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:propstat"}})
	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:response"}})

	// 2. Response for each calendar (project)
	for _, p := range projects {
		projectURL := fmt.Sprintf("%s/%s/", ProjectBasePath, strconv.FormatInt(p.ID, 10))
		// Using addPropfindResponse for consistency, ensure it includes all necessary properties for a calendar collection
		addPropfindResponse(enc, projectURL, p.Title, true, p.Updated, projectEtag(p))
	}

	_ = enc.EncodeToken(xml.EndElement{Name: xml.Name{Local: "d:multistatus"}})
	_ = enc.Flush()

	return c.XMLBlob(http.StatusMultiStatus, buf.Bytes())
}
