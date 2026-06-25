import {test, expect} from '../../support/fixtures'
import dayjs from 'dayjs'

import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

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
})
