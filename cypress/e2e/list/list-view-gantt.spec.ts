import {formatISO, format} from 'date-fns'
import {TaskFactory} from '../../factories/task'
import {prepareLists} from './prepareLists'

import '../../support/authenticateUser'

describe('List View Gantt', () => {
	prepareLists()

	it('Hides tasks with no dates', () => {
		const tasks = TaskFactory.create(1)
		cy.visit('/lists/1/gantt')

		cy.get('.g-gantt-rows-container')
			.should('not.contain', tasks[0].title)
	})

	it('Shows tasks from the current and next month', () => {
		const now = Date.UTC(2022, 8, 25)
		cy.clock(now, ['Date'])

		const nextMonth = new Date(now)
		nextMonth.setDate(1)
		nextMonth.setMonth(9)

		cy.visit('/lists/1/gantt')

		cy.get('.g-timeunits-container')
			.should('contain', format(now, 'MMMM'))
			.should('contain', format(nextMonth, 'MMMM'))
	})

	it('Shows tasks with dates', () => {
		const now = new Date()
		const tasks = TaskFactory.create(1, {
			start_date: formatISO(now),
			end_date: formatISO(now.setDate(now.getDate() + 4)),
		})
		cy.visit('/lists/1/gantt')

		cy.get('.g-gantt-rows-container')
			.should('not.be.empty')
			.should('contain', tasks[0].title)
	})

	it('Shows tasks with no dates after enabling them', () => {
		const tasks = TaskFactory.create(1, {
			start_date: null,
			end_date: null,
		})
		cy.visit('/lists/1/gantt')

		cy.get('.gantt-options .fancycheckbox')
			.contains('Show tasks which don\'t have dates set')
			.click()

		cy.get('.g-gantt-rows-container')
			.should('not.be.empty')
			.should('contain', tasks[0].title)
	})

	it('Drags a task around', () => {
		cy.intercept('**/api/v1/tasks/*')
			.as('taskUpdate')
		
		const now = new Date()
		TaskFactory.create(1, {
			start_date: formatISO(now),
			end_date: formatISO(now.setDate(now.getDate() + 4)),
		})
		cy.visit('/lists/1/gantt')

		cy.get('.g-gantt-rows-container .g-gantt-row .g-gantt-row-bars-container div .g-gantt-bar')
			.first()
			.trigger('mousedown', {which: 1})
			.trigger('mousemove', {clientX: 500, clientY: 0})
			.trigger('mouseup', {force: true})
		cy.wait('@taskUpdate')
	})
})