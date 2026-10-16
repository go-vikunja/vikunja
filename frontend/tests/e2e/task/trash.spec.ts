import {test, expect} from '../../support/fixtures'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {TimeEntryFactory} from '../../factories/time_entry'
import {LicenseFactory} from '../../factories/license'
import {createDefaultViews} from '../project/prepareProjects'

function trashedTask(id: number, title: string) {
	return TaskFactory.create(1, {
		id,
		title,
		project_id: 1,
		deleted_at: new Date().toISOString(),
		deleted_by_id: 1,
	}, false)
}

test.describe('Trash', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
		const views = await createDefaultViews(1)
		await BucketFactory.create(1, {project_view_id: views[3].id})
		await TaskFactory.truncate()
	})

	test('undo after deleting a task restores it', async ({authenticatedPage: page, apiContext, userToken}) => {
		await TaskFactory.create(1, {id: 1, project_id: 1, title: 'Undo me'}, false)
		await page.goto('/tasks/1')

		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'}).click()
		const toast = page.locator('.global-notification').filter({hasText: 'moved to the trash'})
		await toast.getByRole('button', {name: 'Undo'}).click()
		await expect(page.locator('.global-notification')).toContainText('The task has been restored.')
		await expect(page.locator('.tasks .task').filter({hasText: 'Undo me'})).toBeVisible()

		const stored = await apiContext.get('tasks/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.ok()).toBe(true)
	})

	test('lists trashed tasks and restores one', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashedTask(1, 'Trashed task')
		await page.goto('/trash')

		await expect(page).toHaveURL(/\/trash\/tasks/)
		const row = page.locator('[data-cy="trashedTask"]').filter({hasText: 'Trashed task'})
		await expect(row).toContainText('Deleted by')
		await expect(row).toContainText('in 30 days')

		await row.locator('[data-cy="restoreTask"]').click()
		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(0)
		await page.reload()
		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()

		const stored = await apiContext.get('tasks/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.ok()).toBe(true)
	})

	test('deletes a trashed task forever', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashedTask(1, 'Gone for good')
		await page.goto('/trash/tasks')

		await page.locator('[data-cy="trashedTask"] [data-cy="deleteTaskForever"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Delete this task forever'})
		await expect(dialog).toContainText('Gone for good')
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(0)
		const stored = await apiContext.get('tasks/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.status()).toBe(404)
		expect((await stored.json()).code).toBe(4002)
	})

	test('empties the trash', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashedTask(1, 'First trashed')
		await trashedTask(2, 'Second trashed')
		await page.goto('/trash/tasks')
		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(2)

		await page.locator('[data-cy="emptyTaskTrash"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Empty trash'})
		await expect(dialog).toContainText('2 tasks')
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()
		await page.reload()
		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()
		for (const id of [1, 2]) {
			const stored = await apiContext.get(`tasks/${id}`, {headers: {Authorization: `Bearer ${userToken}`}})
			expect((await stored.json()).code).toBe(4002)
		}
	})

	test('opening a trashed task shows a notice with restore', async ({authenticatedPage: page}) => {
		await trashedTask(1, 'Linked trashed task')
		await page.goto('/tasks/1')

		const notice = page.locator('[data-cy="taskInTrashNotice"]')
		await expect(notice).toContainText('This task is in the trash.')
		await expect(page).toHaveURL(/\/tasks\/1/)

		await notice.locator('[data-cy="restoreTrashedTask"]').click()
		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Linked trashed task')
		await page.reload()
		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Linked trashed task')
	})

	test.describe('time tracking', () => {
		test.beforeEach(async () => {
			await LicenseFactory.enable(['time_tracking'])
		})

		test.afterEach(async () => {
			await LicenseFactory.disable()
		})

		test('time entries of a trashed task are read-only', async ({authenticatedPage: page}) => {
			await trashedTask(1, 'Tracked then trashed')
			await TimeEntryFactory.create(1, {task_id: 1, comment: 'tracked before trash'})
			await page.goto('/time-tracking')

			const row = page.locator('[data-cy="timeEntry"]').filter({hasText: 'tracked before trash'})
			await expect(row.locator('[data-cy="timeEntryInTrash"]')).toBeVisible()
			await expect(row.locator('[data-cy="editTimeEntry"]')).toHaveCount(0)
			await expect(row.locator('[data-cy="deleteTimeEntry"]')).toHaveCount(0)
		})
	})
})
