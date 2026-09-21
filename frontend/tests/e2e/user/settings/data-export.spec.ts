import {readFile} from 'node:fs/promises'
import {test, expect} from '../../../support/fixtures'
import {gotoUserSettings} from '../../../support/userSettings'
import {TEST_PASSWORD} from '../../../support/constants'

test.describe('Data export', () => {
	test('requests and downloads a stored export', async ({authenticatedPage: page, apiContext, userToken}) => {
		await gotoUserSettings(page, 'data-export')
		await page.locator('#currentPasswordDataExport').fill(TEST_PASSWORD)

		const resp = page.waitForResponse(r => r.url().includes('/user/export/request'))
		await page.getByRole('button', {name: /request/i}).click()
		const r = await resp
		expect(r.ok()).toBe(true)
		await expect(page.locator('.global-notification .vue-notification.success')).toBeVisible()
		await expect.poll(async () => {
			const status = await apiContext.get('user/export', {headers: {Authorization: `Bearer ${userToken}`}})
			return (await status.json())?.id ?? 0
		}, {timeout: 30000}).toBeGreaterThan(0)
		await page.reload()
		await page.getByRole('link', {name: 'Download', exact: true}).click()
		await page.locator('#currentPasswordDataExport').fill(TEST_PASSWORD)
		const downloaded = page.waitForEvent('download')
		await page.getByRole('button', {name: 'Download', exact: true}).click()
		const download = await downloaded
		expect(download.suggestedFilename()).toBe('vikunja-export.zip')
		const bytes = await readFile((await download.path())!)
		expect(bytes.subarray(0, 2).toString()).toBe('PK')

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
