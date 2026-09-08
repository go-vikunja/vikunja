import {test, expect} from '../../support/fixtures'
import dayjs from 'dayjs'
import type {Page} from '@playwright/test'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {createDefaultViews} from '../project/prepareProjects'

const TASK_ID = 1

// Makes the page overflow the 375x740 viewport, so the scroll-lock test has something to scroll.
const LONG_DESCRIPTION = Array.from({length: 60}, (_, i) => `<p>Description line ${i + 1}</p>`).join('')

async function nextFrame(page: Page) {
	await page.evaluate(() => new Promise(requestAnimationFrame))
}

async function openDueDateSheet(page: Page) {
	const [project] = await ProjectFactory.create(1)
	await createDefaultViews(project.id)
	await TaskFactory.create(1, {
		id: TASK_ID,
		project_id: project.id,
		done: false,
		description: LONG_DESCRIPTION,
	})

	await page.goto(`/tasks/${TASK_ID}`)

	const setDueDateButton = page.locator('.task-view .action-buttons .button').filter({hasText: 'Set Due Date'})
	await expect(setDueDateButton).toBeVisible({timeout: 10000})
	expect(await page.evaluate(() => document.documentElement.scrollHeight > window.innerHeight)).toBe(true)
	await setDueDateButton.click()

	const datepickerShow = page.locator('.task-view .columns.details .column').filter({hasText: 'Due Date'}).locator('.date-input .datepicker .show')
	await expect(datepickerShow).toBeVisible()

	const panel = page.locator('.bottom-sheet__panel')
	await expect(panel).toBeVisible()

	return panel
}

// The panel covers the dialog's centre, so click the strip it leaves free at the top.
async function closeViaScrim(page: Page) {
	await page.locator('.bottom-sheet .modal-container').click({position: {x: 10, y: 5}})
	await expect(page.locator('.bottom-sheet')).toHaveCount(0)
}

test.describe('Mobile bottom sheet', () => {
	test.use({viewport: {width: 375, height: 740}})

	test('Opens the due date picker as a sheet and saves the picked day', async ({authenticatedPage: page, apiContext, userToken}) => {
		const tomorrow = dayjs().add(1, 'day').format('YYYY-MM-DD')

		const panel = await openDueDateSheet(page)
		await expect(panel.locator('.datepicker-popup .calendar-month')).toBeVisible()

		await panel.locator(`.calendar-month__day[data-date="${tomorrow}"]`).click()

		await closeViaScrim(page)
		await expect(page.locator('.global-notification')).toContainText('Success')

		const response = await apiContext.get(`tasks/${TASK_ID}`, {
			headers: {'Authorization': `Bearer ${userToken}`},
		})
		expect(response.ok()).toBe(true)
		const task = await response.json()
		expect(dayjs(task.due_date).format('YYYY-MM-DD')).toBe(tomorrow)
	})

	test('Locks page scrolling while the sheet is open', async ({authenticatedPage: page}) => {
		await openDueDateSheet(page)

		const scrollY = await page.evaluate(() => window.scrollY)

		// The strip left free at the top by the panel (see closeViaScrim).
		await page.mouse.move(10, 5)
		await page.mouse.wheel(0, 400)
		await nextFrame(page)
		expect(await page.evaluate(() => window.scrollY)).toBe(scrollY)

		await closeViaScrim(page)

		await page.mouse.wheel(0, 400)
		await expect.poll(() => page.evaluate(() => window.scrollY)).toBeGreaterThan(scrollY)
	})
})
