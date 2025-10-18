import {createFakeUserAndLogin} from '../../support/authenticateUser'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import { createProjects } from './prepareProjects'

describe('Filter Persistence Across Views', () => {
	createFakeUserAndLogin()
	
	const openAndSetFilters = () => {
		cy.get('.filter-container button')
			.contains('Filters')
			.click()
		cy.get('.filter-popup')
			.should('be.visible')
		cy.get('.filter-popup .filter-input')
			.type('done = true')
		cy.get('.filter-popup button')
			.contains('Show results')
			.click()
	}

	beforeEach(() => {
		createProjects()
		TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
			title: 'Test Task {increment}'
		})
		cy.visit('/projects/1/1')
	})

	it('should persist filters in List view after page refresh', () => {
		openAndSetFilters()

		cy.url().should('include', 'filter=')

		cy.reload()

		cy.url().should('include', 'filter=')
	})

	it('should persist filters in Table view after page refresh', () => {
		cy.visit('/projects/1/3')

		openAndSetFilters()

		cy.url().should('include', 'filter=')

		cy.reload()

		cy.url().should('include', 'filter=')
	})

	it('should persist filters in Kanban view after page refresh', () => {
		cy.visit('/projects/1/4')

		openAndSetFilters()

		cy.url().should('include', 'filter=')

		cy.reload()

		cy.url().should('include', 'filter=')
	})

	it('should handle URL sharing with filters', () => {
		// Visit URL with pre-existing filter parameters
		cy.visit('/projects/1/4?filter=done%3Dtrue&s=Test')

		// Verify URL parameters are preserved
		cy.url().should('include', 'filter=done%3Dtrue')
		cy.url().should('include', 's=Test')

		// Switch views and verify parameters persist
		cy.visit('/projects/1/3?filter=done%3Dtrue&s=Test')
		cy.url().should('include', 'filter=done%3Dtrue')
		cy.url().should('include', 's=Test')
	})

	it('should apply sort settings in List view from filter modal', () => {
		// Clear existing tasks and create new ones with specific priorities
		TaskFactory.truncate()

		const now = new Date()
		const tasks = [
			{
				id: 1,
				project_id: 1,
				title: 'Low Priority Task',
				priority: 1,
				done: false,
				created: now.toISOString(),
				updated: now.toISOString(),
				created_by_id: 1,
			},
			{
				id: 2,
				project_id: 1,
				title: 'Medium Priority Task',
				priority: 3,
				done: false,
				created: now.toISOString(),
				updated: now.toISOString(),
				created_by_id: 1,
			},
			{
				id: 3,
				project_id: 1,
				title: 'High Priority Task',
				priority: 5,
				done: false,
				created: now.toISOString(),
				updated: now.toISOString(),
				created_by_id: 1,
			},
		]
		cy.request({
			method: 'PATCH',
			url: 'http://127.0.0.1:3456/api/v1/test/tasks?truncate=true',
			headers: {
				'Authorization': Cypress.env('TEST_SECRET'),
			},
			body: tasks,
		})

		cy.visit('/projects/1/1')

		// Wait for tasks to load
		cy.get('.tasks .task').should('have.length', 3)

		// Open filter modal
		cy.get('.filter-container button')
			.contains('Filters')
			.click()

		cy.get('.filter-popup')
			.should('be.visible')

		// Change sort to priority
		cy.get('.filter-popup .select select')
			.should('be.visible')
			.select('priority')

		// Toggle to descending (high priority first)
		cy.get('.filter-popup .field')
			.contains('label', 'Sort by')
			.parent()
			.find('.has-addons .button')
			.should('be.visible')
			.click()

		cy.get('.filter-popup button')
			.contains('Show results')
			.click()

		// Verify tasks are sorted by priority descending (High Priority Task should be first)
		cy.get('.tasks .task')
			.first()
			.should('contain.text', 'High Priority Task')
	})
})