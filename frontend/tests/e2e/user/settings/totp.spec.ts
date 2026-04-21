import {test, expect} from '../../../support/fixtures'
import {authenticator} from 'otplib'
import {gotoUserSettings} from '../../../support/userSettings'

test.describe('TOTP', () => {
	test('enrolls TOTP and forces re-login', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'totp')

		await page.getByRole('button', {name: 'Enroll'}).click()

		// Secret is rendered in <strong> after enroll()
		const secret = await page.locator('.card strong').first().innerText()
		expect(secret).toMatch(/^[A-Z2-7]+$/) // base32

		const code = authenticator.generate(secret)
		await page.locator('#totpConfirmPasscode').fill(code)
		await page.getByRole('button', {name: 'Confirm'}).click()

		// TOTP.vue:152 calls authStore.logout() on confirm success.
		await expect(page).toHaveURL(/\/login/)
	})
})
