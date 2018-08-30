package swagger

import "code.vikunja.io/api/models"

// Message
// swagger:response Message
type swaggerResponseMessage struct {
	// in:body
	Body models.Message `json:"body"`
}

// ================
// User definitions
// ================

// User Object
// swagger:response User
type swaggerResponseUser struct {
	// in:body
	Body models.User `json:"body"`
}

// Token
// swagger:response Token
type swaggerResponseToken struct {
	// The body message
	// in:body
	Body struct {
		// The token
		//
		// Required: true
		Token string `json:"token"`
	} `json:"body"`
}

// ================
// List definitions
// ================

// List
// swagger:response List
type swaggerResponseLIst struct {
	// in:body
	Body models.List `json:"body"`
}

// ListTask
// swagger:response ListTask
type swaggerResponseLIstTask struct {
	// in:body
	Body models.ListTask `json:"body"`
}

// ================
// Namespace definitions
// ================

// Namespace
// swagger:response Namespace
type swaggerResponseNamespace struct {
	// in:body
	Body models.Namespace `json:"body"`
}

// ================
// Team definitions
// ================

// Team
// swagger:response Team
type swaggerResponseTeam struct {
	// in:body
	Body models.Team `json:"body"`
}

// TeamMember
// swagger:response TeamMember
type swaggerResponseTeamMember struct {
	// in:body
	Body models.TeamMember `json:"body"`
}

// TeamList
// swagger:response TeamList
type swaggerResponseTeamList struct {
	// in:body
	Body models.TeamList `json:"body"`
}

// TeamNamespace
// swagger:response TeamNamespace
type swaggerResponseTeamNamespace struct {
	// in:body
	Body models.TeamNamespace `json:"body"`
}
