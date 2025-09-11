import {createFakeUserAndLogin} from '../../support/authenticateUser'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

describe('Filter Persistence Across Views', () => {
	createFakeUserAndLogin()
	
	// Helper function to open filters popup
	const openFilters = () => {
		cy.get('button').contains('Filters').click()
		cy.get('.filter-popup').should('be.visible')
	}
	
	beforeEach(() => {
		// Create test project and tasks
		ProjectFactory.create(1, {
			id: 1,
			title: 'Test Project'
		})
		TaskFactory.create(5, {
			id: '{increment}',
			project_id: 1,
			title: 'Test Task {increment}'
		})
		cy.visit('/projects/1/list')
	})

	it('should persist filters in List view after page refresh', () => {
		// Apply filter in List view
		openFilters()
		cy.get('input[placeholder*="Filter"]').type('Test Task 1')
		cy.get('button').contains('Show results').click()
		
		// Verify query parameters appear
		cy.url().should('include', 'filter=')
		
		// Refresh page
		cy.reload()
		
		// Verify filter persists in URL
		cy.url().should('include', 'filter=')
	})

	it('should persist filters in Table view after page refresh', () => {
		// Switch to Table view
		cy.visit('/projects/1/table')
		
		// Apply filter in Table view
		openFilters()
		cy.get('input[placeholder*="Filter"]').type('Test Task 2')
		cy.get('button').contains('Show results').click()
		
		// Verify query parameters appear
		cy.url().should('include', 'filter=')
		
		// Refresh page
		cy.reload()
		
		// Verify filter persists in URL
		cy.url().should('include', 'filter=')
	})

	it('should persist filters in Kanban view after page refresh', () => {
		// Switch to Kanban view
		cy.visit('/projects/1/kanban')
		
		// Apply filter in Kanban view
		openFilters()
		cy.get('input[placeholder*="Filter"]').type('Test Task 3')
		cy.get('button').contains('Show results').click()
		
		// Verify query parameters appear
		cy.url().should('include', 'filter=')
		
		// Refresh page
		cy.reload()
		
		// Verify filter persists in URL
		cy.url().should('include', 'filter=')
	})

	it('should handle URL sharing with filters', () => {
		// Visit URL with pre-existing filter parameters
		cy.visit('/projects/1/kanban?filter=Test%20Task&s=Test')
		
		// Verify URL parameters are preserved
		cy.url().should('include', 'filter=Test%20Task')
		cy.url().should('include', 's=Test')
		
		// Switch views and verify parameters persist
		cy.visit('/projects/1/table?filter=Test%20Task&s=Test')
		cy.url().should('include', 'filter=Test%20Task')
		cy.url().should('include', 's=Test')
	})
})