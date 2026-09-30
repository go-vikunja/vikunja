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
	"crypto/ecdh"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/metrics"
	"code.vikunja.io/api/pkg/utils"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/google/uuid"
	"xorm.io/xorm"
)

const (
	webPushBatchSize       = 1
	webPushLeaseDuration   = time.Minute
	webPushPollInterval    = 30 * time.Second
	webPushInitialBackoff  = 30 * time.Second
	webPushMaximumBackoff  = 15 * time.Minute
	webPushMaximumErrorLen = 1000
)

// WebPushSubscription is one browser Push API subscription owned by a user session.
// Endpoint and keys are capabilities and are never serialized back to clients.
type WebPushSubscription struct {
	ID             int64      `xorm:"bigint autoincr not null unique pk" json:"id" readOnly:"true"`
	UserID         int64      `xorm:"bigint not null index unique(user_device)" json:"-"`
	SessionID      string     `xorm:"varchar(36) not null index" json:"-"`
	DeviceID       string     `xorm:"varchar(36) not null unique(user_device)" json:"device_id" readOnly:"true"`
	Endpoint       string     `xorm:"text not null" json:"-"`
	EndpointHash   string     `xorm:"char(64) not null unique" json:"-"`
	P256DH         string     `xorm:"text not null" json:"-"`
	Auth           string     `xorm:"text not null" json:"-"`
	ExpirationTime *time.Time `xorm:"datetime null" json:"-"`
	Created        time.Time  `xorm:"created not null" json:"created" readOnly:"true"`
	Updated        time.Time  `xorm:"updated not null" json:"updated" readOnly:"true"`
}

func (*WebPushSubscription) TableName() string { return "web_push_subscriptions" }

type WebPushDelivery struct {
	ID             int64      `xorm:"bigint autoincr not null unique pk" json:"-"`
	SubscriptionID int64      `xorm:"bigint not null index unique(subscription_delivery)" json:"-"`
	DeliveryKey    string     `xorm:"varchar(255) not null unique(subscription_delivery)" json:"-"`
	Payload        string     `xorm:"text not null" json:"-"`
	Attempts       int        `xorm:"not null default 0" json:"-"`
	NextAttemptAt  time.Time  `xorm:"datetime not null index" json:"-"`
	LeaseOwner     string     `xorm:"varchar(36) null" json:"-"`
	LeaseUntil     *time.Time `xorm:"datetime null index" json:"-"`
	LastError      string     `xorm:"text null" json:"-"`
	ExpiresAt      time.Time  `xorm:"datetime not null index" json:"-"`
	Created        time.Time  `xorm:"created not null" json:"-"`
	Updated        time.Time  `xorm:"updated not null" json:"-"`
}

func (*WebPushDelivery) TableName() string { return "web_push_deliveries" }

type WebPushMessage struct {
	Title          string `json:"title"`
	Body           string `json:"body"`
	URL            string `json:"url"`
	Tag            string `json:"tag"`
	NotificationID int64  `json:"notification_id,omitempty"`
	Test           bool   `json:"test,omitempty"`
}

// WebPushable is implemented only by notification types that are safe to show on a lock screen.
type WebPushable interface {
	ToWebPush(lang string) *WebPushMessage
}

type WebPushSubscriptionInput struct {
	Endpoint       string
	P256DH         string
	Auth           string
	ExpirationTime *time.Time
}

type webPushSessionState struct {
	ID            string
	UserID        int64
	IsLongSession bool
	LastActive    time.Time
}

func (*webPushSessionState) TableName() string { return "sessions" }

var (
	webPushWorkerMu     sync.Mutex
	webPushWorkerCancel context.CancelFunc
	webPushWorkerDone   chan struct{}
	webPushWake         = make(chan struct{}, 1)
	webPushHTTPClient   webpush.HTTPClient
)

type retryableWebPushError struct{ error }

type retryableWebPushClient struct{ client webpush.HTTPClient }

func (c retryableWebPushClient) Do(request *http.Request) (*http.Response, error) {
	response, err := c.client.Do(request)
	if err != nil {
		return nil, retryableWebPushError{error: errors.New("push transport failed")}
	}
	return response, nil
}

func endpointFingerprint(endpoint string) string {
	hash := sha256.Sum256([]byte(endpoint))
	return hex.EncodeToString(hash[:])
}

