import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

test.describe('Task recurrence', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1, {id: 1})
	})

	test('sets repeat-every-day via preset button', async ({authenticatedPage: page, apiContext, userToken}) => {
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: new Date(Date.now() + 86_400_000).toISOString(),
		}, false)
		await page.goto(`/tasks/${task.id}`)

		// Reveal the RepeatAfter component (hidden until the user activates it)
		await page.getByRole('button', {name: 'Set Repeating Interval'}).click()

		const save = page.waitForResponse(r =>
			new URL(r.url()).pathname.endsWith(`/tasks/${task.id}`) && r.request().method() === 'PATCH',
		)
		await page.getByRole('button', {name: 'Every Day'}).click()
		expect((await save).ok()).toBeTruthy()
		const response = await apiContext.get(`/api/v2/tasks/${task.id}`, {headers: {Authorization: `Bearer ${userToken}`}})
		expect((await response.json()).repeat_after).toBe(86400)
	})

	test('undo restores the previous occurrence of a recurring task', async ({
		authenticatedPage: page, apiContext, userToken,
	}) => {
		const originalDue = new Date(Math.floor(Date.now() / 1000) * 1000 + 86_400_000)
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: originalDue.toISOString(),
			repeat_after: 86400,
		}, false)

		await page.goto(`/tasks/${task.id}`)

		const headers = {Authorization: `Bearer ${userToken}`}
		const original = await apiContext.get(`tasks/${task.id}`, {headers})
		expect(original.ok()).toBe(true)
		const before = await original.json()

		const completed = page.waitForResponse(r =>
			new URL(r.url()).pathname.endsWith(`/tasks/${task.id}/recurrence`) && r.request().method() === 'POST',
		)
		await page.getByRole('button', {name: 'Mark task done!'}).click()
		expect((await completed).ok()).toBe(true)

		const advanced = await apiContext.get(`tasks/${task.id}`, {headers})
		expect(advanced.ok()).toBe(true)
		const afterCompletion = await advanced.json()
		expect(afterCompletion.done).toBe(false)
		expect(new Date(afterCompletion.due_date).getTime()).toBe(originalDue.getTime() + 86_400_000)

		const undone = page.waitForResponse(r =>
			new URL(r.url()).pathname.endsWith(`/tasks/${task.id}/recurrence`) && r.request().method() === 'POST',
		)
		await page.getByRole('button', {name: 'Undo', exact: true}).click()
		expect((await undone).ok()).toBe(true)
		await page.reload()
		await expect(page.getByRole('button', {name: 'Mark task done!'})).toBeVisible()

		const restored = await apiContext.get(`tasks/${task.id}`, {headers})
		expect(restored.ok()).toBe(true)
		expect(await restored.json()).toMatchObject({done: false, due_date: before.due_date, done_at: before.done_at})
	})

	test('monthly repeat mode hides the amount field', async ({authenticatedPage: page}) => {
		const [task] = await TaskFactory.create(1, {id: 1, project_id: 1}, false)
		await page.goto(`/tasks/${task.id}`)

		// Reveal the RepeatAfter component (hidden until the user activates it)
		await page.getByRole('button', {name: 'Set Repeating Interval'}).click()

		await expect(page.locator('#repeatMode')).toBeVisible()
		// Amount input is visible in the default repeat mode
		await expect(page.locator('input[placeholder*="amount" i]')).toHaveCount(1)

		await page.locator('#repeatMode').selectOption({label: 'Monthly'})

		await expect(page.locator('input[placeholder*="amount" i]')).toHaveCount(0)
	})
})
