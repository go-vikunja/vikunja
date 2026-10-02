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
	"strings"
	"time"
	"unicode/utf8"

	"code.vikunja.io/api/pkg/user"

	"xorm.io/xorm"
)

// TaskExternalRef remembers which task an integration (a bot user) created for one of its own
// references, so a retried call returns the task instead of creating a second one.
type TaskExternalRef struct {
	ID         int64     `xorm:"bigint autoincr not null unique pk"`
	BotID      int64     `xorm:"bigint not null unique(task_external_ref)"`
	ExternalID string    `xorm:"varchar(250) not null unique(task_external_ref)"`
	TaskID     int64     `xorm:"bigint not null index"`
	Created    time.Time `xorm:"created not null"`
}

// TableName returns the table name of TaskExternalRef.
func (*TaskExternalRef) TableName() string { return "task_external_refs" }

const (
	maxAssignTitleLength      = 250
	maxAssignExternalIDLength = 250
	maxAssignPriority         = 5
)

// AssignTaskByEmailInput is what an integration sends.
type AssignTaskByEmailInput struct {
	AssigneeEmail string
	Title         string
	// Description is already HTML.
	Description string
	DueDate     time.Time
	Priority    int64
	// ProjectID is optional. Without it the task goes to the assignee's inbox.
	ProjectID int64
	// ExternalID is optional and makes the call idempotent per bot.
	ExternalID string
	// AssignorEmail is optional. Without it the assignor is the owner of ProjectID, or the
	// assignee if no project was given.
	AssignorEmail string
}

// AssignTaskByEmailResult is the outcome of AssignTaskByEmail.
type AssignTaskByEmailResult struct {
	// Created is false when the external id was seen before and the existing task is returned.
	Created  bool
	Task     *Task
	Assignee *user.User
	Assignor *user.User
}

// resolveAssignor decides who the task is stored as created by: the given email address, else the
// owner of the project when one was given, else the assignee. The assignor needs no access to the
// project, only to be an active, non-bot user.
func resolveAssignor(s *xorm.Session, in *AssignTaskByEmailInput, assignee *user.User, project *Project) (*user.User, error) {
	if strings.TrimSpace(in.AssignorEmail) != "" {
		u, err := resolveAssigneeByEmail(s, in.AssignorEmail)
		switch {
		case IsErrAssigneeEmailNotFound(err):
			return nil, ErrAssignorNotFound{}
		case IsErrAssigneeNotActive(err):
			return nil, ErrAssignorNotActive{}
		case IsErrAssigneeAmbiguous(err):
			return nil, ErrAssignorAmbiguous{}
		case err != nil:
			return nil, err
		}
		return u, nil
	}

	if in.ProjectID == 0 {
		return assignee, nil
	}

	owner, err := user.GetUserByID(s, project.OwnerID)
	if err != nil {
		if user.IsErrUserStatusError(err) {
			return nil, ErrAssignorNotActive{}
		}
		if user.IsErrUserDoesNotExist(err) {
			return nil, ErrAssignorNotFound{}
		}
		return nil, err
	}
	if owner.IsBot() || owner.Status != user.StatusActive {
		return nil, ErrAssignorNotActive{}
	}
	return owner, nil
}

// resolveAssigneeByEmail finds the one non-bot user with this email address, case-insensitive.
// Email is only unique for local users, so two matches are an error instead of a guess.
func resolveAssigneeByEmail(s *xorm.Session, email string) (*user.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))
	if email == "" {
		return nil, ErrAssigneeEmailNotFound{}
	}

	var users []*user.User
	err := s.
		Where("lower(email) = ? AND (bot_owner_id IS NULL OR bot_owner_id = 0)", email).
		Limit(2).
		Find(&users)
	if err != nil {
		return nil, err
	}

	switch len(users) {
	case 0:
		return nil, ErrAssigneeEmailNotFound{}
	case 1:
		if users[0].Status != user.StatusActive {
			return nil, ErrAssigneeNotActive{}
		}
		return users[0], nil
	default:
		return nil, ErrAssigneeAmbiguous{}
	}
}

// assigneeInbox returns the project a task without a project goes to: the assignee's default
// project, created on the spot if they have none. A default project the assignee does not own
// (they picked a shared project) or an archived one is refused. An integration must not write into
// somebody else's project with the privileges of the inbox path.
func assigneeInbox(s *xorm.Session, assignee *user.User) (*Project, error) {
	if assignee.DefaultProjectID == 0 {
		err := CreateNewProjectForUser(s, assignee)
		if err != nil {
			return nil, err
		}
	}

	project, err := GetProjectSimpleByID(s, assignee.DefaultProjectID)
	if err != nil {
		if IsErrProjectDoesNotExist(err) {
			return nil, ErrNoUsableInbox{UserID: assignee.ID}
		}
		return nil, err
	}
	if project.OwnerID != assignee.ID || project.IsArchived {
		return nil, ErrNoUsableInbox{UserID: assignee.ID}
	}
	// An archived parent archives the inbox too.
	if project.CheckIsArchived(s) != nil {
		return nil, ErrNoUsableInbox{UserID: assignee.ID}
	}
	return project, nil
}

