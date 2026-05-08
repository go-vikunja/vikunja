import {test, expect} from '../../support/fixtures'

test.describe('Quick actions search', () => {
	test.beforeEach(async ({authenticatedPage: page}) => {
		await page.goto('/')
		await page.waitForLoadState('networkidle')
	})

	test('search input is focused when opening via button click', async ({authenticatedPage: page}) => {
		await page.getByTitle('Open the search/quick action bar').click()

		// The <dialog> uses showModal() + autofocus to focus the input.
		// This was a regression in cef03cb2a (native dialog refactor): v-focus
		// fired before showModal() opened the dialog and silently no-oped.
		await expect(page.locator('dialog.modal-dialog')).toBeVisible()
		await expect(
			page.locator('dialog.modal-dialog .action-input input.input'),
		).toBeFocused()
	})

	test('search input is focused when opening via keyboard shortcut', async ({authenticatedPage: page}) => {
		// Mirror isAppleDevice() from src/helpers/isAppleDevice.ts so the modifier
		// matches what the component expects, regardless of how the test runner
		// platform (process.platform) compares to the browser's navigator.userAgent.
		const isAppleDevice = await page.evaluate(() =>
			navigator.userAgent.includes('Mac') ||
			['iPad Simulator', 'iPhone Simulator', 'iPad', 'iPhone', 'iPod'].includes(navigator.platform),
		)
		const modifier = isAppleDevice ? 'Meta' : 'Control'
		await page.keyboard.press(`${modifier}+k`)

		await expect(page.locator('dialog.modal-dialog')).toBeVisible()
		await expect(
			page.locator('dialog.modal-dialog .action-input input.input'),
		).toBeFocused()
	})

	test('search input is focused when modal is closed and reopened', async ({authenticatedPage: page}) => {
		const input = page.locator('dialog.modal-dialog .action-input input.input')

		await page.getByTitle('Open the search/quick action bar').click()
		await expect(page.locator('dialog.modal-dialog')).toBeVisible()
		await expect(input).toBeFocused()

		await page.keyboard.press('Escape')
		await expect(page.locator('dialog.modal-dialog')).not.toBeVisible()

		await page.getByTitle('Open the search/quick action bar').click()
		await expect(page.locator('dialog.modal-dialog')).toBeVisible()
		await expect(input).toBeFocused()
	})
})
