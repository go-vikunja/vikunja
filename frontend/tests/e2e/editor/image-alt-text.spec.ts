import {test, expect} from '../../support/fixtures'
import type {Page} from '@playwright/test'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

// The image extension is configured with allowBase64:false, so it strips data:
// URIs on parse — the <img> src has to be a real same-origin URL the browser can
// fetch. The test owns that image: a Playwright route serves a tiny PNG for this
// path, so nothing outside the test can move or rename it.
const IMAGE_SRC = '/e2e-test-image.png'

// Smallest valid 1x1 transparent PNG.
const TEST_PNG = Buffer.from(
	'iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAAC0lEQVR42mNk+M9QDwADhgGAWjR9awAAAABJRU5ErkJggg==',
	'base64',
)

async function serveTestImage(page: Page) {
	await page.route(`**${IMAGE_SRC}`, route => route.fulfill({
		contentType: 'image/png',
		body: TEST_PNG,
	}))
}

test.describe('Editor image alt text', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	test('sets alt text on a selected image via the image bubble menu', async ({authenticatedPage: page}) => {
		await serveTestImage(page)

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
		await expect(description.locator('h2 span.is-small.has-text-success')).toContainText('Saved!')

		// The alt text must survive a round-trip through the backend.
		await page.reload()
		await page.waitForLoadState('networkidle')
		await expect(description.locator('img[alt="A helpful description"]')).toBeVisible({timeout: 10000})
	})

	test('does not wipe existing alt text when the prompt is cancelled', async ({authenticatedPage: page}) => {
		await serveTestImage(page)

		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: `<p>before</p><img src="${IMAGE_SRC}" alt="existing alt text"><p>after</p>`,
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const description = page.locator('.task-view .details.content.description')

		const editButton = description.locator('.tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const editor = description.locator('.tiptap__editor .tiptap.ProseMirror')
		await expect(editor).toBeVisible()

		const image = editor.locator('img')
		await expect(image).toBeVisible()
		await image.click()

		await description.locator('.editor-bubble__button', {hasText: 'Alt text'}).click()

		const altInput = page.locator('input.input[placeholder="Describe this image"]')
		await expect(altInput).toBeVisible()
		await expect(altInput).toHaveValue('existing alt text')
		await altInput.press('Escape')

		await expect(editor.locator('img[alt="existing alt text"]')).toBeVisible()
	})
})
