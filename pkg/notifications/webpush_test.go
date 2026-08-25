// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package notifications

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"xorm.io/builder"
)

type webPushHTTPClientFunc func(*http.Request) (*http.Response, error)

func (f webPushHTTPClientFunc) Do(request *http.Request) (*http.Response, error) {
	return f(request)
}

func setupWebPushTest(t *testing.T) WebPushSubscriptionInput {
	t.Helper()
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)

	oldEnabled := config.WebPushEnabled.GetBool()
	oldPublicKey := config.WebPushPublicKey.GetString()
	oldPrivateKey := config.WebPushPrivateKey.GetString()
	oldPublicURL := config.ServicePublicURL.GetString()
	config.WebPushEnabled.Set(true)
	config.WebPushPublicKey.Set(publicKey)
	config.WebPushPrivateKey.Set(privateKey)
	config.ServicePublicURL.Set("https://vikunja.example.com/")
	t.Cleanup(func() {
		config.WebPushEnabled.Set(oldEnabled)
		config.WebPushPublicKey.Set(oldPublicKey)
		config.WebPushPrivateKey.Set(oldPrivateKey)
		config.ServicePublicURL.Set(oldPublicURL)
		webPushHTTPClient = nil
	})

	s := db.NewSession()
	defer s.Close()
	_, err = s.Exec("DELETE FROM web_push_deliveries")
	require.NoError(t, err)
	_, err = s.Exec("DELETE FROM web_push_subscriptions")
	require.NoError(t, err)
	_, err = s.Exec("DELETE FROM sessions")
	require.NoError(t, err)
	_, err = s.Insert(&webPushSessionState{
		ID:         "550e8400-e29b-41d4-a716-446655440001",
		UserID:     1,
		LastActive: time.Now(),
	})
	require.NoError(t, err)
	_, err = s.Insert(&webPushSessionState{
		ID:         "550e8400-e29b-41d4-a716-446655440002",
		UserID:     2,
		LastActive: time.Now(),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())

	return WebPushSubscriptionInput{
		Endpoint: "https://push.example.com/subscription/one",
		P256DH:   publicKey,
		Auth:     base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef")),
	}
}

func TestWebPushSubscriptionLifecycle(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	defer s.Close()

	first, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()
	require.Positive(t, first.ID)

	s = db.NewSession()
	refreshed, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()
	assert.Equal(t, first.ID, refreshed.ID, "upsert must preserve the device row")

	s = db.NewSession()
	_, err = UpsertWebPushSubscription(s, 2, "550e8400-e29b-41d4-a716-446655440002", "ec35ec57-c51d-4d5f-a1d3-219982df56c0", input)
	require.ErrorAs(t, err, &ErrWebPushEndpointOwned{})
	require.NoError(t, s.Rollback())
	s.Close()

	serialized := string(mustMarshalForTest(t, first))
	assert.NotContains(t, serialized, input.Endpoint)
	assert.NotContains(t, serialized, input.P256DH)
	assert.NotContains(t, serialized, input.Auth)

	s = db.NewSession()
	require.NoError(t, DeleteWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", first.DeviceID))
	require.NoError(t, s.Commit())
	s.Close()
	db.AssertMissing(t, "web_push_subscriptions", map[string]any{"id": first.ID})
}

func mustMarshalForTest(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err)
	return data
}

func TestWebPushQueueDurabilityAndDeduplication(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	message := &WebPushMessage{Title: "Vikunja", Body: "A task changed", URL: "/tasks/1"}
	s = db.NewSession()
	require.NoError(t, enqueueWebPush(s, 1, "notification:123", message, time.Hour))
	require.NoError(t, enqueueWebPush(s, 1, "notification:123", message, time.Hour))
	require.NoError(t, s.Commit())
	s.Close()
	db.AssertCount(t, "web_push_deliveries", builder.Eq{"delivery_key": "notification:123"}, 1)

	s = db.NewSession()
	require.NoError(t, enqueueWebPush(s, 1, "notification:rollback", message, time.Hour))
	require.NoError(t, s.Rollback())
	s.Close()
	db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:rollback"})
}

