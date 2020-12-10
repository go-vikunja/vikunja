
// This authenticates a user and puts the token in local storage which allows us to perform authenticated requests.
// Built after https://github.com/cypress-io/cypress-example-recipes/tree/bd2d6ffb33214884cab343d38e7f9e6ebffb323f/examples/logging-in__jwt

import {UserFactory} from '../factories/user'

let token

before(() => {
	const users = UserFactory.create(1)

	cy.request('POST', `${Cypress.env('API_URL')}/login`, {
		username: users[0].username,
		password: '1234',
	})
		.its('body')
		.then(r => {
			token = r.token
		})
})

beforeEach(() => {
	cy.log(`Using token ${token} to make authenticated requests`)
	cy.visit('/', {
		onBeforeLoad(win) {
			win.localStorage.setItem('token', token)
		},
	})
})
