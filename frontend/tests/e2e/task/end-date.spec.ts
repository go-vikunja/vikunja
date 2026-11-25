import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import type {Page} from '@playwright/test'

const startDate = '2030-11-10T09:00:00Z'
const dueDate = '2030-11-15T17:00:00Z'
const endDate = '2030-11-12T17:00:00Z'

function waitForSave(page: Page) {
	return page.waitForResponse(response =>
		new URL(response.url()).pathname.endsWith('/tasks/1') && response.request().method() === 'PATCH',
	)
}

async function selectTomorrow(page: Page, property: 'start-date' | 'due-date') {
	const row = page.locator(`.task-view .property-${property}`)
	await row.locator('.datepicker .show').click()
	await row.locator('.datepicker-popup button').filter({hasText: 'Tomorrow'}).click()
	const saved = waitForSave(page)
	await row.locator('[data-cy="closeDatepicker"]').filter({hasText: 'Confirm'}).click()
	expect((await saved).ok()).toBe(true)
}

test.describe('Task end date', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1, {id: 1})
	})

	test('clearing the end date persists after reload and later task edits', async ({
		authenticatedPage: page, apiContext, userToken,
	}) => {
		await TaskFactory.create(1, {id: 1, project_id: 1, start_date: startDate, due_date: dueDate, end_date: endDate}, false)
		const headers = {Authorization: `Bearer ${userToken}`}
		const readTask = async () => {
			const response = await apiContext.get('/api/v2/tasks/1', {headers})
			expect(response.ok()).toBe(true)
			return response.json()
		}
		const before = await readTask()
		await page.goto('/tasks/1')

		const cleared = waitForSave(page)
		await page.getByRole('button', {name: 'Remove end date', exact: true}).click()
		expect((await cleared).ok()).toBe(true)
		await page.reload()
		await expect(page.locator('.property-end-date .datepicker .show')).toContainText('Not set')
		expect(await readTask()).toMatchObject({start_date: before.start_date, due_date: before.due_date})
		expect(Date.parse((await readTask()).end_date)).toBeLessThan(0)

		const edited = waitForSave(page)
		await page.locator('.property-priority select').selectOption('Urgent')
		expect((await edited).ok()).toBe(true)
		await selectTomorrow(page, 'due-date')
		await selectTomorrow(page, 'start-date')
		await page.reload()

		await expect(page.locator('.property-end-date .datepicker .show')).toContainText('Not set')
		expect(Date.parse((await readTask()).end_date)).toBeLessThan(0)
	})

	for (const missing of ['start-date', 'due-date'] as const) {
		test(`sets an initial end date when adding the missing ${missing}`, async ({
			authenticatedPage: page, apiContext, userToken,
		}) => {
			await TaskFactory.create(1, {
				id: 1,
				project_id: 1,
				...(missing === 'start-date' ? {due_date: dueDate} : {start_date: startDate}),
			}, false)
			await page.goto('/tasks/1')
			await selectTomorrow(page, missing)
			await page.reload()

			const response = await apiContext.get('/api/v2/tasks/1', {
				headers: {Authorization: `Bearer ${userToken}`},
			})
			expect(response.ok()).toBe(true)
			const task = await response.json()
			expect(Date.parse(task.due_date)).toBeGreaterThan(0)
			expect(task.end_date).toBe(task.due_date)
		})
	}
})
