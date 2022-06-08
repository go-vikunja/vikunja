import {TaskFactory} from '../../factories/task'

import '../../support/authenticateUser'

describe('List View Table', () => {
	it('Should show a table with tasks', () => {
		const tasks = TaskFactory.create(1)
		cy.visit('/lists/1/table')

		cy.get('.list-table table.table')
			.should('exist')
		cy.get('.list-table table.table')
			.should('contain', tasks[0].title)
	})

	it('Should have working column switches', () => {
		TaskFactory.create(1)
		cy.visit('/lists/1/table')

		cy.get('.list-table .filter-container .items .button')
			.contains('Columns')
			.click()
		cy.get('.list-table .filter-container .card.columns-filter .card-content .fancycheckbox .check')
			.contains('Priority')
			.click()
		cy.get('.list-table .filter-container .card.columns-filter .card-content .fancycheckbox .check')
			.contains('Done')
			.click()

		cy.get('.list-table table.table th')
			.contains('Priority')
			.should('exist')
		cy.get('.list-table table.table th')
			.contains('Done')
			.should('not.exist')
	})

	it('Should navigate to the task when the title is clicked', () => {
		const tasks = TaskFactory.create(5, {
			id: '{increment}',
			list_id: 1,
		})
		cy.visit('/lists/1/table')

		cy.get('.list-table table.table')
			.contains(tasks[0].title)
			.click()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`)
	})
})