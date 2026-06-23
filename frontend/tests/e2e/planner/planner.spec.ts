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
})
