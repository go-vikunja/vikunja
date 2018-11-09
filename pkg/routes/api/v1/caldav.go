package v1

import (
	"code.vikunja.io/api/pkg/caldav"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/routes/crud"
	"github.com/labstack/echo"
	"net/http"
	"time"
)

// Caldav returns a caldav-readable format with all tasks
func Caldav(c echo.Context) error {

	// swagger:operation GET /tasks/caldav list caldavTasks
	// ---
	// summary: Get all tasks as caldav
	// responses:
	//   "200":
	//     "$ref": "#/responses/Message"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	// Request basic auth
	user, pass, ok := c.Request().BasicAuth()

	// Check credentials
	creds := &models.UserLogin{
		Username: user,
		Password: pass,
	}
	u, err := models.CheckUserCredentials(creds)

	if !ok || err != nil {
		c.Response().Header().Set("WWW-Authenticate", `Basic realm="Vikunja cal"`)
		return c.String(http.StatusUnauthorized, "Unauthorized.")
	}

	// Get all tasks for that user
	tasks, err := models.GetTasksByUser("", &u, -1)
	if err != nil {
		return crud.HandleHTTPError(err)
	}

	hour := int64(time.Hour.Seconds())
	var caldavTasks []*caldav.Event
	for _, t := range tasks {
		if t.DueDateUnix != 0 {
			event := &caldav.Event{
				Summary:       t.Text,
				Description:   t.Description,
				UID:           "",
				TimestampUnix: t.Updated,
				StartUnix:     t.DueDateUnix,
				EndUnix:       t.DueDateUnix + hour,
			}

			if t.ReminderUnix != 0 {
				event.Alarms = append(event.Alarms, caldav.Alarm{TimeUnix: t.ReminderUnix})
			}

			caldavTasks = append(caldavTasks, event)
		}
	}

	caldavConfig := &caldav.Config{
		Name:   "Vikunja Calendar for " + u.Username,
		ProdID: "Vikunja Todo App",
	}

	return c.String(http.StatusOK, caldav.ParseEvents(caldavConfig, caldavTasks))
}
