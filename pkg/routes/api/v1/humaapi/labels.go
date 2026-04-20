package humaapi

import (
	"code.vikunja.io/api/pkg/models"

	"github.com/danielgtaylor/huma/v2"
)

// RegisterLabelRoutes wires Huma-flavoured Label CRUD operations onto the
// given Huma API. Runs alongside (not replacing) the legacy swag-driven
// routes for the duration of the spike.
func RegisterLabelRoutes(api huma.API) {
	Register[*models.Label](api, Config[*models.Label, SingleID]{
		Tag:      "labels",
		BasePath: "/labels",
		ItemPath: "/labels/{id}",
		New:      func() *models.Label { return &models.Label{} },
		ApplyPath: func(l *models.Label, p SingleID) {
			l.ID = p.ID
		},
	})
}
