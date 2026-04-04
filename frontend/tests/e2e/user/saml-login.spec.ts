import {test, expect} from '../../support/fixtures'

test.describe('SAML Login', () => {
	test('logs in via SAML provider', async ({page}) => {
		await page.goto('/login')
		await page.locator('text=Test SAML IDP').click()

		// Wait for navigation to SimpleSAMLphp IDP
		await page.waitForURL(/saml-idp/)

		// Fill in the SimpleSAMLphp login form
		await page.locator('input[name="username"]').fill('user1')
		await page.locator('input[name="password"]').fill('user1pass')
		await page.locator('input[type="submit"], button[type="submit"]').first().click()

		// Should redirect back to the app after SAML ACS processing
		await expect(page).toHaveURL(/\//)
		await expect(page.locator('.show-tasks h3')).toContainText('Current Tasks')
	})
})
