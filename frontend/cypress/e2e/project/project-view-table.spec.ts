import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'

describe('Project View Table', () => {
	createFakeUserAndLogin()
	prepareProjects()

	it('Should show a table with tasks', () => {
		const tasks = TaskFactory.create(1, {
			project_id: 1,
		})

		// Set up comprehensive API intercepts for all possible task loading endpoints
		cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')
		cy.visit('/projects/1/3')
		cy.wait('@loadTasks', {timeout: 30000})

		// Wait for the table to be visible
		cy.get('.project-table table.table', {timeout: 30000})
			.should('be.visible')

		// Wait for the table to contain the task
		cy.get('.project-table table.table tbody', {timeout: 30000})
			.should('be.visible')
			.should('contain', tasks[0].title)
	})

	it('Should have working column switches', () => {
		TaskFactory.create(1, {
			project_id: 1,
		})

		// Set up comprehensive API intercepts for all possible task loading endpoints
		cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')
		cy.visit('/projects/1/3')
		cy.wait('@loadTasks', {timeout: 30000})

		// Wait for the table to load
		cy.get('.project-table table.table', {timeout: 30000})
			.should('be.visible')

		// Open columns filter
		cy.get('.project-table .filter-container .button')
			.contains('Columns')
			.should('be.visible')
			.click()

		// Enable Priority column
		cy.get('.project-table .filter-container .card.columns-filter .card-content .fancy-checkbox')
			.contains('Priority')
			.should('be.visible')
			.click()

		// Disable Done column
		cy.get('.project-table .filter-container .card.columns-filter .card-content .fancy-checkbox')
			.contains('Done')
			.should('be.visible')
			.click()

		// Verify Priority column is visible
		cy.get('.project-table table.table th')
			.contains('Priority')
			.should('be.visible')
			.should('exist')

		// Verify Done column is hidden
		cy.get('.project-table table.table th')
			.contains('Done')
			.should('not.exist')
	})

	it('Should navigate to the task when the title is clicked', () => {
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
		})

		// Set up comprehensive API intercepts for all possible task loading endpoints
		cy.intercept('GET', '**/api/v1/projects/*/views/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/projects/*/tasks**').as('loadTasks')
		cy.intercept('GET', '**/api/v1/tasks/all**').as('loadTasks')
		cy.visit('/projects/1/3')
		cy.wait('@loadTasks', {timeout: 30000})

		// Wait for the table to be visible and contain tasks
		cy.get('.project-table table.table tbody', {timeout: 30000})
			.should('be.visible')
			.and('contain', tasks[0].title)

		// Click on the task title to navigate
		cy.get('.project-table table.table')
			.contains(tasks[0].title)
			.should('be.visible')
			.click()

		// Verify navigation to task detail page
		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`, {timeout: 30000})
	})
})