func TestWebPushWorkerOutcomes(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, enqueueWebPush(s, 1, "notification:accepted", &WebPushMessage{Title: "Vikunja", Body: "Accepted", URL: "/"}, time.Hour))
	require.NoError(t, s.Commit())
	s.Close()

	webPushHTTPClient = webPushHTTPClientFunc(func(request *http.Request) (*http.Response, error) {
		assert.Equal(t, "POST", request.Method)
		assert.NotEmpty(t, request.Header.Get("Authorization"))
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})
	claimed, err := claimWebPushDeliveries("worker-one", time.Now())
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	contended, err := claimWebPushDeliveries("worker-two", time.Now())
	require.NoError(t, err)
	assert.Empty(t, contended, "an active lease must prevent a second worker from claiming the row")
	processWebPushDelivery(context.Background(), claimed[0])
	db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:accepted"})

	s = db.NewSession()
	require.NoError(t, enqueueWebPush(s, 1, "notification:retry", &WebPushMessage{Title: "Vikunja", Body: "Retry", URL: "/"}, time.Hour))
	require.NoError(t, s.Commit())
	s.Close()
	webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("temporary network failure")
	})
	claimed, err = claimWebPushDeliveries("worker-two", time.Now())
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	processWebPushDelivery(context.Background(), claimed[0])
	s = db.NewSession()
	retried := &WebPushDelivery{}
	found, err := s.Where("delivery_key = ?", "notification:retry").Get(retried)
	require.NoError(t, err)
	require.True(t, found)
	assert.Equal(t, 1, retried.Attempts)
	assert.Nil(t, retried.LeaseUntil)
	s.Close()
}

func TestWebPushLeaseRecoveryAndRelease(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, enqueueWebPush(s, 1, "notification:restart", &WebPushMessage{Title: "Vikunja", Body: "Restart", URL: "/"}, time.Hour))
	require.NoError(t, s.Commit())
	s.Close()

	claimed, err := claimWebPushDeliveries("stopped-worker", time.Now())
	require.NoError(t, err)
	require.Len(t, claimed, 1)

	releaseWebPushLeases("stopped-worker")
	recovered, err := claimWebPushDeliveries("new-worker", time.Now())
	require.NoError(t, err)
	require.Len(t, recovered, 1, "a graceful stop must make the delivery immediately recoverable")

	s = db.NewSession()
	past := time.Now().Add(-time.Minute)
	_, err = s.ID(recovered[0].ID).Cols("lease_until").Update(&WebPushDelivery{LeaseUntil: &past})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	recoveredAfterCrash, err := claimWebPushDeliveries("restart-worker", time.Now())
	require.NoError(t, err)
	require.Len(t, recoveredAfterCrash, 1, "an expired lease must recover work after an ungraceful restart")
}

func TestSendWebPushTestAccepted(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})
	s = db.NewSession()
	require.NoError(t, SendWebPushTest(context.Background(), s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", WebPushMessage{Title: "Vikunja", Body: "Test", URL: "/"}))
	require.NoError(t, s.Rollback())
	s.Close()
}

func TestWebPushWorkerHTTPResponses(t *testing.T) {
	t.Run("retry after", func(t *testing.T) {
		input := setupWebPushTest(t)
		s := db.NewSession()
		_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
		require.NoError(t, err)
		require.NoError(t, enqueueWebPush(s, 1, "notification:rate-limited", &WebPushMessage{Title: "Vikunja", Body: "Retry", URL: "/"}, time.Hour))
		require.NoError(t, s.Commit())
		s.Close()

		webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusTooManyRequests, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{"Retry-After": []string{"120"}}}, nil
		})
		claimed, err := claimWebPushDeliveries("worker", time.Now())
		require.NoError(t, err)
		require.Len(t, claimed, 1)
		before := time.Now()
		processWebPushDelivery(context.Background(), claimed[0])

		s = db.NewSession()
		retried := &WebPushDelivery{}
		found, err := s.Where("delivery_key = ?", "notification:rate-limited").Get(retried)
		require.NoError(t, err)
		require.True(t, found)
		assert.False(t, retried.NextAttemptAt.Before(before.Add(119*time.Second)))
		s.Close()
	})

	t.Run("permanent response keeps subscription", func(t *testing.T) {
		input := setupWebPushTest(t)
		s := db.NewSession()
		subscription, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
		require.NoError(t, err)
		require.NoError(t, enqueueWebPush(s, 1, "notification:bad-message", &WebPushMessage{Title: "Vikunja", Body: "Bad", URL: "/"}, time.Hour))
		require.NoError(t, s.Commit())
		s.Close()

		webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusBadRequest, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
		})
		claimed, err := claimWebPushDeliveries("worker", time.Now())
		require.NoError(t, err)
		require.Len(t, claimed, 1)
		processWebPushDelivery(context.Background(), claimed[0])
		db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:bad-message"})
		db.AssertExists(t, "web_push_subscriptions", map[string]any{"id": subscription.ID}, false)
	})

	t.Run("gone response removes subscription", func(t *testing.T) {
		input := setupWebPushTest(t)
		s := db.NewSession()
		subscription, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
		require.NoError(t, err)
		require.NoError(t, enqueueWebPush(s, 1, "notification:gone", &WebPushMessage{Title: "Vikunja", Body: "Gone", URL: "/"}, time.Hour))
		require.NoError(t, s.Commit())
		s.Close()

		webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
			return &http.Response{StatusCode: http.StatusGone, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
		})
		claimed, err := claimWebPushDeliveries("worker", time.Now())
		require.NoError(t, err)
		require.Len(t, claimed, 1)
		processWebPushDelivery(context.Background(), claimed[0])
		db.AssertMissing(t, "web_push_subscriptions", map[string]any{"id": subscription.ID})
		db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:gone"})
	})

	t.Run("expired delivery is discarded before claim", func(t *testing.T) {
		input := setupWebPushTest(t)
		s := db.NewSession()
		_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
		require.NoError(t, err)
		require.NoError(t, enqueueWebPush(s, 1, "notification:expired", &WebPushMessage{Title: "Vikunja", Body: "Expired", URL: "/"}, -time.Second))
		require.NoError(t, s.Commit())
		s.Close()

		claimed, err := claimWebPushDeliveries("worker", time.Now())
		require.NoError(t, err)
		assert.Empty(t, claimed)
		db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:expired"})
	})
}

