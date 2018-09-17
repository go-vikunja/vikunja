package swagger

import "code.vikunja.io/api/models"

// not actually a response, just a hack to get go-swagger to include definitions
// of the various XYZOption structs

// parameterBodies
// swagger:response parameterBodies
type swaggerParameterBodies struct {
	// in:body
	UserLogin models.UserLogin

	// in:body
	APIUserPassword models.APIUserPassword

	// in:body
	List models.List

	// in:body
	ListTask models.ListTask

	// in:body
	Namespace models.Namespace

	// in:body
	Team models.Team

	// in:body
	TeamMember models.TeamMember

	// in:body
	TeamList models.TeamList

	// in:body
	TeamNamespace models.TeamNamespace

	// in:body
	ListUser models.ListUser

	// in:body
	NamespaceUser models.NamespaceUser
}
