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

package apiv2

import (
	"context"
	"fmt"
	"net/http"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/web/handler"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/conditional"
)

// RiskListQueryParams is the filter, sort and search block of every risk list. It must stay
// EXPORTED: Huma only binds the params of an anonymous embed whose type name is exported.
type RiskListQueryParams struct {
	ListParams
	Status    []string `query:"status,explode" enum:"open,mitigating,accepted,closed" doc:"Only risks with one of these statuses. Repeatable. Without it risks of every status are returned, closed ones included."`
	Rating    []string `query:"rating,explode" enum:"low,medium,high,critical" doc:"Only risks with one of these ratings (low 1-4, medium 5-9, high 10-16, critical 17-25 of probability times impact). Repeatable."`
	OwnerID   []int64  `query:"owner_id,explode" doc:"Only risks owned by one of these users. Repeatable. 0 matches risks without an owner."`
	Category  []string `query:"category,explode" doc:"Only risks with one of these categories, compared without regard to case. Repeatable."`
	Overdue   bool     `query:"overdue" doc:"If true, only risks that are not closed and whose due date has passed."`
	SortBy    []string `query:"sort_by,explode" enum:"score,probability,impact,due_date,status,title,created,updated,id" doc:"What to sort by. Repeatable, pairs positionally with order_by. Defaults to score, worst first."`
	OrderBy   []string `query:"order_by,explode" enum:"asc,desc" doc:"The sort order per sort_by field. Repeatable, defaults to asc."`
	projectID int64
}

type riskListAllInput struct {
	RiskListQueryParams
	// Declared on the input, they only make sense across projects.
	ProjectID       []int64 `query:"project_id,explode" doc:"Only risks of these projects. Repeatable. Projects you cannot read are never included, whatever is asked for."`
	IncludeArchived bool    `query:"include_archived" doc:"If true, risks of archived projects are listed too. Archived projects are left out of a list over all projects by default; choosing projects with project_id always shows them."`
}

type riskListProjectInput struct {
	ProjectID int64 `path:"project_id" doc:"The numeric id of the project."`
	RiskListQueryParams
}

type riskListBody struct {
	Body Paginated[*models.Risk]
}

type riskHistoryBody struct {
	Body Paginated[*models.RiskStatusHistory]
}

// riskReadBody is the read shape. Update takes it as well, so that AutoPatch's GET to PUT echo of
// max_permission validates.
type riskReadBody struct {
	models.Risk
	MaxPermission models.Permission `json:"max_permission" readOnly:"true" doc:"The maximum permission the requesting user has on this risk (0=read, 1=read/write, 2=admin)."`
}

type riskStatusRequest struct {
	Status string `json:"status" enum:"open,mitigating,accepted,closed" doc:"The new status. Any status can follow any other, so a closed risk can be reopened by moving it back to open (or any other status)."`
	Note   string `json:"note,omitempty" maxLength:"20000" doc:"An optional note that is kept in the history. When closing, it is stored as the resolution of the risk."`
}

func init() { AddRouteRegistrar(RegisterRiskRoutes) }

// RegisterRiskRoutes registers the risk register of projects. Access follows the project: read
// access reads, write access creates, edits, changes the status of and deletes.
func RegisterRiskRoutes(api huma.API) {
	tags := []string{"risks"}

	Register(api, huma.Operation{
		OperationID: "risks-list",
		Summary:     "List risks of all projects",
		Description: "Returns the risks of every project you can read, paginated, worst first by default. Narrow it with status, rating, owner_id, category, overdue and project_id, search with q. Risks of archived projects are only included with include_archived or when their project is chosen with project_id. Link shares see none.",
		Method:      http.MethodGet,
		Path:        "/risks",
		Tags:        tags,
	}, risksList)

	Register(api, huma.Operation{
		OperationID: "project-risks-list",
		Summary:     "List the risks of a project",
		Description: "Returns the risks of one project, paginated, with the same filters as the list over all projects. An inaccessible or unknown project yields an empty list, not an error.",
		Method:      http.MethodGet,
		Path:        "/projects/{project_id}/risks",
		Tags:        tags,
	}, projectRisksList)

	Register(api, huma.Operation{
		OperationID: "project-risks-create",
		Summary:     "Create a risk",
		Description: "Adds a risk to the project. It starts as open, whatever status is sent. Probability and impact default to 3. The owner must be an active user who can read the project. Needs write access to the project.",
		Method:      http.MethodPost,
		Path:        "/projects/{project_id}/risks",
		Tags:        tags,
	}, risksCreate)

	Register(api, huma.Operation{
		OperationID: "risks-read",
		Summary:     "Get a risk",
		Description: "Returns one risk. Sends an ETag; pass it as If-None-Match on a later read to get a 304 Not Modified.",
		Method:      http.MethodGet,
		Path:        "/risks/{id}",
		Tags:        tags,
	}, risksRead)

	Register(api, huma.Operation{
		OperationID: "risks-update",
		Summary:     "Update a risk",
		Description: "Replaces the editable fields: title, description, category, probability, impact, owner, mitigation, contingency, identified and due date. The project and the status are not changed by it, whatever is sent: use the status operation, so that the history stays complete. PUT replaces all editable fields; use PATCH for a partial update. Needs write access to the project.",
		Method:      http.MethodPut,
		Path:        "/risks/{id}",
		Tags:        tags,
	}, risksUpdate)

	Register(api, huma.Operation{
		OperationID: "risks-delete",
		Summary:     "Delete a risk",
		Description: "Deletes the risk and its history. To only mark a risk as resolved, close it instead. Needs write access to the project.",
		Method:      http.MethodDelete,
		Path:        "/risks/{id}",
		Tags:        tags,
	}, risksDelete)

	Register(api, huma.Operation{
		OperationID:   "risks-status",
		Summary:       "Change the status of a risk",
		Description:   "Moves a risk to open, mitigating, accepted or closed, from any status to any other, and records who did it, when and the note. Closing sets the closed date and stores the note as the resolution. Moving a closed risk to another status reopens it and clears the closed date and resolution; the history keeps them. The same status again changes nothing. Returns 409 if somebody else changed the status at the same moment. Needs write access to the project.",
		Method:        http.MethodPost,
		Path:          "/risks/{id}/status",
		DefaultStatus: http.StatusOK,
		Tags:          tags,
	}, risksStatus)

	Register(api, huma.Operation{
		OperationID: "risks-history",
		Summary:     "List the status changes of a risk",
		Description: "Returns who changed the status of the risk, when, from which status to which, and the note, newest first. Needs read access to the project.",
		Method:      http.MethodGet,
		Path:        "/risks/{id}/history",
		Tags:        tags,
	}, risksHistory)
}

