import {test, expect} from '../../support/fixtures'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {TaskBucketFactory} from '../../factories/task_buckets'
import {TimeEntryFactory} from '../../factories/time_entry'
import {LicenseFactory} from '../../factories/license'
import {UserFactory} from '../../factories/user'
import {UserProjectFactory} from '../../factories/users_project'
import {createDefaultViews} from '../project/prepareProjects'

function trashedTask(id: number, title: string, deletedBy = 1) {
	return TaskFactory.create(1, {
		id,
		title,
		project_id: 1,
		deleted_at: new Date().toISOString(),
		deleted_by_id: deletedBy,
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
		await page.clock.install()

		await page.locator('.task-view .action-buttons .button').filter({hasText: 'Delete'}).click()
		const toast = page.locator('.global-notification').filter({hasText: 'moved to the trash'})
		const undo = toast.getByRole('button', {name: 'Undo'})
		await expect(undo).toBeFocused()
		await page.clock.fastForward(11_000)
		await expect(undo).toBeVisible()
		await expect(undo).toBeFocused()
		await undo.click()
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
		await expect(row).toContainText('Deleted by you')
		await expect(row).toContainText('30 days left')

		await row.locator('[data-cy="restoreTask"]').click()
		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(0)
		await page.reload()
		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()

		const stored = await apiContext.get('tasks/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.ok()).toBe(true)
	})

	test('restoring with the keyboard moves focus to the next row', async ({authenticatedPage: page}) => {
		await trashedTask(1, 'First trashed')
		await trashedTask(2, 'Second trashed')
		await page.goto('/trash/tasks')
		const rows = page.locator('[data-cy="trashedTask"]')
		await expect(rows).toHaveCount(2)
		const remainingTitle = (await rows.nth(1).innerText()).includes('First trashed')
			? 'First trashed'
			: 'Second trashed'

		await rows.first().locator('[data-cy="restoreTask"]').focus()
		await page.keyboard.press('Enter')

		await expect(rows).toHaveCount(1)
		await expect(rows.first()).toContainText(remainingTitle)
		await expect(rows.first().locator('[data-cy="restoreTask"]')).toBeFocused()
	})

	test('shows who else deleted a task', async ({authenticatedPage: page}) => {
		const [other] = await UserFactory.create(1, {id: 2}, false)
		await trashedTask(1, 'Trashed by someone else', 2)
		await page.goto('/trash/tasks')

		const row = page.locator('[data-cy="trashedTask"]').filter({hasText: 'Trashed by someone else'})
		await expect(row).toContainText(`Deleted by ${other.username}`)
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

	test('deleting forever with the keyboard moves focus to the next row', async ({authenticatedPage: page}) => {
		await trashedTask(1, 'First trashed')
		await trashedTask(2, 'Second trashed')
		await page.goto('/trash/tasks')
		const rows = page.locator('[data-cy="trashedTask"]')
		await expect(rows).toHaveCount(2)

		await rows.first().locator('[data-cy="deleteTaskForever"]').focus()
		await page.keyboard.press('Enter')
		const dialog = page.getByRole('dialog', {name: 'Delete this task forever'})
		await dialog.getByRole('button', {name: 'Do it!'}).focus()
		await page.keyboard.press('Enter')

		await expect(rows).toHaveCount(1)
		await expect(dialog).toHaveCount(0)
		await expect(rows.first().locator('[data-cy="deleteTaskForever"]')).toBeFocused()
	})

	test('empties the trash', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashedTask(1, 'First trashed')
		await trashedTask(2, 'Second trashed')
		await page.goto('/trash/tasks')
		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(2)

		await page.locator('[data-cy="emptyTaskTrash"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Empty trash'})
		await expect(dialog).toContainText('2 tasks')
		await expect(dialog.locator('[data-cy="emptyTrashOthersWarning"]')).toHaveCount(0)
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()
		await page.reload()
		await expect(page.getByText('There are no tasks in the trash.')).toBeVisible()
		for (const id of [1, 2]) {
			const stored = await apiContext.get(`tasks/${id}`, {headers: {Authorization: `Bearer ${userToken}`}})
			expect((await stored.json()).code).toBe(4002)
		}
	})

	test('emptying warns about tasks other people trashed and can keep them', async ({authenticatedPage: page, apiContext, userToken}) => {
		await UserFactory.create(1, {id: 2}, false)
		await trashedTask(1, 'Trashed by me')
		await trashedTask(2, 'Trashed by someone else', 2)
		await page.goto('/trash/tasks')
		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(2)

		await page.locator('[data-cy="emptyTaskTrash"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Empty trash'})
		await expect(dialog.locator('[data-cy="emptyTrashText"]')).toContainText('2 tasks')
		await expect(dialog.locator('[data-cy="emptyTrashOthersWarning"]')).toContainText('1 of them was moved to the trash by someone else.')
		const includeOthers = dialog.locator('[data-cy="emptyTrashIncludeOthers"]')
		await expect(includeOthers.locator('input')).toBeChecked()

		await includeOthers.click()
		await expect(dialog.locator('[data-cy="emptyTrashText"]')).toContainText('1 task.')
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.locator('[data-cy="trashedTask"]')).toHaveCount(1)
		await expect(page.locator('[data-cy="trashedTask"]')).toContainText('Trashed by someone else')
		const mine = await apiContext.get('tasks/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect((await mine.json()).code).toBe(4002)
		const theirs = await apiContext.get('tasks/2', {headers: {Authorization: `Bearer ${userToken}`}})
		expect((await theirs.json()).code).toBe(4040)
	})

	test('opening a trashed task shows a notice with restore', async ({authenticatedPage: page}) => {
		await trashedTask(1, 'Linked trashed task')
		await page.goto('/tasks/1')

		const notice = page.locator('[data-cy="taskInTrashNotice"]')
		await expect(notice).toContainText('This task is in the trash.')
		await expect(page).toHaveURL(/\/tasks\/1/)
		await expect(notice.getByRole('link', {name: 'Open trash'})).toBeVisible()

		await notice.locator('[data-cy="restoreTrashed"]').click()
		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Linked trashed task')
		await expect(page.locator('.task-view-container')).toBeFocused()
		await page.reload()
		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Linked trashed task')
	})

	test('opening a task in a trashed project offers restoring the project', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			title: 'Task in trashed project',
		}, false)
		const headers = {Authorization: `Bearer ${userToken}`}
		const deleted = await apiContext.delete('projects/1', {headers})
		expect(deleted.ok()).toBe(true)
		await page.goto('/tasks/1')

		const notice = page.locator('[data-cy="projectInTrashNotice"]')
		await expect(notice).toContainText('This project is in the trash.')
		await expect(page.locator('[data-cy="taskInTrashNotice"]')).toHaveCount(0)
		await expect(page).toHaveURL(/\/tasks\/1/)
		await expect(page).toHaveTitle('Trash | Vikunja')
		await expect(page.locator('.global-notification')).not.toContainText('This project is in the trash')
		await notice.getByRole('button', {name: 'Restore', exact: true}).click()

		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Task in trashed project')
		await expect(page.locator('.task-view-container')).toBeFocused()
		await page.reload()
		await expect(page.locator('.task-view h1[contenteditable]')).toContainText('Task in trashed project')
		const stored = await apiContext.get('tasks/1', {headers})
		expect(stored.ok()).toBe(true)
	})

	test('a task modal stays open when its project was trashed and restores it', async ({
		authenticatedPage: page,
		apiContext,
		userToken,
	}) => {
		await TaskFactory.create(1, {
			id: 1,
			project_id: 1,
			title: 'Task in trashed project',
		}, false)
		await TaskBucketFactory.create(1, {
			task_id: 1,
			bucket_id: 1,
			project_view_id: 4,
		})
		await page.goto('/projects/1/4')
		const card = page.locator('.kanban .task').filter({hasText: 'Task in trashed project'})
		await expect(card).toBeVisible()
		const headers = {Authorization: `Bearer ${userToken}`}
		const deleted = await apiContext.delete('projects/1', {headers})
		expect(deleted.ok()).toBe(true)
		await card.click()

		const modal = page.locator('.task-view-container.is-modal')
		const notice = modal.locator('[data-cy="projectInTrashNotice"]')
		await expect(notice).toContainText('This project is in the trash.')
		await expect(modal.getByRole('button', {name: 'Close', exact: true})).toBeVisible()
		await notice.getByRole('button', {name: 'Restore', exact: true}).click()

		await expect(modal.locator('h1[contenteditable]')).toContainText('Task in trashed project')
		await expect(modal).toBeFocused()
		await page.getByRole('dialog', {name: 'Task detail'}).getByRole('button', {
			name: 'Close dialog',
			exact: true,
		}).click()
		await expect(page).toHaveURL(/\/projects\/1\/4/)
		const stored = await apiContext.get('projects/1', {headers})
		expect(stored.ok()).toBe(true)
	})

	test('a read-only share sees a trashed task without restore', async ({authenticatedPage: page, currentUser}) => {
		const [owner] = await UserFactory.create(1, {id: 2}, false)
		await ProjectFactory.create(1, {id: 2, title: 'Shared Project', owner_id: owner.id}, false)
		await createDefaultViews(2, 5)
		await UserProjectFactory.create(1, {project_id: 2, user_id: currentUser.id, permission: 0})
		await TaskFactory.create(1, {
			id: 1,
			title: 'Shared trashed task',
			project_id: 2,
			deleted_at: new Date().toISOString(),
			deleted_by_id: owner.id,
		}, false)
		await page.goto('/tasks/1')

		const notice = page.locator('[data-cy="taskInTrashNotice"]')
		await expect(notice).toContainText('This task is in the trash.')
		await expect(notice.locator('[data-cy="restoreTrashed"]')).toHaveCount(0)
		await expect(notice.getByRole('link', {name: 'Open trash'})).toHaveCount(0)
	})

	test.describe('time tracking', () => {
		test.beforeEach(async () => {
			await LicenseFactory.enable(['time_tracking'])
		})

		test.afterEach(async () => {
			await LicenseFactory.disable()
		})

		test('time entries of a trashed task are read-only', async ({
			authenticatedPage: page,
			apiContext,
			userToken,
		}) => {
			await trashedTask(1, 'Tracked then trashed')
			const [entry] = await TimeEntryFactory.create(1, {task_id: 1, comment: 'tracked before trash'})
			await page.goto('/time-tracking')

			const row = page.locator('[data-cy="timeEntry"]').filter({hasText: 'tracked before trash'})
			await expect(row.locator('[data-cy="timeEntryInTrash"]')).toBeVisible()
			await expect(row.locator('[data-cy="editTimeEntry"]')).toHaveCount(0)
			await expect(row.locator('[data-cy="deleteTimeEntry"]')).toHaveCount(0)

			const stored = await apiContext.get(`/api/v2/time-entries/${entry.id}`, {
				headers: {Authorization: `Bearer ${userToken}`},
			})
			expect(stored.ok()).toBe(true)
			expect((await stored.json()).in_trash).toBe(true)
		})
	})
})
