
// This authenticates a user and puts the token in local storage which allows us to perform authenticated requests.
// Built after https://github.com/cypress-io/cypress-example-recipes/tree/bd2d6ffb33214884cab343d38e7f9e6ebffb323f/examples/logging-in__jwt

import {UserFactory} from '../factories/user'

export function login(user, cacheAcrossSpecs = false) {
	if (!user) {
		throw new Error('Needs user')
	}
	// Caching session when logging in via page visit
	cy.session(`user__${user.username}`, () => {
		cy.request('POST', `${Cypress.env('API_URL')}/login`, {
			username: user.username,
			password: '1234',
		}).then(({ body }) => {
			window.localStorage.setItem('token', body.token)
		})
	}, {
    cacheAcrossSpecs,
  })
}

export function createFakeUser() {
	return UserFactory.create(1)[0]
}

export function createFakeUserAndLogin() {
	let user
	before(() => {
		user = createFakeUser()
	})

	beforeEach(() => {
		login(user, true)
	})

	return user
}
