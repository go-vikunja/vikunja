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

package mail

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/mail"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/jobs"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"xorm.io/xorm"
)

type delivery struct {
	Message string `json:"message"`
}

func init() {
	jobs.Register("mail.send", "mail", deliver)
	jobs.Register("mail.bulk", "bulk-mail", deliver)
}

func Enqueue(s *xorm.Session, opts *Opts, queue, key string) error {
	if isUnderTest {
		_, err := s.After(func(any) { sentMails = append(sentMails, opts) }).Insert(&jobs.Job{
			Queue:      queue,
			Kind:       "mail.send",
			Version:    1,
			Payload:    `{}`,
			State:      jobs.StateCompleted,
			RunAt:      time.Now(),
			FinishedAt: time.Now(),
		})
		return err
	}
	if !config.MailerEnabled.GetBool() {
		return nil
	}
	var content bytes.Buffer
	if _, err := getMessage(opts).WriteTo(&content); err != nil {
		return err
	}
	payload, err := json.Marshal(delivery{Message: content.String()})
	if err != nil {
		return err
	}
	kind := "mail.send"
	if queue == "bulk-mail" {
		kind = "mail.bulk"
	}
	return jobs.Enqueue(s, jobs.Request{Queue: queue, Kind: kind, Payload: payload, UniqueKey: key})
}

func deliver(msg *message.Message) error {
	if !config.MailerEnabled.GetBool() {
		return nil
	}
	payload := delivery{}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		return err
	}
	parsed, err := mail.ReadMessage(strings.NewReader(payload.Message))
	if err != nil {
		return err
	}
	from, err := parsed.Header.AddressList("From")
	if err != nil {
		return err
	}
	to, err := parsed.Header.AddressList("To")
	if err != nil {
		return err
	}
	client, err := getClient()
	if err != nil {
		return err
	}
	connection, err := client.DialToSMTPClientWithContext(msg.Context())
	if err != nil {
		return err
	}
	defer connection.Close()
	stop := context.AfterFunc(msg.Context(), func() { _ = connection.Close() })
	defer stop()
	if err := connection.Mail((&mail.Address{Address: from[0].Address}).String()); err != nil {
		return err
	}
	for _, recipient := range to {
		if err := connection.Rcpt((&mail.Address{Address: recipient.Address}).String()); err != nil {
			return err
		}
	}
	writer, err := connection.Data()
	if err != nil {
		return err
	}
	if _, err := io.WriteString(writer, payload.Message); err != nil {
		return err
	}
	return writer.Close()
}

func DeliverPending(ctx context.Context) error {
	if !config.MailerEnabled.GetBool() || isUnderTest {
		return nil
	}
	engine, err := db.CreateDBEngine()
	if err != nil {
		return err
	}
	queue := jobs.NewQueue(engine, watermill.NewUUID())
	for {
		batch, err := queue.Claim(ctx, "mail", 1, time.Now(), time.Minute)
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		job := batch[0]
		msg := message.NewMessage(watermill.NewUUID(), []byte(job.Payload))
		msg.SetContext(ctx)
		if err := deliver(msg); err != nil { //nolint:contextcheck // The context is carried by the message.
			return queue.Fail(ctx, job, time.Now(), err, job.MaxAttempts)
		}
		s := engine.NewSession()
		err = jobs.Complete(s, job, time.Now())
		s.Close()
		if err != nil {
			return err
		}
	}
}

func Deliver(ctx context.Context, opts *Opts) error {
	if isUnderTest {
		sentMails = append(sentMails, opts)
		return nil
	}
	if !config.MailerEnabled.GetBool() {
		return nil
	}
	client, err := getClient()
	if err != nil {
		return err
	}
	return client.DialAndSendWithContext(ctx, getMessage(opts))
}
