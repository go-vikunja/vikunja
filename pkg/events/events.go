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

package events

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/getsentry/sentry-go"

	"code.vikunja.io/api/pkg/config"
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

type messageHandleFailedError struct {
	Metadata message.Metadata
}

func (m *messageHandleFailedError) Error() string {
	return fmt.Sprintf("Failed to handle message: %v", m.Metadata)
}

// InitEvents sets up everything needed to work with events
func InitEvents() (err error) {
	logger := log.NewWatermillLogger(config.LogEnabled.GetBool(), config.LogEvents.GetString(), config.LogEventsLevel.GetString(), config.LogFormat.GetString())

	router, err := message.NewRouter(
		message.RouterConfig{},
		logger,
	)
	if err != nil {
		return err
	}

	metricsBuilder := metrics.NewPrometheusMetricsBuilder(vmetrics.GetRegistry(), "", "")
	metricsBuilder.AddPrometheusRouterMetrics(router)

	pubsub = gochannel.NewGoChannel(
		gochannel.Config{
			OutputChannelBuffer: 1024,
		},
		logger,
	)

	poison, err := middleware.PoisonQueue(pubsub, "poison")
	if err != nil {
		return err
	}
	router.AddNoPublisherHandler("poison.logger", "poison", pubsub, func(msg *message.Message) error {
		meta := ""
		for s, m := range msg.Metadata {
			meta += s + "=" + m + ", "
		}
		log.Errorf("Error while handling message %s, %s payload=%s", msg.UUID, meta, string(msg.Payload))

		if config.SentryEnabled.GetBool() {
			sentry.CaptureException(&messageHandleFailedError{
				Metadata: msg.Metadata,
			})
		}
		return nil
	})

	router.AddMiddleware(
		poison,
		middleware.Retry{
			MaxRetries:          5,
			InitialInterval:     time.Millisecond * 100,
			MaxInterval:         time.Hour,
			Multiplier:          2,
			MaxElapsedTime:      0,
			RandomizationFactor: 1,
			Logger:              logger,
		}.Middleware,
		middleware.Recoverer,
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
