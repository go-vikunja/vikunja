import {test, expect} from '../../support/fixtures'
import type {Page} from '@playwright/test'
import dayjs from 'dayjs'

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'
import {updateUserSettings} from '../../support/updateUserSettings'

interface Project {
	id: number
	title: string
}

// A time inside today that is comfortably away from midnight so the block is
// unambiguously a timed block (not all-day).
function todayAt(hour: number): string {
	return dayjs().hour(hour).minute(0).second(0).millisecond(0).toISOString()
}

test.describe('Planner', () => {
	let projects: Project[]

	test.beforeEach(async () => {
		projects = await ProjectFactory.create(1) as Project[]
	})

	test('renders the grid and an unscheduled task in the sidebar', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 901,
			title: 'Unscheduled planner task',
			project_id: projects[0].id,
			start_date: null,
			end_date: null,
			due_date: null,
		}, false)

		await page.goto('/planner')

		await expect(page.locator('.calendar-grid')).toBeVisible()
		await expect(page.locator('.planner-sidebar')).toContainText('Unscheduled planner task')
	})

	test('shows a scheduled task as a timed block', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 902,
			title: 'Scheduled block task',
			project_id: projects[0].id,
			start_date: todayAt(10),
			end_date: todayAt(11),
		}, false)

		await page.goto('/planner')

		await expect(page.locator('.calendar-block')).toContainText('Scheduled block task')
	})

	test('shows a due-only task in the all-day row', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 903,
			title: 'Due only task',
			project_id: projects[0].id,
			start_date: null,
			end_date: null,
			due_date: todayAt(0),
		}, false)

		await page.goto('/planner')

		await expect(page.locator('.all-day-chip')).toContainText('Due only task')
	})

	test('toggles between week and day views', async ({authenticatedPage: page}) => {
		await page.goto('/planner')

		// Week view shows 7 day headers, day view shows 1.
		await expect(page.locator('.day-head')).toHaveCount(7)
		await page.getByRole('button', {name: 'Day', exact: true}).click()
		await expect(page.locator('.day-head')).toHaveCount(1)
	})

	test('offers a sort control for the unscheduled sidebar', async ({authenticatedPage: page}) => {
		await page.goto('/planner')

		const sortSelect = page.locator('.planner-sidebar .sort-select select')
		await expect(sortSelect).toBeVisible()
		// Includes the client-side random shuffle and excludes nonsensical date sorts.
		await expect(sortSelect.locator('option', {hasText: 'Random'})).toHaveCount(1)
	})

	test('double-clicking an empty slot opens the create-task modal', async ({authenticatedPage: page}) => {
		await page.goto('/planner')

		await page.locator('.day-column').first().dblclick({position: {x: 20, y: 200}})

		const dialog = page.locator('.modal-dialog')
		await expect(dialog).toContainText('New task')
		await expect(dialog.locator('textarea')).toBeVisible()
	})

	// AddTask emits its created tasks as one batch; a mismatch between that event
	// name and the planner's listener leaves the modal open and the task off the
	// grid, which is invisible to the "modal opens" test above.
	test('creating a task from the modal closes it and schedules the task on the grid', async ({authenticatedPage: page, apiContext}) => {
		await page.goto('/planner')

		// The planner is cross-project, so AddTask has no projectId from the route
		// and falls back to the default project — without one it only errors.
		const token = await page.evaluate(() => localStorage.getItem('token'))
		await updateUserSettings(apiContext, token!, {
			// Snake case, as the settings endpoint and the auth store both spell it.
			// A camelCase key is accepted with a 2xx and then dropped, so the write
			// looks fine and AddTask still finds no default project.
			default_project_id: projects[0].id,
			// Required-and-validated by the settings endpoint; omitting it 400s the
			// whole payload and the default project silently never lands.
			overdue_tasks_reminders_time: '9:00',
		})
		await page.reload()
		// The store must have the new setting before AddTask reads it.
		await page.waitForLoadState('networkidle')

		await page.locator('.day-column').first().dblclick({position: {x: 20, y: 200}})

		const dialog = page.locator('.modal-dialog')
		await dialog.locator('textarea').fill('Painted planner task')
		await dialog.getByRole('button', {name: 'Add'}).click()

		await expect(dialog).toBeHidden()
		await expect(page.locator('.calendar-block')).toContainText('Painted planner task')
		// It must land on the grid, not fall back into the unscheduled sidebar.
		await expect(page.locator('.planner-sidebar')).not.toContainText('Painted planner task')
	})

	test('double-clicking the all-day row opens the create-task modal', async ({authenticatedPage: page}) => {
		await page.goto('/planner')

		await page.locator('.all-day-cell').first().dblclick()

		const dialog = page.locator('.modal-dialog')
		await expect(dialog).toContainText('All day')
		await expect(dialog.locator('textarea')).toBeVisible()
	})

	test('renders the configured number of days in rolling mode', async ({authenticatedPage: page}) => {
		// Seed planner settings before the app reads them (mergeDefaults fills the rest).
		await page.addInitScript(() => {
			localStorage.setItem('planner-settings', JSON.stringify({fullWeek: false, daysToShow: 10}))
		})

		await page.goto('/planner')

		await expect(page.locator('.day-head')).toHaveCount(10)
	})

	test('projects a daily recurring task into the next week', async ({authenticatedPage: page}) => {
		await TaskFactory.create(1, {
			id: 904,
			title: 'Daily standup',
			project_id: projects[0].id,
			start_date: todayAt(10),
			end_date: todayAt(11),
			repeat_after: 86400, // one day, in seconds
		}, false)

		await page.goto('/planner')
		// Next week contains no stored instance — only projected occurrences.
		await page.getByRole('button', {name: 'Next', exact: true}).click()

		await expect(page.locator('.calendar-block', {hasText: 'Daily standup'}).first()).toBeVisible()
	})

	test.describe('month view', () => {
		// Anchored to the middle of the current month so these days are always on
		// the grid, whichever day of the month the suite happens to run on.
		function thisMonthAt(dayOfMonth: number, hour: number): string {
			return dayjs().date(dayOfMonth).hour(hour).minute(0).second(0).millisecond(0).toISOString()
		}

		function cellFor(page: Page, dayOfMonth: number) {
			return page.locator(`.month-cell[data-day="${dayjs().date(dayOfMonth).format('YYYY-MM-DD')}"]`)
		}

		async function openMonth(page: Page) {
			await page.goto('/planner')
			await page.getByRole('button', {name: 'Month', exact: true}).click()
			await expect(page.locator('.month-grid')).toBeVisible()
		}

		test('renders whole weeks under seven weekday headers', async ({authenticatedPage: page}) => {
			await openMonth(page)

			await expect(page.locator('.month-weekday')).toHaveCount(7)
			// 4 to 6 whole weeks, never a partial row.
			const cells = await page.locator('.month-cell').count()
			expect(cells % 7).toBe(0)
			expect(cells).toBeGreaterThanOrEqual(28)
			expect(cells).toBeLessThanOrEqual(42)
		})

		// The month grid has no time axis, so it lists timed tasks as chips where
		// the week grid draws them as blocks in the day columns.
		test('lists a timed task as a chip on its day', async ({authenticatedPage: page}) => {
			await TaskFactory.create(1, {
				id: 910,
				title: 'Timed month task',
				project_id: projects[0].id,
				start_date: thisMonthAt(10, 10),
				end_date: thisMonthAt(10, 11),
			}, false)

			await openMonth(page)

			await expect(cellFor(page, 10).locator('.month-chip')).toContainText('Timed month task')
			await expect(page.locator('.calendar-block')).toHaveCount(0)
		})

		test('spreads a multi-day task across every day it covers', async ({authenticatedPage: page}) => {
			await TaskFactory.create(1, {
				id: 911,
				title: 'Three day push',
				project_id: projects[0].id,
				start_date: thisMonthAt(10, 9),
				end_date: thisMonthAt(12, 17),
			}, false)

			await openMonth(page)

			for (const dayOfMonth of [10, 11, 12]) {
				await expect(cellFor(page, dayOfMonth).locator('.month-chip'))
					.toContainText('Three day push')
			}
			await expect(page.locator('.month-chip', {hasText: 'Three day push'})).toHaveCount(3)
		})

		test('collapses a crowded day behind "+N more" and expands it on click', async ({authenticatedPage: page}) => {
			await TaskFactory.create(15, {
				project_id: projects[0].id,
				title: (i: number) => `Crowded task ${i}`,
				start_date: thisMonthAt(15, 9),
				end_date: thisMonthAt(15, 10),
			}, false)

			await openMonth(page)

			const cell = cellFor(page, 15)
			const more = cell.locator('.month-more')

			// However many chips the cell measures room for, it is not all fifteen.
			await expect(more).toBeVisible()
			expect(await cell.locator('.month-chip').count()).toBeLessThan(15)

			await more.click()

			await expect(cell.locator('.month-chip')).toHaveCount(15)
			await expect(more).toBeHidden()
		})

		test('hides the zoom controls, which have no time axis to act on', async ({authenticatedPage: page}) => {
			await page.goto('/planner')
			await expect(page.locator('.zoom-controls')).toBeVisible()

			await page.getByRole('button', {name: 'Month', exact: true}).click()

			await expect(page.locator('.zoom-controls')).toBeHidden()
		})

		test('double-clicking a day opens the create-task modal for that day', async ({authenticatedPage: page}) => {
			await openMonth(page)

			await cellFor(page, 10).dblclick()

			const dialog = page.locator('.modal-dialog')
			await expect(dialog).toContainText('All day')
			await expect(dialog.locator('textarea')).toBeVisible()
		})
	})
})
