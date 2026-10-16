import type {Page} from '@playwright/test'
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

test.describe('Editor toolbar overflow menu', () => {
	test.beforeEach(async () => {
		await ProjectFactory.create(1)
	})

	async function openNarrowEditor(page: Page) {
		await page.setViewportSize({width: 375, height: 800})
		const tasks = await TaskFactory.create(1, {
			id: 1,
			description: '<p>some text</p>',
		})

		await page.goto(`/tasks/${tasks[0].id}`)
		await page.waitForLoadState('networkidle')

		const description = page.locator('.task-view .details.content.description')
		const editButton = description.locator('.tiptap button.done-edit').filter({hasText: 'Edit'})
		await expect(editButton).toBeVisible({timeout: 10000})
		await editButton.click()
		await expect(description.locator('.tiptap__editor .tiptap.ProseMirror')).toBeVisible()

		const toolbar = description.locator('.editor-toolbar')
		const moreButton = toolbar.getByRole('button', {name: 'More formatting options'})
		await expect(moreButton).toBeVisible()
		return {toolbar, moreButton}
	}

	test('arrow keys skip buttons moved into the menu', async ({authenticatedPage: page}) => {
		const {toolbar, moreButton} = await openNarrowEditor(page)

		const overflowing = toolbar.locator('.editor-toolbar__row .is-overflowing')
		await expect(overflowing.first()).toBeAttached()
		await expect(overflowing.first()).toBeHidden()

		const visibleCount = await toolbar.locator('.editor-toolbar__row button:not(.is-overflowing)').count()
		const firstButton = toolbar.locator('button.editor-toolbar__button').first()
		await firstButton.focus()

		// Walk the whole toolbar: every stop must be a visible button, ending on "More".
		for (let i = 0; i < visibleCount; i++) {
			await page.keyboard.press('ArrowRight')
			const focusedIsOverflowing = await page.evaluate(() => document.activeElement?.classList.contains('is-overflowing'))
			expect(focusedIsOverflowing).toBe(false)
		}
		await expect(moreButton).toBeFocused()

		await page.keyboard.press('ArrowRight')
		await expect(firstButton).toBeFocused()

		await page.keyboard.press('ArrowLeft')
		await expect(moreButton).toBeFocused()

		await page.keyboard.press('Home')
		await expect(firstButton).toBeFocused()
		await page.keyboard.press('End')
		await expect(moreButton).toBeFocused()

		// Still a single Tab stop, and none of the hidden buttons can hold it.
		await expect(toolbar.locator('button[tabindex="0"]')).toHaveCount(1)
		await expect(moreButton).toHaveAttribute('tabindex', '0')
	})

	test('the menu opens from the keyboard and keeps its items out of the toolbar tab order', async ({authenticatedPage: page}) => {
		const {toolbar, moreButton} = await openNarrowEditor(page)

		await moreButton.focus()
		await page.keyboard.press('Enter')
		await expect(moreButton).toHaveAttribute('aria-expanded', 'true')

		const menuItems = toolbar.locator('.dropdown-menu .dropdown-item')
		await expect(menuItems.first()).toBeFocused()

		await page.keyboard.press('ArrowDown')
		await expect(menuItems.nth(1)).toBeFocused()
		await page.keyboard.press('ArrowUp')
		await expect(menuItems.first()).toBeFocused()
		await page.keyboard.press('ArrowUp')
		await expect(menuItems.last()).toBeFocused()

		// Left/right belong to the toolbar, not the open menu.
		await page.keyboard.press('ArrowRight')
		await expect(menuItems.last()).toBeFocused()

		await page.keyboard.press('Escape')
		await expect(moreButton).toHaveAttribute('aria-expanded', 'false')
		await expect(moreButton).toBeFocused()

		await expect(toolbar.locator('button[tabindex="0"]')).toHaveCount(1)
		await page.keyboard.press('ArrowRight')
		await expect(toolbar.locator('button.editor-toolbar__button').first()).toBeFocused()
	})

	test('formats applied from the menu show as active', async ({authenticatedPage: page}) => {
		const {toolbar, moreButton} = await openNarrowEditor(page)

		await moreButton.click()
		const bulletList = toolbar.locator('.dropdown-menu .dropdown-item').filter({hasText: 'Bullet list'})
		await expect(bulletList).toHaveAttribute('aria-pressed', 'false')
		await bulletList.click()

		await expect(toolbar.locator('.dropdown-menu')).toBeHidden()
		await moreButton.click()
		await expect(bulletList).toHaveAttribute('aria-pressed', 'true')
		await expect(bulletList.locator('.editor-toolbar__menu-check')).toBeVisible()
	})
})
