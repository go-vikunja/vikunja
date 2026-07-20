import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

test.describe('Editor toolbar keyboard navigation', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	test('roving tabindex: arrow keys move between buttons, Tab leaves the toolbar', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: '<p>some text</p>',
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const description = page.locator('.task-view .details.content.description')

		// Entering edit mode is what renders (v-show) the toolbar.
		const editButton = description.locator('.tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		const editor = description.locator('.tiptap__editor .tiptap.ProseMirror')
		await expect(editor).toBeVisible()

		const toolbar = description.locator('.editor-toolbar')
		await expect(toolbar).toHaveAttribute('role', 'toolbar')

		// Roving tabindex: the whole toolbar is a single Tab stop.
		const tabbableButtons = toolbar.locator('button[tabindex="0"]')
		await expect(tabbableButtons).toHaveCount(1)

		const buttons = toolbar.locator('button.editor-toolbar__button')
		// nth(0/1) are the always-enabled heading buttons — roving navigation skips disabled ones.
		const firstButton = buttons.nth(0)
		const secondButton = buttons.nth(1)

		await firstButton.focus()
		await expect(firstButton).toBeFocused()
		await expect(firstButton).toHaveAttribute('tabindex', '0')

		await page.keyboard.press('ArrowRight')
		await expect(secondButton).toBeFocused()
		await expect(secondButton).toHaveAttribute('tabindex', '0')
		// The tab stop moved with focus — still exactly one.
		await expect(tabbableButtons).toHaveCount(1)

		// Tab must exit the toolbar into the editor content, not step to the next button.
		await page.keyboard.press('Tab')
		await expect(secondButton).not.toBeFocused()
		const focusStillInToolbar = await toolbar.evaluate(el => el.contains(document.activeElement))
		expect(focusStillInToolbar).toBe(false)
	})
})
