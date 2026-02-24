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

package models

import (
	"fmt"
	"time"

	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/web"

	"xorm.io/xorm"
)

// TaskFromChain holds everything needed to create tasks from a chain template
type TaskFromChain struct {
	// The chain id to create tasks from
	ChainID int64 `json:"-" param:"chain"`
	// The target project id where tasks will be created
	TargetProjectID int64 `json:"target_project_id"`
	// The anchor date — the start date of the first task in the chain.
	// All other task dates are calculated relative to this date.
	AnchorDate time.Time `json:"anchor_date"`
	// Optional title prefix for all tasks (e.g. "Batch #42 - ")
	TitlePrefix string `json:"title_prefix"`

	// The created tasks (returned after creation)
	Tasks []*Task `json:"created_tasks,omitempty"`

	web.Permissions `json:"-"`
	web.CRUDable    `json:"-"`
}

// CanCreate checks if the user can create tasks from this chain
func (tfc *TaskFromChain) CanCreate(s *xorm.Session, a web.Auth) (bool, error) {
	// User must own the chain
	tc := &TaskChain{}
	exists, err := s.Where("id = ?", tfc.ChainID).Get(tc)
	if err != nil {
		return false, err
	}
	if !exists {
		return false, &ErrTaskChainDoesNotExist{ID: tfc.ChainID}
	}
	if tc.OwnerID != a.GetID() {
		return false, nil
	}

	// User must have write access to the target project
	targetProject := &Project{ID: tfc.TargetProjectID}
	return targetProject.CanUpdate(s, a)
}

// Create creates all tasks from a chain template with calculated dates and relations
// @Summary Create tasks from a chain template
// @Description Creates all tasks defined in the chain template with dates calculated
// @Description relative to the anchor date. Tasks are linked with precedes/follows relations.
// @tags task_chains
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param chainID path int true "The chain ID"
// @Param chain body models.TaskFromChain true "The target project, anchor date, and optional title prefix"
// @Success 201 {object} models.TaskFromChain "The created tasks."
// @Failure 400 {object} web.HTTPError "Invalid request."
// @Failure 403 {object} web.HTTPError "Not authorized."
// @Failure 500 {object} models.Message "Internal error"
// @Router /taskchains/{chainID}/tasks [put]
func (tfc *TaskFromChain) Create(s *xorm.Session, doer web.Auth) (err error) {

	// Load the chain with steps
	tc := &TaskChain{ID: tfc.ChainID}
	err = tc.ReadOne(s, doer)
	if err != nil {
		return err
	}

	if len(tc.Steps) == 0 {
		return &ErrTaskChainHasNoSteps{ID: tfc.ChainID}
	}

	// Use now as anchor if not specified
	anchorDate := tfc.AnchorDate
	if anchorDate.IsZero() {
		anchorDate = time.Now()
	}

	log.Debugf("Creating %d tasks from chain %d (anchor: %s) in project %d",
		len(tc.Steps), tfc.ChainID, anchorDate.Format("2006-01-02"), tfc.TargetProjectID)

	// Create all tasks — offsets are relative to the previous step
	createdTasks := make([]*Task, 0, len(tc.Steps))
	cumulativeOffset := 0
	for _, step := range tc.Steps {
		// Each step's offset_days is relative to the previous step's start
		cumulativeOffset += step.OffsetDays

		// Calculate dates
		startDate := anchorDate.AddDate(0, 0, cumulativeOffset)
		endDate := startDate.AddDate(0, 0, step.DurationDays)

		// Build title
		title := step.Title
		if tfc.TitlePrefix != "" {
			title = tfc.TitlePrefix + title
		}

		// Build description with chain info
		description := step.Description
		chainNote := `<p><small style="color:#888">🔗 Chain: ` + tc.Title +
			` (step ` + fmt.Sprintf("%d", step.Sequence+1) + `/` + fmt.Sprintf("%d", len(tc.Steps)) + `)</small></p>`
		if description != "" {
			description = description + chainNote
		} else {
			description = chainNote
		}

		newTask := &Task{
			Title:       title,
			Description: description,
			StartDate:   startDate,
			EndDate:     endDate,
			Priority:    step.Priority,
			HexColor:    step.HexColor,
			ProjectID:   tfc.TargetProjectID,
		}

		err = createTask(s, newTask, doer, false, true)
		if err != nil {
			return err
		}

		log.Debugf("Created chain task %d: %s (start: %s, end: %s)",
			newTask.ID, newTask.Title, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))

		// Apply labels from step
		if len(step.LabelIDs) > 0 {
			for _, labelID := range step.LabelIDs {
				lt := &LabelTask{
					TaskID:  newTask.ID,
					LabelID: labelID,
				}
				if _, insertErr := s.Insert(lt); insertErr != nil {
					log.Debugf("Could not add label %d to chain task %d: %v", labelID, newTask.ID, insertErr)
					continue
				}
			}
		}

		createdTasks = append(createdTasks, newTask)
	}

	// Create precedes/follows relations between consecutive tasks (both directions)
	for i := 0; i < len(createdTasks)-1; i++ {
		forwardRelation := &TaskRelation{
			TaskID:       createdTasks[i].ID,
			OtherTaskID:  createdTasks[i+1].ID,
			RelationKind: RelationKindPreceeds,
			CreatedByID:  doer.GetID(),
		}
		inverseRelation := &TaskRelation{
			TaskID:       createdTasks[i+1].ID,
			OtherTaskID:  createdTasks[i].ID,
			RelationKind: RelationKindFollows,
			CreatedByID:  doer.GetID(),
		}
		if _, err := s.Insert(&[]*TaskRelation{forwardRelation, inverseRelation}); err != nil {
			log.Debugf("Could not create chain relation between task %d and %d: %v",
				createdTasks[i].ID, createdTasks[i+1].ID, err)
		}
	}

	log.Debugf("Created %d chain relations for chain %d", len(createdTasks)-1, tfc.ChainID)

	// Read back full tasks
	for i, task := range createdTasks {
		_ = task.ReadOne(s, doer)
		createdTasks[i] = task
	}

	tfc.Tasks = createdTasks
	return nil
}

// Stub methods to satisfy CRUDable interface
func (tfc *TaskFromChain) ReadOne(_ *xorm.Session, _ web.Auth) error { return nil }
func (tfc *TaskFromChain) ReadAll(_ *xorm.Session, _ web.Auth, _ string, _ int, _ int) (interface{}, int, int64, error) {
	return nil, 0, 0, nil
}
func (tfc *TaskFromChain) Update(_ *xorm.Session, _ web.Auth) error { return nil }
func (tfc *TaskFromChain) Delete(_ *xorm.Session, _ web.Auth) error { return nil }

// --- Errors ---

type ErrTaskChainHasNoSteps struct {
	ID int64
}

func IsErrTaskChainHasNoSteps(err error) bool {
	_, ok := err.(*ErrTaskChainHasNoSteps)
	return ok
}

func (err *ErrTaskChainHasNoSteps) Error() string {
	return fmt.Sprintf("Task chain has no steps [ID: %d]", err.ID)
}

func (err *ErrTaskChainHasNoSteps) HTTPError() web.HTTPError {
	return web.HTTPError{
		HTTPCode: 400,
		Code:     12011,
		Message:  "This task chain has no steps defined.",
	}
}
