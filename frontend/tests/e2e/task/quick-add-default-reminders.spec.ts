import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {UserFactory} from '../../factories/user'
import {createDefaultViews} from '../project/prepareProjects'
import {login} from '../../support/authenticateUser'
import {REMINDER_PERIOD_RELATIVE_TO_TYPES} from '../../../src/types/IReminderPeriodRelativeTo'
import {SECONDS_A_HOUR} from '../../../src/constants/date'

test.describe('Quick add default reminders', () => {
	test('Auto-attaches default reminder when quick add task has a due date', async ({page, apiContext}) => {
		const user = (await UserFactory.create(1, {
			frontend_settings: JSON.stringify({
				quickAddDefaultReminders: [
					{
						reminder: null,
						relativePeriod: -2 * SECONDS_A_HOUR,
						relativeTo: REMINDER_PERIOD_RELATIVE_TO_TYPES.DUEDATE,
					},
				],
			}),
		}))[0]
		const project = (await ProjectFactory.create(1, {owner_id: user.id}))[0]
		await createDefaultViews(project.id)

		await login(page, apiContext, user)
		await page.goto(`/projects/${project.id}/1`)

		await page.locator('.input[placeholder="Add a task…"]').fill('Buy milk tomorrow')

		const createTaskPromise = page.waitForResponse(response =>
			response.url().includes('/projects/') &&
			response.url().includes('/tasks') &&
			response.request().method() === 'PUT',
		)
		await page.locator('.button').filter({hasText: 'Add'}).click()
		await createTaskPromise

		const taskLink = page.locator('.tasks .task').filter({hasText: 'Buy milk'}).first().locator('a.task-link')
		await expect(taskLink).toBeVisible({timeout: 10000})
		await taskLink.click()

		// Reminders section auto-expands when the task already has reminders.
		const reminderInput = page.locator('.task-view .columns.details .reminder-input').first()
		await expect(reminderInput).toBeVisible({timeout: 10000})
		await expect(reminderInput).toContainText('2 hours before Due Date')
	})
})