func TestInactiveOrDeletedSessionStillDelivers(t *testing.T) {
	input := setupWebPushTest(t)
	s := db.NewSession()
	_, err := UpsertWebPushSubscription(s, 1, "550e8400-e29b-41d4-a716-446655440001", "9a34527d-1357-4c64-8171-6d6e25f01d62", input)
	require.NoError(t, err)
	require.NoError(t, enqueueWebPush(s, 1, "notification:idle-session", &WebPushMessage{Title: "Vikunja", Body: "Idle session", URL: "/"}, time.Hour))
	require.NoError(t, s.Commit())
	s.Close()

	// Simulate the session-cleanup cron removing the idle login session. The
	// browser push subscription must survive and keep receiving notifications;
	// it is revoked only on explicit logout, account deletion, disabling push,
	// or a Gone response from the push service.
	s = db.NewSession()
	_, err = s.Exec("DELETE FROM sessions WHERE id = ?", "550e8400-e29b-41d4-a716-446655440001")
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()

	s = db.NewSession()
	hasPush, err := HasWebPushSubscription(s, 1)
	require.NoError(t, err)
	assert.True(t, hasPush, "a subscription must survive its login session going away")
	require.NoError(t, s.Rollback())
	s.Close()

	reached := false
	webPushHTTPClient = webPushHTTPClientFunc(func(*http.Request) (*http.Response, error) {
		reached = true
		return &http.Response{StatusCode: http.StatusCreated, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header)}, nil
	})
	claimed, err := claimWebPushDeliveries("worker", time.Now())
	require.NoError(t, err)
	require.Len(t, claimed, 1)
	processWebPushDelivery(context.Background(), claimed[0])
	assert.True(t, reached, "delivery must reach the push service even without an active session")
	db.AssertExists(t, "web_push_subscriptions", map[string]any{"user_id": 1}, false)
	db.AssertMissing(t, "web_push_deliveries", map[string]any{"delivery_key": "notification:idle-session"})
}

func TestParseRetryAfter(t *testing.T) {
	now := time.Date(2026, time.August, 7, 12, 0, 0, 0, time.UTC)
	assert.Equal(t, 90*time.Second, parseRetryAfter("90", now))
	assert.Equal(t, 2*time.Minute, parseRetryAfter(now.Add(2*time.Minute).Format(http.TimeFormat), now))
	assert.Zero(t, parseRetryAfter("invalid", now))
}

func TestWebPushSubscriber(t *testing.T) {
	oldFrom := config.MailerFromEmail.GetString()
	oldURL := config.ServicePublicURL.GetString()
	t.Cleanup(func() {
		config.MailerFromEmail.Set(oldFrom)
		config.ServicePublicURL.Set(oldURL)
	})

	// A real mailer address is used verbatim: webpush-go adds the single
	// "mailto:" prefix. Pre-prefixing here would produce "mailto:mailto:...",
	// which Apple's push service rejects with 403 BadJwtToken.
	config.MailerFromEmail.Set("noreply@vikunja.example.com")
	config.ServicePublicURL.Set("https://vikunja.example.com/")
	assert.Equal(t, "noreply@vikunja.example.com", webPushSubscriber())
	assert.False(t, strings.HasPrefix(webPushSubscriber(), "mailto:"))

	// The default placeholder address has no dot in the domain and is not a
	// usable contact, so fall back to a synthesized address on the public host.
	config.MailerFromEmail.Set("mail@vikunja")
	config.ServicePublicURL.Set("https://tasks.example.org/")
	assert.Equal(t, "webmaster@tasks.example.org", webPushSubscriber())
}

func TestIsEmailLikeAddress(t *testing.T) {
	assert.True(t, isEmailLikeAddress("noreply@example.com"))
	assert.True(t, isEmailLikeAddress("a@b.co"))
	assert.False(t, isEmailLikeAddress("mail@vikunja"))
	assert.False(t, isEmailLikeAddress("plainstring"))
	assert.False(t, isEmailLikeAddress("@example.com"))
	assert.False(t, isEmailLikeAddress("user@"))
	assert.False(t, isEmailLikeAddress("a b@c.com"))
}