func validateAssignInput(in *AssignTaskByEmailInput) error {
	title := strings.TrimSpace(in.Title)
	if title == "" {
		return ErrTaskCannotBeEmpty{}
	}
	if utf8.RuneCountInString(title) > maxAssignTitleLength {
		return ErrInvalidData{Message: "the title must not be longer than 250 characters"}
	}
	if in.Priority < 0 || in.Priority > maxAssignPriority {
		return ErrInvalidData{Message: "the priority must be between 0 and 5"}
	}
	if utf8.RuneCountInString(in.ExternalID) > maxAssignExternalIDLength {
		return ErrInvalidData{Message: "the external_id must not be longer than 250 characters"}
	}
	return nil
}

// existingExternalTask returns the task a bot already created for the external id, or nil. A
// reference whose task is gone is removed so the id can be used again. The same id for another
// assignee is a conflict.
func existingExternalTask(s *xorm.Session, bot *user.User, externalID string, assignee *user.User) (*Task, error) {
	ref := &TaskExternalRef{}
	has, err := s.Where("bot_id = ? AND external_id = ?", bot.ID, externalID).Get(ref)
	if err != nil || !has {
		return nil, err
	}

	task, err := GetTaskByIDSimple(s, ref.TaskID)
	if err != nil {
		if !IsErrTaskDoesNotExist(err) {
			return nil, err
		}
		_, err = s.Where("id = ?", ref.ID).Delete(&TaskExternalRef{})
		return nil, err
	}

	same, err := s.Where("task_id = ? AND user_id = ?", task.ID, assignee.ID).Exist(&TaskAssginee{})
	if err != nil {
		return nil, err
	}
	if !same {
		return nil, ErrExternalIDConflict{}
	}

	project, err := GetProjectSimpleByID(s, task.ProjectID)
	if err != nil {
		return nil, err
	}
	task.setIdentifier(project)
	return &task, nil
}

// AssignTaskByEmail creates a task and assigns it to the user with the given email address, on
// behalf of an integration. Only bots owned by an instance admin may call it.
//
// The bot needs no access to the assignee's projects, which is the point: without a project the
// task goes into the assignee's inbox, which the bot cannot read, edit or delete afterwards. With
// a project_id the bot needs write access to it like any other task creator.
//
// It does not commit; the caller owns the transaction and dispatches the queued events.
func AssignTaskByEmail(s *xorm.Session, bot *user.User, in *AssignTaskByEmailInput) (*AssignTaskByEmailResult, error) {
	// Only a bot an instance admin owns may put tasks into other people's inboxes and name the
	// assignor. A human's token must not, and neither must the bot of an ordinary user: everybody
	// can create bots, so that would let any user fill any inbox in anybody's name.
	if bot == nil || !bot.IsBot() {
		return nil, ErrGenericForbidden{}
	}
	privileged, err := isPrivilegedActor(s, bot)
	if err != nil {
		return nil, err
	}
	if !privileged {
		return nil, ErrGenericForbidden{}
	}

	err = validateAssignInput(in)
	if err != nil {
		return nil, err
	}

	assignee, err := resolveAssigneeByEmail(s, in.AssigneeEmail)
	if err != nil {
		return nil, err
	}

	if in.ExternalID != "" {
		var existing *Task
		existing, err = existingExternalTask(s, bot, in.ExternalID, assignee)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			// A retry returns what was created, whatever assignor the retry names.
			var creator *user.User
			creator, err = user.GetUserByID(s, existing.CreatedByID)
			if err != nil && !user.IsErrUserStatusError(err) {
				return nil, err
			}
			return &AssignTaskByEmailResult{Created: false, Task: existing, Assignee: assignee, Assignor: creator}, nil
		}
	}

	var project *Project
	if in.ProjectID != 0 {
		project, err = GetProjectSimpleByID(s, in.ProjectID)
		if err != nil {
			return nil, err
		}
		var canWrite bool
		canWrite, err = project.CanWrite(s, bot)
		if err != nil {
			return nil, err
		}
		if !canWrite {
			return nil, ErrGenericForbidden{}
		}
	} else {
		project, err = assigneeInbox(s, assignee)
		if err != nil {
			return nil, err
		}
	}

	assignor, err := resolveAssignor(s, in, assignee, project)
	if err != nil {
		return nil, err
	}

	task := &Task{
		Title:            strings.TrimSpace(in.Title),
		Description:      in.Description,
		DueDate:          in.DueDate,
		Priority:         in.Priority,
		ProjectID:        project.ID,
		Assignees:        []*user.User{assignee},
		assignorOverride: assignor,
	}
	// Same path as a normal task, so index, buckets, positions, the assignee row, notifications,
	// webhooks and the audit log behave identically. The assignor is stored as the creator; the
	// bot stays the doer of the events, and a bot may name any assignor (canSetAssignor).
	err = createTask(s, task, bot, true, true)
	if err != nil {
		return nil, err
	}

	if in.ExternalID != "" {
		_, err = s.Insert(&TaskExternalRef{BotID: bot.ID, ExternalID: in.ExternalID, TaskID: task.ID})
		if err != nil {
			return nil, err
		}
	}

	return &AssignTaskByEmailResult{Created: true, Task: task, Assignee: assignee, Assignor: assignor}, nil
}
