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

package events

import (
	"context"
	"encoding/json"
	"time"

	"code.vikunja.io/api/pkg/log"
	vmetrics "code.vikunja.io/api/pkg/metrics"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/components/metrics"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/ThreeDotsLabs/watermill/pubsub/gochannel"
)

var pubsub *gochannel.GoChannel

// Event represents the event interface used by all events
type Event interface {
	Name() string
}

// InitEvents sets up everything needed to work with events
func InitEvents() (err error) {
	logger := log.NewWatermillLogger()

	router, err := message.NewRouter(
		message.RouterConfig{},
		logger,
	)
	if err != nil {
		return err
	}

	router.AddMiddleware(
		middleware.Retry{
			MaxRetries:      5,
			InitialInterval: time.Millisecond * 100,
			Logger:          logger,
			Multiplier:      2,
		}.Middleware,
		middleware.Recoverer,
	)

	metricsBuilder := metrics.NewPrometheusMetricsBuilder(vmetrics.GetRegistry(), "", "")
	metricsBuilder.AddPrometheusRouterMetrics(router)

	pubsub = gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 1024,
		},
		logger,
	)

	for topic, funcs := range listeners {
		for _, handler := range funcs {
			router.AddNoPublisherHandler(topic+"."+handler.Name(), topic, pubsub, handler.Handle)
		}
	}

	return router.Run(context.Background())
}

// Dispatch dispatches an event
func Dispatch(event Event) error {
	if isUnderTest {
		dispatchedTestEvents = append(dispatchedTestEvents, event)
		return nil
	}

	content, err := json.Marshal(event)
	if err != nil {
		return err
	}

	msg := message.NewMessage(watermill.NewUUID(), content)
	return pubsub.Publish(event.Name(), msg)
}
