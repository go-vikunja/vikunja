import {formatISO, format} from 'date-fns'
import {TaskFactory} from '../../factories/task'

import '../../support/authenticateUser'

describe('List View Gantt', () => {
	it('Hides tasks with no dates', () => {
		const tasks = TaskFactory.create(1)
		cy.visit('/lists/1/gantt')

		cy.get('.gantt-chart .tasks')
			.should('not.contain', tasks[0].title)
	})

	it('Shows tasks from the current and next month', () => {
		const now = new Date()
		const nextMonth = now
		nextMonth.setDate(1)
		nextMonth.setMonth(now.getMonth() + 1)

		cy.visit('/lists/1/gantt')

		cy.get('.gantt-chart .months')
			.should('contain', format(now, 'MMMM'))
			.should('contain', format(nextMonth, 'MMMM'))
	})

	it('Shows tasks with dates', () => {
		const now = new Date()
		const tasks = TaskFactory.create(1, {
			start_date: formatISO(now),
			end_date: formatISO(now.setDate(now.getDate() + 4))
		})
		cy.visit('/lists/1/gantt')

		cy.get('.gantt-chart .tasks')
			.should('not.be.empty')
		cy.get('.gantt-chart .tasks')
			.should('contain', tasks[0].title)
	})

	it('Shows tasks with no dates after enabling them', () => {
		TaskFactory.create(1, {
			start_date: null,
			end_date: null,
		})
		cy.visit('/lists/1/gantt')

		cy.get('.gantt-options .fancycheckbox')
			.contains('Show tasks which don\'t have dates set')
			.click()

		cy.get('.gantt-chart .tasks')
			.should('not.be.empty')
		cy.get('.gantt-chart .tasks .task.nodate')
			.should('exist')
	})

	it('Drags a task around', () => {
		const now = new Date()
		TaskFactory.create(1, {
			start_date: formatISO(now),
			end_date: formatISO(now.setDate(now.getDate() + 4))
		})
		cy.visit('/lists/1/gantt')

		cy.get('.gantt-chart .tasks .task')
			.first()
			.trigger('mousedown', {which: 1})
			.trigger('mousemove', {clientX: 500, clientY: 0})
			.trigger('mouseup', {force: true})
	})
})