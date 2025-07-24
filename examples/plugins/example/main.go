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

package main

import (
	"code.vikunja.io/api/pkg/events"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/plugins"

	"github.com/ThreeDotsLabs/watermill/message"
)

type ExamplePlugin struct{}

func (p *ExamplePlugin) Name() string    { return "example" }
func (p *ExamplePlugin) Version() string { return "1.0.0" }
func (p *ExamplePlugin) Init() error {
	log.Infof("example plugin initialized")

	events.RegisterListener((&models.TaskCreatedEvent{}).Name(), &TestListener{})

	return nil
}
func (p *ExamplePlugin) Shutdown() error { return nil }

func NewPlugin() plugins.Plugin { return &ExamplePlugin{} }

type TestListener struct{}

func (t *TestListener) Handle(msg *message.Message) error {
	log.Infof("TestListener received message: %s", string(msg.Payload))
	return nil
}

func (t *TestListener) Name() string {
	return "TestListener"
}
