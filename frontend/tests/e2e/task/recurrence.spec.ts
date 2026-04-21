import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'

test.describe('Task recurrence', () => {
	test.beforeEach(async ({authenticatedPage}) => {
		await ProjectFactory.create(1, {id: 1})
	})

	test('sets repeat-every-day via preset button', async ({authenticatedPage: page}) => {
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: new Date(Date.now() + 86_400_000).toISOString(),
		}, false)
		await page.goto(`/tasks/${task.id}`)

		// Reveal the RepeatAfter component (hidden until the user activates it)
		await page.getByRole('button', {name: 'Set Repeating Interval'}).click()

		const save = page.waitForResponse(r =>
			r.url().includes(`/tasks/${task.id}`) && r.request().method() === 'POST',
		)
		await page.getByRole('button', {name: 'Every Day'}).click()
		const r = await save
		const body = r.request().postDataJSON()
		expect(body.repeat_after).toBe(86400)
	})

	test('completing a recurring task reopens with advanced due date', async ({
		authenticatedPage: page, apiContext, userToken,
	}) => {
		const originalDue = new Date(Date.now() + 86_400_000)
		const [task] = await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			due_date: originalDue.toISOString(),
			repeat_after: 86400,
		}, false)

		await page.goto(`/tasks/${task.id}`)

		const completed = page.waitForResponse(r =>
			r.url().includes(`/tasks/${task.id}`) && r.request().method() === 'POST',
		)
		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Mark task done!'}).click()
		await completed

		// Fetch fresh state from the API to verify the backend regenerated the task.
		const resp = await apiContext.get(`tasks/${task.id}`, {
			headers: {Authorization: `Bearer ${userToken}`},
		})
		expect(resp.ok()).toBe(true)
		const refreshed = await resp.json()
		expect(refreshed.done).toBe(false)
		const newDue = new Date(refreshed.due_date).getTime()
		// addRepeatIntervalToTime: when the original due date is still in the
		// future, the backend advances it by exactly one interval (86400s here).
		expect(newDue - originalDue.getTime()).toBeCloseTo(86_400_000, -3)
	})
})