func stablePushTag(deliveryKey string) string {
	hash := sha256.Sum256([]byte(deliveryKey))
	return "vikunja-" + base64.RawURLEncoding.EncodeToString(hash[:18])
}

func validateSubscriptionInput(input WebPushSubscriptionInput) error {
	endpoint, err := url.Parse(input.Endpoint)
	if err != nil || endpoint.Scheme != "https" || endpoint.Hostname() == "" || endpoint.User != nil || endpoint.Fragment != "" {
		return errors.New("push endpoint must be an absolute https URL")
	}
	auth, err := base64.RawURLEncoding.DecodeString(input.Auth)
	if err != nil || len(auth) != 16 {
		return errors.New("invalid push authentication key")
	}
	p256dh, err := base64.RawURLEncoding.DecodeString(input.P256DH)
	if err != nil {
		return errors.New("invalid push P-256 key")
	}
	if _, err := ecdh.P256().NewPublicKey(p256dh); err != nil {
		return errors.New("invalid push P-256 key")
	}
	return nil
}

func UpsertWebPushSubscription(s *xorm.Session, userID int64, sessionID, deviceID string, input WebPushSubscriptionInput) (*WebPushSubscription, error) {
	if !config.WebPushEnabled.GetBool() {
		return nil, errors.New("web push is disabled")
	}
	if _, err := uuid.Parse(deviceID); err != nil {
		return nil, ErrWebPushInputInvalid{errors.New("device id must be a UUID")}
	}
	if sessionID == "" {
		return nil, errors.New("web push requires a user session")
	}
	session := &webPushSessionState{}
	foundSession, err := s.Where("id = ? AND user_id = ?", sessionID, userID).Get(session)
	if err != nil {
		return nil, err
	}
	if !foundSession || !webPushSessionIsActive(session, time.Now()) {
		return nil, ErrWebPushSessionInvalid{}
	}
	if err := validateSubscriptionInput(input); err != nil {
		return nil, ErrWebPushInputInvalid{err}
	}

	hash := endpointFingerprint(input.Endpoint)
	byEndpoint := &WebPushSubscription{}
	foundEndpoint, err := s.Where("endpoint_hash = ?", hash).Get(byEndpoint)
	if err != nil {
		return nil, err
	}
	if foundEndpoint && (byEndpoint.UserID != userID || byEndpoint.DeviceID != deviceID) {
		return nil, ErrWebPushEndpointOwned{}
	}

	subscription := &WebPushSubscription{}
	foundDevice, err := s.Where("user_id = ? AND device_id = ?", userID, deviceID).Get(subscription)
	if err != nil {
		return nil, err
	}
	if foundDevice {
		if subscription.Endpoint != input.Endpoint || subscription.P256DH != input.P256DH || subscription.Auth != input.Auth {
			if _, err = s.Where("subscription_id = ?", subscription.ID).Delete(&WebPushDelivery{}); err != nil {
				return nil, err
			}
		}
		subscription.SessionID = sessionID
		subscription.Endpoint = input.Endpoint
		subscription.EndpointHash = hash
		subscription.P256DH = input.P256DH
		subscription.Auth = input.Auth
		subscription.ExpirationTime = input.ExpirationTime
		_, err = s.ID(subscription.ID).
			Cols("session_id", "endpoint", "endpoint_hash", "p256dh", "auth", "expiration_time", "updated").
			Nullable("expiration_time").
			Update(subscription)
		return subscription, err
	}

	subscription = &WebPushSubscription{
		UserID:         userID,
		SessionID:      sessionID,
		DeviceID:       deviceID,
		Endpoint:       input.Endpoint,
		EndpointHash:   hash,
		P256DH:         input.P256DH,
		Auth:           input.Auth,
		ExpirationTime: input.ExpirationTime,
	}
	_, err = s.Insert(subscription)
	return subscription, err
}

type ErrWebPushInputInvalid struct{ error }

type ErrWebPushSubscriptionNotFound struct{}

func (ErrWebPushSubscriptionNotFound) Error() string { return "web push subscription not found" }

type ErrWebPushEndpointOwned struct{}

func (ErrWebPushEndpointOwned) Error() string { return "push endpoint belongs to another account" }

type ErrWebPushSessionInvalid struct{}

func (ErrWebPushSessionInvalid) Error() string { return "web push requires an active user session" }

type ErrWebPushSubscriptionInvalid struct{}

