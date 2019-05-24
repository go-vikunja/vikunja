package caldav

import (
	"github.com/samedi/caldav-go/data"
	"github.com/samedi/caldav-go/global"
)

// SetupStorage sets the storage to be used by the server. The storage is where the resources data will be fetched from.
// You can provide a custom storage for your own purposes (which might be looking for data in the cloud, DB, etc).
// Just make sure it implements the `data.Storage` interface.
func SetupStorage(stg data.Storage) {
	global.Storage = stg
}

// SetupUser sets the current user which is currently interacting with the calendar.
// It is used, for example, in some of the CALDAV responses, when rendering the path where to find the user's resources.
func SetupUser(username string) {
	global.User = &data.CalUser{Name: username}
}

// SetupSupportedComponents sets all components which are supported by this storage implementation.
func SetupSupportedComponents(components []string) {
	global.SupportedComponents = components
}
