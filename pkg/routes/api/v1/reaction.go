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

package v1

import (
	"net/http"
	"strconv"

	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/services"
	"code.vikunja.io/api/pkg/user"
	"code.vikunja.io/api/pkg/web/handler"
	"github.com/labstack/echo/v4"
	"xorm.io/xorm"
)

// ReactionRoutes defines all reaction API routes with their explicit permission scopes.
var ReactionRoutes = []APIRoute{
	{Method: "GET", Path: "/:entitykind/:entityid/reactions", Handler: handler.WithDBAndUser(getAllReactionsLogic, false), PermissionScope: "read_all"},
	{Method: "PUT", Path: "/:entitykind/:entityid/reactions", Handler: handler.WithDBAndUser(createReactionLogic, true), PermissionScope: "create"},
	{Method: "POST", Path: "/:entitykind/:entityid/reactions/delete", Handler: handler.WithDBAndUser(deleteReactionLogic, true), PermissionScope: "delete"},
}

// RegisterReactions registers all reaction routes
func RegisterReactions(a *echo.Group) {
	registerRoutes(a, ReactionRoutes)
}

// getAllReactionsLogic handles retrieving all reactions for an entity (task or comment).
//
// @Summary Get all reactions
// @Description Returns all reactions for an entity grouped by reaction value
// @tags reactions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entitykind path string true "Entity kind (tasks or comments)"
// @Param entityid path int true "Entity ID"
// @Success 200 {object} models.ReactionMap "The reactions grouped by value"
// @Failure 400 {object} models.Message "Invalid entity kind"
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 404 {object} web.HTTPError "The entity does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{entitykind}/{entityid}/reactions [get]
func getAllReactionsLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse entity ID
	entityID, err := strconv.ParseInt(c.Param("entityid"), 10, 64)
	if err != nil {
		return &models.ErrInvalidEntityID{ID: c.Param("entityid")}
	}

	// Parse and validate entity kind
	entityKind, err := parseEntityKind(c.Param("entitykind"))
	if err != nil {
		return err
	}

	// Check access permissions based on entity type
	if err := checkEntityAccess(s, u, entityID, entityKind); err != nil {
		return err
	}

	// Get reactions from service
	reactionsService := services.NewReactionsService(s.Engine())
	reactions, err := reactionsService.GetAll(s, entityID, entityKind)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, reactions)
}

// createReactionLogic handles creating a new reaction.
//
// @Summary Create a reaction
// @Description Add a reaction to an entity. Will do nothing if the reaction already exists (idempotent).
// @tags reactions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entitykind path string true "Entity kind (tasks or comments)"
// @Param entityid path int true "Entity ID"
// @Param reaction body models.Reaction true "The reaction to create"
// @Success 200 {object} models.Reaction "The created reaction"
// @Failure 400 {object} models.Message "Invalid entity kind or reaction data"
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 404 {object} web.HTTPError "The entity does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{entitykind}/{entityid}/reactions [put]
func createReactionLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse entity ID
	entityID, err := strconv.ParseInt(c.Param("entityid"), 10, 64)
	if err != nil {
		return &models.ErrInvalidEntityID{ID: c.Param("entityid")}
	}

	// Parse and validate entity kind
	entityKind, err := parseEntityKind(c.Param("entitykind"))
	if err != nil {
		return err
	}

	// Check access permissions based on entity type
	if err := checkEntityAccess(s, u, entityID, entityKind); err != nil {
		return err
	}

	// Parse reaction from request body
	reaction := &models.Reaction{}
	if err := c.Bind(reaction); err != nil {
		return &models.ErrInvalidReactionValue{Value: ""}
	}

	// Set entity details
	reaction.EntityID = entityID
	reaction.EntityKind = entityKind

	// Validate reaction value
	if reaction.Value == "" || len(reaction.Value) > 20 {
		return &models.ErrInvalidReactionValue{Value: reaction.Value}
	}

	// Create reaction via service
	reactionsService := services.NewReactionsService(s.Engine())
	if err := reactionsService.Create(s, reaction, u); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, reaction)
}

// deleteReactionLogic handles deleting a reaction.
//
// @Summary Delete a reaction
// @Description Removes the user's reaction from an entity. Only the user who created the reaction can delete it.
// @tags reactions
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param entitykind path string true "Entity kind (tasks or comments)"
// @Param entityid path int true "Entity ID"
// @Param reaction body models.Reaction true "The reaction to delete (only value field is required)"
// @Success 200 {object} models.Message "The reaction was successfully removed"
// @Failure 400 {object} models.Message "Invalid entity kind or reaction data"
// @Failure 403 {object} web.HTTPError "The user does not have access to the entity"
// @Failure 404 {object} web.HTTPError "The entity does not exist"
// @Failure 500 {object} models.Message "Internal error"
// @Router /{entitykind}/{entityid}/reactions/delete [post]
func deleteReactionLogic(s *xorm.Session, u *user.User, c echo.Context) error {
	// Parse entity ID
	entityID, err := strconv.ParseInt(c.Param("entityid"), 10, 64)
	if err != nil {
		return &models.ErrInvalidEntityID{ID: c.Param("entityid")}
	}

	// Parse and validate entity kind
	entityKind, err := parseEntityKind(c.Param("entitykind"))
	if err != nil {
		return err
	}

	// Check access permissions based on entity type
	if err := checkEntityAccess(s, u, entityID, entityKind); err != nil {
		return err
	}

	// Parse reaction from request body
	reaction := &models.Reaction{}
	if err := c.Bind(reaction); err != nil {
		return &models.ErrInvalidReactionValue{Value: ""}
	}

	if reaction.Value == "" {
		return &models.ErrInvalidReactionValue{Value: ""}
	}

	// Delete reaction via service
	reactionsService := services.NewReactionsService(s.Engine())
	if err := reactionsService.Delete(s, entityID, u.ID, reaction.Value, entityKind); err != nil {
		return err
	}

	return c.JSON(http.StatusOK, models.Message{Message: "The reaction was successfully removed."})
}

// parseEntityKind converts a string entity kind to ReactionKind
func parseEntityKind(kindStr string) (models.ReactionKind, error) {
	switch kindStr {
	case "tasks":
		return models.ReactionKindTask, nil
	case "comments":
		return models.ReactionKindComment, nil
	default:
		return 0, &models.ErrInvalidReactionEntityKind{Kind: kindStr}
	}
}

// checkEntityAccess verifies that the user has access to the specified entity
func checkEntityAccess(s *xorm.Session, u *user.User, entityID int64, entityKind models.ReactionKind) error {
	switch entityKind {
	case models.ReactionKindTask:
		// Check if task exists and user has access
		task := &models.Task{ID: entityID}
		can, _, err := task.CanRead(s, u)
		if err != nil {
			return err
		}
		if !can {
			return models.ErrGenericForbidden{}
		}
		// Verify task exists by trying to get it
		exists, err := s.Where("id = ?", entityID).Exist(&models.Task{})
		if err != nil {
			return err
		}
		if !exists {
			return models.ErrTaskDoesNotExist{ID: entityID}
		}
	case models.ReactionKindComment:
		// Check if comment exists and user has access
		comment := &models.TaskComment{ID: entityID}
		can, _, err := comment.CanRead(s, u)
		if err != nil {
			return err
		}
		if !can {
			return models.ErrGenericForbidden{}
		}
		// Verify comment exists
		exists, err := s.Where("id = ?", entityID).Exist(&models.TaskComment{})
		if err != nil {
			return err
		}
		if !exists {
			return models.ErrTaskCommentDoesNotExist{ID: entityID}
		}
	}
	return nil
}
