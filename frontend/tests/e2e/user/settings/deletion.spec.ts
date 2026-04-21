import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {TEST_PASSWORD} from '../../../support/constants'

test.describe('Account deletion', () => {
	test('blocks deletion request with no password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'deletion')
		await page.locator('.card .button.is-danger').click()
		await expect(page.locator('.help.is-danger')).toContainText(/password/i)
	})

	test('schedules deletion with correct password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'deletion')
		await page.locator('#currentPasswordAccountDelete').fill(TEST_PASSWORD)

		const resp = page.waitForResponse(r => r.url().includes('/user/deletion/request'))
		await page.locator('.card .button.is-danger').click()
		const r = await resp
		expect(r.ok()).toBe(true)
		await expect(page.locator('.global-notification .vue-notification.success')).toBeVisible()
	})
})
