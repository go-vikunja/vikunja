// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

package migration

import (
	"time"

	"src.techknowlogick.com/xormigrate"
	"xorm.io/xorm"
)

type webPushSubscription20260807120000 struct {
	ID             int64      `xorm:"bigint autoincr not null unique pk"`
	UserID         int64      `xorm:"bigint not null index unique(user_device)"`
	SessionID      string     `xorm:"varchar(36) not null index"`
	DeviceID       string     `xorm:"varchar(36) not null unique(user_device)"`
	Endpoint       string     `xorm:"text not null"`
	EndpointHash   string     `xorm:"char(64) not null unique"`
	P256DH         string     `xorm:"text not null"`
	Auth           string     `xorm:"text not null"`
	ExpirationTime *time.Time `xorm:"datetime null"`
	Created        time.Time  `xorm:"created not null"`
	Updated        time.Time  `xorm:"updated not null"`
}

func (*webPushSubscription20260807120000) TableName() string { return "web_push_subscriptions" }

type webPushDelivery20260807120000 struct {
	ID             int64      `xorm:"bigint autoincr not null unique pk"`
	SubscriptionID int64      `xorm:"bigint not null index unique(subscription_delivery)"`
	DeliveryKey    string     `xorm:"varchar(255) not null unique(subscription_delivery)"`
	Payload        string     `xorm:"text not null"`
	Attempts       int        `xorm:"not null default 0"`
	NextAttemptAt  time.Time  `xorm:"datetime not null index"`
	LeaseOwner     string     `xorm:"varchar(36) null"`
	LeaseUntil     *time.Time `xorm:"datetime null index"`
	LastError      string     `xorm:"text null"`
	ExpiresAt      time.Time  `xorm:"datetime not null index"`
	Created        time.Time  `xorm:"created not null"`
	Updated        time.Time  `xorm:"updated not null"`
}

func (*webPushDelivery20260807120000) TableName() string { return "web_push_deliveries" }

func init() {
	migrations = append(migrations, &xormigrate.Migration{
		ID:          "20260807120000",
		Description: "Add durable Web Push subscriptions and deliveries",
		Migrate: func(tx *xorm.Engine) error {
			return tx.Sync2(&webPushSubscription20260807120000{}, &webPushDelivery20260807120000{}) //nolint:forbidigo // brand-new tables
		},
		Rollback: func(tx *xorm.Engine) error {
			return nil
		},
	})
}
