import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'
import {
	createTasksWithPriorities,
	createTasksWithSearch,
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

		cy.visit('/projects/1/3?filter=priority%20>=%204')

		cy.url()
			.should('include', 'filter=priority')

		cy.contains('.project-table table.table', highPriorityTasks[0].title, {timeout: 10000})
			.should('exist')

		cy.get('.project-table table.table')
			.should('contain', highPriorityTasks[0].title)
		cy.get('.project-table table.table')
			.should('contain', highPriorityTasks[1].title)

		cy.get('.project-table table.table')
			.should('not.contain', lowPriorityTasks[0].title)
		cy.get('.project-table table.table')
			.should('not.contain', lowPriorityTasks[1].title)
	})

	it('Should respect search query parameter from URL', () => {
		const {searchableTask} = createTasksWithSearch()

		cy.visit('/projects/1/3?s=meeting')

		cy.url()
			.should('include', 's=meeting')

		cy.contains('.project-table table.table', searchableTask.title, {timeout: 10000})
			.should('exist')

		cy.get('.project-table table.table')
			.should('contain', searchableTask.title)

		cy.get('.project-table table.table tbody tr')
			.should('have.length', 1)
	})
})
