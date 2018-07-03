package v1

import (
	"git.kolaente.de/konrad/list/models"
	"github.com/labstack/echo"
	"net/http"
	"strconv"
	"fmt"
)

func AddNamespace(c echo.Context) error {
	// swagger:operation PUT /namespaces namespaces addNamespace
	// ---
	// summary: Creates a new namespace owned by the currently logged in user
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/Namespace"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return addOrUpdateNamespace(c)
}

func UpdateNamespace(c echo.Context) error {
	// swagger:operation POST /namespaces/{namespaceID} namespaces upadteNamespace
	// ---
	// summary: Updates a namespace
	// consumes:
	// - application/json
	// produces:
	// - application/json
	// parameters:
	// - name: namespaceID
	//   in: path
	//   description: ID of the namespace to update
	//   type: string
	//   required: true
	// - name: body
	//   in: body
	//   schema:
	//     "$ref": "#/definitions/Namespace"
	// responses:
	//   "200":
	//     "$ref": "#/responses/Namespace"
	//   "400":
	//     "$ref": "#/responses/Message"
	//   "403":
	//     "$ref": "#/responses/Message"
	//   "500":
	//     "$ref": "#/responses/Message"

	return addOrUpdateNamespace(c)
}

// AddOrUpdateNamespace Adds or updates a new namespace
func addOrUpdateNamespace(c echo.Context) error {

	// Get the namespace
	var namespace *models.Namespace

	if err := c.Bind(&namespace); err != nil {
		return c.JSON(http.StatusBadRequest, models.Message{"No namespace model provided."})
	}

	// Check if we have an ID other than the one in the struct
	id := c.Param("id")
	if id != "" {
		// Make int
		namespaceID, err := strconv.ParseInt(id, 10, 64)

		if err != nil {
			return c.JSON(http.StatusBadRequest, models.Message{"Invalid ID."})
		}
		namespace.ID = namespaceID
	}

	// Check if the namespace exists
	// ID = 0 means new namespace, no error
	if namespace.ID != 0 {
		_, err := models.GetNamespaceByID(namespace.ID)
		if err != nil {
			if models.IsErrNamespaceDoesNotExist(err) {
				return c.JSON(http.StatusBadRequest, models.Message{"The namespace does not exist."})
			}
			return c.JSON(http.StatusInternalServerError, models.Message{"Could not check if the namespace exists."})
		}
	}

	// Get the current user for later checks
	user, err := models.GetCurrentUser(c)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
	}
	namespace.Owner = user

	// update or create...
	if namespace.ID == 0 {
		err = models.CreateOrUpdateNamespace(namespace)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	} else {
		// Check if the user has admin access to the namespace
		oldNamespace, err := models.GetNamespaceByID(namespace.ID)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
		has, err := user.IsNamespaceAdmin(&oldNamespace)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
		if !has {
			return c.JSON(http.StatusForbidden, models.Message{"You need to be namespace admin to edit a namespace."})
		}

		fmt.Println(namespace)

		err = models.CreateOrUpdateNamespace(namespace)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, models.Message{"An error occured."})
		}
	}

	return c.JSON(http.StatusOK, namespace)
}
