import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'

describe('Project History', () => {
	createFakeUserAndLogin()

	it('should show a project history on the home page', () => {
		const projects = ProjectFactory.create(7)
		ProjectViewFactory.truncate()
		projects.forEach(p => ProjectViewFactory.create(1, {
			id: p.id,
			project_id: p.id,
		}, false))

		cy.visit('/')
		cy.get('body')
			.should('not.contain', 'Last viewed')

		// Visit each project to create history
		cy.visit(`/projects/${projects[0].id}/${projects[0].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[1].id}/${projects[1].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[2].id}/${projects[2].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[3].id}/${projects[3].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[4].id}/${projects[4].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[5].id}/${projects[5].id}`)
		cy.get('.project-title').should('be.visible')

		cy.visit(`/projects/${projects[6].id}/${projects[6].id}`)
		cy.get('.project-title').should('be.visible')

		// Navigate back to home
		cy.visit('/')

		cy.get('body')
			.should('contain', 'Last viewed')
		cy.get('[data-cy="projectCardGrid"]')
			.should('not.contain', projects[0].title)
			.should('contain', projects[1].title)
			.should('contain', projects[2].title)
			.should('contain', projects[3].title)
			.should('contain', projects[4].title)
			.should('contain', projects[5].title)
			.should('contain', projects[6].title)
	})
})
