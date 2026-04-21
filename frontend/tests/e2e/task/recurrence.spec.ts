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
})
