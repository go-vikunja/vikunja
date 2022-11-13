import {createFakeUserAndLogin} from '../../support/authenticateUser'
import {createProjects} from '../project/prepareProjects'

function logout() {
	cy.get('.navbar .username-dropdown-trigger')
		.click()
	cy.get('.navbar .dropdown-item')
		.contains('Logout')
		.click()
}

describe('Log out', () => {
	createFakeUserAndLogin()

	it('Logs the user out', () => {
		cy.visit('/')

		expect(localStorage.getItem('token')).to.not.eq(null)

		logout()

		cy.url()
			.should('contain', '/login')
			.then(() => {
				expect(localStorage.getItem('token')).to.eq(null)
			})
	})
	
	it.skip('Should clear the project history after logging the user out', () => {
		const projects = createProjects()
		cy.visit(`/projects/${projects[0].id}`)
			.then(() => {
				expect(localStorage.getItem('projectHistory')).to.not.eq(null)
			})

		logout()

		cy.wait(1000) // This makes re-loading of the project and associated entities (and the resulting error) visible
		
		cy.url()
			.should('contain', '/login')
			.then(() => {
				expect(localStorage.getItem('projectHistory')).to.eq(null)
			})
	})
})
