import {test, expect} from '../../support/fixtures'
import dayjs from 'dayjs'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'

test.describe('Project View Gantt', () => {
	test('Hides tasks with no dates', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const tasks = await TaskFactory.create(1)
		await page.goto('/projects/1/2')

		await expect(page.locator('.gantt-rows')).not.toContainText(tasks[0].title)
	})

	test('Shows tasks from the current and next month', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const now = Date.UTC(2022, 8, 25)
		await page.clock.install({time: new Date(now)})

		const nextMonth = new Date(now)
		nextMonth.setDate(1)
		nextMonth.setMonth(9)

		await page.goto('/projects/1/2')

		await expect(page.locator('.gantt-timeline-months')).toContainText(dayjs(now).format('MMMM YYYY'))
		await expect(page.locator('.gantt-timeline-months')).toContainText(dayjs(nextMonth).format('MMMM YYYY'))
	})

	test('Shows tasks with dates', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const now = new Date()
		const tasks = await TaskFactory.create(1, {
			start_date: now.toISOString(),
			end_date: new Date(new Date(now).setDate(now.getDate() + 4)).toISOString(),
		})
		await page.goto('/projects/1/2')

		await expect(page.locator('.gantt-rows')).not.toBeEmpty()
		await expect(page.locator('.gantt-rows')).toContainText(tasks[0].title)
	})

	test('Shows tasks with no dates after enabling them', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const tasks = await TaskFactory.create(1, {
			start_date: null,
			end_date: null,
		})
		await page.goto('/projects/1/2')

		await page.locator('.gantt-options .fancy-checkbox').filter({hasText: 'Show tasks without date'}).click()

		await expect(page.locator('.gantt-rows')).not.toBeEmpty()
		await expect(page.locator('.gantt-rows')).toContainText(tasks[0].title)
	})

	test('Drags a task around', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const taskUpdatePromise = page.waitForResponse(response =>
			response.url().includes('/tasks/') && response.request().method() === 'POST',
		)

		const now = new Date()
		await TaskFactory.create(1, {
			start_date: now.toISOString(),
			end_date: new Date(new Date(now).setDate(now.getDate() + 4)).toISOString(),
		})
		await page.goto('/projects/1/2')

		const bar = page.locator('.gantt-rows .gantt-row-bars .gantt-bar').first()
		const barBox = await bar.boundingBox()

		if (barBox) {
			const startX = barBox.x + barBox.width / 2
			const startY = barBox.y + barBox.height / 2

			// Trigger pointer events
			await bar.dispatchEvent('pointerdown', {clientX: startX, clientY: startY, pointerId: 1, which: 1})
			await page.waitForTimeout(100)
			await bar.dispatchEvent('pointermove', {clientX: startX + 10, clientY: startY, pointerId: 1})
			await bar.dispatchEvent('pointermove', {clientX: startX + 150, clientY: startY, pointerId: 1})
			await bar.dispatchEvent('pointerup', {clientX: startX + 150, clientY: startY, pointerId: 1})
		}

		await taskUpdatePromise
	})

	test('Should change the query parameters when selecting a date range', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const now = Date.UTC(2022, 10, 9)
		await page.clock.install({time: new Date(now)})

		await page.goto('/projects/1/2')

		await page.locator('.project-gantt .gantt-options .field .control input.input.form-control').click()
		await page.locator('.flatpickr-calendar .flatpickr-innerContainer .dayContainer .flatpickr-day').first().click()
		await page.locator('.flatpickr-calendar .flatpickr-innerContainer .dayContainer .flatpickr-day').last().click()

		await expect(page).toHaveURL(/dateFrom=2022-09-25/)
		await expect(page).toHaveURL(/dateTo=2022-11-05/)
	})

	test('Should change the date range based on date query parameters', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		await page.goto('/projects/1/2?dateFrom=2022-09-25&dateTo=2022-11-05')

		await expect(page.locator('.gantt-timeline-months')).toContainText('September 2022')
		await expect(page.locator('.gantt-timeline-months')).toContainText('October 2022')
		await expect(page.locator('.gantt-timeline-months')).toContainText('November 2022')
		await expect(page.locator('.project-gantt .gantt-options .field .control input.input.form-control')).toHaveValue('25 Sep 2022 to 5 Nov 2022')
	})

	test('Should open a task when double clicked on it', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		await ProjectViewFactory.create(1, {id: 2, project_id: 1, view_kind: 1})
		const now = new Date()
		const tasks = await TaskFactory.create(1, {
			start_date: dayjs(now).format(),
			end_date: dayjs(now.setDate(now.getDate() + 4)).format(),
		})
		await page.goto('/projects/1/2')

		await page.locator('.gantt-container .gantt-row-bars .gantt-bar').dblclick()

		await expect(page).toHaveURL(new RegExp(`/tasks/${tasks[0].id}`))
	})
})
