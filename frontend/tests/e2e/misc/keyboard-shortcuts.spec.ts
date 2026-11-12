import type {Page} from '@playwright/test'
import {test, expect} from '../../support/fixtures'
import {TaskFactory} from '../../factories/task'
import {createProjects} from '../project/prepareProjects'
import {apiRootUrl} from '../../support/apiUrl'

async function pressLayoutKey(page: Page, init: KeyboardEventInit) {
	// Playwright's keyboard uses a US layout; supply the actual layout's key/code pair.
	await page.locator('body').dispatchEvent('keydown', init)
}

test.describe('Keyboard layouts', () => {
	test('opens help with a Danish question mark, but not underscore', async ({authenticatedPage: page}) => {
		await page.goto('/')
		await expect(page.locator('.menu-container')).toBeVisible()
		const help = page.locator('.keyboard-shortcuts')
		await pressLayoutKey(page, {
			key: '_',
			code: 'Slash',
			shiftKey: true,
		})
		await expect(help).not.toBeVisible()
		await pressLayoutKey(page, {
			key: '?',
			code: 'Minus',
			shiftKey: true,
		})
		await expect(help).toBeVisible()
		await page.keyboard.press('Escape')
		await expect(help).not.toBeVisible()
	})

	for (const [layout, key, code] of [['Dvorak', 'k', 'KeyV'], ['Russian', 'л', 'KeyK']]) {
		test(`opens quick search on ${layout}`, async ({authenticatedPage: page}) => {
			await page.goto('/')
			await expect(page.locator('.menu-container')).toBeVisible()
			await pressLayoutKey(page, {
				key,
				code,
				ctrlKey: true,
			})
			await expect(page.locator('.quick-actions input')).toBeFocused()
		})
	}

	test('navigates with a Dvorak sequence', async ({authenticatedPage: page}) => {
		await page.goto('/projects')
		await expect(page.locator('.menu-container')).toBeVisible()
		await pressLayoutKey(page, {
			key: 'g',
			code: 'KeyU',
		})
		await pressLayoutKey(page, {
			key: 'o',
			code: 'KeyS',
		})
		await expect(page).toHaveURL(/\/$/)
	})

	test('uses Dvorak list navigation and persists marking a task done', async ({authenticatedPage: page, apiContext, userToken}) => {
		await createProjects(1)
		const [task] = await TaskFactory.create(1)
		await page.goto('/projects/1/1')
		const row = page.locator('.tasks .single-task').first()
		await expect(row).toBeVisible()
		await page.locator('.project-title').click()
		await pressLayoutKey(page, {
			key: 'j',
			code: 'KeyC',
		})
		await expect(row).toBeFocused()
		await page.keyboard.press('Enter')
		await expect(page).toHaveURL(new RegExp(`/tasks/${task.id}`))
		const doneButton = page.locator('.button--mark-done')
		await expect(doneButton).toBeVisible()
		await doneButton.focus()
		const saved = page.waitForResponse(response =>
			response.url().endsWith(`/api/v2/tasks/${task.id}`) && response.request().method() === 'PATCH',
		)
		await pressLayoutKey(page, {
			key: 't',
			code: 'KeyK',
		})
		expect((await saved).ok()).toBe(true)
		await page.reload()
		await expect(doneButton).not.toHaveClass(/is-pending/)
		const response = await apiContext.get(`${apiRootUrl(process.env.API_URL || 'http://localhost:3456')}/api/v2/tasks/${task.id}`, {
			headers: {Authorization: `Bearer ${userToken}`},
		})
		expect(response.ok()).toBe(true)
		expect((await response.json()).done).toBe(true)
	})
})
