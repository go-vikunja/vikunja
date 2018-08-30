package v1

// swagger:operation DELETE /tasks/{taskID} lists deleteListTask
// ---
// summary: Deletes a list task
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: taskID
//   in: path
//   description: ID of the list task to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "404":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /lists/{listID} lists deleteList
// ---
// summary: Deletes a list with all tasks on it
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "404":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /lists/{listID} lists addListTask
// ---
// summary: Adds an task to a list
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to use
//   type: string
//   required: true
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/ListTask"
// responses:
//   "200":
//     "$ref": "#/responses/ListTask"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation POST /tasks/{taskID} lists updateListTask
// ---
// summary: Updates a list task
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: taskID
//   in: path
//   description: ID of the task to update
//   type: string
//   required: true
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/ListTask"
// responses:
//   "200":
//     "$ref": "#/responses/ListTask"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /lists/{listID} lists getList
// ---
// summary: gets one list with all todo tasks
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to show
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/List"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /namespaces/{namespaceID}/lists lists addList
// ---
// summary: Creates a new list owned by the currently logged in user in that namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace that list should belong to
//   type: string
//   required: true
// - name: body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/List"
// responses:
//   "200":
//     "$ref": "#/responses/List"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /namespaces/{namespaceID}/teams teams addTeamToNamespace
// ---
// summary: Gives a team access to a namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace that list should belong to
//   type: string
//   required: true
// - name: body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/TeamNamespace"
// responses:
//   "200":
//     "$ref": "#/responses/TeamNamespace"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /namespaces/{namespaceID}/teams/{teamID} teams removeTeamFromNamespace
// ---
// summary: Removes a team from a namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace
//   type: string
//   required: true
// - name: teamID
//   in: path
//   description: ID of the team you want to remove
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation POST /lists/{listID} lists upadteList
// ---
// summary: Updates a list
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to update
//   type: string
//   required: true
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/List"
// responses:
//   "200":
//     "$ref": "#/responses/List"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /lists lists getLists
// ---
// summary: Gets all lists owned by the current user
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   "200":
//     "$ref": "#/responses/List"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /namespaces namespaces addNamespace
// ---
// summary: Creates a new namespace owned by the currently logged in user
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/Namespace"
// responses:
//   "200":
//     "$ref": "#/responses/Namespace"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation POST /namespaces/{namespaceID} namespaces upadteNamespace
// ---
// summary: Updates a namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace to update
//   type: string
//   required: true
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/Namespace"
// responses:
//   "200":
//     "$ref": "#/responses/Namespace"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /namespaces/{namespaceID} namespaces deleteNamespace
// ---
// summary: Deletes a namespace with all lists
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "404":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /namespaces/{namespaceID} namespaces getNamespace
// ---
// summary: gets one namespace with all todo tasks
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace to show
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Namespace"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /namespaces/{namespaceID}/lists lists getNamespaceLists
// ---
// summary: gets all lists in that namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace to show
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/List"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /namespaces/{namespaceID}/teams teams getNamespaceTeams
// ---
// summary: gets all teams which have access to that namespace
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: namespaceID
//   in: path
//   description: ID of the namespace to show
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /namespaces namespaces getNamespaces
// ---
// summary: Get all namespaces the currently logged in user has at least read access
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   "200":
//     "$ref": "#/responses/Namespace"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /lists/{listID}/teams teams getTeamsByList
// ---
// summary: gets all teams which have access to the list
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to show
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /lists/{listID}/teams teams addTeamToList
// ---
// summary: adds a team to a list
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list to show
//   type: string
//   required: true
// - name: body
//   in: body
//   schema:
//     "$ref": "#/definitions/TeamList"
// responses:
//   "200":
//     "$ref": "#/responses/TeamList"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /lists/{listID}/teams/{teamID} teams deleteTeamFromList
// ---
// summary: removes a team from a list
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: listID
//   in: path
//   description: ID of the list
//   type: string
//   required: true
// - name: teamID
//   in: path
//   description: ID of the team to remove
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /teams teams getTeams
// ---
// summary: gets all teams the current user is part of
// consumes:
// - application/json
// produces:
// - application/json
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation GET /teams/{teamID} teams getTeamByID
// ---
// summary: gets infos about the team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: teamID
//   in: path
//   description: ID of the team
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /teams teams createTeam
// ---
// summary: Creates a team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/Team"
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation POST /teams/{teamID} teams updateTeam
// ---
// summary: Updates a team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: teamID
//   in: path
//   description: ID of the team you want to update
//   type: string
//   required: true
// - name: body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/Team"
// responses:
//   "200":
//     "$ref": "#/responses/Team"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /teams/{teamID} teams deleteTeam
// ---
// summary: Deletes a team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: teamID
//   in: path
//   description: ID of the team you want to delete
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation PUT /teams/{teamID}/members teams addTeamMember
// ---
// summary: Adds a member to a team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: teamID
//   in: path
//   description: ID of the team you want to add a member to
//   type: string
//   required: true
// - name: body
//   in: body
//   required: true
//   schema:
//     "$ref": "#/definitions/TeamMember"
// responses:
//   "200":
//     "$ref": "#/responses/TeamMember"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"

// swagger:operation DELETE /teams/{teamID}/members/{userID} teams removeTeamMember
// ---
// summary: Removes a member from a team
// consumes:
// - application/json
// produces:
// - application/json
// parameters:
// - name: teamID
//   in: path
//   description: ID of the team you want to delete a member
//   type: string
//   required: true
// - name: userID
//   in: path
//   description: ID of the user you want to remove from the team
//   type: string
//   required: true
// responses:
//   "200":
//     "$ref": "#/responses/Message"
//   "400":
//     "$ref": "#/responses/Message"
//   "403":
//     "$ref": "#/responses/Message"
//   "500":
//     "$ref": "#/responses/Message"