func (ErrWebPushSubscriptionInvalid) Error() string {
	return "web push subscription is no longer valid"
}

func DeleteWebPushSubscription(s *xorm.Session, userID int64, sessionID, deviceID string) error {
	subscription := &WebPushSubscription{}
	found, err := s.Where("user_id = ? AND session_id = ? AND device_id = ?", userID, sessionID, deviceID).Get(subscription)
	if err != nil || !found {
		return err
	}
	return deleteWebPushSubscription(s, subscription.ID)
}

func deleteWebPushSubscription(s *xorm.Session, subscriptionID int64) error {
	if _, err := s.Where("subscription_id = ?", subscriptionID).Delete(&WebPushDelivery{}); err != nil {
		return err
	}
	_, err := s.ID(subscriptionID).Delete(&WebPushSubscription{})
	return err
}

func DeleteWebPushSubscriptionsForSession(s *xorm.Session, sessionID string) error {
	if sessionID == "" {
		return nil
	}
	return deleteWebPushSubscriptionsWhere(s, "session_id = ?", sessionID)
}

func DeleteWebPushSubscriptionsForUser(s *xorm.Session, userID int64) error {
	return deleteWebPushSubscriptionsWhere(s, "user_id = ?", userID)
}

func deleteWebPushSubscriptionsWhere(s *xorm.Session, query string, arg any) error {
	subscriptions := []*WebPushSubscription{}
	if err := s.Where(query, arg).Find(&subscriptions); err != nil {
		return err
	}
	for _, subscription := range subscriptions {
		if err := deleteWebPushSubscription(s, subscription.ID); err != nil {
			return err
		}
	}
	return nil
}

// Push remains useful after login inactivity; only explicit revocation ends the subscription.
func HasWebPushSubscription(s *xorm.Session, userID int64) (bool, error) {
	// Compare in UTC: stored timestamps are UTC, but xorm does not apply the
	// database time zone to a time.Time bound into a raw Where string, so a
	// local time.Time would skew string comparisons on SQLite.
	now := time.Now().UTC()
	return s.Table("web_push_subscriptions").
		Where("user_id = ?", userID).
		And("(expiration_time IS NULL OR expiration_time > ?)", now).
		Exist(&WebPushSubscription{})
}

func webPushSessionCutoffs(now time.Time) (shortCutoff, longCutoff time.Time) {
	shortMaxAge := time.Duration(config.ServiceJWTTTL.GetInt64()) * time.Second
	longMaxAge := time.Duration(config.ServiceJWTTTLLong.GetInt64()) * time.Second
	return now.Add(-shortMaxAge), now.Add(-longMaxAge)
}

func webPushSessionIsActive(session *webPushSessionState, now time.Time) bool {
	shortCutoff, longCutoff := webPushSessionCutoffs(now)
	if session.IsLongSession {
		return !session.LastActive.Before(longCutoff)
	}
	return !session.LastActive.Before(shortCutoff)
}

func enqueueWebPush(s *xorm.Session, userID int64, deliveryKey string, message *WebPushMessage, ttl time.Duration) error {
	if !config.WebPushEnabled.GetBool() || message == nil || deliveryKey == "" {
		return nil
	}
	message.Tag = stablePushTag(deliveryKey)
	body := []rune(message.Body)
	if len(body) > 256 {
		message.Body = string(body[:256]) + "…"
	}
	payload, err := json.Marshal(message)
	if err != nil {
		return err
	}
	if len(payload) > 3000 {
		return fmt.Errorf("web push payload exceeds safe encrypted size")
	}

	// UTC throughout: stored timestamps are UTC and xorm does not convert a
	// time.Time bound into a raw Where string to the database time zone.
	now := time.Now().UTC()
	subscriptions := []*WebPushSubscription{}
	err = s.Where("user_id = ?", userID).
		And("(expiration_time IS NULL OR expiration_time > ?)", now).
		Find(&subscriptions)
	if err != nil {
		return err
	}
	for _, subscription := range subscriptions {
		delivery := &WebPushDelivery{
			SubscriptionID: subscription.ID,
			DeliveryKey:    deliveryKey,
			Payload:        string(payload),
			NextAttemptAt:  now,
			ExpiresAt:      now.Add(ttl),
		}
		if err := insertWebPushDeliveryIfMissing(s, delivery); err != nil {
			return err
		}
	}
	return nil
}

