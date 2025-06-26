import dayjs from 'dayjs'

import {createFakeUserAndLogin} from '../../support/authenticateUser'

import {TaskFactory} from '../../factories/task'
import {prepareProjects} from './prepareProjects'

describe('Project View Gantt', () => {
	createFakeUserAndLogin()
	prepareProjects()

	it('Hides tasks with no dates', () => {
		const tasks = TaskFactory.create(1)
		cy.visit('/projects/1/2')

		cy.get('.g-gantt-rows-container')
			.should('not.contain', tasks[0].title)
	})

	it('Shows tasks from the current and next month', () => {
		const now = Date.UTC(2022, 8, 25)
		cy.clock(now, ['Date'])

		const nextMonth = new Date(now)
		nextMonth.setDate(1)
		nextMonth.setMonth(9)

		cy.visit('/projects/1/2')

		cy.get('.g-timeunits-container')
			.should('contain', dayjs(now).format('MMMM'))
			.should('contain', dayjs(nextMonth).format('MMMM'))
	})

	it('Shows tasks with dates', () => {
		const now = new Date()
		const tasks = TaskFactory.create(1, {
			start_date: now.toISOString(),
			end_date: new Date(new Date(now).setDate(now.getDate() + 4)).toISOString(),
		})
		cy.visit('/projects/1/2')

		cy.get('.g-gantt-rows-container')
			.should('not.be.empty')
			.should('contain', tasks[0].title)
	})

	it('Shows tasks with no dates after enabling them', () => {
		const tasks = TaskFactory.create(1, {
			start_date: null,
			end_date: null,
		})
		cy.visit('/projects/1/2')

		cy.get('.gantt-options .fancy-checkbox')
			.contains('Show tasks without dates')
			.click()

		cy.get('.g-gantt-rows-container')
			.should('not.be.empty')
			.should('contain', tasks[0].title)
	})

	it('Drags a task around', () => {
		cy.intercept(Cypress.env('API_URL') + '/tasks/*').as('taskUpdate')

		const now = new Date()
		TaskFactory.create(1, {
			start_date: now.toISOString(),
			end_date: new Date(new Date(now).setDate(now.getDate() + 4)).toISOString(),
		})
		cy.visit('/projects/1/2')

		cy.get('.g-gantt-rows-container .g-gantt-row .g-gantt-row-bars-container div .g-gantt-bar')
			.first()
			.trigger('mousedown', {which: 1})
			.trigger('mousemove', {clientX: 500, clientY: 0})
			.trigger('mouseup', {force: true})
		cy.wait('@taskUpdate')
	})

	it('Should change the query parameters when selecting a date range', () => {
		const now = Date.UTC(2022, 10, 9)
		cy.clock(now, ['Date'])

		cy.visit('/projects/1/2')

		cy.get('.project-gantt .gantt-options .field .control input.input.form-control')
			.click()
		cy.get('.flatpickr-calendar .flatpickr-innerContainer .dayContainer .flatpickr-day')
			.first()
			.click()
		cy.get('.flatpickr-calendar .flatpickr-innerContainer .dayContainer .flatpickr-day')
			.last()
			.click()

		cy.url().should('contain', 'dateFrom=2022-09-25')
		cy.url().should('contain', 'dateTo=2022-11-05')
	})

	it('Should change the date range based on date query parameters', () => {
		cy.visit('/projects/1/2?dateFrom=2022-09-25&dateTo=2022-11-05')

		cy.get('.g-timeunits-container')
			.should('contain', 'September 2022')
			.should('contain', 'October 2022')
			.should('contain', 'November 2022')
		cy.get('.project-gantt .gantt-options .field .control input.input.form-control')
			.should('have.value', '25 Sep 2022 to 5 Nov 2022')
	})

	it('Should open a task when double clicked on it', () => {
		const now = new Date()
		const tasks = TaskFactory.create(1, {
			start_date: dayjs(now).format(),
			end_date: dayjs(now.setDate(now.getDate() + 4)).format(),
		})
		cy.visit('/projects/1/2')

		cy.get('.gantt-container .g-gantt-chart .g-gantt-row-bars-container .g-gantt-bar')
			.dblclick()

		cy.url()
			.should('contain', `/tasks/${tasks[0].id}`)
	})
})
