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

	"code.vikunja.io/api/pkg/files"
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
	// Optional custom step list. When provided, these steps are used instead
	// of the chain's saved steps. This allows the caller to add, remove, or
	// modify steps at creation time without changing the template.
	CustomSteps []*TaskChainStep `json:"custom_steps"`

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
// @Description If custom_steps is provided, those steps are used instead of the chain's
// @Description saved steps, allowing on-the-fly add/remove/edit without changing the template.
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

	// Load the chain (for metadata like title, and saved step attachments)
	tc := &TaskChain{ID: tfc.ChainID}
	err = tc.ReadOne(s, doer)
	if err != nil {
		return err
	}

	// Decide which steps to use: custom overrides or saved template steps
	steps := tc.Steps
	if len(tfc.CustomSteps) > 0 {
		steps = tfc.CustomSteps
		log.Debugf("Using %d custom steps (template has %d)", len(steps), len(tc.Steps))
	}

	if len(steps) == 0 {
		return &ErrTaskChainHasNoSteps{ID: tfc.ChainID}
	}

	// Use now as anchor if not specified
	anchorDate := tfc.AnchorDate
	if anchorDate.IsZero() {
		anchorDate = time.Now()
	}

	log.Debugf("Creating %d tasks from chain %d (anchor: %s) in project %d",
		len(steps), tfc.ChainID, anchorDate.Format("2006-01-02"), tfc.TargetProjectID)

	// Build a map of saved step IDs for template attachment lookup
	savedStepBySequence := map[int]*TaskChainStep{}
	for _, savedStep := range tc.Steps {
		savedStepBySequence[savedStep.Sequence] = savedStep
	}

	// Create all tasks — offsets are relative to the previous step
	createdTasks := make([]*Task, 0, len(steps))
	cumulativeOffset := 0
	for stepIndex, step := range steps {
		// Each step's offset_days is relative to the previous step's start
		cumulativeOffset += step.OffsetDays

		// Calculate dates
		startDate := anchorDate.AddDate(0, 0, cumulativeOffset)
		endDate := startDate.AddDate(0, 0, step.DurationDays)

		// Build title
		title := step.Title
		if title == "" {
			title = fmt.Sprintf("Step %d", stepIndex+1)
		}
		if tfc.TitlePrefix != "" {
			// Auto-append a separator if the prefix doesn't end with one
			p := tfc.TitlePrefix
			lastChar := p[len(p)-1]
			if lastChar != ' ' && lastChar != '_' && lastChar != '-' && lastChar != ':' && lastChar != '/' && lastChar != '.' {
				p = p + "_"
			}
			title = p + title
		}

		// Build description
		description := step.Description

		chainNote := `<p><small style="color:#888">🔗 Chain: ` + tc.Title +
			` (step ` + fmt.Sprintf("%d", stepIndex+1) + `/` + fmt.Sprintf("%d", len(steps)) + `)</small></p>`
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

		// Copy step template attachments to the new task.
		// Look up by step ID if the step came from the template, or by sequence
		// match if using custom steps (to preserve template attachments).
		var lookupStepID int64
		if step.ID > 0 {
			lookupStepID = step.ID
		} else if savedStep, ok := savedStepBySequence[stepIndex]; ok {
			lookupStepID = savedStep.ID
		}

		if lookupStepID > 0 {
			stepAttachments := []*TaskChainStepAttachment{}
			if err := s.Where("step_id = ?", lookupStepID).Find(&stepAttachments); err == nil {
				for _, att := range stepAttachments {
					srcFile := &files.File{ID: att.FileID}
					if err := srcFile.LoadFileMetaByID(); err != nil {
						continue
					}
					if err := srcFile.LoadFileByID(); err != nil {
						continue
					}
					copiedFile, err := files.Create(srcFile.File, att.FileName, uint64(srcFile.Size), doer)
					if err != nil {
						continue
					}
					taskAtt := &TaskAttachment{
						TaskID:      newTask.ID,
						FileID:      copiedFile.ID,
						CreatedByID: doer.GetID(),
					}
					_, _ = s.Insert(taskAtt)
				}
			}
		}
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
