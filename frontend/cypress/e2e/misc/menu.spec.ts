import {createFakeUserAndLogin} from '../../support/authenticateUser'

describe('The Menu', () => {
	createFakeUserAndLogin()

	beforeEach(() => {
		cy.visit('/')
	})

	it('Is visible by default on desktop', () => {
		cy.get('.menu-container')
			.should('have.class', 'is-active')
	})

	it('Can be hidden on desktop', () => {
		cy.get('button.menu-show-button:visible')
			.click()
		cy.get('.menu-container')
			.should('not.have.class', 'is-active')
	})

	it('Is hidden by default on mobile', () => {
		cy.viewport('iphone-8')
		cy.get('.menu-container')
			.should('not.have.class', 'is-active')
	})

	it('Is can be shown on mobile', () => {
		cy.viewport('iphone-8')
		cy.get('button.menu-show-button:visible')
			.click()
		cy.get('.menu-container')
			.should('have.class', 'is-active')
	})
})
