import {UserFactory, type UserAttributes} from '../../factories/user'
import {TokenFactory, type TokenAttributes} from '../../factories/token'

context('Password Reset', () => {
	let user: UserAttributes

	beforeEach(() => {
		UserFactory.truncate()
		TokenFactory.truncate()
		user = UserFactory.create(1)[0] as UserAttributes
	})

	it('Should allow a user to reset their password with a valid token', () => {
		const tokenArray = TokenFactory.create(1, {user_id: user.id as number, kind: 1})
		const token: TokenAttributes = tokenArray[0] as TokenAttributes

		cy.visit(`/?userPasswordReset=${token.token}`)
		cy.url().should('include', `/password-reset?userPasswordReset=${token.token}`)

		const newPassword = 'newSecurePassword123'
		cy.get('input[id=password]').type(newPassword)
		cy.get('button').contains('Reset your password').click()

		cy.get('.message.success').should('contain', 'The password was updated successfully.')
		cy.get('.button').contains('Login').click()
		cy.url().should('include', '/login')

		// Try to login with the new password
		cy.get('input[id=username]').type(user.username)
		cy.get('input[id=password]').type(newPassword)
		cy.get('.button').contains('Login').click()
		cy.url().should('not.include', '/login')
	})

	it('Should show an error for an invalid token', () => {
		cy.visit('/?userPasswordReset=invalidtoken123')
		cy.url().should('include', '/password-reset?userPasswordReset=invalidtoken123')

		// Attempt to reset password
		const newPassword = 'newSecurePassword123'
		cy.get('input[id=password]').type(newPassword)
		cy.get('button').contains('Reset your password').click()

		cy.get('.message').should('contain', 'Invalid token')
	})

	it('Should redirect to login if no token is present in query param when visiting /password-reset directly', () => {
		cy.visit('/password-reset')
		cy.url().should('not.include', '/password-reset')
		cy.wait(1000) // Wait for the redirect to happen - this seems to be flaky in CI
		cy.url().should('include', '/login')
	})

	it('Should redirect to login if userPasswordReset token is not present in query param when visiting root', () => {
		cy.visit('/')
		cy.url().should('include', '/login')
	})
}) 
