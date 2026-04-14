import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {TaskFactory} from '../../factories/task'
import {BucketFactory} from '../../factories/bucket'
import {createDefaultViews} from '../project/prepareProjects'

test.describe('Emoji Autocomplete', () => {
	let projectId: number
	test.beforeEach(async () => {
		const projects = await ProjectFactory.create(1) as Array<{id: number}>
		projectId = projects[0].id
		const views = await createDefaultViews(projectId)
		await BucketFactory.create(1, {
			project_view_id: views[3].id,
		})
	})

	test('Inserts the unicode glyph when selecting from the popup', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			project_id: projectId,
			description: 'x',
			index: 1,
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const editButton = page.locator('.task-view .details.content.description .tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const editor = page.locator('.task-view .details.content.description [contenteditable="true"]').first()
		await expect(editor).toBeVisible({timeout: 10000})
		await editor.click()
		await page.keyboard.press('ControlOrMeta+a')
		await page.keyboard.press('Delete')

		await editor.pressSequentially(':smi', {delay: 50})

		const popup = page.locator('.emoji-items')
		await expect(popup).toBeVisible()

		await page.keyboard.press('Enter')

		await expect(popup).toBeHidden()
		await expect(editor).not.toContainText(':smi')
		const text = await editor.innerText()
		let hasNonAscii = false
		for (let i = 0; i < text.length; i++) {
			if (text.charCodeAt(i) > 127) {
				hasNonAscii = true
				break
			}
		}
		expect(hasNonAscii).toBe(true)
	})

	test('Does not open the popup when ":" follows a non-whitespace char', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 2,
			project_id: projectId,
			description: 'x',
			index: 1,
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const editButton = page.locator('.task-view .details.content.description .tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const editor = page.locator('.task-view .details.content.description [contenteditable="true"]').first()
		await expect(editor).toBeVisible({timeout: 10000})
		await editor.click()
		await page.keyboard.press('ControlOrMeta+a')
		await page.keyboard.press('Delete')

		await editor.pressSequentially('word:foo', {delay: 50})

		const popup = page.locator('.emoji-items')
		await expect(popup).toBeHidden()
	})
})
