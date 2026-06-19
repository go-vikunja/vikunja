import {test, expect} from '../../../support/fixtures'
import {authenticator} from 'otplib'
import {gotoUserSettings} from '../../../support/userSettings'
import {TotpFactory} from '../../../factories/totp'
import {TEST_PASSWORD} from '../../../support/constants'

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

	test('disables an enabled TOTP', async ({authenticatedPage: page, currentUser}) => {
		await TotpFactory.create(1, {user_id: currentUser.id}, false)
		await gotoUserSettings(page, 'totp')

		await page.getByRole('button', {name: 'Disable'}).click()
		await page.locator('#currentPassword').fill(TEST_PASSWORD)

		// Second "Disable" inside the form submits it.
		await page.locator('.card').getByRole('button', {name: 'Disable'}).last().click()

		await expect(page.locator('.global-notification')).toContainText('Success')
		await expect(page.getByRole('button', {name: 'Enroll'})).toBeVisible()
	})

	test('rejects wrong passcode during enrollment', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'totp')
		await page.getByRole('button', {name: 'Enroll'}).click()
		await page.locator('#totpConfirmPasscode').fill('000000')
		await page.getByRole('button', {name: 'Confirm'}).click()

		await expect(page.locator('.global-notification .vue-notification.error')).toBeVisible()
		await expect(page).not.toHaveURL(/\/login/)
	})
})
