import '../../support/authenticateUser'

const setHours = hours => {
	const date = new Date()
	date.setHours(hours)
	cy.clock(+date)
}

describe('Home Page', () => {
	it('shows the right salutation in the night', () => {
		setHours(4)
		cy.visit('/')
		cy.get('h2').should('contain', 'Good Night')
	})
	it('shows the right salutation in the morning', () => {
		setHours(8)
		cy.visit('/')
		cy.get('h2').should('contain', 'Good Morning')
	})
	it('shows the right salutation in the day', () => {
		setHours(13)
		cy.visit('/')
		cy.get('h2').should('contain', 'Hi')
	})
	it('shows the right salutation in the night', () => {
		setHours(20)
		cy.visit('/')
		cy.get('h2').should('contain', 'Good Evening')
	})
	it('shows the right salutation in the night again', () => {
		setHours(23)
		cy.visit('/')
		cy.get('h2').should('contain', 'Good Night')
	})
})