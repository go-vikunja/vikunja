// This test assumes no mailer is set up and all users are activated immediately.

import {UserFactory} from '../../factories/user'

context('Registration', () => {
	beforeEach(() => {
		UserFactory.create(1, {
			username: 'test',
		})
		cy.visit('/', {
			onBeforeLoad(win) {
				win.localStorage.removeItem('token')
			},
		})
	})

	it('Should work without issues', () => {
		const fixture = {
			username: 'testuser',
			password: '12345678',
			email: 'testuser@example.com',
		}

		cy.visit('/register')
		cy.get('#username').type(fixture.username)
		cy.get('#email').type(fixture.email)
		cy.get('#password').type(fixture.password)
		cy.get('#register-submit').click()
		cy.url().should('include', '/')
		cy.clock(1625656161057) // 13:00
		cy.get('h2').should('contain', `Hi ${fixture.username}!`)
	})

	it('Should fail', () => {
		const fixture = {
			username: 'test',
			password: '12345678',
			email: 'testuser@example.com',
		}

		cy.visit('/register')
		cy.get('#username').type(fixture.username)
		cy.get('#email').type(fixture.email)
		cy.get('#password').type(fixture.password)
		cy.get('#register-submit').click()
		cy.get('div.message.danger').contains('A user with this username already exists.')
	})
})
