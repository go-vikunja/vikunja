import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {ProjectFactory} from '../../factories/project'
import {prepareProjects} from './prepareProjects'
import {ProjectViewFactory} from '../../factories/project_view'

describe('Project History', () => {
	createFakeUserAndLogin()
	prepareProjects()
	
	it('should show a project history on the home page', () => {
		cy.intercept(Cypress.env('API_URL') + '/projects*').as('loadProjectArray')
		cy.intercept(Cypress.env('API_URL') + '/projects/*').as('loadProject')
		
		const projects = ProjectFactory.create(7)
		ProjectViewFactory.truncate()
		projects.forEach(p => ProjectViewFactory.create(1, {
			id: p.id,
			project_id: p.id,
		}, false))

		cy.visit('/')
		cy.wait('@loadProjectArray')
		cy.get('body')
			.should('not.contain', 'Last viewed')

		cy.visit(`/projects/${projects[0].id}/${projects[0].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[1].id}/${projects[1].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[2].id}/${projects[2].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[3].id}/${projects[3].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[4].id}/${projects[4].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[5].id}/${projects[5].id}`)
		cy.wait('@loadProject')
		cy.visit(`/projects/${projects[6].id}/${projects[6].id}`)
		cy.wait('@loadProject')

		// cy.visit('/')
		// Not using cy.visit here to work around the redirect issue fixed in #1337
		cy.get('nav.menu.top-menu a')
			.contains('Overview')
			.click()
		
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
