import {test, expect} from '../../support/fixtures'
import {ProjectFactory} from '../../factories/project'
import {ProjectViewFactory} from '../../factories/project_view'
import {BucketFactory} from '../../factories/bucket'
import {TaskFactory} from '../../factories/task'
import {TaskBucketFactory} from '../../factories/task_buckets'

// Regression test for #2940: in the Kanban task popup the description editor is
// rendered inside a native <dialog> opened via showModal() (browser top-layer).
// The link prompt used to be appended to document.body, so it was painted behind
// the dialog and unfocusable through its focus trap, making "set link" a no-op.
test.describe('Editor link prompt inside the Kanban task popup', () => {
	test('creates a link in the description when opened as the Kanban popup', async ({authenticatedPage: page}) => {
		const projects = await ProjectFactory.create(1)
		const views = await ProjectViewFactory.create(1, {
			id: 1,
			project_id: projects[0].id,
			view_kind: 3,
			bucket_configuration_mode: 1,
		})
		const buckets = await BucketFactory.create(1, {
			project_view_id: views[0].id,
		})
		const tasks = await TaskFactory.create(1, {
			project_id: projects[0].id,
			description: 'link me',
			index: 1,
		})
		await TaskBucketFactory.create(1, {
			task_id: tasks[0].id,
			bucket_id: buckets[0].id,
			project_view_id: views[0].id,
		})

		await page.goto(`/projects/${projects[0].id}/${views[0].id}`)

		const card = page.locator('.kanban .bucket .tasks .task').filter({hasText: tasks[0].title})
		await expect(card).toBeVisible()
		await card.click()

		// The task popup must be a native <dialog> in the top layer.
		const dialog = page.locator('dialog[open]')
		await expect(dialog).toBeVisible()
		await expect(dialog.locator('.task-view')).toBeVisible()

		const editButton = dialog.locator('.details.content.description .tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const description = dialog.locator('.details.content.description')
		const editor = description.locator('[contenteditable="true"]').first()
		await expect(editor).toBeVisible({timeout: 10000})
		await editor.click()
		await page.keyboard.press('ControlOrMeta+a')

		await description.locator('.editor-toolbar__button').filter({hasText: 'Link'}).click()

		const urlInput = dialog.locator('input.input[placeholder="URL"]')
		await expect(urlInput).toBeVisible()
		await urlInput.fill('https://vikunja.io')
		await urlInput.press('Enter')

		const link = editor.locator('a[href="https://vikunja.io"]')
		await expect(link).toBeVisible()
		await expect(link).toHaveText('link me')
	})
})