func insertWebPushDeliveryIfMissing(s *xorm.Session, delivery *WebPushDelivery) error {
	exists, err := s.Where("subscription_id = ? AND delivery_key = ?", delivery.SubscriptionID, delivery.DeliveryKey).Exist(&WebPushDelivery{})
	if err != nil || exists {
		return err
	}
	_, err = s.Insert(delivery)
	if err == nil {
		wakeWebPushWorker()
	}
	return err
}

func SendWebPushTest(ctx context.Context, s *xorm.Session, userID int64, sessionID, deviceID string, message WebPushMessage) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	subscription := &WebPushSubscription{}
	found, err := s.Where("user_id = ? AND session_id = ? AND device_id = ?", userID, sessionID, deviceID).Get(subscription)
	if err != nil {
		return err
	}
	if !found {
		return ErrWebPushSubscriptionNotFound{}
	}
	message.Test = true
	message.Tag = stablePushTag("test:" + strconv.FormatInt(subscription.ID, 10))
	payload, err := json.Marshal(&message)
	if err != nil {
		return err
	}
	response, err := sendWebPush(ctx, subscription, payload, message.Tag, time.Minute)
	if err != nil {
		log.Errorf("Web Push test transport error for subscription %d (%s): %s", subscription.ID, endpointHost(subscription.Endpoint), err)
		return err
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 2048))
	if response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone {
		if err := deleteWebPushSubscription(s, subscription.ID); err != nil {
			return err
		}
		return ErrWebPushSubscriptionInvalid{}
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {

		return fmt.Errorf("push service returned status %d", response.StatusCode)
	}
	return nil
}

func endpointHost(endpoint string) string {
	if parsed, err := url.Parse(endpoint); err == nil && parsed.Host != "" {
		return parsed.Host
	}
	return "push service"
}

func sendWebPush(ctx context.Context, subscription *WebPushSubscription, payload []byte, topic string, ttl time.Duration) (*http.Response, error) {
	client := webPushHTTPClient
	if client == nil {
		client = utils.NewSSRFSafeHTTPClient()
	}
	return webpush.SendNotificationWithContext(ctx, payload, &webpush.Subscription{
		Endpoint: subscription.Endpoint,
		Keys: webpush.Keys{
			P256dh: subscription.P256DH,
			Auth:   subscription.Auth,
		},
	}, &webpush.Options{
		HTTPClient:      retryableWebPushClient{client: client},
		Subscriber:      webPushSubscriber(),
		VAPIDPublicKey:  config.WebPushPublicKey.GetString(),
		VAPIDPrivateKey: config.WebPushPrivateKey.GetString(),
		Topic:           strings.TrimPrefix(topic, "vikunja-"),
		TTL:             max(0, int(ttl.Seconds())),
	})
}

// webpush-go adds mailto itself; passing a prefixed address breaks Apple VAPID validation.
func webPushSubscriber() string {
	if email := config.MailerFromEmail.GetString(); isEmailLikeAddress(email) {
		return email
	}
	if parsed, err := url.Parse(config.ServicePublicURL.GetString()); err == nil && parsed.Hostname() != "" {
		return "webmaster@" + parsed.Hostname()
	}
	return "webmaster@localhost"
}

func isEmailLikeAddress(value string) bool {
	at := strings.IndexByte(value, '@')
	if at <= 0 || at == len(value)-1 {
		return false
	}
	if strings.ContainsAny(value, " \t\r\n") {
		return false
	}
	return strings.Contains(value[at+1:], ".")
}

func StartWebPushWorker() {
	if !config.WebPushEnabled.GetBool() {
		return
	}
	webPushWorkerMu.Lock()
	defer webPushWorkerMu.Unlock()
	if webPushWorkerCancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	webPushWorkerCancel = cancel
	webPushWorkerDone = make(chan struct{})
	workerID := uuid.NewString()
	done := webPushWorkerDone
	go func() {
		defer close(done)
		runWebPushWorker(ctx, workerID)
	}()
	wakeWebPushWorker()
}

func StopWebPushWorker() {
	webPushWorkerMu.Lock()
	cancel := webPushWorkerCancel
	done := webPushWorkerDone
	webPushWorkerCancel = nil
	webPushWorkerDone = nil
	webPushWorkerMu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		log.Errorf("Timed out waiting for Web Push worker to stop")
	}
}

