// Package global defines the globally accessible variables in the caldav server
// and the interface to setup them.
package global

import (
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/lib"
)

// Storage represents the global storage used in the CRUD operations of resources. Default storage is the `data.FileStorage`.
var Storage data.Storage = new(data.FileStorage)

// User defines the current caldav user, which is the user currently interacting with the calendar.
var User *data.CalUser

// SupportedComponents contains all components which are supported by the current storage implementation
var SupportedComponents = []string{lib.VCALENDAR, lib.VEVENT}
