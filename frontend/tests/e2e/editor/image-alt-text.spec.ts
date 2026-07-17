import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

// The image extension is configured with allowBase64:false, so it drops data:
// URIs on parse; use a static asset the frontend actually serves instead. It is
// same-origin, so it loads without any network dependency.
const IMAGE_SRC = '/images/icons/apple-touch-icon-76x76.png'

test.describe('Editor image alt text', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	test('sets alt text on a selected image via the image bubble menu', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: `<p>before</p><img src="${IMAGE_SRC}"><p>after</p>`,
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const description = page.locator('.task-view .details.content.description')

		const editButton = description.locator('.tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const editor = description.locator('.tiptap__editor .tiptap.ProseMirror')
		await expect(editor).toBeVisible()

		// Selecting the image node is what makes the image bubble menu appear.
		const image = editor.locator('img')
		await expect(image).toBeVisible()
		await image.click()

		await description.locator('.editor-bubble__button', {hasText: 'Alt text'}).click()

		const altInput = page.locator('input.input[placeholder="Describe this image"]')
		await expect(altInput).toBeVisible()
		await altInput.fill('A helpful description')
		await altInput.press('Enter')

		await expect(editor.locator('img[alt="A helpful description"]')).toBeVisible()

		const saveButton = page.locator('[data-cy="saveEditor"]').filter({hasText: 'Save'})
		await expect(saveButton).toBeVisible()
		await saveButton.click()
		await expect(description.locator('h3 span.is-small.has-text-success')).toContainText('Saved!')

		// The alt text must survive a round-trip through the backend.
		await page.reload()
		await page.waitForLoadState('networkidle')
		await expect(description.locator('img[alt="A helpful description"]')).toBeVisible({timeout: 10000})
	})
})
