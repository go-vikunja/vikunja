import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {TEST_PASSWORD} from '../../../support/constants'

test.describe('Data export', () => {
	test('requests an export with correct password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'data-export')
		await page.locator('#currentPasswordDataExport').fill(TEST_PASSWORD)

		const resp = page.waitForResponse(r => r.url().includes('/user/export/request'))
		await page.getByRole('button', {name: /request/i}).click()
		const r = await resp
		expect(r.ok()).toBe(true)
		await expect(page.locator('.global-notification .vue-notification.success')).toBeVisible()
	})

	test('rejects export with wrong password', async ({authenticatedPage: page}) => {
		await gotoUserSettings(page, 'data-export')
		await page.locator('#currentPasswordDataExport').fill('WRONG')

		const resp = page.waitForResponse(r => r.url().includes('/user/export/request'))
		await page.getByRole('button', {name: /request/i}).click()
		const r = await resp
		expect(r.ok()).toBe(false)
		await expect(page.locator('.global-notification .vue-notification.error')).toBeVisible()
	})
})
