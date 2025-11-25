import {UserFactory} from '../../factories/user'
import {TokenFactory} from '../../factories/token'

context('Email Confirmation', () => {
	let user
	let confirmationToken

	beforeEach(() => {
		UserFactory.truncate()
		TokenFactory.truncate()

		// Create a user with status = 1 (StatusEmailConfirmationRequired)
		user = UserFactory.create(1, {
			username: 'unconfirmeduser',
			email: 'unconfirmed@example.com',
			password: '$2a$14$dcadBoMBL9jQoOcZK8Fju.cy0Ptx2oZECkKLnaa8ekRoTFe1w7To.', // 1234
			status: 1, // StatusEmailConfirmationRequired
		})[0]

		// Create an email confirmation token for this user
		// kind: 2 = TokenEmailConfirm
		confirmationToken = 'test-email-confirm-token-12345678901234567890123456789012'
		TokenFactory.create(1, {
			user_id: user.id,
			kind: 2,
			token: confirmationToken,
		})
	})

	it('Should fail login before email is confirmed', () => {
		cy.visit('/login')
		cy.get('input[id=username]').type(user.username)
		cy.get('input[id=password]').type('1234')
		cy.get('.button').contains('Login').click()

		cy.get('div.message.danger').contains('Email address of the user not confirmed')
	})

	it('Should confirm email and allow login', () => {
		// Intercept the confirmation API call
		cy.intercept('POST', '**/user/confirm').as('confirmEmail')

		// Manually set the token in localStorage before visiting the page
		// This simulates what happens when the user clicks the email link
		cy.visit('/login', {
			onBeforeLoad(win) {
				win.localStorage.setItem('emailConfirmToken', confirmationToken)
			},
		})

		// Wait for the confirmation API call to complete
		cy.wait('@confirmEmail', {timeout: 10000}).its('response.statusCode').should('eq', 200)

		// Should show success message
		cy.get('.message.success', {timeout: 10000}).should('be.visible')
		cy.get('.message.success').contains('You successfully confirmed your email')

		// Now login should work
		cy.get('input[id=username]').type(user.username)
		cy.get('input[id=password]').type('1234')
		cy.get('.button').contains('Login').click()

		// Should successfully log in
		cy.url().should('include', '/')
		cy.url().should('not.include', '/login')
		// Check that the username appears in the greeting
		cy.contains(user.username)
	})

	it('Should fail with invalid confirmation token', () => {
		// Intercept the confirmation API call
		cy.intercept('POST', '**/user/confirm').as('confirmEmail')

		// Try to confirm with an invalid token
		const invalidToken = 'invalid-token-that-does-not-exist-in-database'
		cy.visit('/login', {
			onBeforeLoad(win) {
				win.localStorage.setItem('emailConfirmToken', invalidToken)
			},
		})

		// Wait for the confirmation API call to fail
		cy.wait('@confirmEmail', {timeout: 10000})

		// Should show error message
		cy.get('.message.danger', {timeout: 10000}).should('be.visible')

		// Login should still fail
		cy.get('input[id=username]').type(user.username)
		cy.get('input[id=password]').type('1234')
		cy.get('.button').contains('Login').click()

		cy.get('div.message.danger').contains('Email address of the user not confirmed')
	})

	it('Should not allow using the same token twice', () => {
		// Intercept the confirmation API call
		cy.intercept('POST', '**/user/confirm').as('confirmEmail')

		// First confirmation - should work
		cy.visit('/login', {
			onBeforeLoad(win) {
				win.localStorage.setItem('emailConfirmToken', confirmationToken)
			},
		})
		cy.wait('@confirmEmail', {timeout: 10000}).its('response.statusCode').should('eq', 200)
		cy.get('.message.success', {timeout: 10000}).should('be.visible')
		cy.get('.message.success').contains('You successfully confirmed your email')

		// Try to use the same token again - should fail
		cy.visit('/login', {
			onBeforeLoad(win) {
				win.localStorage.setItem('emailConfirmToken', confirmationToken)
			},
		})
		cy.wait('@confirmEmail', {timeout: 10000})
		cy.get('.message.danger', {timeout: 10000}).should('be.visible')
	})

	it('Should confirm email when clicking link from email (via query parameter)', () => {
		// Intercept the confirmation API call
		cy.intercept('POST', '**/user/confirm').as('confirmEmail')

		// Simulate clicking the email confirmation link with query parameter
		// This is what happens when a user clicks the link in their email
		cy.visit(`/?userEmailConfirm=${confirmationToken}`)

		// Should redirect to login page
		cy.url().should('include', '/login')

		// Wait for the confirmation API call to complete
		cy.wait('@confirmEmail', {timeout: 10000}).its('response.statusCode').should('eq', 200)

		// Should show success message
		cy.get('.message.success', {timeout: 10000}).should('be.visible')
		cy.get('.message.success').contains('You successfully confirmed your email')

		// Now login should work
		cy.get('input[id=username]').type(user.username)
		cy.get('input[id=password]').type('1234')
		cy.get('.button').contains('Login').click()

		// Should successfully log in
		cy.url().should('include', '/')
		cy.url().should('not.include', '/login')
		// Check that the username appears in the greeting
		cy.contains(user.username)
	})
})
