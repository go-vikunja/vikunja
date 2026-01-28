import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {ProjectFactory} from '../../factories/project'

test.describe('TipTap Editor Save', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await ProjectFactory.create(1)
		await TaskFactory.truncate()
	})

	/**
	 * Regression test for https://github.com/go-vikunja/vikunja/issues/1770
	 *
	 * When saving the description editor, a race condition between Vue's DOM
	 * reconciliation and tiptap's internal DOM manipulation during unmount
	 * caused "Cannot read properties of null (reading 'insertBefore')" error.
	 *
	 * The fix uses v-show instead of v-if for EditorToolbar and BubbleMenu
	 * to avoid unmounting them during edit mode transitions.
	 */
	test('Should not crash when saving description (issue #1770)', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: 'Initial description',
		})

		// Collect any page errors and console errors that occur
		const pageErrors: Error[] = []
		const consoleErrors: string[] = []

		page.on('pageerror', (error) => {
			pageErrors.push(error)
		})

		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text())
			}
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		// Click edit button to enter edit mode
		const editButton = page.locator('.task-view .details.content.description .tiptap button.done-edit')
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()

		// Wait for editor to be visible and editable
		const editor = page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror')
		await expect(editor).toBeVisible()

		// Make an edit
		await editor.fill('Updated description text')

		// Save the description - this triggers the mode transition that could crash
		const saveButton = page.locator('[data-cy="saveEditor"]').filter({hasText: 'Save'})
		await expect(saveButton).toBeVisible()
		await saveButton.click()

		// Wait for save confirmation
		await expect(page.locator('.task-view .details.content.description h3 span.is-small.has-text-success')).toContainText('Saved!')

		// Give time for mode transition and any async errors to surface
		await page.waitForTimeout(2000)

		// Check for errors - either the DOM crashes or the edit button should appear
		const insertBeforeErrors = [
			...pageErrors.filter(e =>
				e.message.includes('insertBefore') ||
				e.message.includes("Cannot read properties of null")
			),
			...consoleErrors.filter(msg =>
				msg.includes('insertBefore') ||
				msg.includes("Cannot read properties of null")
			),
		]

		// If there are DOM manipulation errors, fail the test
		if (insertBeforeErrors.length > 0) {
			throw new Error(`DOM manipulation errors detected (issue #1770): ${JSON.stringify(insertBeforeErrors)}`)
		}

		// If no errors, the edit button should be visible (mode transition completed)
		await expect(page.locator('.task-view .details.content.description .tiptap button.done-edit')).toBeVisible({timeout: 5000})
	})

	test('Should not crash when rapidly toggling edit mode (issue #1770)', async ({authenticatedPage: page}) => {
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: 'Test description for rapid toggle',
		})

		// Collect any page errors and console errors that occur
		const pageErrors: Error[] = []
		const consoleErrors: string[] = []

		page.on('pageerror', (error) => {
			pageErrors.push(error)
		})

		page.on('console', (msg) => {
			if (msg.type() === 'error') {
				consoleErrors.push(msg.text())
			}
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		// Perform multiple edit/save cycles to stress test the mode transitions
		for (let i = 0; i < 3; i++) {
			// Enter edit mode - click edit button or double-click the editor
			const editButton = page.locator('.task-view .details.content.description .tiptap button.done-edit')
			const isEditButtonVisible = await editButton.isVisible().catch(() => false)

			if (isEditButtonVisible) {
				await editButton.click()
			} else {
				// Already in edit mode or need to double-click to enter
				const editorArea = page.locator('.task-view .details.content.description .tiptap__editor')
				await editorArea.dblclick()
			}

			// Wait for editor to be editable
			const editor = page.locator('.task-view .details.content.description .tiptap__editor .tiptap.ProseMirror')
			await expect(editor).toBeVisible()

			// Make a small edit
			await editor.fill(`Cycle ${i + 1} description`)

			// Save
			const saveButton = page.locator('[data-cy="saveEditor"]').filter({hasText: 'Save'})
			await expect(saveButton).toBeVisible()
			await saveButton.click()

			// Wait for save confirmation
			await expect(page.locator('.task-view .details.content.description h3 span.is-small.has-text-success')).toContainText('Saved!')

			// Give time for mode transition
			await page.waitForTimeout(2000)

			// Check for errors after each cycle
			const domErrors = [
				...pageErrors.filter(e =>
					e.message.includes('insertBefore') ||
					e.message.includes("Cannot read properties of null")
				),
				...consoleErrors.filter(msg =>
					msg.includes('insertBefore') ||
					msg.includes("Cannot read properties of null")
				),
			]

			// If there are DOM manipulation errors, fail the test
			if (domErrors.length > 0) {
				throw new Error(`DOM manipulation errors detected in cycle ${i + 1} (issue #1770): ${JSON.stringify(domErrors)}`)
			}

			// Verify mode transition completed (edit button should be visible)
			await expect(page.locator('.task-view .details.content.description .tiptap button.done-edit')).toBeVisible({timeout: 5000})
		}
	})
})