func wakeWebPushWorker() {
	select {
	case webPushWake <- struct{}{}:
	default:
	}
}

func runWebPushWorker(ctx context.Context, workerID string) {
	ticker := time.NewTicker(webPushPollInterval)
	defer ticker.Stop()
	defer releaseWebPushLeases(workerID)
	for {
		select {
		case <-ctx.Done():
			return
		case <-webPushWake:
			processWebPushBatch(ctx, workerID)
		case <-ticker.C:
			processWebPushBatch(ctx, workerID)
		}
	}
}

func processWebPushBatch(ctx context.Context, workerID string) {
	for {
		if ctx.Err() != nil {
			return
		}
		deliveries, err := claimWebPushDeliveries(workerID, time.Now())
		if err != nil {
			log.Errorf("Could not claim Web Push deliveries: %s", err)
			return
		}
		if len(deliveries) == 0 {
			return
		}
		for _, delivery := range deliveries {
			processWebPushDelivery(ctx, delivery)
		}
	}
}

func claimWebPushDeliveries(workerID string, now time.Time) ([]*WebPushDelivery, error) {
	// Stored timestamps are UTC; compare in UTC so SQLite string comparisons
	// are not skewed by the process time zone.
	now = now.UTC()
	s := db.NewSession()
	defer s.Close()
	expired, err := s.Where("expires_at <= ?", now).Delete(&WebPushDelivery{})
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	if expired > 0 {
		metrics.WebPushDeliveryOutcomes.WithLabelValues("expired").Add(float64(expired))
	}
	candidates := []*WebPushDelivery{}
	err = s.Where("next_attempt_at <= ? AND expires_at > ? AND (lease_until IS NULL OR lease_until < ?)", now, now, now).
		OrderBy("next_attempt_at, id").Limit(webPushBatchSize).Find(&candidates)
	if err != nil {
		_ = s.Rollback()
		return nil, err
	}
	claimed := make([]*WebPushDelivery, 0, len(candidates))
	leaseUntil := now.Add(webPushLeaseDuration)
	for _, delivery := range candidates {
		affected, updateErr := s.ID(delivery.ID).
			Where("lease_until IS NULL OR lease_until < ?", now).
			Cols("lease_owner", "lease_until", "updated").
			Update(&WebPushDelivery{LeaseOwner: workerID, LeaseUntil: &leaseUntil})
		if updateErr != nil {
			_ = s.Rollback()
			return nil, updateErr
		}
		if affected == 1 {
			delivery.LeaseOwner = workerID
			delivery.LeaseUntil = &leaseUntil
			claimed = append(claimed, delivery)
		}
	}
	if err := s.Commit(); err != nil {
		return nil, err
	}
	return claimed, nil
}

func processWebPushDelivery(ctx context.Context, delivery *WebPushDelivery) {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	s := db.NewSession()
	subscription := &WebPushSubscription{}
	found, err := s.ID(delivery.SubscriptionID).Get(subscription)
	if err == nil && found && subscription.ExpirationTime != nil && !subscription.ExpirationTime.After(time.Now()) {
		found = false
	}
	_ = s.Rollback()
	s.Close()
	if err != nil {
		rescheduleWebPushDelivery(delivery, err, "")
		return
	}
	if !found {
		metrics.WebPushDeliveryOutcomes.WithLabelValues("invalid_subscription").Inc()
		deleteWebPushSubscriptionByID(delivery.SubscriptionID)
		return
	}

	ttl := time.Until(delivery.ExpiresAt)
	response, sendErr := sendWebPush(ctx, subscription, []byte(delivery.Payload), stablePushTag(delivery.DeliveryKey), ttl)
	if sendErr != nil {
		var retryable retryableWebPushError
		if errors.As(sendErr, &retryable) {
			metrics.WebPushDeliveryOutcomes.WithLabelValues("retry").Inc()
			rescheduleWebPushDelivery(delivery, sendErr, "")
			return
		}
		metrics.WebPushDeliveryOutcomes.WithLabelValues("permanent_failure").Inc()
		log.Errorf("Dropping permanent Web Push delivery %d: %s", delivery.ID, sendErr)
		deleteWebPushDelivery(delivery)
		return
	}
	defer response.Body.Close()
	_, _ = io.Copy(io.Discard, io.LimitReader(response.Body, 2048))

	switch {
	case response.StatusCode >= 200 && response.StatusCode < 300:
		metrics.WebPushDeliveryOutcomes.WithLabelValues("accepted").Inc()
		log.Debugf("Web Push delivery %d accepted by push service", delivery.ID)
		deleteWebPushDelivery(delivery)
	case response.StatusCode == http.StatusNotFound || response.StatusCode == http.StatusGone:
		metrics.WebPushDeliveryOutcomes.WithLabelValues("invalid_subscription").Inc()
		deleteWebPushSubscriptionByID(delivery.SubscriptionID)
	case response.StatusCode == http.StatusRequestTimeout || response.StatusCode == http.StatusTooManyRequests || response.StatusCode >= 500:
		metrics.WebPushDeliveryOutcomes.WithLabelValues("retry").Inc()
		rescheduleWebPushDelivery(delivery, fmt.Errorf("push service returned status %d", response.StatusCode), response.Header.Get("Retry-After"))
	default:
		metrics.WebPushDeliveryOutcomes.WithLabelValues("permanent_failure").Inc()
		log.Errorf("Dropping permanent Web Push delivery %d after status %d", delivery.ID, response.StatusCode)
		deleteWebPushDelivery(delivery)
	}
}