func (p RiskListQueryParams) toRisk() *models.Risk {
	return &models.Risk{
		ProjectID:  p.projectID,
		Statuses:   p.Status,
		Ratings:    p.Rating,
		OwnerIDs:   p.OwnerID,
		Categories: p.Category,
		Overdue:    p.Overdue,
		SortBy:     p.SortBy,
		OrderBy:    p.OrderBy,
	}
}

func riskListResponse(result any, total int64, page, perPage int) (*riskListBody, error) {
	items, ok := result.([]*models.Risk)
	if !ok {
		return nil, fmt.Errorf("risks.ReadAll returned unexpected type %T (expected []*models.Risk)", result)
	}
	return &riskListBody{Body: NewPaginated(items, total, page, perPage)}, nil
}

func risksList(ctx context.Context, in *riskListAllInput) (*riskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	m := in.RiskListQueryParams.toRisk()
	m.ProjectIDs = in.ProjectID
	m.IncludeArchived = in.IncludeArchived
	result, _, total, err := handler.DoReadAll(ctx, m, a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return riskListResponse(result, total, in.Page, in.PerPage)
}

func projectRisksList(ctx context.Context, in *riskListProjectInput) (*riskListBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.RiskListQueryParams.projectID = in.ProjectID
	result, _, total, err := handler.DoReadAll(ctx, in.RiskListQueryParams.toRisk(), a, in.Q, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return riskListResponse(result, total, in.Page, in.PerPage)
}

func risksCreate(ctx context.Context, in *struct {
	ProjectID int64 `path:"project_id" doc:"The numeric id of the project."`
	Body      models.Risk
}) (*singleBody[models.Risk], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	in.Body.ProjectID = in.ProjectID // the URL decides, not the body
	if err := handler.DoCreate(ctx, &in.Body, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Risk]{Body: &in.Body}, nil
}

func risksRead(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the risk."`
	conditional.Params
}) (*singleReadBody[riskReadBody], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	risk := &models.Risk{ID: in.ID}
	maxPermission, err := handler.DoReadOne(ctx, risk, a)
	if err != nil {
		return nil, translateDomainError(err)
	}
	body := &riskReadBody{Risk: *risk, MaxPermission: models.Permission(maxPermission)}
	return conditionalReadResponse(&in.Params, body, risk.Updated, maxPermission)
}

func risksUpdate(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric id of the risk."`
	Body riskReadBody
}) (*singleBody[models.Risk], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	risk := &in.Body.Risk
	risk.ID = in.ID // the URL decides, not the body
	if err := handler.DoUpdate(ctx, risk, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &singleBody[models.Risk]{Body: risk}, nil
}

func risksDelete(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the risk."`
}) (*emptyBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}
	if err := handler.DoDelete(ctx, &models.Risk{ID: in.ID}, a); err != nil {
		return nil, translateDomainError(err)
	}
	return &emptyBody{}, nil
}

// risksStatus is a custom action: it owns its session, and the model checks write access itself.
func risksStatus(ctx context.Context, in *struct {
	ID   int64 `path:"id" doc:"The numeric id of the risk."`
	Body riskStatusRequest
}) (*singleBody[models.Risk], error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewSession()
	defer s.Close()

	risk, err := models.ChangeRiskStatus(s, a, in.ID, in.Body.Status, in.Body.Note)
	if err != nil {
		_ = s.Rollback()
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	if err := s.Commit(); err != nil {
		events.CleanupPending(s)
		return nil, translateDomainError(err)
	}
	events.DispatchPending(ctx, s)
	return &singleBody[models.Risk]{Body: risk}, nil
}

func risksHistory(ctx context.Context, in *struct {
	ID int64 `path:"id" doc:"The numeric id of the risk."`
	ListParams
}) (*riskHistoryBody, error) {
	a, err := authFromCtx(ctx)
	if err != nil {
		return nil, err
	}

	s := db.NewReadSession()
	defer s.Close()

	entries, total, err := models.ListRiskHistory(s, a, in.ID, in.Page, in.PerPage)
	if err != nil {
		return nil, translateDomainError(err)
	}
	return &riskHistoryBody{Body: NewPaginated(entries, total, in.Page, in.PerPage)}, nil
}
