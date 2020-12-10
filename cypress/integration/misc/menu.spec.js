import '../../support/authenticateUser'

describe('The Menu', () => {
	it('Is visible by default on desktop', () => {
		cy.get('.namespace-container')
			.should('have.class', 'is-active')
	})

	it('Can be hidden on desktop', () => {
		cy.get('a.menu-show-button:visible')
			.click()
		cy.get('.namespace-container')
			.should('not.have.class', 'is-active')
	})

	it('Is hidden by default on mobile', () => {
		cy.viewport('iphone-8')
		cy.get('.namespace-container')
			.should('not.have.class', 'is-active')
	})

	it('Is can be shown on mobile', () => {
		cy.viewport('iphone-8')
		cy.get('a.menu-show-button:visible')
			.click()
		cy.get('.namespace-container')
			.should('have.class', 'is-active')
	})
})