func WakeWebPushWorker() {
	wakeWebPushWorker()
}

func rescheduleWebPushDelivery(delivery *WebPushDelivery, deliveryErr error, retryAfter string) {
	now := time.Now().UTC()
	if !delivery.ExpiresAt.After(now) {
		deleteWebPushDelivery(delivery)
		return
	}
	delay := webPushInitialBackoff << min(delivery.Attempts, 5)
	if delay > webPushMaximumBackoff {
		delay = webPushMaximumBackoff
	}
	if parsed := parseRetryAfter(retryAfter, now); parsed > delay {
		delay = parsed
	}
	next := now.Add(delay)
	if next.After(delivery.ExpiresAt) {
		deleteWebPushDelivery(delivery)
		return
	}
	message := deliveryErr.Error()
	if len(message) > webPushMaximumErrorLen {
		message = message[:webPushMaximumErrorLen]
	}
	s := db.NewSession()
	defer s.Close()
	_, err := s.ID(delivery.ID).
		Where("lease_owner = ?", delivery.LeaseOwner).
		Cols("attempts", "next_attempt_at", "lease_owner", "lease_until", "last_error", "updated").
		Nullable("lease_until").
		Update(&WebPushDelivery{Attempts: delivery.Attempts + 1, NextAttemptAt: next, LeaseOwner: "", LeaseUntil: nil, LastError: message})
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Could not reschedule Web Push delivery %d: %s", delivery.ID, err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit Web Push retry %d: %s", delivery.ID, err)
	}
}

func parseRetryAfter(value string, now time.Time) time.Duration {
	if value == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}
	when, err := http.ParseTime(value)
	if err != nil || !when.After(now) {
		return 0
	}
	return when.Sub(now)
}

func deleteWebPushDelivery(delivery *WebPushDelivery) {
	s := db.NewSession()
	defer s.Close()
	if _, err := s.ID(delivery.ID).Where("lease_owner = ?", delivery.LeaseOwner).Delete(&WebPushDelivery{}); err != nil {
		_ = s.Rollback()
		log.Errorf("Could not delete Web Push delivery %d: %s", delivery.ID, err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit Web Push delivery deletion %d: %s", delivery.ID, err)
	}
}

func deleteWebPushSubscriptionByID(id int64) {
	s := db.NewSession()
	defer s.Close()
	if err := deleteWebPushSubscription(s, id); err != nil {
		_ = s.Rollback()
		log.Errorf("Could not remove invalid Web Push subscription %d: %s", id, err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit Web Push subscription deletion %d: %s", id, err)
	}
}

func releaseWebPushLeases(workerID string) {
	s := db.NewSession()
	defer s.Close()
	_, err := s.Where("lease_owner = ?", workerID).
		Cols("lease_owner", "lease_until", "updated").
		Nullable("lease_until").
		Update(&WebPushDelivery{LeaseOwner: "", LeaseUntil: nil})
	if err != nil {
		_ = s.Rollback()
		log.Errorf("Could not release Web Push leases: %s", err)
		return
	}
	if err := s.Commit(); err != nil {
		log.Errorf("Could not commit released Web Push leases: %s", err)
	}
}
