import {test, expect} from '../../support/fixtures'
import type {APIRequestContext} from '@playwright/test'

import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {TimeEntryFactory} from '../../factories/time_entry'
import {LicenseFactory} from '../../factories/license'
import {createDefaultViews, createProjects} from './prepareProjects'

async function addChildProject() {
	await ProjectFactory.create(1, {id: 2, title: 'Child project', parent_project_id: 1}, false)
	await createDefaultViews(2, 5)
}

async function trashProject(apiContext: APIRequestContext, userToken: string, id: number) {
	const response = await apiContext.delete(`projects/${id}`, {headers: {Authorization: `Bearer ${userToken}`}})
	expect(response.ok()).toBe(true)
}

test.describe('Project trash', () => {
	test.beforeEach(async () => {
		await createProjects(1)
		await TaskFactory.truncate()
	})

	test('deleting a project moves it to the trash and undo restores it', async ({authenticatedPage: page, apiContext, userToken}) => {
		await TaskFactory.create(2, {project_id: 1}, false)
		await page.goto('/projects/1/settings/delete')

		const dialog = page.getByRole('dialog', {name: 'Delete this project'})
		await expect(dialog.locator('[data-cy="projectDeleteNotice"]')).toContainText('Moves this project and 2 tasks to the trash.')
		await dialog.getByRole('button', {name: 'Do it!'}).click()

		const toast = page.locator('.global-notification').filter({hasText: 'moved to the trash'})
		await toast.getByRole('button', {name: 'Undo'}).click()
		await expect(page.locator('.global-notification')).toContainText('The project has been restored.')

		const stored = await apiContext.get('projects/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.ok()).toBe(true)
	})

	test('the delete dialog counts sub-projects and time entries', async ({authenticatedPage: page}) => {
		await LicenseFactory.enable(['time_tracking'])
		try {
			await addChildProject()
			await TimeEntryFactory.create(1, {project_id: 1})
			await page.goto('/projects/1/settings/delete')

			const dialog = page.getByRole('dialog', {name: 'Delete this project'})
			await expect(dialog.locator('[data-cy="projectDeleteNotice"]')).toContainText('Moves this project, 1 sub-project and no tasks to the trash.')
			await expect(dialog.locator('[data-cy="projectDeleteTimeEntries"]')).toContainText('1 time entries from 1 people')
		} finally {
			await LicenseFactory.disable()
		}
	})

	test('lists a trashed project with its sub-projects and restores it', async ({authenticatedPage: page, apiContext, userToken}) => {
		await addChildProject()
		await trashProject(apiContext, userToken, 1)
		await page.goto('/trash/projects')

		const row = page.locator('[data-cy="trashedProject"]').filter({hasText: 'First Project'})
		await expect(row).toContainText('+1 sub-project')
		await expect(row).toContainText('Deleted by')
		await expect(page.locator('[data-cy="trashedProject"]')).toHaveCount(1)

		await row.locator('[data-cy="restoreProject"]').click()
		await expect(page.locator('[data-cy="trashedProject"]')).toHaveCount(0)
		await expect(page.locator('.menu-container').getByText('First Project')).toBeVisible()

		for (const id of [1, 2]) {
			const stored = await apiContext.get(`projects/${id}`, {headers: {Authorization: `Bearer ${userToken}`}})
			expect(stored.ok()).toBe(true)
		}
	})

	test('deletes a trashed project forever', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashProject(apiContext, userToken, 1)
		await page.goto('/trash/projects')

		await page.locator('[data-cy="trashedProject"] [data-cy="deleteProjectForever"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Delete this project forever'})
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.getByText('There are no projects in the trash.')).toBeVisible()
		const stored = await apiContext.get('projects/1', {headers: {Authorization: `Bearer ${userToken}`}})
		expect(stored.status()).toBe(404)
		expect((await stored.json()).code).not.toBe(3018)
	})

	test('empties the project trash', async ({authenticatedPage: page, apiContext, userToken}) => {
		await ProjectFactory.create(1, {id: 3, title: 'Second Project'}, false)
		await trashProject(apiContext, userToken, 1)
		await trashProject(apiContext, userToken, 3)
		await page.goto('/trash/projects')
		await expect(page.locator('[data-cy="trashedProject"]')).toHaveCount(2)

		await page.locator('[data-cy="emptyProjectTrash"]').click()
		const dialog = page.locator('dialog[open]').filter({hasText: 'Empty trash'})
		await expect(dialog).toContainText('2 projects')
		await dialog.locator('.actions .button').filter({hasText: 'Do it!'}).click()

		await expect(page.getByText('There are no projects in the trash.')).toBeVisible()
		await page.reload()
		await expect(page.getByText('There are no projects in the trash.')).toBeVisible()
	})

	test('opening a trashed project shows a notice with restore', async ({authenticatedPage: page, apiContext, userToken}) => {
		await trashProject(apiContext, userToken, 1)
		await page.goto('/projects/1/1')

		const notice = page.locator('[data-cy="projectInTrashNotice"]')
		await expect(notice).toContainText('This project is in the trash.')
		await expect(page.locator('.global-notification')).not.toContainText('This project is in the trash')

		await notice.locator('[data-cy="restoreTrashedProject"]').click()
		await expect(notice).toHaveCount(0)
		await expect(page.locator('.project-title')).toContainText('First Project')
	})
})
