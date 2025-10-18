import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'

describe('Project View Table', () => {
	createFakeUserAndLogin()

	let projects: any[]
	prepareProjects((newProjects) => (projects = newProjects))

	it('Should show a table with tasks', () => {
		// Create multiple tasks like test 3 does
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			index: '{increment}',
			project_id: projects[0].id,
		})
		// View 2 is the table view (index 2 in the views array)
		const tableViewId = projects[0].views[2].id
		cy.visit(`/projects/${projects[0].id}/${tableViewId}`)

		// Wait for the project table to load
		cy.get('.project-table', {timeout: 10000})
			.should('be.visible')

		// Check if table exists and contains our task
		cy.get('.project-table table.table', {timeout: 10000})
			.should('exist')
			.should('contain', tasks[0].title)
	})

	it('Should have working column switches', () => {
		TaskFactory.create(1, {
			project_id: projects[0].id,
		})
		const tableViewId = projects[0].views[2].id
		cy.visit(`/projects/${projects[0].id}/${tableViewId}`)

		cy.get('.project-table .filter-container .button')
			.contains('Columns')
			.click()
		cy.get('.project-table .filter-container .card.columns-filter .card-content .fancy-checkbox')
			.contains('Priority')
			.click()
		cy.get('.project-table .filter-container .card.columns-filter .card-content .fancy-checkbox')
			.contains('Done')
			.click()

		cy.get('.project-table table.table th')
			.contains('Priority')
			.should('exist')
		cy.get('.project-table table.table th')
			.contains('Done')
			.should('not.exist')
	})

	it('Should navigate to the task when the title is clicked', () => {
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			project_id: projects[0].id,
		})
		const tableViewId = projects[0].views[2].id
		cy.visit(`/projects/${projects[0].id}/${tableViewId}`)

		cy.get('.project-table table.table')
			.contains(tasks[0].title)
			.click()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`)
	})
})
