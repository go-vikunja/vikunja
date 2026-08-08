// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package models

import (
	"encoding/base64"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/notifications"
	"code.vikunja.io/api/pkg/user"

	webpush "github.com/SherClockHolmes/webpush-go"
	"github.com/stretchr/testify/require"
)

func TestDeletingSessionRemovesWebPushSubscriptions(t *testing.T) {
	db.LoadAndAssertFixtures(t)
	privateKey, publicKey, err := webpush.GenerateVAPIDKeys()
	require.NoError(t, err)
	config.WebPushEnabled.Set(true)
	config.WebPushPublicKey.Set(publicKey)
	config.WebPushPrivateKey.Set(privateKey)
	t.Cleanup(func() {
		config.WebPushEnabled.Set(false)
		config.WebPushPublicKey.Set("")
		config.WebPushPrivateKey.Set("")
	})

	s := db.NewSession()
	session := &Session{}
	found, err := s.Where("user_id = ?", 1).Get(session)
	require.NoError(t, err)
	require.True(t, found)
	_, err = notifications.UpsertWebPushSubscription(s, 1, session.ID, "9a34527d-1357-4c64-8171-6d6e25f01d62", notifications.WebPushSubscriptionInput{
		Endpoint: "https://push.example.test/session-cleanup",
		P256DH:   publicKey,
		Auth:     base64.RawURLEncoding.EncodeToString([]byte("0123456789abcdef")),
	})
	require.NoError(t, err)
	require.NoError(t, s.Commit())
	s.Close()
	db.AssertExists(t, "web_push_subscriptions", map[string]any{"session_id": session.ID}, false)

	// A non-owner cannot remove either the session or its subscription, even
	// when Delete is called without the normal permission wrapper.
	s = db.NewSession()
	require.NoError(t, (&Session{ID: session.ID}).Delete(s, &user.User{ID: 2}))
	require.NoError(t, s.Commit())
	s.Close()
	db.AssertExists(t, "web_push_subscriptions", map[string]any{"session_id": session.ID}, false)

	s = db.NewSession()
	require.NoError(t, (&Session{ID: session.ID}).Delete(s, &user.User{ID: 1}))
	require.NoError(t, s.Commit())
	s.Close()
	db.AssertMissing(t, "web_push_subscriptions", map[string]any{"session_id": session.ID})
}
