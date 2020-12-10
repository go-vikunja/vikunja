import '../../support/authenticateUser'

describe('Log out', () => {
	it('Logs the user out', () => {
		cy.visit('/')

		cy.get('.navbar .user .username')
			.click()
		cy.get('.navbar .user .dropdown-menu a.dropdown-item')
			.contains('Logout')
			.click()

		cy.url()
			.should('contain', '/login')
	})
})
