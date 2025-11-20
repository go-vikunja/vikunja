package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/auth"
	"github.com/labstack/echo/v4"
)

func MarkTaskAsRead(c echo.Context) error {
	s := db.NewSession()
	defer s.Close()

	a, err := auth.GetAuthFromClaims(c)
	if err != nil {
		return err
	}

	projecttaskParam := c.Param("projecttask")

	projecttask, err := strconv.Atoi(projecttaskParam)
	if err != nil {
		return err
	}

	task := &models.Task{
		ID: int64(projecttask),
	}

	err = task.MarkTaskAsRead(s, a)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "success"})
}
