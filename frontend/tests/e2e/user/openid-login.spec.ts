import {test, expect} from '../../support/fixtures'

test.describe.skip('OpenID Login', () => {
	test('logs in via Dex provider', async ({page}) => {
		await page.goto('/login')
		await page.locator('text=Dex').click()

		// Wait for navigation to Dex origin
		await page.waitForURL('**/dex/**')

		// Fill in the Dex login form
		await page.locator('#login').fill('test@example.com')
		await page.locator('#password').fill('12345678')
		await page.locator('#submit-login').click()

		// Should redirect back to the app
		await expect(page).toHaveURL(/\//)
		await expect(page.locator('main.app-content .content h2')).toContainText('test!')
		await expect(page.locator('.show-tasks h3')).toContainText('Current Tasks')
	})
})
