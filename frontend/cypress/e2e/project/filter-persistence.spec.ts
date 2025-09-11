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
		createProjects(1)
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
})