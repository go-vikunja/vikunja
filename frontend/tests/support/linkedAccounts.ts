import {expect, type Page} from '@playwright/test'

export async function linkDexAccount(page: Page) {
	await page.getByRole('button', {name: 'Link Dex', exact: true}).click()

	await expect(page.locator('h2')).toContainText('Log in to Your Account')
	await page.locator('#login').fill('test@example.com')
	await page.locator('#password').fill('12345678')
	await page.locator('#submit-login').click()

	await expect(page).toHaveURL(/\/auth\/openid\/dex/)
}
