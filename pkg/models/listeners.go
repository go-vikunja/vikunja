// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
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

package models

import (
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/modules/keyvalue"
	"github.com/ThreeDotsLabs/watermill/message"
)

// RegisterListeners registers all event listeners
func RegisterListeners() {
	events.RegisterListener((&ListCreatedEvent{}).Name(), &IncreaseListCounter{})
	events.RegisterListener((&ListDeletedEvent{}).Name(), &DecreaseListCounter{})
	events.RegisterListener((&NamespaceCreatedEvent{}).Name(), &IncreaseNamespaceCounter{})
	events.RegisterListener((&NamespaceDeletedEvent{}).Name(), &DecreaseNamespaceCounter{})
	events.RegisterListener((&TaskCreatedEvent{}).Name(), &IncreaseTaskCounter{})
	events.RegisterListener((&TaskDeletedEvent{}).Name(), &DecreaseTaskCounter{})
	events.RegisterListener((&TeamDeletedEvent{}).Name(), &DecreaseTeamCounter{})
	events.RegisterListener((&TeamCreatedEvent{}).Name(), &IncreaseTeamCounter{})
}

//////
// Task Events

// IncreaseTaskCounter  represents a listener
type IncreaseTaskCounter struct {
}

// Name defines the name for the IncreaseTaskCounter listener
func (s *IncreaseTaskCounter) Name() string {
	return "task.counter.increase"
}

// Hanlde is executed when the event IncreaseTaskCounter listens on is fired
func (s *IncreaseTaskCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.TaskCountKey, 1)
}

// DecreaseTaskCounter  represents a listener
type DecreaseTaskCounter struct {
}

// Name defines the name for the DecreaseTaskCounter listener
func (s *DecreaseTaskCounter) Name() string {
	return "task.counter.decrease"
}

// Hanlde is executed when the event DecreaseTaskCounter listens on is fired
func (s *DecreaseTaskCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.TaskCountKey, 1)
}

///////
// List Event Listeners

type IncreaseListCounter struct {
}

func (s *IncreaseListCounter) Name() string {
	return "list.counter.increase"
}

func (s *IncreaseListCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.ListCountKey, 1)
}

type DecreaseListCounter struct {
}

func (s *DecreaseListCounter) Name() string {
	return "list.counter.decrease"
}

func (s *DecreaseListCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.ListCountKey, 1)
}

//////
// Namespace events

// IncreaseNamespaceCounter  represents a listener
type IncreaseNamespaceCounter struct {
}

// Name defines the name for the IncreaseNamespaceCounter listener
func (s *IncreaseNamespaceCounter) Name() string {
	return "namespace.counter.increase"
}

// Hanlde is executed when the event IncreaseNamespaceCounter listens on is fired
func (s *IncreaseNamespaceCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.NamespaceCountKey, 1)
}

// DecreaseNamespaceCounter  represents a listener
type DecreaseNamespaceCounter struct {
}

// Name defines the name for the DecreaseNamespaceCounter listener
func (s *DecreaseNamespaceCounter) Name() string {
	return "namespace.counter.decrease"
}

// Hanlde is executed when the event DecreaseNamespaceCounter listens on is fired
func (s *DecreaseNamespaceCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.NamespaceCountKey, 1)
}

///////
// Team Events

// IncreaseTeamCounter  represents a listener
type IncreaseTeamCounter struct {
}

// Name defines the name for the IncreaseTeamCounter listener
func (s *IncreaseTeamCounter) Name() string {
	return "team.counter.increase"
}

// Hanlde is executed when the event IncreaseTeamCounter listens on is fired
func (s *IncreaseTeamCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.IncrBy(metrics.TeamCountKey, 1)
}

// DecreaseTeamCounter  represents a listener
type DecreaseTeamCounter struct {
}

// Name defines the name for the DecreaseTeamCounter listener
func (s *DecreaseTeamCounter) Name() string {
	return "team.counter.decrease"
}

// Hanlde is executed when the event DecreaseTeamCounter listens on is fired
func (s *DecreaseTeamCounter) Handle(payload message.Payload) (err error) {
	return keyvalue.DecrBy(metrics.TeamCountKey, 1)
}
