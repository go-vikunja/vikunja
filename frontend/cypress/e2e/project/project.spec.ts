import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {prepareProjects} from './prepareProjects'

describe('Projects', () => {
	createFakeUserAndLogin()

	let projects
	prepareProjects((newProjects) => (projects = newProjects))

	it.skip('Should show all projects on the projects page', () => {
		const projects = ProjectFactory.create(10)

		cy.visit('/projects')

		projects.forEach(p => {
			cy.get('[data-cy="projects-list"]')
				.should('contain', p.title)
		})
	})

})
