import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'
import {
	createTasksWithPriorities,
	createTasksWithSearch,
	createTasksWithPriorityAndSearch,
} from '../../support/filterTestHelpers'

describe('Project View Table', () => {
	createFakeUserAndLogin()
	prepareProjects()

	it('Should show a table with tasks', () => {
		const tasks = TaskFactory.create(1)
		cy.visit('/projects/1/3')

		cy.get('.project-table table.table')
			.should('exist')
		cy.get('.project-table table.table')
			.should('contain', tasks[0].title)
	})

	it('Should have working column switches', () => {
		TaskFactory.create(1)
		cy.visit('/projects/1/3')

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
			project_id: 1,
		})
		cy.visit('/projects/1/3')

		cy.get('.project-table table.table')
			.contains(tasks[0].title)
			.click()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`)
	})

	it('Should respect filter query parameter from URL', () => {
		const {highPriorityTasks, lowPriorityTasks} = createTasksWithPriorities()

		// First verify tasks exist without filter
		cy.visit('/projects/1/3')
		cy.get('.project-table table.table').should('contain', highPriorityTasks[0].title)
		cy.get('.project-table table.table').should('contain', lowPriorityTasks[0].title)

		// Visit with filter parameter for priority >= 4
		cy.visit('/projects/1/3?filter=priority%20%3E%3D%204')

		// URL should retain the filter parameter
		cy.url()
			.should('include', 'filter=priority%20%3E%3D%204')

		// Table should show high priority tasks
		cy.get('.project-table table.table')
			.should('contain', highPriorityTasks[0].title)
		cy.get('.project-table table.table')
			.should('contain', highPriorityTasks[1].title)

		// Table should not show low priority tasks
		cy.get('.project-table table.table')
			.should('not.contain', lowPriorityTasks[0].title)
		cy.get('.project-table table.table')
			.should('not.contain', lowPriorityTasks[1].title)
	})

	it('Should respect search query parameter from URL', () => {
		const {searchableTask} = createTasksWithSearch()

		// Visit with search parameter
		cy.visit('/projects/1/3?s=meeting')

		// URL should retain the search parameter
		cy.url()
			.should('include', 's=meeting')

		// Table should show the searchable task
		cy.get('.project-table table.table')
			.should('contain', searchableTask.title)

		// Table should not show other tasks (assuming they don't contain "meeting")
		cy.get('.project-table table.table tbody tr')
			.should('have.length', 1)
	})

	it('Should respect both filter and search query parameters from URL', () => {
		const {matchingTask, nonMatchingTask1, nonMatchingTask2} = createTasksWithPriorityAndSearch()

		// Visit with both filter and search parameters
		cy.visit('/projects/1/3?filter=priority%20%3E%3D%205&s=meeting')

		// URL should retain both parameters
		cy.url()
			.should('include', 'filter=priority%20%3E%3D%205')
			.and('include', 's=meeting')

		// Table should show only the matching task
		cy.get('.project-table table.table')
			.should('contain', matchingTask.title)
		cy.get('.project-table table.table')
			.should('not.contain', nonMatchingTask1.title)
		cy.get('.project-table table.table')
			.should('not.contain', nonMatchingTask2.title)

		// Should have exactly 1 task row
		cy.get('.project-table table.table tbody tr')
			.should('have.length', 1)
	})
})
